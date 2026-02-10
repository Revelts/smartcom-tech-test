# Docker Deployment on Your GCP VMs

Deploy your microservices using Docker instead of systemd on your existing VMs.

## Your VMs

- **VM1 (Middleware)**: `34.128.100.247` - Port `8080`
- **VM2 (External Endpoint)**: `34.50.103.62` - Port `8081`
- **Zone**: `asia-southeast2-a` (Jakarta)

---

## Step 1: Deploy External Endpoint (VM2) with Docker

### 1.1 SSH into External Endpoint VM

```bash
gcloud compute ssh external-ep --zone=asia-southeast2-a
```

### 1.2 Install Docker

```bash
# Update system
sudo apt-get update

# Install Docker
sudo apt-get install -y docker.io

# Add your user to docker group (no need for sudo)
sudo usermod -aG docker $USER

# Apply group changes
newgrp docker

# Verify Docker installation
docker --version
```

### 1.3 Navigate to Your Code

```bash
cd ~/smartcom-tech-test
```

### 1.4 Build Docker Image

```bash
# Build External Endpoint image
docker build -f services/external-endpoint/Dockerfile -t external-endpoint .
```

This will:
- Use the Dockerfile you already have
- Build from the project root (needed for `pkg/` folder)
- Tag the image as `external-endpoint`

### 1.5 Run Docker Container

```bash
# Run External Endpoint container
docker run -d \
  --name external-endpoint \
  --restart always \
  -p 8081:8081 \
  -e PORT=8081 \
  external-endpoint
```

**Explanation**:
- `-d`: Run in background (detached)
- `--name external-endpoint`: Container name
- `--restart always`: Auto-restart on VM reboot or crash
- `-p 8081:8081`: Map host port 8081 to container port 8081 (accessible from outside!)
- `-e PORT=8081`: Set environment variable

### 1.6 Verify Container is Running

```bash
# Check container status
docker ps

# Check logs
docker logs external-endpoint

# Follow logs in real-time
docker logs -f external-endpoint
```

### 1.7 Test Locally

```bash
# Test from inside VM
curl http://localhost:8081/health

# Exit VM
exit
```

### 1.8 Test from Outside

From your local machine:

```bash
# Test from outside (should work!)
curl http://34.50.103.62:8081/health
```

✅ **VM2 is done!**

---

## Step 2: Deploy Middleware (VM1) with Docker

### 2.1 SSH into Middleware VM

```bash
gcloud compute ssh middleware --zone=asia-southeast2-a
```

### 2.2 Install Docker

```bash
# Update system
sudo apt-get update

# Install Docker
sudo apt-get install -y docker.io

# Add your user to docker group
sudo usermod -aG docker $USER

# Apply group changes
newgrp docker

# Verify
docker --version
```

### 2.3 Navigate to Your Code

```bash
cd ~/smartcom-tech-test
```

### 2.4 Build Docker Image

```bash
# Build Middleware image
docker build -f services/middleware/Dockerfile -t middleware .
```

### 2.5 Run Docker Container

```bash
# Run Middleware container with environment variables
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
```

**Note**: Using internal IP `10.184.0.4` for better performance since both VMs are in the same zone.

### 2.6 Verify Container is Running

```bash
# Check container status
docker ps

# Check logs
docker logs middleware

# Follow logs
docker logs -f middleware
```

### 2.7 Test Locally

```bash
# Test from inside VM
curl http://localhost:8080/health

# Exit VM
exit
```

### 2.8 Test from Outside

From your local machine:

```bash
# Test from outside (should work!)
curl http://34.128.100.247:8080/health
```

✅ **VM1 is done!**

---

## Step 3: End-to-End Testing

Now test the complete flow from your local machine:

```bash
# Send a test event to Middleware
curl -X POST http://34.128.100.247:8080/integrations/events \
  -H "Content-Type: application/json" \
  -d '{
    "source": "docker-deployment",
    "event_type": "test_alert",
    "severity": "critical",
    "message": "Testing Docker deployment on GCP VMs",
    "metadata": {
      "deployment_type": "docker",
      "region": "jakarta",
      "timestamp": "2024-02-10T12:00:00Z"
    }
  }'
```

Expected response:
```json
{
  "status": "accepted",
  "event_id": "...",
  "correlation_id": "..."
}
```

### Check Logs

```bash
# Check Middleware processed the event
gcloud compute ssh middleware --zone=asia-southeast2-a --command="docker logs middleware | tail -20"

# Check External Endpoint received the alert
gcloud compute ssh external-ep --zone=asia-southeast2-a --command="docker logs external-endpoint | tail -20"
```

---

## Managing Docker Containers

### View Container Status

```bash
# SSH into VM
gcloud compute ssh middleware --zone=asia-southeast2-a

# List running containers
docker ps

# List all containers (including stopped)
docker ps -a
```

### View Logs

```bash
# View all logs
docker logs middleware

# View last 50 lines
docker logs middleware --tail 50

# Follow logs in real-time
docker logs -f middleware

# View logs with timestamps
docker logs -t middleware
```

### Restart Container

```bash
# Restart container
docker restart middleware

# Or restart External Endpoint
docker restart external-endpoint
```

### Stop Container

```bash
# Stop container
docker stop middleware

# Stop External Endpoint
docker stop external-endpoint
```

### Start Container

```bash
# Start stopped container
docker start middleware

# Start External Endpoint
docker start external-endpoint
```

### Remove Container

```bash
# Stop and remove container
docker stop middleware
docker rm middleware

# Then recreate with docker run command
```

---

## Updating Your Services

When you make code changes:

### Update External Endpoint

```bash
# SSH into VM
gcloud compute ssh external-ep --zone=asia-southeast2-a

# Navigate to code
cd ~/smartcom-tech-test

# Pull latest changes
git pull

# Stop and remove old container
docker stop external-endpoint
docker rm external-endpoint

# Rebuild image
docker build -f services/external-endpoint/Dockerfile -t external-endpoint .

# Run new container
docker run -d \
  --name external-endpoint \
  --restart always \
  -p 8081:8081 \
  -e PORT=8081 \
  external-endpoint

# Verify
docker logs external-endpoint

exit
```

### Update Middleware

```bash
# SSH into VM
gcloud compute ssh middleware --zone=asia-southeast2-a

# Navigate to code
cd ~/smartcom-tech-test

# Pull latest changes
git pull

# Stop and remove old container
docker stop middleware
docker rm middleware

# Rebuild image
docker build -f services/middleware/Dockerfile -t middleware .

# Run new container
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

# Verify
docker logs middleware

exit
```

---

## Using Docker Compose (Alternative Method)

If you prefer using Docker Compose, here's how:

### On Each VM, Install Docker Compose

```bash
# Install Docker Compose
sudo apt-get install -y docker-compose

# Verify
docker-compose --version
```

### For External Endpoint VM

Create `docker-compose.yml`:

```bash
cd ~/smartcom-tech-test
cat > docker-compose-external.yml <<'EOF'
version: '3.8'

services:
  external-endpoint:
    build:
      context: .
      dockerfile: services/external-endpoint/Dockerfile
    container_name: external-endpoint
    ports:
      - "8081:8081"
    environment:
      - PORT=8081
    restart: always
EOF

# Run with Docker Compose
docker-compose -f docker-compose-external.yml up -d

# View logs
docker-compose -f docker-compose-external.yml logs -f
```

### For Middleware VM

Create `docker-compose.yml`:

```bash
cd ~/smartcom-tech-test
cat > docker-compose-middleware.yml <<'EOF'
version: '3.8'

services:
  middleware:
    build:
      context: .
      dockerfile: services/middleware/Dockerfile
    container_name: middleware
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - EXTERNAL_ENDPOINT_URL=http://10.184.0.4:8081/external/alerts
      - QUEUE_SIZE=1000
      - WORKER_COUNT=10
      - HTTP_TIMEOUT=3s
      - MAX_RETRIES=3
      - BASE_DELAY=500ms
    restart: always
EOF

# Run with Docker Compose
docker-compose -f docker-compose-middleware.yml up -d

# View logs
docker-compose -f docker-compose-middleware.yml logs -f
```

---

## Comparison: Docker vs Systemd

| Feature | Docker | Systemd |
|---------|--------|---------|
| **Isolation** | Full container isolation | Process-level |
| **Updates** | Rebuild image, replace container | Rebuild binary, restart service |
| **Portability** | Same everywhere | Platform-specific |
| **Logs** | `docker logs` | `journalctl` |
| **Restart Policy** | `--restart always` | systemd unit |
| **Resource Limits** | Easy with Docker flags | Requires systemd config |
| **Learning Curve** | Docker knowledge needed | Systemd knowledge needed |

**Recommendation**: Docker is better for modern deployments and easier updates!

---

## Quick Reference Commands

### Docker Commands (Run on VM)

```bash
# List running containers
docker ps

# View logs
docker logs -f middleware

# Restart container
docker restart middleware

# Stop container
docker stop middleware

# Start container
docker start middleware

# Remove container
docker rm middleware

# View container stats
docker stats middleware

# Execute command in running container
docker exec -it middleware sh
```

### Remote Commands (From Local Machine)

```bash
# View Middleware logs
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker logs middleware --tail 50"

# Restart Middleware
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker restart middleware"

# View External Endpoint logs
gcloud compute ssh external-ep --zone=asia-southeast2-a \
  --command="docker logs external-endpoint --tail 50"

# Check container status
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker ps"
```

---

## Troubleshooting

### Container Won't Start

```bash
# Check container logs
docker logs middleware

# Check if port is already in use
sudo netstat -tlnp | grep 8080

# Remove old container if exists
docker rm -f middleware

# Try running in foreground to see errors
docker run --rm -it \
  -p 8080:8080 \
  -e PORT=8080 \
  middleware
```

### Cannot Access from Outside

```bash
# Check if container is running
docker ps

# Check port mapping
docker port middleware

# Check firewall rules
gcloud compute firewall-rules list --filter="name~middleware"

# Test from inside VM
curl http://localhost:8080/health

# Check if container is listening
docker exec middleware netstat -tlnp
```

### Build Fails

```bash
# Clean Docker cache
docker system prune -a

# Rebuild without cache
docker build --no-cache -f services/middleware/Dockerfile -t middleware .

# Check if you're in the right directory
pwd  # Should be ~/smartcom-tech-test
```

---

## Automatic Cleanup Script

Create a cleanup script on each VM:

```bash
cat > ~/cleanup-docker.sh <<'EOF'
#!/bin/bash
# Stop and remove containers
docker stop middleware external-endpoint 2>/dev/null
docker rm middleware external-endpoint 2>/dev/null

# Remove images
docker rmi middleware external-endpoint 2>/dev/null

# Clean up unused resources
docker system prune -f

echo "Cleanup complete!"
EOF

chmod +x ~/cleanup-docker.sh
```

---

## Service URLs

After deployment, your services are accessible at:

- **Middleware**: http://34.128.100.247:8080
- **External Endpoint**: http://34.50.103.62:8081

---

## Complete Deployment Script

Want to automate the entire deployment? Save this on each VM:

### For External Endpoint VM

```bash
cat > ~/deploy-external.sh <<'EOF'
#!/bin/bash
set -e

cd ~/smartcom-tech-test

echo "Pulling latest code..."
git pull

echo "Stopping old container..."
docker stop external-endpoint 2>/dev/null || true
docker rm external-endpoint 2>/dev/null || true

echo "Building new image..."
docker build -f services/external-endpoint/Dockerfile -t external-endpoint .

echo "Starting new container..."
docker run -d \
  --name external-endpoint \
  --restart always \
  -p 8081:8081 \
  -e PORT=8081 \
  external-endpoint

echo "Checking status..."
sleep 2
docker ps | grep external-endpoint
docker logs external-endpoint --tail 10

echo "✅ External Endpoint deployed successfully!"
EOF

chmod +x ~/deploy-external.sh
```

### For Middleware VM

```bash
cat > ~/deploy-middleware.sh <<'EOF'
#!/bin/bash
set -e

cd ~/smartcom-tech-test

echo "Pulling latest code..."
git pull

echo "Stopping old container..."
docker stop middleware 2>/dev/null || true
docker rm middleware 2>/dev/null || true

echo "Building new image..."
docker build -f services/middleware/Dockerfile -t middleware .

echo "Starting new container..."
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

echo "Checking status..."
sleep 2
docker ps | grep middleware
docker logs middleware --tail 10

echo "✅ Middleware deployed successfully!"
EOF

chmod +x ~/deploy-middleware.sh
```

### Usage

```bash
# On each VM, just run:
~/deploy-external.sh
# or
~/deploy-middleware.sh
```

---

## Summary

✅ **Docker is now managing your services!**
✅ **Containers auto-restart on VM reboot**
✅ **Services are accessible from outside via mapped ports**
✅ **Easy to update: rebuild image + replace container**
✅ **Better isolation than systemd**

**Your services are production-ready with Docker!** 🐳🚀
