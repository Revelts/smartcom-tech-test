# 🐳 Docker Deployment Quick Start

Deploy your services with Docker on your existing GCP VMs - accessible from outside!

## Your VMs
- **VM1 (Middleware)**: `34.128.100.247:8080`
- **VM2 (External Endpoint)**: `34.50.103.62:8081`

---

## Super Quick Deploy (5 Minutes)

### Step 1: Deploy External Endpoint

```bash
# SSH into External Endpoint VM
gcloud compute ssh external-ep --zone=asia-southeast2-a

# Install Docker (first time only)
sudo apt-get update && sudo apt-get install -y docker.io
sudo usermod -aG docker $USER
newgrp docker

# Navigate to your code
cd ~/smartcom-tech-test

# Build and run
docker build -f services/external-endpoint/Dockerfile -t external-endpoint .

docker run -d \
  --name external-endpoint \
  --restart always \
  -p 8081:8081 \
  -e PORT=8081 \
  external-endpoint

# Verify
docker logs external-endpoint
curl http://localhost:8081/health

# Exit
exit
```

### Step 2: Deploy Middleware

```bash
# SSH into Middleware VM
gcloud compute ssh middleware --zone=asia-southeast2-a

# Install Docker (first time only)
sudo apt-get update && sudo apt-get install -y docker.io
sudo usermod -aG docker $USER
newgrp docker

# Navigate to your code
cd ~/smartcom-tech-test

# Build and run
docker build -f services/middleware/Dockerfile -t middleware .

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
curl http://localhost:8080/health

# Exit
exit
```

### Step 3: Test from Your Computer

```bash
# Test External Endpoint
curl http://34.50.103.62:8081/health

# Test Middleware
curl http://34.128.100.247:8080/health

# Send test event
curl -X POST http://34.128.100.247:8080/integrations/events \
  -H "Content-Type: application/json" \
  -d '{"source":"docker-test","event_type":"alert","severity":"critical","message":"Testing Docker deployment!"}'
```

✅ **Done! Your services are running in Docker containers and accessible from outside!**

---

## Using Deployment Scripts (Even Easier)

### Upload Scripts to VMs

```bash
# Upload to External Endpoint VM
gcloud compute scp deploy-external-docker.sh external-ep:~ --zone=asia-southeast2-a

# Upload to Middleware VM
gcloud compute scp deploy-middleware-docker.sh middleware:~ --zone=asia-southeast2-a
```

### Run on Each VM

```bash
# On External Endpoint VM
gcloud compute ssh external-ep --zone=asia-southeast2-a
./deploy-external-docker.sh

# On Middleware VM
gcloud compute ssh middleware --zone=asia-southeast2-a
./deploy-middleware-docker.sh
```

---

## Essential Docker Commands

### View Logs

```bash
# Real-time logs
docker logs -f middleware

# Last 50 lines
docker logs middleware --tail 50
```

### Restart Container

```bash
docker restart middleware
docker restart external-endpoint
```

### Stop Container

```bash
docker stop middleware
docker stop external-endpoint
```

### Start Container

```bash
docker start middleware
docker start external-endpoint
```

### Check Status

```bash
docker ps
```

### View Container Stats

```bash
docker stats middleware
```

---

## Updating Your Code

When you make changes:

```bash
# SSH into VM
gcloud compute ssh middleware --zone=asia-southeast2-a

# Pull latest code
cd ~/smartcom-tech-test
git pull

# Stop old container
docker stop middleware && docker rm middleware

# Rebuild image
docker build -f services/middleware/Dockerfile -t middleware .

# Run new container
docker run -d --name middleware --restart always \
  -p 8080:8080 \
  -e PORT=8080 \
  -e EXTERNAL_ENDPOINT_URL=http://10.184.0.4:8081/external/alerts \
  -e QUEUE_SIZE=1000 -e WORKER_COUNT=10 \
  -e HTTP_TIMEOUT=3s -e MAX_RETRIES=3 -e BASE_DELAY=500ms \
  middleware

exit
```

---

## Why Docker?

✅ **Accessible from outside** - Port mapping with `-p` flag
✅ **Auto-restart** - Containers restart on crash or VM reboot
✅ **Easy updates** - Rebuild image, replace container
✅ **Isolation** - Each service in its own container
✅ **Portable** - Same image works anywhere
✅ **Clean** - No system dependencies to manage

---

## Remote Management

Manage containers from your local machine:

```bash
# View Middleware logs
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker logs middleware --tail 20"

# Restart Middleware
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker restart middleware"

# Check container status
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker ps"

# View External Endpoint logs
gcloud compute ssh external-ep --zone=asia-southeast2-a \
  --command="docker logs external-endpoint --tail 20"
```

---

## Troubleshooting

### Container won't start?

```bash
# View error logs
docker logs middleware

# Check if port is busy
sudo netstat -tlnp | grep 8080

# Remove and recreate
docker rm -f middleware
# Then run docker run command again
```

### Can't access from outside?

```bash
# Check firewall rules
gcloud compute firewall-rules list --filter="name~middleware"

# Make sure container is running
docker ps | grep middleware

# Test from inside VM first
curl http://localhost:8080/health
```

### Build fails?

```bash
# Make sure you're in the right directory
cd ~/smartcom-tech-test
pwd  # Should show: /home/your-username/smartcom-tech-test

# Clean and rebuild
docker system prune -f
docker build --no-cache -f services/middleware/Dockerfile -t middleware .
```

---

## Complete Documentation

For more details, see:
- **DOCKER_VM_DEPLOYMENT.md** - Complete Docker deployment guide
- **VM_DEPLOYMENT_GUIDE.md** - Your VM-specific guide

---

## Quick Reference

| Command | Description |
|---------|-------------|
| `docker ps` | List running containers |
| `docker logs -f <name>` | Follow logs |
| `docker restart <name>` | Restart container |
| `docker stop <name>` | Stop container |
| `docker start <name>` | Start container |
| `docker rm -f <name>` | Force remove container |
| `docker exec -it <name> sh` | Access container shell |

---

**Your services are now running in Docker and accessible from anywhere!** 🐳✨

Test URLs:
- Middleware: http://34.128.100.247:8080
- External Endpoint: http://34.50.103.62:8081
