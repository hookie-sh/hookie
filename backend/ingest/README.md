# Ingest service

Receives webhook payloads and publishes them to Redis Streams.

## Environment

See [.env.example](.env.example). Main vars: `PORT` (default `4000`), `REDIS_ADDR` (default `localhost:6379`).

## Run

From repo root: `pnpm dev`. Or from this directory: `go run main.go`.

## Endpoints

- **POST /topics/{topicId}** — Authenticated webhooks (topic from Supabase, rate limits by plan).
- **POST /anon/{anonTopicID}** — Anonymous webhooks (e.g. playground); no auth; event dropped if no clients connected.

Both accept any HTTP method and body; request is captured and published to Redis (`topics:{topicId}` or `anon:topics:{anonTopicID}`).
