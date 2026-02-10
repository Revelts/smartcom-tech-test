# Docker VM Deployment Checklist

Quick checklist for deploying services with Docker to existing GCP VMs.

## Pre-Deployment Checklist

- [ ] VMs created in GCP
  - [ ] External Endpoint VM (e.g., `external-ep`)
  - [ ] Middleware VM (e.g., `middleware`)

- [ ] VM information collected:
  - [ ] Zone: `_____________`
  - [ ] Middleware External IP: `_____________`
  - [ ] Middleware Internal IP: `_____________`
  - [ ] External Endpoint External IP: `_____________`
  - [ ] External Endpoint Internal IP: `_____________`

- [ ] Code repository on VMs
  - [ ] Git repository cloned to `~/smartcom-tech-test`
  - [ ] Latest code pulled

- [ ] Tools installed locally:
  - [ ] `gcloud` CLI installed
  - [ ] Authenticated with `gcloud auth login`

## Deployment Steps

### Step 1: Upload Deployment Scripts

- [ ] Upload External Endpoint script:
  ```bash
  gcloud compute scp deploy-external-docker.sh external-ep:~ --zone=YOUR_ZONE
  ```

- [ ] Upload Middleware script:
  ```bash
  gcloud compute scp deploy-middleware-docker.sh middleware:~ --zone=YOUR_ZONE
  ```

### Step 2: Deploy External Endpoint (VM2)

- [ ] SSH into External Endpoint VM:
  ```bash
  gcloud compute ssh external-ep --zone=YOUR_ZONE
  ```

- [ ] Make script executable:
  ```bash
  chmod +x deploy-external-docker.sh
  ```

- [ ] Run deployment script:
  ```bash
  ./deploy-external-docker.sh
  ```

- [ ] Verify deployment:
  ```bash
  docker ps | grep external-endpoint
  docker logs external-endpoint --tail 20
  curl http://localhost:8081/health
  ```

- [ ] Exit VM:
  ```bash
  exit
  ```

### Step 3: Deploy Middleware (VM1)

- [ ] SSH into Middleware VM:
  ```bash
  gcloud compute ssh middleware --zone=YOUR_ZONE
  ```

- [ ] Make script executable:
  ```bash
  chmod +x deploy-middleware-docker.sh
  ```

- [ ] Run deployment script:
  ```bash
  ./deploy-middleware-docker.sh
  ```

- [ ] Verify deployment:
  ```bash
  docker ps | grep middleware
  docker logs middleware --tail 20
  curl http://localhost:8080/health
  ```

- [ ] Exit VM:
  ```bash
  exit
  ```

## Post-Deployment Testing

### Test from Local Machine

- [ ] Test External Endpoint health:
  ```bash
  curl http://EXTERNAL_ENDPOINT_IP:8081/health
  ```
  Expected: `{"status":"ok"}`

- [ ] Test Middleware health:
  ```bash
  curl http://MIDDLEWARE_IP:8080/health
  ```
  Expected: `{"status":"ok"}`

- [ ] Send test event:
  ```bash
  curl -X POST http://MIDDLEWARE_IP:8080/integrations/events \
    -H "Content-Type: application/json" \
    -d '{
      "source": "deployment-test",
      "event_type": "test_event",
      "severity": "critical",
      "message": "Testing Docker deployment"
    }'
  ```
  Expected: `{"status":"accepted","event_id":"...","correlation_id":"..."}`

### Verify Event Flow

- [ ] Check Middleware logs for event processing:
  ```bash
  gcloud compute ssh middleware --zone=YOUR_ZONE \
    --command="docker logs middleware --tail 50"
  ```
  Look for: "event received", "worker processing", "event sent"

- [ ] Check External Endpoint logs for received alert:
  ```bash
  gcloud compute ssh external-ep --zone=YOUR_ZONE \
    --command="docker logs external-endpoint --tail 50"
  ```
  Look for: "alert received", matching correlation_id

## Firewall Configuration

- [ ] Check firewall rules exist:
  ```bash
  gcloud compute firewall-rules list --filter="name~(middleware|external-endpoint)"
  ```

- [ ] If missing, create firewall rule for External Endpoint:
  ```bash
  gcloud compute firewall-rules create allow-external-endpoint \
    --direction=INGRESS --priority=1000 --network=default \
    --action=ALLOW --rules=tcp:8081 --source-ranges=0.0.0.0/0 \
    --target-tags=external-endpoint
  ```

- [ ] If missing, create firewall rule for Middleware:
  ```bash
  gcloud compute firewall-rules create allow-middleware \
    --direction=INGRESS --priority=1000 --network=default \
    --action=ALLOW --rules=tcp:8080 --source-ranges=0.0.0.0/0 \
    --target-tags=middleware
  ```

- [ ] Add network tags to External Endpoint VM:
  ```bash
  gcloud compute instances add-tags external-ep \
    --tags=external-endpoint --zone=YOUR_ZONE
  ```

- [ ] Add network tags to Middleware VM:
  ```bash
  gcloud compute instances add-tags middleware \
    --tags=middleware --zone=YOUR_ZONE
  ```

## Verification Checklist

### External Endpoint Container

- [ ] Container is running: `docker ps | grep external-endpoint`
- [ ] Container has status "Up"
- [ ] Port 8081 is mapped: `-p 8081:8081`
- [ ] Health endpoint responds: `curl http://localhost:8081/health`
- [ ] Accessible from outside: `curl http://EXTERNAL_IP:8081/health`
- [ ] Logs show no errors: `docker logs external-endpoint`
- [ ] Auto-restart enabled: `--restart always` in `docker inspect`

### Middleware Container

- [ ] Container is running: `docker ps | grep middleware`
- [ ] Container has status "Up"
- [ ] Port 8080 is mapped: `-p 8080:8080`
- [ ] Health endpoint responds: `curl http://localhost:8080/health`
- [ ] Accessible from outside: `curl http://EXTERNAL_IP:8080/health`
- [ ] Logs show no errors: `docker logs middleware`
- [ ] Auto-restart enabled: `--restart always` in `docker inspect`
- [ ] Environment variables set correctly

### End-to-End Flow

- [ ] Middleware accepts events
- [ ] Events are queued
- [ ] Workers process events
- [ ] HTTP requests sent to External Endpoint
- [ ] External Endpoint receives and logs alerts
- [ ] Correlation IDs match between services

## Troubleshooting

### If External Endpoint not accessible:

- [ ] Check Docker is installed: `docker --version`
- [ ] Check container is running: `docker ps`
- [ ] Check container logs: `docker logs external-endpoint`
- [ ] Check port binding: `docker port external-endpoint`
- [ ] Test from inside VM: `curl http://localhost:8081/health`
- [ ] Check firewall rules: `gcloud compute firewall-rules list`
- [ ] Check VM network tags: `gcloud compute instances describe external-ep --zone=YOUR_ZONE --format="get(tags.items)"`

### If Middleware not accessible:

- [ ] Check Docker is installed: `docker --version`
- [ ] Check container is running: `docker ps`
- [ ] Check container logs: `docker logs middleware`
- [ ] Check port binding: `docker port middleware`
- [ ] Test from inside VM: `curl http://localhost:8080/health`
- [ ] Check firewall rules: `gcloud compute firewall-rules list`
- [ ] Check VM network tags: `gcloud compute instances describe middleware --zone=YOUR_ZONE --format="get(tags.items)"`

### If containers won't start:

- [ ] Check Docker service: `sudo systemctl status docker`
- [ ] Check disk space: `df -h`
- [ ] Check memory: `free -h`
- [ ] Rebuild image: `docker build --no-cache ...`
- [ ] Check for port conflicts: `sudo netstat -tlnp | grep 8080`

### If events not flowing:

- [ ] Check Middleware can reach External Endpoint:
  ```bash
  gcloud compute ssh middleware --zone=YOUR_ZONE \
    --command="curl -v http://INTERNAL_IP:8081/health"
  ```
- [ ] Check EXTERNAL_ENDPOINT_URL is correct in Middleware
- [ ] Check both services are running
- [ ] Check logs for errors

## Post-Deployment

- [ ] Document VM IPs and zone
- [ ] Save deployment commands
- [ ] Set up monitoring (optional)
- [ ] Configure alerts (optional)
- [ ] Schedule VM snapshots (optional)
- [ ] Review security settings

## Quick Commands Reference

```bash
# View logs
docker logs -f middleware

# Restart container
docker restart middleware

# Check status
docker ps

# Update service (after code changes)
cd ~/smartcom-tech-test && git pull
docker stop middleware && docker rm middleware
./deploy-middleware-docker.sh

# SSH into container
docker exec -it middleware sh
```

## Success Criteria

✅ Both containers running with status "Up"
✅ Both health endpoints responding
✅ Test event flows end-to-end
✅ Logs show no errors
✅ Services accessible from outside
✅ Auto-restart enabled

---

**Deployment Complete!** 🎉

Your microservices are now running in Docker containers on GCP VMs, accessible from anywhere!

**Service URLs:**
- Middleware: `http://YOUR_MIDDLEWARE_IP:8080`
- External Endpoint: `http://YOUR_EXTERNAL_IP:8081`

**Next Steps:**
1. Monitor logs regularly
2. Test with production traffic
3. Set up alerts
4. Configure backups
