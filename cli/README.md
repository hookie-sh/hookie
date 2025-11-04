# Hookie CLI

Command-line tool for authenticating with Clerk, listing applications/topics, and streaming webhook events in real-time.

## Setup

1. Install dependencies:

```bash
cd cli
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

3. Build the CLI:

```bash
go build -o hookie main.go
```

4. Set environment variables (optional, for development):

```bash
export CLERK_PUBLISHABLE_KEY=your_clerk_publishable_key
export CLERK_SECRET_KEY=your_clerk_secret_key  # Only needed for CLI auth verification
export HOOKIE_RELAY_URL=localhost:50051  # Default relay URL
```

Note: For production/distributed CLI, users will configure these via their system environment or the CLI will prompt for them during `hookie login`.

## Usage

### Authentication

```bash
# Login (opens browser)
hookie login

# Logout
hookie logout
```

### List Commands

```bash
# List all applications
hookie apps list

# List applications for a specific organization
hookie apps list --org-id <org-id>

# List topics for an application
hookie topics <app-id>
```

### Listen to Events

```bash
# Listen to all topics for an application
hookie listen --app-id <app-id>

# Listen to a specific topic
hookie listen --app-id <app-id> --topic-id <topic-id>

# Listen to an organization-owned application
hookie listen --app-id <app-id> --org-id <org-id>

# Set organization globally for all commands
hookie --org-id <org-id> listen --app-id <app-id>

# Run multiple listeners (in separate terminals)
hookie listen --app-id app1 --topic-id topic1
hookie listen --app-id app2 --org-id org1  # Org-owned app
```

## Configuration

The CLI stores authentication tokens in `~/.hookie/config.json`. This file contains:

- `token`: Clerk session token
- `user_id`: Authenticated user ID
- `relay_url`: Optional relay service URL override
