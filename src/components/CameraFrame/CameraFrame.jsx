import { useEffect, useRef, useState, useCallback } from 'react'
import Hls from 'hls.js'
import './CameraFrame.css'

function CameraFrame({ camera, onStatusChange, onClick, isFullscreen, onEdit, onDelete, isAdmin = false, isRecordingActive = false, onRecordClick }) {
    const { id, name, location, isRecording, streamUrl, aspectRatio = 'auto' } = camera
    const videoRef = useRef(null)
    const hlsRef = useRef(null)
    const pcRef = useRef(null)
    const retryTimeoutRef = useRef(null)

    const [isConnected, setIsConnected] = useState(false)
    const [isLoading, setIsLoading] = useState(false)
    const [errorMsg, setErrorMsg] = useState('')

    const RETRY_INTERVAL = 5000

    const isWebRTCUrl = (url) => url && url.includes('/webrtc/') && url.includes('/whep')

    const destroyAll = useCallback(() => {
        if (hlsRef.current) {
            hlsRef.current.destroy()
            hlsRef.current = null
        }
        if (pcRef.current) {
            pcRef.current.close()
            pcRef.current = null
        }
        if (retryTimeoutRef.current) {
            clearTimeout(retryTimeoutRef.current)
            retryTimeoutRef.current = null
        }
    }, [])

    const connectWebRTC = useCallback(async () => {
        const video = videoRef.current
        if (!video) return

        try {
            console.log(`[${name}] Connecting WebRTC to:`, streamUrl)

            const pc = new RTCPeerConnection({
                iceServers: []
            })
            pcRef.current = pc

            pc.ontrack = (event) => {
                console.log(`[${name}] Received WebRTC track`)
                video.srcObject = event.streams[0]
                setIsConnected(true)
                setIsLoading(false)
                onStatusChange?.(id, true)
            }

            pc.oniceconnectionstatechange = () => {
                console.log(`[${name}] ICE state:`, pc.iceConnectionState)
                if (pc.iceConnectionState === 'failed' || pc.iceConnectionState === 'closed') {
                    setIsConnected(false)
                    setErrorMsg('WebRTC disconnected')
                    onStatusChange?.(id, false)
                    retryTimeoutRef.current = setTimeout(connectStream, RETRY_INTERVAL)
                }
            }

            pc.addTransceiver('video', { direction: 'recvonly' })
            pc.addTransceiver('audio', { direction: 'recvonly' })

            const offer = await pc.createOffer()
            await pc.setLocalDescription(offer)

            const response = await fetch(streamUrl, {
                method: 'POST',
                headers: { 'Content-Type': 'application/sdp' },
                body: offer.sdp
            })

            if (!response.ok) throw new Error(`WHEP failed: ${response.status}`)

            const answer = await response.text()
            await pc.setRemoteDescription(new RTCSessionDescription({ type: 'answer', sdp: answer }))

        } catch (err) {
            console.error(`[${name}] WebRTC error:`, err)
            setIsConnected(false)
            setIsLoading(false)
            setErrorMsg('Lỗi WebRTC, đang thử lại...')
            onStatusChange?.(id, false)
            retryTimeoutRef.current = setTimeout(connectStream, RETRY_INTERVAL)
        }
    }, [streamUrl, id, name, onStatusChange])

    const connectStream = useCallback(() => {
        const video = videoRef.current
        if (!video || !streamUrl) {
            setIsConnected(false)
            setIsLoading(false)
            return
        }

        console.log(`[${name}] Connecting to:`, streamUrl)
        setIsLoading(true)
        setErrorMsg('')
        destroyAll()

        // Check if WebRTC WHEP
        if (isWebRTCUrl(streamUrl)) {
            connectWebRTC()
            return
        }

        if (Hls.isSupported()) {
            const hls = new Hls({
                debug: false,
                enableWorker: true,
                lowLatencyMode: true,
                backBufferLength: 90,
                maxBufferLength: 30,
                maxMaxBufferLength: 60,
                startLevel: -1,
                manifestLoadingMaxRetry: 6,
                manifestLoadingRetryDelay: 1000,
                levelLoadingMaxRetry: 6,
                levelLoadingRetryDelay: 1000,
                fragLoadingMaxRetry: 6,
                fragLoadingRetryDelay: 1000,
            })

            hls.loadSource(streamUrl)
            hls.attachMedia(video)

            hls.on(Hls.Events.MANIFEST_PARSED, (event, data) => {
                console.log(`[${name}] Manifest parsed`)
                setIsConnected(true)
                setIsLoading(false)
                setErrorMsg('')
                onStatusChange?.(id, true)
                video.play().catch(err => console.log(`[${name}] Autoplay prevented:`, err))
            })

            hls.on(Hls.Events.FRAG_LOADED, () => {
                if (!isConnected) {
                    setIsConnected(true)
                    setIsLoading(false)
                    onStatusChange?.(id, true)
                }
            })

            hls.on(Hls.Events.ERROR, (event, data) => {
                if (data.fatal) {
                    setIsConnected(false)
                    setIsLoading(false)
                    onStatusChange?.(id, false)

                    switch (data.type) {
                        case Hls.ErrorTypes.NETWORK_ERROR:
                            setErrorMsg('Lỗi mạng, đang thử lại...')
                            hls.startLoad()
                            retryTimeoutRef.current = setTimeout(connectStream, RETRY_INTERVAL)
                            break
                        case Hls.ErrorTypes.MEDIA_ERROR:
                            setErrorMsg('Lỗi media, đang khôi phục...')
                            hls.recoverMediaError()
                            break
                        default:
                            setErrorMsg('Lỗi kết nối, đang thử lại...')
                            hls.destroy()
                            retryTimeoutRef.current = setTimeout(connectStream, RETRY_INTERVAL)
                            break
                    }
                }
            })

            hlsRef.current = hls
        } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
            video.src = streamUrl
            video.addEventListener('loadedmetadata', () => {
                setIsConnected(true)
                setIsLoading(false)
                onStatusChange?.(id, true)
                video.play().catch(err => console.log(`[${name}] Autoplay prevented:`, err))
            })
            video.addEventListener('error', () => {
                setIsConnected(false)
                setIsLoading(false)
                setErrorMsg('Lỗi kết nối')
                onStatusChange?.(id, false)
                retryTimeoutRef.current = setTimeout(connectStream, RETRY_INTERVAL)
            })
        } else {
            setErrorMsg('Trình duyệt không hỗ trợ HLS')
            setIsLoading(false)
        }
    }, [streamUrl, id, name, onStatusChange, destroyAll, isConnected, connectWebRTC])

    useEffect(() => {
        if (streamUrl) {
            connectStream()
        } else {
            setIsConnected(false)
            setIsLoading(false)
            destroyAll()
        }
        return () => destroyAll()
    }, [streamUrl])

    useEffect(() => {
        const video = videoRef.current
        if (!video) return

        const handlePlaying = () => {
            setIsConnected(true)
            setIsLoading(false)
            onStatusChange?.(id, true)
        }

        video.addEventListener('playing', handlePlaying)
        return () => video.removeEventListener('playing', handlePlaying)
    }, [id, onStatusChange])

    // Handle click - chỉ khi không ở fullscreen mode
    const handleClick = () => {
        if (!isFullscreen && onClick) {
            onClick()
        }
    }

    return (
        <div
            className={`camera-frame ${isConnected ? 'online' : 'offline'} ${isFullscreen ? 'fullscreen' : ''} aspect-${aspectRatio.replace(':', '-')}`}
            onClick={handleClick}
            style={{ cursor: isFullscreen ? 'default' : 'pointer' }}
        >
            {/* Video Stream Area */}
            <div className="camera-stream">
                {streamUrl ? (
                    <>
                        <video
                            ref={videoRef}
                            autoPlay
                            muted
                            playsInline
                            className={`stream-video ${isConnected ? 'visible' : 'hidden'}`}
                        />
                        {isLoading && (
                            <div className="stream-loading">
                                <div className="loading-spinner"></div>
                                <span>Đang kết nối...</span>
                            </div>
                        )}
                        {!isConnected && !isLoading && (
                            <div className="stream-placeholder">
                                <div className="no-signal">{errorMsg || 'Đang chờ tín hiệu...'}</div>
                            </div>
                        )}
                    </>
                ) : (
                    <div className="stream-placeholder">
                        <div className="no-signal">Chưa cấu hình stream</div>
                    </div>
                )}
            </div>

            {/* Click hint - chỉ hiện khi không fullscreen và hover */}
            {!isFullscreen && (
                <div className="click-hint">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z" />
                    </svg>
                    <span>Nhấn để phóng to</span>
                </div>
            )}

            {/* Top Overlay */}
            <div className="camera-overlay-top">
                <div className="camera-info-left">
                    <span className="camera-name">{name}</span>
                    <span className={`status-indicator ${isConnected ? 'online' : 'offline'}`}></span>
                </div>
                <div className="camera-info-right">
                    <span className="camera-location">{location}</span>
                    {/* Nút Edit/Delete camera - chỉ hiện với admin */}
                    {isAdmin && (
                        <div className="camera-action-buttons">
                            <button
                                className="camera-action-btn edit-btn"
                                onClick={(e) => {
                                    e.stopPropagation()
                                    onEdit?.()
                                }}
                                title="Chỉnh sửa camera"
                            >
                                <svg viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z" />
                                </svg>
                            </button>
                            <button
                                className="camera-action-btn delete-btn"
                                onClick={(e) => {
                                    e.stopPropagation()
                                    onDelete?.()
                                }}
                                title="Xóa camera"
                            >
                                <svg viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z" />
                                </svg>
                            </button>
                        </div>
                    )}
                </div>
            </div>

            {/* Bottom Overlay */}
            <div className="camera-overlay-bottom">
                <div className="camera-icons">
                    {/* Nút Record/Stop */}
                    <button
                        className={`icon-btn record-btn ${isRecordingActive ? 'recording' : ''}`}
                        onClick={(e) => {
                            e.stopPropagation()
                            onRecordClick?.()
                        }}
                        title={isRecordingActive ? 'Dừng ghi hình' : 'Bắt đầu ghi hình'}
                    >
                        {isRecordingActive ? (
                            <svg viewBox="0 0 24 24" fill="currentColor">
                                <rect x="6" y="6" width="12" height="12" rx="2" />
                            </svg>
                        ) : (
                            <svg viewBox="0 0 24 24" fill="currentColor">
                                <circle cx="12" cy="12" r="8" />
                            </svg>
                        )}
                    </button>
                    <span className={`icon-btn wifi ${isConnected ? 'active' : ''}`} title={isConnected ? 'Đang kết nối' : 'Mất kết nối'}>
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M1 9l2 2c4.97-4.97 13.03-4.97 18 0l2-2C16.93 2.93 7.08 2.93 1 9zm8 8l3 3 3-3c-1.65-1.66-4.34-1.66-6 0zm-4-4l2 2c2.76-2.76 7.24-2.76 10 0l2-2C15.14 9.14 8.87 9.14 5 13z" />
                        </svg>
                    </span>
                </div>
                <div className="camera-status">
                    {isRecordingActive && (
                        <span className="rec-indicator recording-active">
                            <span className="rec-dot"></span>
                            REC
                        </span>
                    )}
                </div>
            </div>
        </div>
    )
}

export default CameraFrame
