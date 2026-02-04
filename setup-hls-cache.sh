#!/bin/bash

echo "========================================"
echo "HLS Cache Setup - Camera Dashboard"
echo "========================================"
echo ""

echo "[1/3] Building Nginx Cache image..."
docker build -t camera-nginx-cache:latest ./nginx
if [ $? -ne 0 ]; then
    echo "ERROR: Failed to build nginx image"
    exit 1
fi

echo ""
echo "[2/3] Starting Nginx Cache service..."
docker-compose -f docker-compose.prod.yml up -d nginx-cache
if [ $? -ne 0 ]; then
    echo "ERROR: Failed to start nginx-cache"
    exit 1
fi

echo ""
echo "[3/3] Checking service health..."
sleep 5
docker ps --filter "name=camera-nginx-cache" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "========================================"
echo "Setup completed successfully!"
echo "========================================"
echo ""
echo "Nginx Cache is running on: http://localhost:8082"
echo ""
echo "Update your camera stream URLs:"
echo "FROM: http://camera-ip:port/stream/index.m3u8"
echo "TO:   http://localhost:8082/hls/camera-ip:port/stream/index.m3u8"
echo ""
echo "Check cache status:"
echo "  curl http://localhost:8082/nginx-health"
echo ""
echo "View logs:"
echo "  docker logs camera-nginx-cache -f"
echo ""
