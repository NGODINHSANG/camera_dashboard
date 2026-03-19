@echo off
REM Camera Dashboard - Build and Export Docker Images
REM Run this on Windows to build images and export to tar files

echo === Building Docker Images ===

REM Build frontend
echo Building frontend image...
docker build -t camera-dashboard-frontend:v2.1 .
if %errorlevel% neq 0 (
    echo Failed to build frontend!
    exit /b 1
)

REM Build backend
echo Building backend image...
docker build -t camera-dashboard-backend:v2.1 ./backend
if %errorlevel% neq 0 (
    echo Failed to build backend!
    exit /b 1
)

REM Build mediamtx (with wget for webhooks)
echo Building mediamtx image...
docker build -t camera-mediamtx:v2.1 ./mediamtx
if %errorlevel% neq 0 (
    echo Failed to build mediamtx!
    exit /b 1
)

echo.
echo === Exporting Images to TAR ===

REM Create deploy folder
if not exist "deploy" mkdir deploy

REM Export images
echo Exporting frontend image...
docker save camera-dashboard-frontend:v2.1 -o deploy/frontend.tar

echo Exporting backend image...
docker save camera-dashboard-backend:v2.1 -o deploy/backend.tar

echo Exporting mediamtx image...
docker save camera-mediamtx:v2.1 -o deploy/mediamtx.tar

REM Copy configs for server
copy docker-compose.prod.yml deploy\docker-compose.yml
copy .env.production deploy\.env.production
copy mediamtx.yml deploy\mediamtx.yml
if not exist "deploy\nginx" mkdir deploy\nginx
copy nginx\nginx-proxy.conf deploy\nginx\nginx-proxy.conf

echo.
echo === Done! ===
echo.
echo Files created in deploy folder:
dir deploy
echo.

REM Zip everything
echo === Creating ZIP file ===
powershell -Command "Compress-Archive -Path 'deploy\*' -DestinationPath 'camera-dashboard.zip' -Force"
echo.
echo Created: camera-dashboard.zip
echo.
echo Next steps:
echo 1. Upload to server:
echo    scp camera-dashboard.zip user@your-server:/opt/
echo.
echo 2. SSH to server and extract:
echo    cd /opt
echo    unzip camera-dashboard.zip -d camera-dashboard
echo    cd camera-dashboard
echo.
echo 3. Edit .env.production:
echo    nano .env.production
echo    # Change SERVER_IP to your public IP
echo.
echo 4. Load images and run:
echo    docker load -i frontend.tar
echo    docker load -i backend.tar
echo    docker load -i mediamtx.tar
echo    docker-compose -f docker-compose.yml up -d
echo.
echo 5. Access: http://YOUR_PUBLIC_IP:54543
