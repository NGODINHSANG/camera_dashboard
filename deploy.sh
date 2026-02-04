#!/bin/bash
set -e

echo "=========================================="
echo "Camera Dashboard - Production Deployment"
echo "=========================================="

# Check if .env.production exists
if [ ! -f .env.production ]; then
    echo "ERROR: .env.production not found!"
    echo "Copy .env.production.example and configure it first"
    exit 1
fi

# Load environment
export $(cat .env.production | xargs)

echo "Server IP: $SERVER_IP"
echo ""

# Build images
echo "[1/4] Building Docker images..."
docker-compose -f docker-compose.prod.yml build

# Stop existing containers
echo ""
echo "[2/4] Stopping existing containers..."
docker-compose -f docker-compose.prod.yml down

# Start services
echo ""
echo "[3/4] Starting services..."
docker-compose -f docker-compose.prod.yml up -d

# Wait for services
echo ""
echo "[4/4] Waiting for services to be ready..."
sleep 10

# Show status
echo ""
echo "=========================================="
echo "Deployment Complete!"
echo "=========================================="
docker-compose -f docker-compose.prod.yml ps

echo ""
echo "Services:"
echo "  - Frontend: http://$SERVER_IP:8081"
echo "  - Backend API: http://$SERVER_IP:8080"
echo "  - HLS Cache: http://$SERVER_IP:8082"
echo "  - WebRTC SFU: http://$SERVER_IP:3000"
echo "  - MediaMTX: http://$SERVER_IP:8889"
echo ""
echo "View logs: docker-compose -f docker-compose.prod.yml logs -f"
echo ""
