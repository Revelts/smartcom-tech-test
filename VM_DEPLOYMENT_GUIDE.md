# Your VM Deployment Guide

Quick reference for your existing VMs in **asia-southeast2-a (Jakarta)**.

## Your VMs

### VM 1: Middleware Service
- **Name**: `middleware`
- **Zone**: `asia-southeast2-a`
- **Internal IP**: `10.184.0.3`
- **External IP**: `34.128.100.247`
- **Port**: `8080`
- **Status**: ✅ Running

### VM 2: External Endpoint Service
- **Name**: `external-ep`
- **Zone**: `asia-southeast2-a`
- **Internal IP**: `10.184.0.4`
- **External IP**: `34.50.103.62`
- **Port**: `8081`
- **Status**: Running (partial)

---

## Quick Deploy Commands

### 1. Deploy External Endpoint Service (VM: external-ep)

```bash
# SSH into External Endpoint VM
gcloud compute ssh external-ep --zone=asia-southeast2-a

# Once inside the VM:
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install Go
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Clone or upload your code
git clone https://github.com/your-username/smartcom-tech-test.git
cd smartcom-tech-test

# Build External Endpoint
cd services/external-endpoint
go mod download
go build -o external-endpoint ./cmd/main.go
sudo mv external-endpoint /usr/local/bin/

# Create .env file
cat > .env <<EOF
PORT=8081
EOF

# Create systemd service
sudo tee /etc/systemd/system/external-endpoint.service > /dev/null <<EOF
[Unit]
Description=External Endpoint Service
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=/home/$USER/smartcom-tech-test/services/external-endpoint
EnvironmentFile=/home/$USER/smartcom-tech-test/services/external-endpoint/.env
ExecStart=/usr/local/bin/external-endpoint
Restart=always
RestartSec=10
StandardOutput=append:/var/log/external-endpoint.log
StandardError=append:/var/log/external-endpoint-error.log

[Install]
WantedBy=multi-user.target
EOF

# Start service
sudo systemctl daemon-reload
sudo systemctl enable external-endpoint
sudo systemctl start external-endpoint
sudo systemctl status external-endpoint

# Test locally
curl http://localhost:8081/health

# Exit VM
exit
```

### 2. Deploy Middleware Service (VM: middleware)

```bash
# SSH into Middleware VM
gcloud compute ssh middleware --zone=asia-southeast2-a

# Once inside the VM:
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install Go
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Clone or upload your code
git clone https://github.com/your-username/smartcom-tech-test.git
cd smartcom-tech-test

# Build Middleware
cd services/middleware
go mod download
go build -o middleware ./cmd/main.go
sudo mv middleware /usr/local/bin/

# Create .env file (using internal IP for better performance)
cat > .env <<EOF
PORT=8080
EXTERNAL_ENDPOINT_URL=http://10.184.0.4:8081/external/alerts
QUEUE_SIZE=1000
WORKER_COUNT=10
HTTP_TIMEOUT=3s
MAX_RETRIES=3
BASE_DELAY=500ms
EOF

# Create systemd service
sudo tee /etc/systemd/system/middleware.service > /dev/null <<EOF
[Unit]
Description=Middleware Integration Service
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=/home/$USER/smartcom-tech-test/services/middleware
EnvironmentFile=/home/$USER/smartcom-tech-test/services/middleware/.env
ExecStart=/usr/local/bin/middleware
Restart=always
RestartSec=10
StandardOutput=append:/var/log/middleware.log
StandardError=append:/var/log/middleware-error.log

[Install]
WantedBy=multi-user.target
EOF

# Start service
sudo systemctl daemon-reload
sudo systemctl enable middleware
sudo systemctl start middleware
sudo systemctl status middleware

# Test locally
curl http://localhost:8080/health

# Exit VM
exit
```

---

## Testing Your Deployment

### Test from Your Local Machine

```bash
# Test External Endpoint
curl http://34.50.103.62:8081/health

# Test Middleware
curl http://34.128.100.247:8080/health

# Send a test event
curl -X POST http://34.128.100.247:8080/integrations/events \
  -H "Content-Type: application/json" \
  -d '{
    "source": "production",
    "event_type": "server_alert",
    "severity": "critical",
    "message": "Testing Jakarta deployment",
    "metadata": {
      "region": "asia-southeast2",
      "zone": "jakarta",
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

---

## Viewing Logs

```bash
# View Middleware logs
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="sudo journalctl -u middleware -f"

# View External Endpoint logs
gcloud compute ssh external-ep --zone=asia-southeast2-a \
  --command="sudo journalctl -u external-endpoint -f"

# View last 50 lines
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="sudo journalctl -u middleware -n 50"
```

---

## Managing Services

```bash
# Restart Middleware
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="sudo systemctl restart middleware"

# Restart External Endpoint
gcloud compute ssh external-ep --zone=asia-southeast2-a \
  --command="sudo systemctl restart external-endpoint"

# Check service status
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="sudo systemctl status middleware"

# Stop service
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="sudo systemctl stop middleware"

# Start service
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="sudo systemctl start middleware"
```

---

## Uploading Code from Local Machine

If you prefer to upload code instead of using git:

```bash
# Upload to External Endpoint VM
gcloud compute scp --recurse ~/smartcom-tech-test external-ep:~ \
  --zone=asia-southeast2-a

# Upload to Middleware VM
gcloud compute scp --recurse ~/smartcom-tech-test middleware:~ \
  --zone=asia-southeast2-a
```

---

## Updating Services

When you make code changes:

### Update External Endpoint

```bash
gcloud compute ssh external-ep --zone=asia-southeast2-a

# Pull latest changes
cd ~/smartcom-tech-test
git pull

# Rebuild
cd services/external-endpoint
go build -o /tmp/external-endpoint ./cmd/main.go

# Replace binary
sudo systemctl stop external-endpoint
sudo mv /tmp/external-endpoint /usr/local/bin/external-endpoint
sudo systemctl start external-endpoint

# Verify
sudo systemctl status external-endpoint

exit
```

### Update Middleware

```bash
gcloud compute ssh middleware --zone=asia-southeast2-a

# Pull latest changes
cd ~/smartcom-tech-test
git pull

# Rebuild
cd services/middleware
go build -o /tmp/middleware ./cmd/main.go

# Replace binary
sudo systemctl stop middleware
sudo mv /tmp/middleware /usr/local/bin/middleware
sudo systemctl start middleware

# Verify
sudo systemctl status middleware

exit
```

---

## Firewall Rules Needed

Make sure these firewall rules exist:

```bash
# Allow traffic to External Endpoint (port 8081)
gcloud compute firewall-rules create allow-external-endpoint \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:8081 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=external-endpoint

# Allow traffic to Middleware (port 8080)
gcloud compute firewall-rules create allow-middleware \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:8080 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=middleware

# Verify rules exist
gcloud compute firewall-rules list --filter="name~(middleware|external-endpoint)"
```

---

## VM Network Tags

Make sure your VMs have the correct network tags:

```bash
# Add tag to External Endpoint VM
gcloud compute instances add-tags external-ep \
  --tags=external-endpoint \
  --zone=asia-southeast2-a

# Add tag to Middleware VM
gcloud compute instances add-tags middleware \
  --tags=middleware \
  --zone=asia-southeast2-a

# Verify tags
gcloud compute instances describe external-ep --zone=asia-southeast2-a --format="get(tags.items)"
gcloud compute instances describe middleware --zone=asia-southeast2-a --format="get(tags.items)"
```

---

## Quick Reference

### VM Access

```bash
# SSH into Middleware
gcloud compute ssh middleware --zone=asia-southeast2-a

# SSH into External Endpoint
gcloud compute ssh external-ep --zone=asia-southeast2-a
```

### Service URLs

- **Middleware**: http://34.128.100.247:8080
- **External Endpoint**: http://34.50.103.62:8081

### Environment Variables

| Service | Variable | Value |
|---------|----------|-------|
| Middleware | PORT | 8080 |
| Middleware | EXTERNAL_ENDPOINT_URL | http://10.184.0.4:8081/external/alerts |
| Middleware | QUEUE_SIZE | 1000 |
| Middleware | WORKER_COUNT | 10 |
| External Endpoint | PORT | 8081 |

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="sudo journalctl -u middleware -n 100"

# Check if binary exists
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="ls -la /usr/local/bin/middleware"

# Test binary manually
gcloud compute ssh middleware --zone=asia-southeast2-a
PORT=8080 /usr/local/bin/middleware
```

### Cannot Connect from Outside

```bash
# Test from inside VM
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="curl http://localhost:8080/health"

# Check firewall rules
gcloud compute firewall-rules list

# Check if service is listening
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="sudo netstat -tlnp | grep 8080"
```

### Middleware Can't Reach External Endpoint

```bash
# Test connectivity from Middleware VM
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="curl -v http://10.184.0.4:8081/health"

# If internal IP doesn't work, try external IP
gcloud compute ssh middleware --zone=asia-southeast2-a \
  --command="curl -v http://34.50.103.62:8081/health"
```

---

## Cost Estimate

**Current Setup** (2 x e2-small VMs in Jakarta):
- ~$30-35 per month if running 24/7

**To Save Money**:
```bash
# Stop VMs when not in use
gcloud compute instances stop middleware external-ep --zone=asia-southeast2-a

# Start when needed
gcloud compute instances start middleware external-ep --zone=asia-southeast2-a
```

---

## Next Steps

1. ✅ Deploy External Endpoint service first
2. ✅ Deploy Middleware service second
3. ✅ Test end-to-end with curl commands above
4. ✅ Set up monitoring and alerts
5. ✅ Configure backups (VM snapshots)
6. ✅ Review security settings

---

**Your VMs are ready! Follow the deploy commands above to get your services running.** 🚀
