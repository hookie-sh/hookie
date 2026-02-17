# Ingest Service

A Go-based webhook ingestion service that receives webhook payloads and publishes them to Redis Streams for multi-consumer processing.

## Features

- Accepts webhook requests via `/topics/{topicId}` (authenticated) and `/anon/{anonTopicID}` (anonymous) endpoints
- Supports all HTTP methods (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD)
- Captures complete request context for replay:
  - HTTP method, URL, path, query parameters
  - All request headers
  - Request body (handles binary content via base64 encoding)
  - Client IP address
  - Timestamps
- Publishes events to Redis Streams for reliable multi-consumer processing
- Runs as part of the turborepo with `pnpm dev`

## Prerequisites

- Go 1.21 or later
- Redis instance running (default: `localhost:6379`)
- pnpm (for turborepo integration)

## Configuration

Environment variables:

- `PORT` - Server port (default: `4000`)
- `REDIS_ADDR` - Redis connection address (default: `localhost:6379`)

## Development

Run the service as part of the monorepo:

```bash
pnpm dev
```

Or run individually:

```bash
go run main.go
```

## API

Base URL for webhooks is typically `NEXT_PUBLIC_INGEST_BASE_URL` (e.g. the web app builds URLs as `NEXT_PUBLIC_INGEST_BASE_URL/topics/{id}`).

### POST /topics/{topicId}

Accepts webhook payloads for **authenticated** topics and publishes them to Redis Streams. Requires a valid topic (resolved via Supabase); rate limits depend on the org’s plan.

**Path Parameters:**

- `topicId` - Topic identifier

**Request:**

- Accepts any HTTP method
- Accepts any content type
- Request body is read as-is (no validation)

**Response:**

- `200 OK` - Successfully published to Redis Stream
- `400 Bad Request` - Invalid request (missing/invalid path parameters, body read failure)
- `404 Not Found` - Topic not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Redis connection/publish failures

**Success Response:**

```json
{
  "status": "ok"
}
```

**Error Response:**

```json
{
  "error": "Brief error message"
}
```

### POST /anon/{anonTopicID}

Accepts webhook payloads for **anonymous** topics (e.g. playground or share links). No auth required. Events are published to Redis only when at least one client is connected to the topic; otherwise the request is accepted but the event is dropped.

**Path Parameters:**

- `anonTopicID` - Anonymous topic identifier

**Request:**

- Same as `/topics/{topicId}`: any HTTP method, any content type, body read as-is.

**Response:**

- `200 OK` - Published to Redis, or accepted but dropped (no clients connected). When dropped, response includes `"dropped": "no clients connected"`.
- `400 Bad Request` - Invalid path or body read failure
- `429 Too Many Requests` - Rate limit exceeded (anon tier: 10 req/min burst, 100/day, 64 KB max body)
- `500 Internal Server Error` - Redis/publish failures

**Success Response (published):**

```json
{
  "status": "ok"
}
```

**Success Response (dropped, no clients):**

```json
{
  "status": "ok",
  "dropped": "no clients connected"
}
```

**Error Response:** Same JSON shape as `/topics/{topicId}`.

Anonymous streams use the key `anon:topics:{anonTopicID}` (see Redis Stream Format below).

## Redis Stream Format

Events are published to Redis Streams with the following key format:

- **Authenticated topics:** `topics:{topicId}`
- **Anonymous topics:** `anon:topics:{anonTopicID}`

Each stream entry contains the following fields:

- `method` - HTTP method
- `url` - Full request URL including query parameters
- `path` - Request path
- `query` - Query parameters as JSON object
- `headers` - All request headers as JSON object
- `body` - Request body (base64 encoded)
- `content_type` - Content-Type header value
- `content_length` - Content-Length header value
- `remote_addr` - Client IP address
- `timestamp` - Unix timestamp in nanoseconds
- `topic_id` - Topic ID from URL path

This format ensures complete request reconstruction for downstream services.
