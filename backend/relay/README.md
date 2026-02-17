# Relay service

Consumes webhook events from Redis and streams them to CLI clients via gRPC.

## Environment

Copy [.env.example](.env.example) to `.env`. Required: `REDIS_ADDR`, `GRPC_ADDR`, `CLERK_SECRET_KEY`, `SUPABASE_URL`, `SUPABASE_SECRET_KEY`.

## Run

From repo root: `pnpm dev`. Or from this directory: `go run main.go` (after `go mod download` and protobuf generation if needed).

## API (gRPC)

- `Subscribe` — Stream events for an app/topic
- `ListApplications` — List user’s applications
- `ListTopics` — List topics for an application
