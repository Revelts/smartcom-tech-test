#!/bin/bash
# Deploy External Endpoint with Docker on VM: external-ep (34.50.103.62)
# Run this script ON the external-ep VM

set -e

echo "🚀 Deploying External Endpoint with Docker..."

# Navigate to code directory
cd ~/smartcom-tech-test

# Install Docker if not installed
if ! command -v docker &> /dev/null; then
    echo "📦 Installing Docker..."
    sudo apt-get update
    sudo apt-get install -y docker.io
    sudo usermod -aG docker $USER
    echo "✅ Docker installed. Please log out and log back in, then run this script again."
    exit 0
fi

# Pull latest code
echo "📥 Pulling latest code..."
git pull || true

# Stop and remove old container
echo "🛑 Stopping old container..."
docker stop external-endpoint 2>/dev/null || true
docker rm external-endpoint 2>/dev/null || true

# Build new image
echo "🔨 Building Docker image..."
docker build -f services/external-endpoint/Dockerfile -t external-endpoint .

# Run container
echo "▶️  Starting container..."
docker run -d \
  --name external-endpoint \
  --restart always \
  -p 8081:8081 \
  -e PORT=8081 \
  external-endpoint

# Wait for container to start
sleep 3

# Check status
echo ""
echo "📊 Container Status:"
docker ps | grep external-endpoint

echo ""
echo "📝 Recent Logs:"
docker logs external-endpoint --tail 10

echo ""
echo "✅ External Endpoint deployed successfully!"
echo ""
echo "Test from VM:"
echo "  curl http://localhost:8081/health"
echo ""
echo "Test from outside:"
echo "  curl http://34.50.103.62:8081/health"
echo ""
echo "View logs:"
echo "  docker logs -f external-endpoint"
