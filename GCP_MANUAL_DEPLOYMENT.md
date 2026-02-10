# GCP Manual Deployment Guide

Complete step-by-step manual deployment guide for deploying microservices to separate Google Cloud VMs without automation scripts.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Architecture Overview](#architecture-overview)
3. [Step 1: Initial Setup](#step-1-initial-setup)
4. [Step 2: Create VMs](#step-2-create-vms)
5. [Step 3: Configure Firewall Rules](#step-3-configure-firewall-rules)
6. [Step 4: Deploy External Endpoint Service (VM2)](#step-4-deploy-external-endpoint-service-vm2)
7. [Step 5: Deploy Middleware Service (VM1)](#step-5-deploy-middleware-service-vm1)
8. [Step 6: Testing](#step-6-testing)
9. [Step 7: Monitoring and Logs](#step-7-monitoring-and-logs)
10. [Troubleshooting](#troubleshooting)
11. [Updating Services](#updating-services)
12. [Cleanup](#cleanup)

---

## Prerequisites

### Required Tools

1. **Google Cloud SDK (gcloud)**
   ```bash
   # macOS
   brew install --cask google-cloud-sdk
   
   # Linux
   curl https://sdk.cloud.google.com | bash
   exec -l $SHELL
   
   # Windows
   # Download from: https://cloud.google.com/sdk/docs/install
   ```

2. **Authenticate with Google Cloud**
   ```bash
   gcloud auth login
   gcloud auth application-default login
   ```

3. **Project Requirements**
   - Active GCP project with billing enabled
   - Compute Engine API enabled

### Information You'll Need

- GCP Project ID
- Preferred region (e.g., `asia-southeast1`)
- Preferred zone (e.g., `asia-southeast1-a`)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Internet/Clients                     │
└───────────────────────┬─────────────────────────────────┘
                        │
         ┌──────────────┴──────────────┐
         │                             │
    ┌────▼────────────┐         ┌──────▼─────────────┐
    │      VM1        │         │       VM2          │
    │  (Middleware)   │────────>│(External Endpoint) │
    │                 │         │                    │
    │  Port: 8080     │         │  Port: 8081        │
    │  Public IP      │         │  Public IP         │
    └─────────────────┘         └────────────────────┘
```

**VM1 (Middleware Service)**:
- Receives events from external clients
- Processes and normalizes events
- Forwards to External Endpoint

**VM2 (External Endpoint Service)**:
- Receives processed alerts
- Logs and responds to events

---

## Step 1: Initial Setup

### 1.1 Set Your Project

```bash
# Replace with your actual project ID
export PROJECT_ID="your-project-id"
gcloud config set project $PROJECT_ID
```

### 1.2 Set Your Region and Zone

Choose based on your location (see recommendations below):

```bash
# For Southeast Asia (Singapore) - Recommended
export REGION="asia-southeast1"
export ZONE="asia-southeast1-a"

# For Japan (Tokyo)
# export REGION="asia-northeast1"
# export ZONE="asia-northeast1-a"

# For India (Mumbai)
# export REGION="asia-south1"
# export ZONE="asia-south1-a"

# For Hong Kong
# export REGION="asia-east2"
# export ZONE="asia-east2-a"
```

### 1.3 Enable Required APIs

```bash
gcloud services enable compute.googleapis.com
```

### 1.4 Verify Setup

```bash
gcloud config list
```

Expected output:
```
[core]
account = your-email@example.com
project = your-project-id
```

---

## Step 2: Create VMs

### 2.1 Create VM2 (External Endpoint) First

```bash
gcloud compute instances create external-endpoint-vm \
  --project=$PROJECT_ID \
  --zone=$ZONE \
  --machine-type=e2-small \
  --network-interface=network-tier=PREMIUM,subnet=default \
  --maintenance-policy=MIGRATE \
  --provisioning-model=STANDARD \
  --tags=http-server,external-endpoint \
  --create-disk=auto-delete=yes,boot=yes,device-name=external-endpoint-vm,image=projects/ubuntu-os-cloud/global/images/ubuntu-2204-jammy-v20250110,mode=rw,size=20,type=pd-balanced \
  --shielded-vtpm \
  --shielded-integrity-monitoring \
  --labels=service=external-endpoint
```

**Wait for creation** (takes about 30-60 seconds):
```bash
# Check status
gcloud compute instances list --filter="name=external-endpoint-vm"
```

### 2.2 Create VM1 (Middleware)

```bash
gcloud compute instances create middleware-service-vm \
  --project=$PROJECT_ID \
  --zone=$ZONE \
  --machine-type=e2-small \
  --network-interface=network-tier=PREMIUM,subnet=default \
  --maintenance-policy=MIGRATE \
  --provisioning-model=STANDARD \
  --tags=http-server,middleware \
  --create-disk=auto-delete=yes,boot=yes,device-name=middleware-vm,image=projects/ubuntu-os-cloud/global/images/ubuntu-2204-jammy-v20250110,mode=rw,size=20,type=pd-balanced \
  --shielded-vtpm \
  --shielded-integrity-monitoring \
  --labels=service=middleware
```

### 2.3 Get VM IP Addresses

```bash
# Get External Endpoint IP
export EXTERNAL_IP=$(gcloud compute instances describe external-endpoint-vm \
  --zone=$ZONE \
  --format='get(networkInterfaces[0].accessConfigs[0].natIP)')

echo "External Endpoint IP: $EXTERNAL_IP"

# Get Middleware IP
export MIDDLEWARE_IP=$(gcloud compute instances describe middleware-service-vm \
  --zone=$ZONE \
  --format='get(networkInterfaces[0].accessConfigs[0].natIP)')

echo "Middleware IP: $MIDDLEWARE_IP"
```

**Save these IPs** - you'll need them throughout the deployment!

---

## Step 3: Configure Firewall Rules

### 3.1 Create Firewall Rule for External Endpoint (Port 8081)

```bash
gcloud compute firewall-rules create allow-external-endpoint \
  --project=$PROJECT_ID \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:8081 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=external-endpoint \
  --description="Allow traffic on port 8081 for External Endpoint service"
```

### 3.2 Create Firewall Rule for Middleware (Port 8080)

```bash
gcloud compute firewall-rules create allow-middleware \
  --project=$PROJECT_ID \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:8080 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=middleware \
  --description="Allow traffic on port 8080 for Middleware service"
```

### 3.3 Verify Firewall Rules

```bash
gcloud compute firewall-rules list --filter="name~(middleware|external-endpoint)"
```

**Security Note**: For production, restrict `--source-ranges` to specific IP addresses instead of `0.0.0.0/0`.

---

## Step 4: Deploy External Endpoint Service (VM2)

### 4.1 SSH into External Endpoint VM

```bash
gcloud compute ssh external-endpoint-vm --zone=$ZONE
```

You're now inside the VM. Run all commands in steps 4.2-4.8 inside the VM.

### 4.2 Update System

```bash
sudo apt-get update
sudo apt-get upgrade -y
```

### 4.3 Install Go

```bash
# Download Go 1.23.5
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz

# Remove any previous Go installation and extract
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin

# Verify installation
go version
```

Expected output: `go version go1.23.5 linux/amd64`

### 4.4 Install Git (if needed)

```bash
sudo apt-get install -y git
```

### 4.5 Upload Your Code

**Option A: Using Git (Recommended)**

```bash
# Clone your repository
cd ~
git clone https://github.com/your-username/smartcom-tech-test.git
cd smartcom-tech-test
```

**Option B: Upload from Local Machine**

Exit the VM first (type `exit`), then from your local machine:

```bash
# From your local machine
gcloud compute scp --recurse ~/smartcom-tech-test external-endpoint-vm:~ --zone=$ZONE
```

Then SSH back in:
```bash
gcloud compute ssh external-endpoint-vm --zone=$ZONE
cd ~/smartcom-tech-test
```

### 4.6 Build External Endpoint Service

```bash
cd ~/smartcom-tech-test/services/external-endpoint

# Download dependencies
go mod download

# Build the binary
go build -o external-endpoint ./cmd/main.go

# Move binary to system path
sudo mv external-endpoint /usr/local/bin/external-endpoint

# Verify binary
/usr/local/bin/external-endpoint --help
```

### 4.7 Create Systemd Service

```bash
sudo tee /etc/systemd/system/external-endpoint.service > /dev/null <<'EOF'
[Unit]
Description=External Endpoint Service
After=network.target

[Service]
Type=simple
User=YOUR_USERNAME
WorkingDirectory=/home/YOUR_USERNAME/smartcom-tech-test/services/external-endpoint
Environment="PORT=8081"
ExecStart=/usr/local/bin/external-endpoint
Restart=always
RestartSec=10
StandardOutput=append:/var/log/external-endpoint.log
StandardError=append:/var/log/external-endpoint-error.log

[Install]
WantedBy=multi-user.target
EOF
```

**Replace YOUR_USERNAME** with your actual username:
```bash
# Get your username
whoami

# Edit the service file and replace YOUR_USERNAME
sudo nano /etc/systemd/system/external-endpoint.service
# Or use sed:
sudo sed -i "s/YOUR_USERNAME/$(whoami)/g" /etc/systemd/system/external-endpoint.service
```

### 4.8 Start External Endpoint Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable external-endpoint

# Start the service
sudo systemctl start external-endpoint

# Check status
sudo systemctl status external-endpoint
```

Expected output should show `active (running)`.

### 4.9 Test External Endpoint Locally

```bash
# Test health endpoint
curl http://localhost:8081/health

# Expected response: {"status":"ok"}
```

### 4.10 Exit VM2

```bash
exit
```

### 4.11 Test External Endpoint from Outside

From your local machine:

```bash
# Test from outside
curl http://$EXTERNAL_IP:8081/health
```

If this works, VM2 is successfully deployed! ✅

---

## Step 5: Deploy Middleware Service (VM1)

### 5.1 SSH into Middleware VM

```bash
gcloud compute ssh middleware-service-vm --zone=$ZONE
```

### 5.2 Update System

```bash
sudo apt-get update
sudo apt-get upgrade -y
```

### 5.3 Install Go

```bash
# Download Go
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz

# Install Go
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin

# Verify
go version
```

### 5.4 Install Git

```bash
sudo apt-get install -y git
```

### 5.5 Upload Your Code

**Option A: Using Git**

```bash
cd ~
git clone https://github.com/your-username/smartcom-tech-test.git
cd smartcom-tech-test
```

**Option B: Upload from Local Machine**

Exit VM, then from your local machine:

```bash
gcloud compute scp --recurse ~/smartcom-tech-test middleware-service-vm:~ --zone=$ZONE
```

SSH back in:
```bash
gcloud compute ssh middleware-service-vm --zone=$ZONE
cd ~/smartcom-tech-test
```

### 5.6 Build Middleware Service

```bash
cd ~/smartcom-tech-test/services/middleware

# Download dependencies
go mod download

# Build the binary
go build -o middleware ./cmd/main.go

# Move to system path
sudo mv middleware /usr/local/bin/middleware

# Verify
/usr/local/bin/middleware --help
```

### 5.7 Get External Endpoint IP

You need the External Endpoint VM's IP address. From VM1, run:

```bash
# Get the IP (replace with the IP you saved earlier)
export EXTERNAL_ENDPOINT_IP="<PASTE-EXTERNAL-IP-HERE>"

# Or get it programmatically
export EXTERNAL_ENDPOINT_IP=$(gcloud compute instances describe external-endpoint-vm \
  --zone=$ZONE \
  --format='get(networkInterfaces[0].accessConfigs[0].natIP)')

echo "External Endpoint URL: http://$EXTERNAL_ENDPOINT_IP:8081/external/alerts"
```

### 5.8 Create Systemd Service for Middleware

**Important**: Replace the EXTERNAL_ENDPOINT_IP with the actual IP!

```bash
sudo tee /etc/systemd/system/middleware.service > /dev/null <<EOF
[Unit]
Description=Middleware Integration Service
After=network.target

[Service]
Type=simple
User=$(whoami)
WorkingDirectory=/home/$(whoami)/smartcom-tech-test/services/middleware
Environment="PORT=8080"
Environment="EXTERNAL_ENDPOINT_URL=http://$EXTERNAL_ENDPOINT_IP:8081/external/alerts"
Environment="QUEUE_SIZE=1000"
Environment="WORKER_COUNT=10"
Environment="HTTP_TIMEOUT=3s"
Environment="MAX_RETRIES=3"
Environment="BASE_DELAY=500ms"
ExecStart=/usr/local/bin/middleware
Restart=always
RestartSec=10
StandardOutput=append:/var/log/middleware.log
StandardError=append:/var/log/middleware-error.log

[Install]
WantedBy=multi-user.target
EOF
```

### 5.9 Verify Configuration

```bash
# Check the service file
cat /etc/systemd/system/middleware.service

# Make sure EXTERNAL_ENDPOINT_URL has the correct IP
```

### 5.10 Start Middleware Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service
sudo systemctl enable middleware

# Start service
sudo systemctl start middleware

# Check status
sudo systemctl status middleware
```

Should show `active (running)`.

### 5.11 Test Middleware Locally

```bash
# Test health endpoint
curl http://localhost:8080/health

# Expected: {"status":"ok"}
```

### 5.12 Check Logs

```bash
# View middleware logs
sudo journalctl -u middleware -n 50

# Follow logs in real-time
sudo journalctl -u middleware -f
```

Press `Ctrl+C` to stop following logs.

### 5.13 Exit VM1

```bash
exit
```

---

## Step 6: Testing

### 6.1 Test Middleware from Your Local Machine

```bash
# Send a test event
curl -X POST http://$MIDDLEWARE_IP:8080/integrations/events \
  -H "Content-Type: application/json" \
  -d '{
    "source": "production-system",
    "event_type": "server_alert",
    "severity": "critical",
    "message": "Server CPU usage exceeded 95%",
    "metadata": {
      "server_id": "prod-web-01",
      "region": "asia-southeast1",
      "timestamp": "2024-02-10T12:00:00Z"
    }
  }'
```

### 6.2 Expected Response

```json
{
  "status": "accepted",
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "correlation_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7"
}
```

If you see `status: "accepted"` and both IDs, the event was successfully received! ✅

### 6.3 Verify Event Processing

Check Middleware logs:

```bash
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="sudo journalctl -u middleware -n 20"
```

Look for:
- Event received log
- Worker processing event
- HTTP request sent to external endpoint

Check External Endpoint logs:

```bash
gcloud compute ssh external-endpoint-vm --zone=$ZONE --command="sudo journalctl -u external-endpoint -n 20"
```

Look for:
- Alert received log
- Correlation ID matching your event

### 6.4 Test Multiple Events

```bash
# Send multiple events to test concurrency
for i in {1..5}; do
  curl -X POST http://$MIDDLEWARE_IP:8080/integrations/events \
    -H "Content-Type: application/json" \
    -d "{
      \"source\": \"load-test\",
      \"event_type\": \"test_event_$i\",
      \"severity\": \"high\",
      \"message\": \"Load test event $i\"
    }"
  echo ""
done
```

### 6.5 Test Different Severity Levels

```bash
# Critical severity
curl -X POST http://$MIDDLEWARE_IP:8080/integrations/events \
  -H "Content-Type: application/json" \
  -d '{
    "source": "monitoring",
    "event_type": "disk_full",
    "severity": "critical",
    "message": "Disk space at 98%"
  }'

# High severity
curl -X POST http://$MIDDLEWARE_IP:8080/integrations/events \
  -H "Content-Type: application/json" \
  -d '{
    "source": "monitoring",
    "event_type": "memory_warning",
    "severity": "high",
    "message": "Memory usage at 85%"
  }'

# Medium severity
curl -X POST http://$MIDDLEWARE_IP:8080/integrations/events \
  -H "Content-Type: application/json" \
  -d '{
    "source": "monitoring",
    "event_type": "slow_response",
    "severity": "medium",
    "message": "API response time increased"
  }'
```

---

## Step 7: Monitoring and Logs

### 7.1 View Real-time Logs

**Middleware logs**:
```bash
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="sudo journalctl -u middleware -f"
```

**External Endpoint logs**:
```bash
gcloud compute ssh external-endpoint-vm --zone=$ZONE --command="sudo journalctl -u external-endpoint -f"
```

Press `Ctrl+C` to stop.

### 7.2 View Last N Lines of Logs

```bash
# Last 50 lines of Middleware logs
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="sudo journalctl -u middleware -n 50"

# Last 50 lines of External Endpoint logs
gcloud compute ssh external-endpoint-vm --zone=$ZONE --command="sudo journalctl -u external-endpoint -n 50"
```

### 7.3 View Log Files Directly

```bash
# Middleware logs
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="sudo tail -f /var/log/middleware.log"

# External Endpoint logs
gcloud compute ssh external-endpoint-vm --zone=$ZONE --command="sudo tail -f /var/log/external-endpoint.log"
```

### 7.4 Check Service Status

```bash
# Middleware status
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="sudo systemctl status middleware"

# External Endpoint status
gcloud compute ssh external-endpoint-vm --zone=$ZONE --command="sudo systemctl status external-endpoint"
```

### 7.5 Check VM Resource Usage

```bash
# Check Middleware VM resources
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="top -bn1 | head -20"

# Check External Endpoint VM resources
gcloud compute ssh external-endpoint-vm --zone=$ZONE --command="top -bn1 | head -20"
```

---

## Troubleshooting

### Issue 1: Cannot SSH into VM

**Problem**: `gcloud compute ssh` fails

**Solution**:
```bash
# Check if VM is running
gcloud compute instances list

# Start VM if stopped
gcloud compute instances start external-endpoint-vm --zone=$ZONE

# Try with explicit SSH key
gcloud compute ssh external-endpoint-vm --zone=$ZONE --ssh-key-file=~/.ssh/google_compute_engine
```

### Issue 2: Service Won't Start

**Problem**: `systemctl start` fails

**Solution**:
```bash
# Check service logs
sudo journalctl -u middleware -n 100

# Check if binary exists
ls -la /usr/local/bin/middleware

# Check binary permissions
sudo chmod +x /usr/local/bin/middleware

# Test binary manually
PORT=8080 /usr/local/bin/middleware

# Check if port is already in use
sudo netstat -tlnp | grep 8080
```

### Issue 3: Cannot Access Service from Outside

**Problem**: `curl http://$MIDDLEWARE_IP:8080` times out

**Solution**:
```bash
# 1. Check firewall rules
gcloud compute firewall-rules list --filter="name~middleware"

# 2. Test from inside the VM
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="curl http://localhost:8080/health"

# 3. Check if service is listening
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="sudo netstat -tlnp | grep 8080"

# 4. Check VM tags
gcloud compute instances describe middleware-service-vm --zone=$ZONE --format="get(tags.items)"
# Should include: middleware

# 5. Recreate firewall rule if needed
gcloud compute firewall-rules delete allow-middleware --quiet
gcloud compute firewall-rules create allow-middleware \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:8080 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=middleware
```

### Issue 4: Middleware Can't Reach External Endpoint

**Problem**: Events are accepted but not processed

**Solution**:
```bash
# 1. Check Middleware logs for errors
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="sudo journalctl -u middleware -n 50 | grep error"

# 2. Test connectivity from Middleware VM
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="curl -v http://$EXTERNAL_IP:8081/health"

# 3. Verify External Endpoint is running
gcloud compute ssh external-endpoint-vm --zone=$ZONE --command="sudo systemctl status external-endpoint"

# 4. Check EXTERNAL_ENDPOINT_URL in Middleware service
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="sudo systemctl show middleware | grep EXTERNAL_ENDPOINT_URL"
```

### Issue 5: Service Crashes After Startup

**Problem**: Service starts but stops immediately

**Solution**:
```bash
# Check detailed logs
sudo journalctl -u middleware -n 100 --no-pager

# Check for panics or errors
sudo journalctl -u middleware | grep -i "panic\|fatal\|error"

# Verify environment variables
sudo systemctl show middleware | grep Environment

# Test binary with environment variables
sudo su - $(whoami) -c "export PORT=8080 && /usr/local/bin/middleware"
```

### Issue 6: Out of Memory

**Problem**: VM runs out of memory

**Solution**:
```bash
# Check memory usage
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="free -h"

# Reduce worker count (edit service file)
gcloud compute ssh middleware-service-vm --zone=$ZONE
sudo nano /etc/systemd/system/middleware.service
# Change: Environment="WORKER_COUNT=5"  # Reduce from 10 to 5
sudo systemctl daemon-reload
sudo systemctl restart middleware

# Or upgrade VM to larger machine type
gcloud compute instances stop middleware-service-vm --zone=$ZONE
gcloud compute instances set-machine-type middleware-service-vm \
  --machine-type=e2-medium \
  --zone=$ZONE
gcloud compute instances start middleware-service-vm --zone=$ZONE
```

---

## Updating Services

### Update Middleware Code

```bash
# 1. SSH into Middleware VM
gcloud compute ssh middleware-service-vm --zone=$ZONE

# 2. Pull latest code
cd ~/smartcom-tech-test
git pull

# 3. Rebuild
cd services/middleware
go build -o /tmp/middleware ./cmd/main.go

# 4. Stop service
sudo systemctl stop middleware

# 5. Replace binary
sudo mv /tmp/middleware /usr/local/bin/middleware

# 6. Start service
sudo systemctl start middleware

# 7. Verify
sudo systemctl status middleware

# 8. Exit
exit
```

### Update External Endpoint Code

```bash
# 1. SSH into External Endpoint VM
gcloud compute ssh external-endpoint-vm --zone=$ZONE

# 2. Pull latest code
cd ~/smartcom-tech-test
git pull

# 3. Rebuild
cd services/external-endpoint
go build -o /tmp/external-endpoint ./cmd/main.go

# 4. Stop service
sudo systemctl stop external-endpoint

# 5. Replace binary
sudo mv /tmp/external-endpoint /usr/local/bin/external-endpoint

# 6. Start service
sudo systemctl start external-endpoint

# 7. Verify
sudo systemctl status external-endpoint

# 8. Exit
exit
```

### Update Configuration

To change environment variables (e.g., QUEUE_SIZE, WORKER_COUNT):

```bash
# 1. SSH into VM
gcloud compute ssh middleware-service-vm --zone=$ZONE

# 2. Edit service file
sudo nano /etc/systemd/system/middleware.service

# 3. Modify Environment variables
# Change: Environment="WORKER_COUNT=20"

# 4. Save and exit (Ctrl+X, Y, Enter)

# 5. Reload and restart
sudo systemctl daemon-reload
sudo systemctl restart middleware

# 6. Verify new configuration
sudo systemctl show middleware | grep Environment

# 7. Exit
exit
```

---

## Cleanup

### Stop Services

```bash
# Stop Middleware
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="sudo systemctl stop middleware"

# Stop External Endpoint
gcloud compute ssh external-endpoint-vm --zone=$ZONE --command="sudo systemctl stop external-endpoint"
```

### Delete VMs

```bash
# Delete both VMs
gcloud compute instances delete middleware-service-vm external-endpoint-vm \
  --zone=$ZONE \
  --quiet
```

### Delete Firewall Rules

```bash
# Delete firewall rules
gcloud compute firewall-rules delete allow-middleware --quiet
gcloud compute firewall-rules delete allow-external-endpoint --quiet
```

### Verify Cleanup

```bash
# Check no VMs exist
gcloud compute instances list --filter="name~(middleware|external-endpoint)"

# Check no firewall rules exist
gcloud compute firewall-rules list --filter="name~(middleware|external-endpoint)"
```

---

## Quick Reference Commands

### VM Management

```bash
# List all VMs
gcloud compute instances list

# Start VM
gcloud compute instances start VM-NAME --zone=$ZONE

# Stop VM
gcloud compute instances stop VM-NAME --zone=$ZONE

# Restart VM
gcloud compute instances reset VM-NAME --zone=$ZONE

# Delete VM
gcloud compute instances delete VM-NAME --zone=$ZONE
```

### Service Management (run inside VM)

```bash
# Start service
sudo systemctl start SERVICE-NAME

# Stop service
sudo systemctl stop SERVICE-NAME

# Restart service
sudo systemctl restart SERVICE-NAME

# Check status
sudo systemctl status SERVICE-NAME

# Enable auto-start on boot
sudo systemctl enable SERVICE-NAME

# Disable auto-start
sudo systemctl disable SERVICE-NAME

# View logs
sudo journalctl -u SERVICE-NAME -f
```

### SSH Commands

```bash
# SSH into VM
gcloud compute ssh VM-NAME --zone=$ZONE

# Run command on VM without interactive session
gcloud compute ssh VM-NAME --zone=$ZONE --command="COMMAND"

# Copy files to VM
gcloud compute scp LOCAL-FILE VM-NAME:REMOTE-PATH --zone=$ZONE

# Copy files from VM
gcloud compute scp VM-NAME:REMOTE-PATH LOCAL-FILE --zone=$ZONE

# Copy directory to VM
gcloud compute scp --recurse LOCAL-DIR VM-NAME:REMOTE-DIR --zone=$ZONE
```

### Monitoring Commands

```bash
# View real-time logs
gcloud compute ssh VM-NAME --zone=$ZONE --command="sudo journalctl -u SERVICE-NAME -f"

# View last N lines
gcloud compute ssh VM-NAME --zone=$ZONE --command="sudo journalctl -u SERVICE-NAME -n 50"

# Check service status
gcloud compute ssh VM-NAME --zone=$ZONE --command="sudo systemctl status SERVICE-NAME"

# Check resource usage
gcloud compute ssh VM-NAME --zone=$ZONE --command="top -bn1 | head -20"
```

---

## Configuration Reference

### Middleware Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `EXTERNAL_ENDPOINT_URL` | Required | URL of external endpoint (VM2) |
| `QUEUE_SIZE` | `1000` | Event queue buffer size |
| `WORKER_COUNT` | `10` | Number of concurrent workers |
| `HTTP_TIMEOUT` | `3s` | HTTP client timeout |
| `MAX_RETRIES` | `3` | Maximum retry attempts |
| `BASE_DELAY` | `500ms` | Initial retry delay |

### External Endpoint Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | HTTP server port |

### VM Specifications

**e2-small (Recommended)**:
- vCPUs: 2
- Memory: 2 GB
- Cost: ~$15-17/month (Asia)

**e2-micro (Budget option for External Endpoint)**:
- vCPUs: 0.25-2 (shared)
- Memory: 1 GB
- Cost: Free tier eligible in some regions

**e2-medium (High traffic)**:
- vCPUs: 2
- Memory: 4 GB
- Cost: ~$30-35/month (Asia)

---

## Security Best Practices

### 1. Use Internal IPs for Inter-Service Communication

Instead of using public IPs between VMs:

```bash
# Get internal IP of External Endpoint
INTERNAL_IP=$(gcloud compute instances describe external-endpoint-vm \
  --zone=$ZONE \
  --format='get(networkInterfaces[0].networkIP)')

# Update Middleware service to use internal IP
sudo nano /etc/systemd/system/middleware.service
# Change: Environment="EXTERNAL_ENDPOINT_URL=http://INTERNAL_IP:8081/external/alerts"
```

### 2. Restrict Firewall Rules

```bash
# Instead of 0.0.0.0/0, use specific IP ranges
gcloud compute firewall-rules update allow-middleware \
  --source-ranges=YOUR_OFFICE_IP/32,YOUR_HOME_IP/32
```

### 3. Use Service Accounts

Create dedicated service accounts with minimal permissions:

```bash
# Create service account
gcloud iam service-accounts create middleware-sa \
  --display-name="Middleware Service Account"

# Attach to VM
gcloud compute instances set-service-account middleware-service-vm \
  --service-account=middleware-sa@$PROJECT_ID.iam.gserviceaccount.com \
  --zone=$ZONE
```

### 4. Enable OS Login

```bash
# Enable OS Login for centralized user management
gcloud compute instances add-metadata middleware-service-vm \
  --zone=$ZONE \
  --metadata=enable-oslogin=TRUE
```

### 5. Regular Updates

```bash
# Set up automatic security updates (inside VM)
sudo apt-get install -y unattended-upgrades
sudo dpkg-reconfigure --priority=low unattended-upgrades
```

---

## Production Checklist

- [ ] VMs created in correct region/zone
- [ ] Firewall rules configured and tested
- [ ] Both services running and healthy
- [ ] End-to-end event flow tested
- [ ] Logs accessible and monitoring set up
- [ ] Services configured to auto-start on boot
- [ ] Backups configured (VM snapshots)
- [ ] Security hardening applied
- [ ] Documentation updated with IP addresses
- [ ] Team trained on maintenance procedures

---

## Support and Resources

### Documentation
- **Main README**: `README.md` - Architecture overview
- **Docker Guide**: `DOCKER.md` - Docker deployment
- **Asia Zones**: `GCP_ASIA_ZONES.md` - Region selection

### GCP Resources
- **Console**: https://console.cloud.google.com
- **Compute Engine**: https://console.cloud.google.com/compute
- **Pricing Calculator**: https://cloud.google.com/products/calculator
- **Documentation**: https://cloud.google.com/compute/docs

### Commands Cheat Sheet

Save this for quick reference:

```bash
# Essential variables
export PROJECT_ID="your-project-id"
export ZONE="asia-southeast1-a"
export MIDDLEWARE_IP="<middleware-vm-ip>"
export EXTERNAL_IP="<external-endpoint-vm-ip>"

# Quick test
curl -X POST http://$MIDDLEWARE_IP:8080/integrations/events \
  -H "Content-Type: application/json" \
  -d '{"source":"test","event_type":"alert","severity":"high","message":"test"}'

# Quick logs
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="sudo journalctl -u middleware -n 20"
gcloud compute ssh external-endpoint-vm --zone=$ZONE --command="sudo journalctl -u external-endpoint -n 20"

# Quick restart
gcloud compute ssh middleware-service-vm --zone=$ZONE --command="sudo systemctl restart middleware"
gcloud compute ssh external-endpoint-vm --zone=$ZONE --command="sudo systemctl restart external-endpoint"
```

---

**Deployment Complete!** 🎉

Your microservices are now running on separate Google Cloud VMs. The system is ready to process events in production.
