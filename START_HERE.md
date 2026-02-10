# 🚀 START HERE - GCP Deployment

**New to this project?** Follow this simple guide to get your microservices deployed to Google Cloud.

---

## ⚡ Super Quick Start (5 Minutes)

If you just want it deployed NOW:

```bash
# 1. Make sure you have gcloud installed and authenticated
gcloud auth login

# 2. Run the deployment script
./deploy-gcp.sh

# 3. Enter your GCP Project ID when prompted
# 4. Press Enter to accept default region (Singapore)
# 5. Wait 5-10 minutes
# 6. Done! ✅
```

The script will give you the IP addresses to test your services.

---

## 📖 Want to Learn? (45 Minutes)

If you want to understand what you're deploying:

```bash
# Read the complete manual deployment guide
open GCP_MANUAL_DEPLOYMENT.md

# Then follow it step-by-step
# You'll learn:
# - How to create VMs
# - How to configure networking
# - How to deploy services
# - How to troubleshoot
```

---

## 🤔 Not Sure? (10 Minutes)

Read the comparison guide first:

```bash
open DEPLOYMENT_OPTIONS.md
```

This will help you decide between:
- **Manual deployment** (full control, learn as you go)
- **Automated deployment** (quick and easy)

---

## 📋 What This Project Does

You're deploying **2 microservices** to **2 separate VMs**:

```
┌─────────────────┐         ┌─────────────────┐
│      VM1        │         │      VM2        │
│   Middleware    │────────>│External Endpoint│
│   Port: 8080    │         │   Port: 8081    │
└─────────────────┘         └─────────────────┘
        ▲
        │
   Your clients
```

**VM1 (Middleware)**: 
- Receives events from external clients
- Processes and validates them
- Forwards to VM2 asynchronously

**VM2 (External Endpoint)**:
- Receives processed alerts
- Logs and responds

---

## ✅ Prerequisites

Before you start, make sure you have:

1. **Google Cloud Account** with billing enabled
   - Go to: https://console.cloud.google.com
   - Enable billing on your project

2. **gcloud CLI installed**
   ```bash
   # macOS
   brew install --cask google-cloud-sdk
   
   # Or download from:
   # https://cloud.google.com/sdk/docs/install
   ```

3. **Authenticated**
   ```bash
   gcloud auth login
   gcloud auth application-default login
   ```

4. **Project ID**
   ```bash
   # Find your project ID at:
   # https://console.cloud.google.com
   ```

---

## 🎯 Choose Your Path

### Path A: "Just Deploy It!"

**Time**: 10 minutes

```bash
./deploy-gcp.sh
```

Follow the prompts. Done!

### Path B: "I Want to Learn"

**Time**: 1 hour

```bash
# 1. Read the overview
cat DEPLOYMENT_OPTIONS.md

# 2. Follow the manual guide
open GCP_MANUAL_DEPLOYMENT.md

# 3. Deploy step-by-step
# (follow commands in the manual guide)
```

### Path C: "I'm Preparing for Production"

**Time**: 2 hours

```bash
# 1. Read deployment options
cat DEPLOYMENT_OPTIONS.md

# 2. Choose best region for your users
cat GCP_ASIA_ZONES.md

# 3. Follow manual deployment
open GCP_MANUAL_DEPLOYMENT.md

# 4. Complete security hardening
# (see Security Best Practices section)

# 5. Use production checklist
cat DEPLOYMENT_CHECKLIST.md
```

---

## 🌏 Where Should I Deploy?

**Default**: Singapore (`asia-southeast1-a`)

**Good for**: Southeast Asia, Australia, India, Japan

**Other options**:
- Tokyo: `asia-northeast1-a` (best for Japan)
- Mumbai: `asia-south1-a` (best for India)
- Hong Kong: `asia-east2-a` (best for Hong Kong/South China)

See full list: [GCP_ASIA_ZONES.md](GCP_ASIA_ZONES.md)

---

## 💰 How Much Will This Cost?

**With defaults (2 x e2-small VMs)**:
- ~$30-40 per month if running 24/7
- ~$1-2 per day

**To reduce costs**:
- Stop VMs when not testing
- Use smaller VMs (e2-micro for External Endpoint)
- See cost optimization tips in [GCP_QUICKSTART.md](GCP_QUICKSTART.md)

---

## 🧪 Testing Your Deployment

After deployment, test with:

```bash
# Replace with your Middleware VM IP
export MIDDLEWARE_IP="<your-vm-ip>"

# Send a test event
curl -X POST http://$MIDDLEWARE_IP:8080/integrations/events \
  -H "Content-Type: application/json" \
  -d '{
    "source": "test",
    "event_type": "test_alert",
    "severity": "high",
    "message": "Testing deployment"
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

## 🔍 Viewing Logs

```bash
# View Middleware logs
gcloud compute ssh middleware-service-vm \
  --zone=asia-southeast1-a \
  --command="sudo journalctl -u middleware -f"

# View External Endpoint logs
gcloud compute ssh external-endpoint-vm \
  --zone=asia-southeast1-a \
  --command="sudo journalctl -u external-endpoint -f"
```

Press `Ctrl+C` to stop.

---

## 🛑 Stopping/Deleting

### Stop VMs (save money, keep data)

```bash
gcloud compute instances stop middleware-service-vm external-endpoint-vm \
  --zone=asia-southeast1-a
```

### Start VMs again

```bash
gcloud compute instances start middleware-service-vm external-endpoint-vm \
  --zone=asia-southeast1-a
```

### Delete Everything (permanent!)

```bash
./cleanup-gcp.sh
```

---

## ❓ Common Questions

### Q: I don't have a GCP account

**A**: Create one at https://console.cloud.google.com
- You get $300 free credits for 90 days
- Credit card required but won't be charged without permission

### Q: Which deployment method should I use?

**A**: 
- First time? → **Manual** (learn as you go)
- In a hurry? → **Automated** (fastest)
- Production? → **Manual** (more control)

See [DEPLOYMENT_OPTIONS.md](DEPLOYMENT_OPTIONS.md) for detailed comparison.

### Q: I'm getting permission errors

**A**: Make sure you've authenticated:
```bash
gcloud auth login
gcloud auth application-default login
```

### Q: The deployment failed

**A**: Check the troubleshooting section in [GCP_MANUAL_DEPLOYMENT.md](GCP_MANUAL_DEPLOYMENT.md)

### Q: How do I update my code?

**A**: See "Updating Services" section in [GCP_MANUAL_DEPLOYMENT.md](GCP_MANUAL_DEPLOYMENT.md)

### Q: Can I deploy to US instead of Asia?

**A**: Yes! When prompted for region, enter:
- `us-central1` and `us-central1-a`
- Or `us-east1` and `us-east1-a`

---

## 📚 All Documentation

Complete list of guides:

1. **START_HERE.md** ← You are here
2. **[DEPLOYMENT_OPTIONS.md](DEPLOYMENT_OPTIONS.md)** - Choose your method
3. **[GCP_MANUAL_DEPLOYMENT.md](GCP_MANUAL_DEPLOYMENT.md)** - Step-by-step manual
4. **[GCP_QUICKSTART.md](GCP_QUICKSTART.md)** - Quick automated setup
5. **[GCP_ASIA_ZONES.md](GCP_ASIA_ZONES.md)** - Region selection
6. **[DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)** - Production checklist
7. **[GCP_DEPLOYMENT_INDEX.md](GCP_DEPLOYMENT_INDEX.md)** - Documentation index
8. **[README.md](README.md)** - Architecture overview

---

## 🎯 Recommended Next Steps

1. ✅ Make sure you meet prerequisites above
2. ✅ Decide: Quick deploy or learn? (see "Choose Your Path")
3. ✅ Follow your chosen path
4. ✅ Test your deployment
5. ✅ Review logs and monitoring
6. ✅ (Optional) Set up production security

---

## 💡 Pro Tips

- **First deployment?** Use manual method to learn
- **Testing quickly?** Use automated script
- **Going to production?** Read security section carefully
- **Multiple regions?** See [GCP_ASIA_ZONES.md](GCP_ASIA_ZONES.md) for multi-region setup
- **Save money?** Stop VMs when not in use

---

## 🚀 Ready to Deploy?

### Quick Deploy (5 min)
```bash
./deploy-gcp.sh
```

### Learn & Deploy (45 min)
```bash
open GCP_MANUAL_DEPLOYMENT.md
```

### Need Help Deciding?
```bash
open DEPLOYMENT_OPTIONS.md
```

---

## 🆘 Need Help?

1. **Deployment issues**: Check [GCP_MANUAL_DEPLOYMENT.md](GCP_MANUAL_DEPLOYMENT.md) → Troubleshooting
2. **Region selection**: See [GCP_ASIA_ZONES.md](GCP_ASIA_ZONES.md)
3. **Understanding architecture**: Read [README.md](README.md)
4. **Production setup**: Use [DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)

---

**That's it! Choose your path and start deploying.** 🎉

**Still not sure?** Read [DEPLOYMENT_OPTIONS.md](DEPLOYMENT_OPTIONS.md) first.
