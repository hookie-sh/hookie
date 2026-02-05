# Hookie

A webhook ingestion and relay platform that allows you to receive, inspect, and stream webhook events in real-time.

## Overview

Hookie provides a complete solution for webhook management:

- **Web Application**: Create applications and topics, manage webhooks through a modern UI
- **Ingest Service**: Receives webhook payloads and publishes them to Redis Streams
- **Relay Service**: Consumes events from Redis and streams them to CLI clients via gRPC
- **CLI Tool**: Stream webhook events in real-time from your terminal

## Architecture

This is a monorepo built with Turborepo containing:

- **`apps/web`** - Next.js web application with authentication, dashboard, and webhook management
- **`backend/ingest`** - Go service that receives webhooks and publishes to Redis Streams
- **`backend/relay`** - Go gRPC service that streams events to CLI clients
- **`cli`** - Go CLI tool for authenticating and streaming webhook events
- **`packages/ui`** - Shared React component library
- **`packages/eslint-config`** - Shared ESLint configurations
- **`packages/typescript-config`** - Shared TypeScript configurations

## Prerequisites

- **Node.js** >= 18
- **Go** 1.21 or later
- **pnpm** (package manager)
- **Redis** (for event streaming)
- **Supabase** account and project
- **Clerk** account (for authentication)
- **Stripe** account (optional, for subscriptions)

## Quick Start

1. **Clone and install dependencies:**

```bash
git clone <repository-url>
cd hookie
pnpm install
```

2. **Set up infrastructure:**

```bash
# Start Redis and Supabase locally
pnpm infra:up
```

3. **Configure environment variables:**

See individual service READMEs for required environment variables:

- [`apps/web/README.md`](apps/web/README.md) - Web application setup
- [`backend/ingest/README.md`](backend/ingest/README.md) - Ingest service setup
- [`backend/relay/README.md`](backend/relay/README.md) - Relay service setup
- [`cli/README.md`](cli/README.md) - CLI setup

4. **Run development servers:**

```bash
# Run all services
pnpm dev

# Or run specific services
pnpm dev --filter=web
```

## Development

### Running Services

All services can be run together with:

```bash
pnpm dev
```

This starts:

- Web app on `http://localhost:3000`
- Ingest service on `http://localhost:4000`
- Relay service on `localhost:50051` (gRPC)

### Building

```bash
# Build all apps and packages
pnpm build

# Build specific package
pnpm build --filter=web
```

### Type Checking

```bash
pnpm check-types
```

### Linting

```bash
pnpm lint
```

## Project Structure

```
hookie/
├── apps/
│   ├── web/              # Next.js web application
│   └── playground/      # Development playground
├── backend/
│   ├── ingest/           # Webhook ingestion service (Go)
│   └── relay/            # gRPC relay service (Go)
├── cli/                  # CLI tool (Go)
├── packages/
│   ├── ui/               # Shared UI components
│   ├── eslint-config/    # ESLint configurations
│   └── typescript-config/ # TypeScript configurations
├── supabase/             # Database migrations and config
└── scripts/              # Utility scripts
```

## Environment Setup

Each service has its own environment configuration. See:

- [`apps/web/.env.example`](apps/web/.env.example)
- [`backend/ingest/.env.example`](backend/ingest/.env.example)
- [`backend/relay/.env.example`](backend/relay/.env.example)

## Deployment

Deployment guides are available for backend services:

- [`backend/ingest/DEPLOY.md`](backend/ingest/DEPLOY.md) - Deploy ingest service to Fly.io
- [`backend/relay/DEPLOY.md`](backend/relay/DEPLOY.md) - Deploy relay service to Fly.io

The web application can be deployed to Vercel or any Next.js-compatible platform.

## Contributing

1. Create a feature branch
2. Make your changes
3. Ensure all tests pass and code is linted
4. Submit a pull request

## License

Apache-2.0
