@echo off
echo ========================================
echo WebRTC Setup - Camera Dashboard
echo ========================================
echo.

echo This will setup WebRTC streaming with:
echo - MediaMTX (RTSP to WebRTC bridge)
echo - MediaSoup SFU (Stream distribution)
echo - TURN server (NAT traversal)
echo.

set /p SERVER_IP="Enter your server IP (LAN or Public): "
if "%SERVER_IP%"=="" (
    echo ERROR: Server IP is required
    pause
    exit /b 1
)

echo.
echo Creating .env file...
echo SERVER_IP=%SERVER_IP% > webrtc\.env

echo.
echo [1/4] Building Docker images...
cd webrtc
docker-compose -f docker-compose.webrtc.yml build
if %errorlevel% neq 0 (
    echo ERROR: Failed to build images
    pause
    exit /b 1
)

echo.
echo [2/4] Starting WebRTC services...
docker-compose -f docker-compose.webrtc.yml up -d
if %errorlevel% neq 0 (
    echo ERROR: Failed to start services
    pause
    exit /b 1
)

echo.
echo [3/4] Waiting for services to be ready...
timeout /t 10 /nobreak > nul

echo.
echo [4/4] Checking service health...
docker-compose -f docker-compose.webrtc.yml ps

echo.
echo ========================================
echo WebRTC Setup Completed!
echo ========================================
echo.
echo Services running:
echo - MediaMTX (RTSP Bridge):  http://%SERVER_IP%:8889
echo - MediaSoup (SFU):         http://%SERVER_IP%:3000
echo - TURN Server:             %SERVER_IP%:3478
echo.
echo Next steps:
echo 1. Configure cameras in webrtc/mediamtx.yml
echo 2. Update frontend to use WebRTCPlayer component
echo 3. Test stream: http://%SERVER_IP%:8889/drone_camera_1
echo.
echo View logs:
echo   docker-compose -f docker-compose.webrtc.yml logs -f
echo.
echo Read full guide: WEBRTC_GUIDE.md
echo.
pause
