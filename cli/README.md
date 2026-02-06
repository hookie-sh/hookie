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

### Global Configuration

The CLI stores authentication tokens securely in your system's keychain (macOS Keychain or equivalent). Your user ID and optional relay URL are stored locally.

### Repository Configuration

You can create a `hookie.yml` file in your repository to configure app_id, forward URLs, and per-topic forwarding. This allows team members to run `hookie listen` without specifying flags.

#### File Format

Create `hookie.yml` in your repository root:

```yaml
app_id: app_xxx
forward: http://localhost:3001/webhooks
topics:
  topic_abc: http://localhost:3002/webhooks/topic-abc
  topic_def: http://localhost:3003/webhooks/topic-def
```

- `app_id`: Application ID to subscribe to (optional, can also use `topic_id`)
- `forward`: Default forward URL for all events (optional)
- `topics`: Map of topic_id -> forward URL for per-topic forwarding (optional)

#### Configuration Discovery

The CLI searches for `hookie.yml` starting from the current working directory and walks up the directory tree until found or reaching the filesystem root. The closest config file takes precedence.

#### Priority Order

When running `hookie listen`, configuration is resolved in this order:
1. CLI flags (`--app-id`, `--forward`, etc.)
2. Repository config (`hookie.yml`)
3. Interactive selector (if no flags or config)

#### Initializing Configuration

Use `hookie init` to interactively create a `hookie.yml` file:

```bash
hookie init
```

This will:
- Prompt you to select an application
- Optionally configure a default forward URL
- Create `hookie.yml` in the current directory

#### Per-Topic Forwarding

You can configure different forward URLs for different topics:

```yaml
app_id: app_xxx
forward: http://localhost:3001/webhooks  # Default for all topics
topics:
  payments: http://localhost:3002/payments  # Specific URL for payments topic
  webhooks: http://localhost:3003/webhooks  # Specific URL for webhooks topic
```

When an event arrives for a topic with a specific forward URL, that URL is used. Otherwise, the default `forward` URL is used (if provided).

#### Dependencies

Repository configuration requires the `gopkg.in/yaml.v3` package, which is included in `go.mod`.
