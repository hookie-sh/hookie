# Relay Service Deployment Guide

This guide covers deploying the Relay service to Fly.io.

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
   cd backend/relay
   fly apps create your-relay-app-name
   ```
   Replace `your-relay-app-name` with your desired app name.

3. **Update `fly.toml`:**
   Edit `fly.toml` and update the `app` field with your app name:
   ```toml
   app = "your-relay-app-name"
   ```

## Environment Variables

Set the required secrets and environment variables:

```bash
fly secrets set CLERK_SECRET_KEY=your_clerk_secret_key
fly secrets set REDIS_ADDR=your_redis_address:6379
fly secrets set SUPABASE_URL=your_supabase_url
fly secrets set SUPABASE_SECRET_KEY=your_supabase_service_role_key
```

Optional environment variables:
- `GRPC_ADDR` - gRPC server address (default: `:50051`)

**Note:** Fly.io secrets are encrypted and only available at runtime. Do not commit secrets to git.

## Deploy

Deploy the service:

```bash
fly deploy
```

The deployment will:
- Build the Go application using the builder in `fly.toml`
- Configure gRPC with TLS termination at Fly.io's edge
- Expose the service on port 443 (HTTPS)

## Verify Deployment

1. **Check app status:**
   ```bash
   fly status
   ```

2. **View logs:**
   ```bash
   fly logs
   ```

3. **Test health (if you add a health endpoint):**
   ```bash
   fly curl http://your-relay-app-name.fly.dev/health
   ```

## Configure CLI

After deployment, configure your CLI to connect to the relay:

```bash
export HOOKIE_RELAY_URL=your-relay-app-name.fly.dev:443
```

Or set it in your CLI config. The client will automatically use TLS for non-localhost connections.

## Fly.io Configuration Details

The `fly.toml` file configures:
- **Port 443 with TLS**: External HTTPS endpoint (Fly.io handles certificate management)
- **ALPN h2**: HTTP/2 support required for gRPC
- **Internal port 50051**: Plain TCP connection to your app (Fly.io terminates TLS)

## Troubleshooting

**Connection refused:**
- Verify the app is running: `fly status`
- Check logs: `fly logs`
- Ensure `GRPC_ADDR` matches internal_port in `fly.toml`

**TLS errors:**
- Verify ALPN h2 is configured in `fly.toml`
- Ensure client uses TLS for non-localhost (already configured)

**Authentication errors:**
- Verify `CLERK_SECRET_KEY` is set correctly
- Check that tokens are being sent in gRPC metadata

## Scaling

Scale the app vertically or horizontally:

```bash
# Scale VM size
fly scale vm shared-cpu-1x

# Scale to multiple instances
fly scale count 2
```

## Monitoring

View metrics and logs:

```bash
# Real-time logs
fly logs

# App metrics
fly status
```

## Rollback

If deployment fails, rollback to previous version:

```bash
fly releases
fly releases rollback <release-id>
```

