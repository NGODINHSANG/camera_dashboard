import { useEffect, useRef, useState, useCallback } from 'react'
import Hls from 'hls.js'
import './CameraFrame.css'

// Sub-component for HLS playback (uses same HLS.js as live streaming)
// Bandwidth optimization: small buffer = only download what's being watched
function PlaybackHLS({ hlsUrl, onEnded, onError, loop = true }) {
    const videoRef = useRef(null)
    const hlsRef = useRef(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState(null)

    useEffect(() => {
        const video = videoRef.current
        if (!video || !hlsUrl) return

        setIsLoading(true)
        setError(null)

        if (Hls.isSupported()) {
            const hls = new Hls({
                // Giảm buffer = giảm upload burst, tiết kiệm băng thông WAN
                maxBufferLength: 10,        // Buffer 10 giây
                maxMaxBufferLength: 20,     // Max 20 giây khi seeking
                maxBufferSize: 15 * 1000 * 1000, // Max 15MB buffer
                maxBufferHole: 0.5,
                // QUAN TRỌNG: Tăng timeout để chờ segment tải xong qua rate limit
                // Không timeout sớm → không retry → không spam router
                fragLoadingTimeOut: 30000,  // 30 giây (đủ cho segment 3MB ở 768KB/s)
                manifestLoadingTimeOut: 15000,
                levelLoadingTimeOut: 15000,
                // Giảm retry để tránh bão request khi mạng chậm
                manifestLoadingMaxRetry: 2,
                manifestLoadingRetryDelay: 2000,
                levelLoadingMaxRetry: 2,
                fragLoadingMaxRetry: 2,
                fragLoadingRetryDelay: 3000,
            })

            hls.loadSource(hlsUrl)
            hls.attachMedia(video)

            hls.on(Hls.Events.MANIFEST_PARSED, () => {
                console.log('[PlaybackHLS] Manifest parsed, starting playback')
                setIsLoading(false)
                video.play().catch(() => { })
            })

            hls.on(Hls.Events.FRAG_LOADED, () => {
                setIsLoading(false)
            })

            hls.on(Hls.Events.ERROR, (_, data) => {
                console.error('[PlaybackHLS] Error:', data.type, data.details, data.fatal)
                if (data.fatal) {
                    setIsLoading(false)
                    switch (data.type) {
                        case Hls.ErrorTypes.NETWORK_ERROR:
                            setError('Lỗi mạng khi tải video')
                            // Try to recover
                            hls.startLoad()
                            break
                        case Hls.ErrorTypes.MEDIA_ERROR:
                            setError('Lỗi media')
                            hls.recoverMediaError()
                            break
                        default:
                            setError('Không thể phát video')
                            onError?.()
                            break
                    }
                }
            })

            hlsRef.current = hls
        } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
            video.src = hlsUrl
            video.onloadeddata = () => setIsLoading(false)
            video.onerror = () => {
                setIsLoading(false)
                setError('Không thể phát video')
            }
        }

        // Handle video ended - for loop and playlist support
        const handleEnded = () => {
            if (loop) {
                // Loop single video: HLS segments already cached, seek to start
                video.currentTime = 0
                video.play().catch(() => { })
            } else if (onEnded) {
                // Playlist mode: notify parent to play next
                onEnded()
            }
        }
        video.addEventListener('ended', handleEnded)

        return () => {
            video.removeEventListener('ended', handleEnded)
            if (hlsRef.current) {
                hlsRef.current.destroy()
                hlsRef.current = null
            }
        }
    }, [hlsUrl, onEnded, onError, loop])

    return (
        <div className="playback-container">
            <video
                ref={videoRef}
                controls
                playsInline
                className="stream-video visible playback-video"
            />
            {isLoading && (
                <div className="playback-loading">
                    <div className="loading-spinner"></div>
                    <span>Đang tải video...</span>
                </div>
            )}
            {error && (
                <div className="playback-error">
                    <span>{error}</span>
                </div>
            )}
        </div>
    )
}

function CameraFrame({
    camera,
    onStatusChange,
    onClick,
    isFullscreen,
    onEdit,
    onDelete,
    isAdmin = false,
    isRecordingActive = false,
    onRecordClick,
    onViewRecordings,
    // Playback mode props
    playbackRecording = null,  // { name, path, streamUrl, size, modTime }
    onExitPlayback = null,
    onVideoEnded = null,       // Called when playback video ends
    playlistInfo = null        // { playlist: [...], currentIndex: 0 }
}) {
    const { id, name, location, isRecording, streamUrl, aspectRatio = 'auto' } = camera
    const videoRef = useRef(null)
    const hlsRef = useRef(null)
    const pcRef = useRef(null)
    const retryTimeoutRef = useRef(null)

    const [isConnected, setIsConnected] = useState(false)
    const [isLoading, setIsLoading] = useState(false)
    const [errorMsg, setErrorMsg] = useState('')
    // Lazy loading: only load stream when activated (fullscreen or user clicks)


    const RETRY_INTERVAL = 5000
    const isPlaybackMode = !!playbackRecording

    // Debug playback
    useEffect(() => {
        if (playbackRecording) {
            console.log('CameraFrame received playbackRecording:', playbackRecording)
        }
    }, [playbackRecording])

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

        // Swap HLS URL path based on view mode:
        // Grid: /hls/grid/<stream> (rate limited to ~1Mbps)
        // Fullscreen: /hls/full/<stream> (rate limited to ~6Mbps)
        let effectiveUrl = streamUrl
        if (streamUrl.includes('/hls/')) {
            if (isFullscreen) {
                effectiveUrl = streamUrl.replace('/hls/', '/hls/full/')
            } else {
                effectiveUrl = streamUrl.replace('/hls/', '/hls/grid/')
            }
        }

        console.log(`[${name}] Connecting to:`, effectiveUrl, isFullscreen ? '(fullscreen)' : '(grid)')
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
                // Grid: small buffer = less bandwidth. Fullscreen: normal buffer
                backBufferLength: isFullscreen ? 90 : 10,
                maxBufferLength: isFullscreen ? 30 : 5,
                maxMaxBufferLength: isFullscreen ? 60 : 10,
                startLevel: -1,
                manifestLoadingMaxRetry: 6,
                manifestLoadingRetryDelay: 1000,
                levelLoadingMaxRetry: 6,
                levelLoadingRetryDelay: 1000,
                fragLoadingMaxRetry: 6,
                fragLoadingRetryDelay: 1000,
            })

            hls.loadSource(effectiveUrl)
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
    }, [streamUrl, id, name, onStatusChange, destroyAll, isConnected, connectWebRTC, isFullscreen])

    useEffect(() => {
        // Don't connect live stream if in playback mode
        if (isPlaybackMode) {
            destroyAll()
            setIsConnected(false)
            setIsLoading(false)
            return
        }

        if (streamUrl) {
            connectStream()
        } else {
            setIsConnected(false)
            setIsLoading(false)
            destroyAll()
        }
        return () => destroyAll()
    }, [streamUrl, isPlaybackMode, isFullscreen])

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

    // Handle click
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
                {isPlaybackMode ? (
                    /* PLAYBACK MODE - Use HLS for instant loading, fallback to MP4 */
                    playbackRecording?.hlsUrl ? (
                        <PlaybackHLS
                            hlsUrl={playbackRecording.hlsUrl}
                            loop={true}
                        />
                    ) : playbackRecording?.streamUrl ? (
                        <video
                            src={playbackRecording.streamUrl}
                            autoPlay
                            controls
                            playsInline
                            loop
                            preload="metadata"
                            className="stream-video visible playback-video"
                        />
                    ) : (
                        <div className="stream-placeholder">
                            <div className="no-signal">Không thể tải video</div>
                        </div>
                    )
                ) : streamUrl ? (
                    /* LIVE MODE - Stream activated */
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
                                <div className="no-signal-container">
                                    <svg viewBox="0 0 24 24" fill="currentColor" className="no-signal-icon">
                                        <path d="M21 6h-7.59l3.29-3.29L16 2l-4 4-4-4-.71.71L10.59 6H3c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h18c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 14H3V8h18v12zM9 10v8l7-4z" />
                                    </svg>
                                    <span className="no-signal-text">Chưa có kết nối live stream</span>
                                    {onViewRecordings && (
                                        <button
                                            className="no-signal-playback-btn"
                                            onClick={(e) => {
                                                e.stopPropagation()
                                                onViewRecordings()
                                            }}
                                        >
                                            <svg viewBox="0 0 24 24" fill="currentColor">
                                                <path d="M4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm16-4H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-8 12.5v-9l6 4.5-6 4.5z" />
                                            </svg>
                                            Xem playback
                                        </button>
                                    )}
                                </div>
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
                    {isPlaybackMode ? (
                        /* Playback mode: show exit button */
                        <button
                            className="icon-btn live-btn"
                            onClick={(e) => {
                                e.stopPropagation()
                                onExitPlayback?.()
                            }}
                            title="Quay lai xem truc tiep"
                        >
                            <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M8.59 16.59L13.17 12 8.59 7.41 10 6l6 6-6 6-1.41-1.41z" />
                            </svg>
                            <span style={{ marginLeft: 4, fontSize: 12 }}>LIVE</span>
                        </button>
                    ) : (
                        /* Live mode: show record and playback buttons */
                        <>
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
                            <button
                                className="icon-btn playback-btn"
                                onClick={(e) => {
                                    e.stopPropagation()
                                    onViewRecordings?.()
                                }}
                                title="Xem lai ban ghi"
                            >
                                <svg viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm16-4H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-8 12.5v-9l6 4.5-6 4.5z" />
                                </svg>
                            </button>
                        </>
                    )}
                    <span className={`icon-btn wifi ${isConnected ? 'active' : ''}`} title={isConnected ? 'Đang kết nối' : 'Mất kết nối'}>
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M1 9l2 2c4.97-4.97 13.03-4.97 18 0l2-2C16.93 2.93 7.08 2.93 1 9zm8 8l3 3 3-3c-1.65-1.66-4.34-1.66-6 0zm-4-4l2 2c2.76-2.76 7.24-2.76 10 0l2-2C15.14 9.14 8.87 9.14 5 13z" />
                        </svg>
                    </span>
                </div>
                <div className="camera-status">
                    {isPlaybackMode ? (
                        <span className="rec-indicator playback-active">
                            <svg viewBox="0 0 24 24" fill="currentColor" style={{ width: 12, height: 12, marginRight: 4 }}>
                                <path d="M4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm16-4H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-8 12.5v-9l6 4.5-6 4.5z" />
                            </svg>
                            {playlistInfo ? (
                                <span>PLAYBACK {playlistInfo.currentIndex + 1}/{playlistInfo.playlist.length}</span>
                            ) : (
                                <span>PLAYBACK</span>
                            )}
                        </span>
                    ) : isRecordingActive ? (
                        <span className="rec-indicator recording-active">
                            <span className="rec-dot"></span>
                            REC
                        </span>
                    ) : null}
                </div>
            </div>
        </div>
    )
}

export default CameraFrame
