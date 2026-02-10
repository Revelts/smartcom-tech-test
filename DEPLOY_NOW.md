# 🚀 Deploy Your Services NOW with Docker

**You have VMs ✅ | Code is on VMs ✅ | Ready to deploy with Docker!**

---

## Your Setup

- ✅ **VM1 (Middleware)**: `34.128.100.247` - Jakarta (asia-southeast2-a)
- ✅ **VM2 (External Endpoint)**: `34.50.103.62` - Jakarta (asia-southeast2-a)
- ✅ **Code**: Already on VMs via git

---

## Deploy in 3 Steps (10 Minutes)

### Step 1: Deploy External Endpoint (5 min)

```bash
# 1.1 SSH into VM
gcloud compute ssh external-ep --zone=asia-southeast2-a

# 1.2 Install Docker (if not already installed)
sudo apt-get update && sudo apt-get install -y docker.io
sudo usermod -aG docker $USER
newgrp docker

# 1.3 Go to your code
cd ~/smartcom-tech-test

# 1.4 Build Docker image
docker build -f services/external-endpoint/Dockerfile -t external-endpoint .

# 1.5 Run container (accessible from outside!)
docker run -d \
  --name external-endpoint \
  --restart always \
  -p 8081:8081 \
  -e PORT=8081 \
  external-endpoint

# 1.6 Verify it's running
docker logs external-endpoint
curl http://localhost:8081/health

# 1.7 Exit
exit
```

**Test from your computer:**
```bash
curl http://34.50.103.62:8081/health
# Should return: {"status":"ok"}
```

✅ **VM2 Done!**

---

### Step 2: Deploy Middleware (5 min)

```bash
# 2.1 SSH into VM
gcloud compute ssh middleware --zone=asia-southeast2-a

# 2.2 Install Docker (if not already installed)
sudo apt-get update && sudo apt-get install -y docker.io
sudo usermod -aG docker $USER
newgrp docker

# 2.3 Go to your code
cd ~/smartcom-tech-test

# 2.4 Build Docker image
docker build -f services/middleware/Dockerfile -t middleware .

# 2.5 Run container (accessible from outside!)
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

# 2.6 Verify it's running
docker logs middleware
curl http://localhost:8080/health

# 2.7 Exit
exit
```

**Test from your computer:**
```bash
curl http://34.128.100.247:8080/health
# Should return: {"status":"ok"}
```

✅ **VM1 Done!**

---

### Step 3: Test End-to-End (1 min)

```bash
# Send a test event
curl -X POST http://34.128.100.247:8080/integrations/events \
  -H "Content-Type: application/json" \
  -d '{
    "source": "docker-production",
    "event_type": "deployment_success",
    "severity": "high",
    "message": "Docker deployment successful in Jakarta!",
    "metadata": {
      "region": "asia-southeast2",
      "deployment_type": "docker"
    }
  }'
```

**Expected response:**
```json
{
  "status": "accepted",
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "correlation_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7"
}
```

**Check logs:**
```bash
# Middleware logs
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker logs middleware --tail 20"

# External Endpoint logs
gcloud compute ssh external-ep --zone=asia-southeast2-a \
  --command="docker logs external-endpoint --tail 20"
```

✅ **Everything is working!**

---

## 🎉 You're Done!

Your services are now:
- ✅ Running in Docker containers
- ✅ Accessible from outside (via ports 8080 and 8081)
- ✅ Auto-restart on crash or VM reboot
- ✅ Easy to update (rebuild image & replace container)

---

## Managing Your Containers

### View Logs

```bash
# From your computer
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker logs -f middleware"

# From inside VM
docker logs -f middleware
docker logs -f external-endpoint
```

### Restart Services

```bash
# From your computer
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker restart middleware"

# From inside VM
docker restart middleware
docker restart external-endpoint
```

### Check Status

```bash
# From your computer
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker ps"

# From inside VM
docker ps
```

---

## Updating Your Code

When you make changes:

```bash
# 1. SSH into VM
gcloud compute ssh middleware --zone=asia-southeast2-a

# 2. Pull latest code
cd ~/smartcom-tech-test && git pull

# 3. Stop old container
docker stop middleware && docker rm middleware

# 4. Rebuild image
docker build -f services/middleware/Dockerfile -t middleware .

# 5. Run new container
docker run -d --name middleware --restart always \
  -p 8080:8080 \
  -e PORT=8080 \
  -e EXTERNAL_ENDPOINT_URL=http://10.184.0.4:8081/external/alerts \
  -e QUEUE_SIZE=1000 -e WORKER_COUNT=10 \
  -e HTTP_TIMEOUT=3s -e MAX_RETRIES=3 -e BASE_DELAY=500ms \
  middleware

# 6. Exit
exit
```

---

## Quick Reference

### Your Service URLs

- **Middleware**: http://34.128.100.247:8080
- **External Endpoint**: http://34.50.103.62:8081

### Essential Commands

```bash
# View logs
docker logs -f middleware

# Restart container
docker restart middleware

# Check status
docker ps

# View container stats
docker stats middleware
```

### Remote Commands

```bash
# View Middleware logs from your computer
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker logs middleware --tail 50"

# Restart Middleware from your computer
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="docker restart middleware"
```

---

## Need Help?

- **Complete Docker guide**: [DOCKER_VM_DEPLOYMENT.md](DOCKER_VM_DEPLOYMENT.md)
- **Quick reference**: [DOCKER_QUICKSTART.md](DOCKER_QUICKSTART.md)
- **Your VMs info**: [VM_DEPLOYMENT_GUIDE.md](VM_DEPLOYMENT_GUIDE.md)
- **Architecture**: [README.md](README.md)

---

## Troubleshooting

### Container won't start?

```bash
docker logs middleware
```

### Can't access from outside?

```bash
# Test from inside VM first
curl http://localhost:8080/health

# Check if container is running
docker ps

# Check firewall rules
gcloud compute firewall-rules list --filter="name~middleware"
```

### Build fails?

```bash
# Make sure you're in the right directory
cd ~/smartcom-tech-test
pwd

# Rebuild without cache
docker build --no-cache -f services/middleware/Dockerfile -t middleware .
```

---

**Your services are production-ready with Docker!** 🐳🚀

**Start deploying now - it only takes 10 minutes!**
