# Hookie CLI

Authenticate with Clerk, list applications/topics, stream webhook events in real time.

## Setup

```bash
go mod download
# Generate protobuf (requires protoc + go plugins)
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/relay.proto
go build -o hookie main.go
```

Env (optional for dev): `CLERK_PUBLISHABLE_KEY`, `CLERK_SECRET_KEY`, `HOOKIE_RELAY_URL` (default `localhost:50051`).

## Usage

```bash
hookie login                    # Auth (opens browser)
hookie apps list                # List applications
hookie topics <app-id>          # List topics
hookie listen --app-id <app-id> # Stream events (optional: --topic-id, --org-id)
hookie init                     # Create hookie.yml in repo
```

**Repository config:** `hookie.yml` in repo root can set `app_id`, `forward`, and `topics` (topic_id → forward URL). CLI discovers it from cwd upward. Flags override config.
