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

Want to contribute? See [CONTRIBUTING.md](CONTRIBUTING.md) for local setup (you'll need your own local Supabase and a development Clerk app) and how to submit changes.

## License

Apache-2.0
