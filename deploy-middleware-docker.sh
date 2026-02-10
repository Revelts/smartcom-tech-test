#!/bin/bash
# Deploy Middleware with Docker on VM: middleware (34.128.100.247)
# Run this script ON the middleware VM

set -e

echo "🚀 Deploying Middleware with Docker..."

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
docker stop middleware 2>/dev/null || true
docker rm middleware 2>/dev/null || true

# Build new image
echo "🔨 Building Docker image..."
docker build -f services/middleware/Dockerfile -t middleware .

# Run container with environment variables
echo "▶️  Starting container..."
docker run -d \
  --name middleware \
  --restart always \
  -p 8080:8080 \
  -e PORT=8080 \
  -e EXTERNAL_ENDPOINT_URL=http://10.184.0.4:8081/external/alerts \
  -e QUEUE_SIZE=1000 \
  -e WORKER_COUNT=10 \
  -e HTTP_TIMEOUT=3s \
  -e MAX_RETRIES=3 \
  -e BASE_DELAY=500ms \
  middleware

# Wait for container to start
sleep 3

# Check status
echo ""
echo "📊 Container Status:"
docker ps | grep middleware

echo ""
echo "📝 Recent Logs:"
docker logs middleware --tail 10

echo ""
echo "✅ Middleware deployed successfully!"
echo ""
echo "Test from VM:"
echo "  curl http://localhost:8080/health"
echo ""
echo "Test from outside:"
echo "  curl http://34.128.100.247:8080/health"
echo ""
echo "Send test event:"
echo '  curl -X POST http://34.128.100.247:8080/integrations/events \'
echo '    -H "Content-Type: application/json" \'
echo '    -d '"'"'{"source":"test","event_type":"alert","severity":"critical","message":"Docker test"}'"'"
echo ""
echo "View logs:"
echo "  docker logs -f middleware"
