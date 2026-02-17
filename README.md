# Hookie

A webhook ingestion and relay platform that allows you to receive, inspect, and stream webhook events in real-time.

## Overview

Hookie provides a complete solution for webhook management:

- **Web Application**: Create applications and topics, manage webhooks through a modern UI
- **Ingest Service**: Receives webhook payloads and publishes them to Redis Streams
- **Relay Service**: Consumes events from Redis and streams them to CLI clients via gRPC
- **CLI Tool**: Stream webhook events in real-time from your terminal

## Getting started

Use the [web app](https://hookie.sh) to create applications and webhooks. See the [documentation](https://docs.hookie.sh) for detailed guides. To use the CLI, see [cli/README.md](cli/README.md) for installation and usage.

## Contributing

We welcome contributions to the CLI, docs, and backend services (ingest/relay). The web app is maintained by the core team; we're not looking for community PRs there. See [CONTRIBUTING.md](CONTRIBUTING.md) for local setup (you'll need your own local Supabase and a development Clerk app) and how to submit changes.

## Self-hosting

You are free to self-host Hookie for your own use. The Hookie team does not provide guides or support for self-hosting. You may not use this repository to self-host or operate a competing product.

## License

Apache-2.0
