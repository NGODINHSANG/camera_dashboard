@echo off
setlocal enabledelayedexpansion

echo ==========================================
echo Camera Dashboard - Production Deployment
echo ==========================================

REM Load .env.production
if not exist .env.production (
    echo ERROR: .env.production not found!
    echo Copy .env.production and configure it first
    pause
    exit /b 1
)

REM Parse .env file
for /f "tokens=1,2 delims==" %%a in (.env.production) do (
    set %%a=%%b
)

echo Server IP: %SERVER_IP%
echo.

echo [1/4] Building Docker images...
docker-compose -f docker-compose.prod.yml build
if %errorlevel% neq 0 (
    echo ERROR: Build failed
    pause
    exit /b 1
)

echo.
echo [2/4] Stopping existing containers...
docker-compose -f docker-compose.prod.yml down

echo.
echo [3/4] Starting services...
docker-compose -f docker-compose.prod.yml up -d
if %errorlevel% neq 0 (
    echo ERROR: Failed to start services
    pause
    exit /b 1
)

echo.
echo [4/4] Waiting for services to be ready...
timeout /t 10 /nobreak > nul

echo.
echo ==========================================
echo Deployment Complete!
echo ==========================================
docker-compose -f docker-compose.prod.yml ps

echo.
echo Services:
echo   - Frontend: http://%SERVER_IP%:8081
echo   - Backend API: http://%SERVER_IP%:8080
echo   - HLS Cache: http://%SERVER_IP%:8082
echo   - WebRTC SFU: http://%SERVER_IP%:3000
echo   - MediaMTX: http://%SERVER_IP%:8889
echo.
echo View logs: docker-compose -f docker-compose.prod.yml logs -f
echo.
pause
