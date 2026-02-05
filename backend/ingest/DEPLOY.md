# Ingest Service Deployment Guide

This guide covers deploying the Ingest service to Fly.io.

## Prerequisites

- [Fly.io CLI](https://fly.io/docs/getting-started/installing-flyctl/) installed
- Fly.io account created
- Redis instance accessible from Fly.io (can be Fly.io Redis, Upstash, or external)

## Initial Setup

1. **Login to Fly.io:**

   ```bash
   fly auth login
   ```

2. **Create the app:**

   ```bash
   cd backend/ingest
   fly apps create your-ingest-app-name
   ```

   Replace `your-ingest-app-name` with your desired app name.

3. **Update `fly.toml`:**
   Edit `fly.toml` and update the `app` field with your app name:
   ```toml
   app = "your-ingest-app-name"
   ```

## Environment Variables

Set the required secrets and environment variables:

```bash
fly secrets set REDIS_ADDR=your_redis_address:6379
```

Optional environment variables:

- `PORT` - Server port (default: `4000`)

**Note:** Fly.io secrets are encrypted and only available at runtime. Do not commit secrets to git.

## Deploy

Deploy the service:

```bash
fly deploy
```

The deployment will:

- Build the Go application using the builder in `fly.toml`
- Configure HTTP/HTTPS with TLS termination at Fly.io's edge
- Expose the service on ports 80 (HTTP) and 443 (HTTPS)
- Auto-scale to zero when idle (cost optimization)

## Verify Deployment

1. **Check app status:**

   ```bash
   fly status
   ```

2. **View logs:**

   ```bash
   fly logs
   ```

3. **Test health endpoint:**

   ```bash
   curl https://your-ingest-app-name.fly.dev/health
   ```

   Should return: `ok`

4. **Test webhook endpoint:**
   ```bash
   curl -X POST https://your-ingest-app-name.fly.dev/webhooks/test-app/test-topic \
     -H "Content-Type: application/json" \
     -d '{"test": "data"}'
   ```
   Should return: `{"status":"ok"}`

## Webhook URLs

After deployment, your webhook URLs will be:

```
https://your-ingest-app-name.fly.dev/webhooks/{appId}/{topicId}
```

Replace `{appId}` and `{topicId}` with your actual application and topic IDs.

## Fly.io Configuration Details

The `fly.toml` file configures:

- **Ports 80 & 443**: HTTP and HTTPS endpoints (Fly.io handles certificate management)
- **Internal port 4000**: Plain HTTP connection to your app (Fly.io terminates TLS)
- **Auto-scaling**: Machines scale to zero when idle to save costs

## Auto-Scaling Behavior

The service is configured to:

- Scale to zero when idle (`min_machines_running = 0`)
- Automatically start when receiving requests
- Automatically stop after inactivity

**Note:** There may be a brief cold start delay (1-2 seconds) when scaling from zero.

To disable auto-scaling and keep the service always running:

```toml
min_machines_running = 1
auto_stop_machines = false
```

## Troubleshooting

**Connection refused:**

- Verify the app is running: `fly status`
- Check logs: `fly logs`
- Ensure `PORT` matches internal_port in `fly.toml` (default: 4000)

**Webhook endpoint not responding:**

- Check logs: `fly logs`
- Verify Redis connection: Ensure `REDIS_ADDR` is correct
- Test health endpoint first

**Cold start delays:**

- If latency is critical, set `min_machines_running = 1` to keep service warm
- Monitor costs vs. latency trade-off

**Redis connection errors:**

- Verify `REDIS_ADDR` is accessible from Fly.io network
- Check Redis firewall/security groups allow Fly.io IPs
- For Fly.io Redis, use internal address format

## Scaling

Scale the app vertically or horizontally:

```bash
# Scale VM size
fly scale vm shared-cpu-1x

# Scale to multiple instances (disable auto-scaling first)
fly scale count 2
```

## Monitoring

View metrics and logs:

```bash
# Real-time logs
fly logs

# App metrics
fly status

# Monitor webhook traffic
fly logs | grep "webhook"
```

## Custom Domain

To use a custom domain:

1. **Add domain to Fly.io:**

   ```bash
   fly certs add yourdomain.com
   ```

2. **Update DNS:** Follow Fly.io's DNS instructions

3. **Update webhook URLs:** Use your custom domain instead of `.fly.dev`

## Rollback

If deployment fails, rollback to previous version:

```bash
fly releases
fly releases rollback <release-id>
```
