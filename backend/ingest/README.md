# Ingest Service

A Go-based webhook ingestion service that receives webhook payloads and publishes them to Redis Streams for multi-consumer processing.

## Features

- Accepts webhook requests via `/webhooks/{appId}/{topicId}` endpoint
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

### POST /webhooks/{appId}/{topicId}

Accepts webhook payloads and publishes them to Redis Streams.

**Path Parameters:**
- `appId` - Application identifier
- `topicId` - Topic identifier

**Request:**
- Accepts any HTTP method
- Accepts any content type
- Request body is read as-is (no validation)

**Response:**
- `200 OK` - Successfully published to Redis Stream
- `400 Bad Request` - Invalid request (missing/invalid path parameters, body read failure)
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

## Redis Stream Format

Events are published to Redis Streams with the key format:
```
webhook:events:{appId}:{topicId}
```

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
- `app_id` - Application ID from URL path
- `topic_id` - Topic ID from URL path

This format ensures complete request reconstruction for downstream services.

