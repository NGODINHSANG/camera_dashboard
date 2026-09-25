import { useState, useRef, useEffect } from 'react';
import './UploadVideo.css';

const API_BASE = import.meta.env.VITE_API_URL || '';

const ALLOWED_EXTENSIONS = ['.mp4', '.avi', '.mkv', '.mov'];
const MAX_FILE_SIZE = 2 * 1024 * 1024 * 1024; // 2GB
const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB per chunk
const PARALLEL_CHUNKS = 3; // Upload 3 chunks in parallel

export default function UploadVideo({ isOpen, onClose, projects, onUploadComplete, preSelectedCamera }) {
    const [selectedProject, setSelectedProject] = useState('');
    const [selectedCamera, setSelectedCamera] = useState('');
    const [cameras, setCameras] = useState([]);
    const [uploadJobs, setUploadJobs] = useState([]);
    const [isDragging, setIsDragging] = useState(false);
    const fileInputRef = useRef(null);
    const pollingRef = useRef(null);
    const abortControllersRef = useRef(new Map()); // Track abort controllers for cancellation

    // Pre-select camera if provided
    useEffect(() => {
        if (preSelectedCamera && isOpen) {
            setSelectedProject(String(preSelectedCamera.project_id));
            setSelectedCamera(String(preSelectedCamera.id));
            setCameras([preSelectedCamera]);
        }
    }, [preSelectedCamera, isOpen]);

    // Fetch cameras when project changes (only if no pre-selected camera)
    useEffect(() => {
        if (selectedProject && !preSelectedCamera) {
            fetchCameras(selectedProject);
        } else if (!selectedProject && !preSelectedCamera) {
            setCameras([]);
            setSelectedCamera('');
        }
    }, [selectedProject, preSelectedCamera]);

    // Poll for job status updates (for processing phase)
    useEffect(() => {
        const processingJobs = uploadJobs.filter(job =>
            job.status === 'merging' || job.status === 'processing'
        );

        if (processingJobs.length > 0) {
            pollingRef.current = setInterval(() => {
                processingJobs.forEach(job => {
                    fetchJobStatus(job.uploadId);
                });
            }, 2000);
        }

        return () => {
            if (pollingRef.current) {
                clearInterval(pollingRef.current);
            }
        };
    }, [uploadJobs]);

    const fetchCameras = async (projectId) => {
        try {
            const token = localStorage.getItem('auth_token');
            const response = await fetch(`${API_BASE}/api/projects/${projectId}/cameras`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (response.ok) {
                const data = await response.json();
                setCameras(data);
            }
        } catch (error) {
            console.error('Failed to fetch cameras:', error);
        }
    };

    const fetchJobStatus = async (uploadId) => {
        try {
            const token = localStorage.getItem('auth_token');
            const response = await fetch(`${API_BASE}/api/projects/chunked-upload/status/${uploadId}`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (response.ok) {
                const data = await response.json();
                setUploadJobs(prev => prev.map(j => {
                    if (j.uploadId === uploadId) {
                        const newStatus = data.status;
                        // Notify parent when job completes
                        if (newStatus === 'completed' && j.status !== 'completed' && onUploadComplete) {
                            onUploadComplete(data);
                        }
                        return {
                            ...j,
                            status: newStatus,
                            progress: newStatus === 'completed' ? 100 : (data.progress || j.progress),
                            error: data.error
                        };
                    }
                    return j;
                }));
            }
        } catch (error) {
            console.error('Failed to fetch job status:', error);
        }
    };

    const validateFile = (file) => {
        const ext = '.' + file.name.split('.').pop().toLowerCase();
        if (!ALLOWED_EXTENSIONS.includes(ext)) {
            return `Invalid file type: ${file.name}. Allowed: ${ALLOWED_EXTENSIONS.join(', ')}`;
        }
        if (file.size > MAX_FILE_SIZE) {
            return `File too large: ${file.name}. Maximum size is 2GB`;
        }
        return null;
    };

    // Chunked upload implementation
    const uploadFileChunked = async (file) => {
        if (!selectedProject || !selectedCamera) {
            alert('Please select a project and camera first');
            return;
        }

        const error = validateFile(file);
        if (error) {
            alert(error);
            return;
        }

        const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
        const tempId = 'temp-' + Date.now();

        // Create a temporary job for UI
        const tempJob = {
            id: tempId,
            uploadId: null,
            original_filename: file.name,
            status: 'initializing',
            progress: 0,
            file_size: file.size,
            uploadedChunks: 0,
            totalChunks: totalChunks
        };
        setUploadJobs(prev => [...prev, tempJob]);

        try {
            const token = localStorage.getItem('auth_token');

            // Step 1: Initialize upload
            const initFormData = new FormData();
            initFormData.append('filename', file.name);
            initFormData.append('totalSize', file.size.toString());
            initFormData.append('totalChunks', totalChunks.toString());
            initFormData.append('chunkSize', CHUNK_SIZE.toString());

            const initResponse = await fetch(
                `${API_BASE}/api/projects/chunked-upload/init/${selectedProject}/${selectedCamera}`,
                {
                    method: 'POST',
                    headers: { 'Authorization': `Bearer ${token}` },
                    body: initFormData
                }
            );

            if (!initResponse.ok) {
                const errorData = await initResponse.json();
                throw new Error(errorData.message || 'Failed to initialize upload');
            }

            const initData = await initResponse.json();
            console.log('[ChunkUpload] Init response:', initData);
            // Response is wrapped: {success: true, data: {uploadId: ...}}
            const uploadId = initData.data?.uploadId || initData.uploadId;
            console.log('[ChunkUpload] Upload ID:', uploadId);

            if (!uploadId) {
                throw new Error('Server did not return uploadId');
            }

            // Update job with real uploadId
            setUploadJobs(prev => prev.map(j =>
                j.id === tempId ? { ...j, uploadId, status: 'uploading' } : j
            ));

            // Create abort controller for this upload
            const abortController = new AbortController();
            abortControllersRef.current.set(uploadId, abortController);

            // Step 2: Upload chunks in parallel (3 at a time)
            let uploadedCount = 0;
            const chunkIndices = Array.from({ length: totalChunks }, (_, i) => i);

            // Process chunks in batches of PARALLEL_CHUNKS
            for (let i = 0; i < chunkIndices.length; i += PARALLEL_CHUNKS) {
                // Check if cancelled
                if (abortController.signal.aborted) {
                    throw new Error('Upload cancelled');
                }

                const batch = chunkIndices.slice(i, i + PARALLEL_CHUNKS);

                await Promise.all(batch.map(async (chunkIndex) => {
                    if (abortController.signal.aborted) return;

                    const start = chunkIndex * CHUNK_SIZE;
                    const end = Math.min(start + CHUNK_SIZE, file.size);
                    const chunk = file.slice(start, end);

                    const chunkFormData = new FormData();
                    chunkFormData.append('chunkIndex', chunkIndex.toString());
                    chunkFormData.append('chunk', chunk);

                    const chunkResponse = await fetch(
                        `${API_BASE}/api/projects/chunked-upload/chunk/${uploadId}`,
                        {
                            method: 'POST',
                            headers: { 'Authorization': `Bearer ${token}` },
                            body: chunkFormData,
                            signal: abortController.signal
                        }
                    );

                    if (!chunkResponse.ok) {
                        const errorData = await chunkResponse.json();
                        throw new Error(errorData.message || `Failed to upload chunk ${chunkIndex}`);
                    }

                    uploadedCount++;
                    const progress = Math.round((uploadedCount / totalChunks) * 100);

                    // Update progress
                    setUploadJobs(prev => prev.map(j =>
                        j.uploadId === uploadId ? { ...j, progress, uploadedChunks: uploadedCount } : j
                    ));
                }));
            }

            // Step 3: Complete upload
            const completeResponse = await fetch(
                `${API_BASE}/api/projects/chunked-upload/complete/${uploadId}`,
                {
                    method: 'POST',
                    headers: { 'Authorization': `Bearer ${token}` }
                }
            );

            if (!completeResponse.ok) {
                const errorData = await completeResponse.json();
                throw new Error(errorData.message || 'Failed to complete upload');
            }

            // Update status to merging/processing
            setUploadJobs(prev => prev.map(j =>
                j.uploadId === uploadId ? { ...j, status: 'merging', progress: 100 } : j
            ));

            // Clean up abort controller
            abortControllersRef.current.delete(uploadId);

        } catch (error) {
            console.error('Upload error:', error);
            if (error.name !== 'AbortError') {
                setUploadJobs(prev => prev.map(j =>
                    j.id === tempId ? { ...j, status: 'failed', error: error.message } : j
                ));
            }
        }
    };

    // Queue để upload tuần tự, tránh xung đột init
    const uploadQueueRef = useRef([]);
    const isProcessingRef = useRef(false);

    const processUploadQueue = async () => {
        if (isProcessingRef.current || uploadQueueRef.current.length === 0) {
            return;
        }

        isProcessingRef.current = true;

        while (uploadQueueRef.current.length > 0) {
            const file = uploadQueueRef.current.shift();
            await uploadFileChunked(file);
        }

        isProcessingRef.current = false;
    };

    const handleFiles = (files) => {
        // Thêm files vào queue
        uploadQueueRef.current.push(...Array.from(files));
        // Bắt đầu xử lý queue
        processUploadQueue();
    };

    const handleDrop = (e) => {
        e.preventDefault();
        setIsDragging(false);
        handleFiles(e.dataTransfer.files);
    };

    const handleDragOver = (e) => {
        e.preventDefault();
        setIsDragging(true);
    };

    const handleDragLeave = (e) => {
        e.preventDefault();
        setIsDragging(false);
    };

    const handleFileSelect = (e) => {
        handleFiles(e.target.files);
        e.target.value = ''; // Reset input
    };

    const cancelJob = async (job) => {
        const uploadId = job.uploadId;

        // If still uploading, abort the fetch requests
        if (uploadId && abortControllersRef.current.has(uploadId)) {
            abortControllersRef.current.get(uploadId).abort();
            abortControllersRef.current.delete(uploadId);
        }

        // If has uploadId, cancel on server
        if (uploadId && !job.id.startsWith('temp-')) {
            try {
                const token = localStorage.getItem('auth_token');
                await fetch(`${API_BASE}/api/projects/chunked-upload/cancel/${uploadId}`, {
                    method: 'DELETE',
                    headers: { 'Authorization': `Bearer ${token}` }
                });
            } catch (error) {
                console.error('Failed to cancel upload:', error);
            }
        }

        // Remove from UI
        setUploadJobs(prev => prev.filter(j => j.id !== job.id));
    };

    const removeCompletedJob = (jobId) => {
        setUploadJobs(prev => prev.filter(j => j.id !== jobId));
    };

    const formatFileSize = (bytes) => {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
        return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
    };

    const getStatusIcon = (status) => {
        switch (status) {
            case 'initializing': return '⏳';
            case 'uploading': return '⬆️';
            case 'merging': return '🔗';
            case 'processing': return '⚙️';
            case 'completed': return '✅';
            case 'failed': return '❌';
            default: return '⏳';
        }
    };

    const getStatusText = (job) => {
        switch (job.status) {
            case 'initializing': return 'Đang khởi tạo...';
            case 'uploading': return 'Đang tải lên...';
            case 'merging': return 'Đang tổng hợp video...';
            case 'processing': return 'Đang xử lý video...';
            case 'completed': return 'Hoàn tất';
            case 'failed': return 'Thất bại';
            default: return job.status;
        }
    };

    if (!isOpen) return null;

    return (
        <div className="upload-modal-overlay" onClick={onClose}>
            <div className="upload-modal" onClick={e => e.stopPropagation()}>
                <div className="upload-modal-header">
                    <h2>Upload Video</h2>
                    <button className="upload-close-btn" onClick={onClose}>&times;</button>
                </div>

                <div className="upload-modal-body">
                    {/* Project & Camera Info */}
                    {preSelectedCamera ? (
                        <div className="upload-info">
                            <div className="upload-info-item">
                                <label>Project</label>
                                <span className="upload-info-value">{preSelectedCamera.projectName || projects[0]?.name || 'Unknown'}</span>
                            </div>
                            <div className="upload-info-item">
                                <label>Camera</label>
                                <span className="upload-info-value">{preSelectedCamera.name}</span>
                            </div>
                        </div>
                    ) : (
                        <div className="upload-selects">
                            <div className="upload-select-group">
                                <label>Project</label>
                                <select
                                    value={selectedProject}
                                    onChange={(e) => setSelectedProject(e.target.value)}
                                >
                                    <option value="">-- Select Project --</option>
                                    {projects.map(p => (
                                        <option key={p.id} value={p.id}>{p.name}</option>
                                    ))}
                                </select>
                            </div>

                            <div className="upload-select-group">
                                <label>Camera</label>
                                <select
                                    value={selectedCamera}
                                    onChange={(e) => setSelectedCamera(e.target.value)}
                                    disabled={!selectedProject}
                                >
                                    <option value="">-- Select Camera --</option>
                                    {cameras.map(c => (
                                        <option key={c.id} value={c.id}>{c.name}</option>
                                    ))}
                                </select>
                            </div>
                        </div>
                    )}

                    {/* Drop Zone */}
                    <div
                        className={`upload-dropzone ${isDragging ? 'dragging' : ''} ${!selectedCamera ? 'disabled' : ''}`}
                        onDrop={handleDrop}
                        onDragOver={handleDragOver}
                        onDragLeave={handleDragLeave}
                        onClick={() => selectedCamera && fileInputRef.current?.click()}
                    >
                        <input
                            ref={fileInputRef}
                            type="file"
                            multiple
                            accept=".mp4,.avi,.mkv,.mov"
                            onChange={handleFileSelect}
                            style={{ display: 'none' }}
                        />
                        <div className="upload-dropzone-content">
                            <span className="upload-icon">📁</span>
                            <p>Drag & drop video files here</p>
                            <p className="upload-hint">or click to select files</p>
                            <p className="upload-formats">MP4, AVI, MKV, MOV - Max 2GB</p>
                        </div>
                    </div>

                    {/* Upload Jobs List */}
                    {uploadJobs.length > 0 && (
                        <div className="upload-jobs">
                            <h3>Upload Progress</h3>
                            {uploadJobs.map(job => (
                                <div key={job.id} className={`upload-job ${job.status}`}>
                                    <div className="upload-job-info">
                                        <span className="upload-job-icon">{getStatusIcon(job.status)}</span>
                                        <div className="upload-job-details">
                                            <span className="upload-job-name">{job.original_filename}</span>
                                            <span className="upload-job-size">{formatFileSize(job.file_size)}</span>
                                        </div>
                                    </div>

                                    <div className="upload-job-status">
                                        {['initializing', 'uploading', 'merging', 'processing'].includes(job.status) && (
                                            <>
                                                <div className="upload-progress-bar">
                                                    <div
                                                        className="upload-progress-fill"
                                                        style={{ width: `${job.progress}%` }}
                                                    />
                                                </div>
                                                <span className="upload-progress-text">
                                                    {job.progress}% - {getStatusText(job)}
                                                </span>
                                                {['initializing', 'uploading'].includes(job.status) && (
                                                    <button
                                                        className="upload-cancel-btn"
                                                        onClick={() => cancelJob(job)}
                                                    >
                                                        Cancel
                                                    </button>
                                                )}
                                            </>
                                        )}

                                        {job.status === 'completed' && (
                                            <>
                                                <span className="upload-status-text success">{getStatusText(job)}</span>
                                                <button
                                                    className="upload-remove-btn"
                                                    onClick={() => removeCompletedJob(job.id)}
                                                >
                                                    Remove
                                                </button>
                                            </>
                                        )}

                                        {job.status === 'failed' && (
                                            <>
                                                <span className="upload-status-text error">
                                                    {job.error || 'Failed'}
                                                </span>
                                                <button
                                                    className="upload-remove-btn"
                                                    onClick={() => removeCompletedJob(job.id)}
                                                >
                                                    Remove
                                                </button>
                                            </>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                <div className="upload-modal-footer">
                    <button className="upload-done-btn" onClick={onClose}>Done</button>
                </div>
            </div>
        </div>
    );
}
