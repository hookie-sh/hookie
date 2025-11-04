# Relay Service

The relay service consumes webhook events from Redis streams and broadcasts them to CLI clients via gRPC.

## Setup

1. Install dependencies:

```bash
cd backend/relay
go mod download
```

2. Generate protobuf code:

```bash
# Install protoc and plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/relay.proto
```

3. Set environment variables (can use .env file):

```bash
# Copy example and edit
cp .env.example .env

# Or set manually:
export REDIS_ADDR=localhost:6379
export GRPC_ADDR=:50051
export CLERK_SECRET_KEY=your_clerk_secret_key
export SUPABASE_URL=your_supabase_url
export SUPABASE_SECRET_KEY=your_supabase_secret_key
```

4. Run the service:

```bash
go run main.go
```

## Architecture

- **Redis Subscriber**: Consumes events from Redis streams using consumer groups
- **gRPC Server**: Provides streaming API for CLI clients
- **Auth Middleware**: Verifies Clerk JWT tokens
- **Access Control**: Verifies user ownership via Supabase queries

## API

The relay service exposes a gRPC API with three methods:

- `Subscribe`: Stream webhook events for an application/topic
- `ListApplications`: List user's applications
- `ListTopics`: List topics for an application
