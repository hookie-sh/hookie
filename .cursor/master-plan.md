# Hookie.sh — Pre-Launch Feature Spec

---

### Context

Anonymous channels let developers experience Hookie without creating an account. They are severely limited, auto-expire, and exist to drive the developer to signup after the "aha moment."

### 3.1 Channel Creation

**Trigger:** The CLI, when run without authentication, creates an anonymous channel.

```bash
# No login, no flags — just works
hookie listen --forward http://localhost:3000/webhooks
```

**CLI flow:**

```
Is the user authenticated (has valid Clerk token)?
  ├── Yes → list their topics, let them pick or create one (existing flow)
  └── No → call POST /api/channels/anonymous → get back channel ID + URL
            → connect to relay via gRPC with channel ID (no auth token)
            → print the URL and start listening
```

**Ingest endpoint:** `POST https://hookie.sh/wh/anon_{channelID}`

The `anon_` prefix lets the ingest server immediately identify anonymous channels without a Supabase lookup.

### 3.2 Server-Side Channel Creation Endpoint

**Endpoint:** `POST /api/channels/anonymous`

**Request:** No body needed. Read client IP from headers.

**Logic:**

```go
func CreateAnonymousChannel(ctx context.Context, r *http.Request, redisClient *redis.Client) (string, error) {
    ip := extractIP(r) // handle X-Forwarded-For from Fly.io proxy

    // Check max anonymous channels per IP (max 3)
    activeCount, err := redisClient.SCard(ctx, "anon:ip:"+ip).Result()
    if err != nil {
        return "", err
    }
    if activeCount >= 3 {
        return "", ErrTooManyAnonChannels
    }

    // Generate channel ID
    channelID := generateKSUID("anon") // e.g., "anon_2fY8xK9m3..."

    expiresAt := time.Now().Add(2 * time.Hour)

    pipe := redisClient.Pipeline()
    // Track in sorted set for expiry cleanup (from Feature 1.3)
    pipe.ZAdd(ctx, "anon:channels", redis.Z{
        Score:  float64(expiresAt.UnixMilli()),
        Member: channelID,
    })
    // Store metadata
    pipe.HSet(ctx, "anon:meta:"+channelID, map[string]interface{}{
        "ip":         ip,
        "created_at": time.Now().UTC().Format(time.RFC3339),
        "expires_at": expiresAt.UTC().Format(time.RFC3339),
    })
    pipe.Expire(ctx, "anon:meta:"+channelID, 2*time.Hour)
    // Track per-IP
    pipe.SAdd(ctx, "anon:ip:"+ip, channelID)
    pipe.Expire(ctx, "anon:ip:"+ip, 2*time.Hour)
    _, err = pipe.Exec(ctx)

    return channelID, err
}
```

**Response:**

```json
{
  "channel_id": "anon_2fY8xK9m3...",
  "url": "https://hookie.sh/wh/anon_2fY8xK9m3...",
  "expires_at": "2026-02-06T20:00:00Z",
  "limits": {
    "requests_per_day": 50,
    "requests_per_minute": 5,
    "max_payload_bytes": 65536
  }
}
```

### 3.3 Relay/gRPC Changes

The relay server currently expects authenticated connections (Clerk JWT). For anonymous channels:

- The CLI connects to the gRPC relay with a metadata header: `x-channel-type: anonymous` and `x-channel-id: anon_xxx`
- The relay validates that the channel ID exists in `anon:channels` sorted set
- **No Clerk JWT required** for anonymous connections
- The relay does NOT write to the `connected_clients` Supabase table for anonymous users (avoid polluting your data model)
- Optionally track anonymous connections in Redis instead: `anon:connected:{channelID}` with a TTL

### 3.4 CLI UX for Anonymous Flow

```
$ hookie listen --forward http://localhost:3000/webhooks

  ⚡ Hookie — anonymous session

  Webhook URL:   https://hookie.sh/wh/anon_2fY8xK9m3
  Forwarding to: http://localhost:3000/webhooks
  Expires in:    2 hours
  Limits:        50 requests/day · 5/min · 64 KB max payload

  ┌──────────────────────────────────────────────────────┐
  │ Sign up for persistent channels and higher limits:   │
  │ https://hookie.sh/signup                             │
  │                                                      │
  │ Or run: hookie auth login                            │
  └──────────────────────────────────────────────────────┘

  Waiting for webhooks...
```

When the session is about to expire (15 min before):

```
  ⚠ Anonymous session expires in 15 minutes.
    Run `hookie auth login` to keep your channels persistent.
```

When a rate limit is hit:

```
  ⚠ Daily limit reached (50/50 requests).
    Upgrade for 10,000+ requests/day: https://hookie.sh/pricing
```

### 3.5 Abuse Protection

| Protection                      | Implementation                                                                            |
| :------------------------------ | :---------------------------------------------------------------------------------------- |
| Max 3 anonymous channels per IP | Redis set `anon:ip:{ip}` with SCARD check                                                 |
| 2-hour TTL hard expiry          | Sorted set + cleanup goroutine (Feature 1.3)                                              |
| No history / no replay          | Anonymous events are never persisted to Supabase                                          |
| Aggressive rate limits          | 50/day, 5/min (Feature 2)                                                                 |
| Blocked IPs                     | Optional: maintain a Redis set `blocked:ips` for known abusers, check on channel creation |

### Todo

- [ ] Create `POST /api/channels/anonymous` endpoint in ingest or a lightweight API server
- [ ] Implement `generateKSUID("anon")` or use existing KSUID function with `anon_` prefix
- [ ] Store anonymous channel metadata in Redis (hash + sorted set + IP set)
- [ ] Enforce max 3 channels per IP
- [ ] Modify ingest handler to recognize `anon_` prefix and skip Supabase lookup
- [ ] Modify relay gRPC to accept unauthenticated connections for anonymous channels
- [ ] Validate anonymous channel ID exists in Redis on relay connect
- [ ] Do NOT write anonymous connections to `connected_clients` in Supabase
- [ ] Update CLI: detect no auth → call anonymous channel endpoint → connect without JWT
- [ ] CLI: display anonymous session banner with URL, limits, expiry
- [ ] CLI: display 15-minute expiry warning
- [ ] CLI: display upgrade nudge on rate limit hit
- [ ] Clean up anonymous channel streams on expiry (ties into Feature 1.3)
- [ ] Clean up IP tracking sets when all channels for an IP expire
- [ ] Test: create anonymous channel without auth, receive a webhook, see it in CLI
- [ ] Test: try creating 4th anonymous channel from same IP, verify rejection
- [ ] Test: verify channel stops working after 2 hours
- [ ] Test: verify Redis keys are cleaned up after expiry

---

## Feature 4: Landing Page on hookie.sh

### Context

The landing page is your storefront. It needs to convert a developer from "what is this" to "let me try it" in under 30 seconds.

### 4.1 Page Structure

This can live in your existing `apps/web` Next.js app as the root `/` route, or be a standalone static page. Keep it **one page, no navigation complexity**.

**Sections in order:**

```
┌─────────────────────────────────────────────┐
│ HERO                                        │
│ "Receive any webhook locally."              │
│ Subline: Open-source webhook relay for      │
│ developers. Like stripe listen, but for     │
│ everything.                                 │
│                                             │
│ [brew install hookie-sh/tap/hookie]  (copy) │
│ [GitHub ★]              [Read the Docs]     │
├─────────────────────────────────────────────┤
│ TERMINAL GIF / DEMO                         │
│ Animated recording showing:                 │
│ 1. hookie listen --forward localhost:3000   │
│ 2. A webhook arrives                        │
│ 3. Pretty output with headers + body        │
├─────────────────────────────────────────────┤
│ HOW IT WORKS                                │
│ 3-step visual:                              │
│ 1. Install the CLI                          │
│ 2. Point your webhook provider at your URL  │
│ 3. See events arrive locally in real-time   │
├─────────────────────────────────────────────┤
│ PRICING                                     │
│ Starter | Pro | Scale | Enterprise          │
│ (table from Feature 2 tier definitions)     │
├─────────────────────────────────────────────┤
│ FOOTER                                      │
│ GitHub · Docs · Status                      │
└─────────────────────────────────────────────┘
```

### 4.2 What to Record in the GIF

Use [vhs](https://github.com/charmbracelet/vhs) to script a reproducible terminal recording. The GIF should show:

1. Run `hookie listen --forward http://localhost:3000/webhooks`
2. See the URL printed
3. In a split pane or second terminal, `curl -X POST https://hookie.sh/wh/xxx -d '{"event":"payment.completed"}'`
4. See the webhook arrive in the CLI with pretty-printed JSON, status code from local server, latency

**Keep it under 20 seconds.** Developers will not watch a longer GIF.

### Todo

- [ ] Design landing page layout (single page, sections as above)
- [ ] Write hero copy and subline
- [ ] Create VHS tape file for terminal recording
- [ ] Record GIF, optimize with `gifsicle` (keep under 2 MB)
- [ ] Build pricing table matching tier definitions
- [ ] Add install commands (brew, go install, binary download)
- [ ] Add GitHub star button / link
- [ ] Deploy as root route on hookie.sh

---

## Feature 5: Distribution (GoReleaser + Homebrew)

### 5.1 GoReleaser Config

Create `.goreleaser.yaml` in the `cli` directory (or repo root):

```yaml
project_name: hookie
builds:
  - main: ./cli
    binary: hookie
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}} -X main.commit={{.ShortCommit}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

brews:
  - repository:
      owner: hookie-sh # your GitHub org
      name: homebrew-tap
    homepage: "https://hookie.sh"
    description: "Receive any webhook locally"
    install: |
      bin.install "hookie"

checksum:
  name_template: "checksums.txt"

changelog:
  use: github-auto
```

### 5.2 Homebrew Tap Setup

1. Create a public GitHub repo: `hookie-sh/homebrew-tap`
2. GoReleaser will auto-push the formula on release
3. Users install with: `brew install hookie-sh/tap/hookie`

### 5.3 GitHub Actions Release Workflow

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

### Todo

- [ ] Create `.goreleaser.yaml` in repo
- [ ] Create `hookie-sh/homebrew-tap` GitHub repo
- [ ] Create `HOMEBREW_TAP_TOKEN` secret (PAT with repo scope for the tap repo)
- [ ] Create `.github/workflows/release.yml`
- [ ] Test locally with `goreleaser release --snapshot --clean`
- [ ] Tag `v0.1.0` and push to trigger first release
- [ ] Verify `brew install hookie-sh/tap/hookie` works
- [ ] Add `go install` command to README for non-Homebrew users
- [ ] Add binary download links to landing page

---

## Master Launch Checklist

### Week 1: Infrastructure Hardening

- [ ] Redis Stream pruning: `MAXLEN` on XADD
- [ ] Redis Stream pruning: stale stream background cleanup
- [ ] Redis Stream pruning: anonymous channel cleanup
- [ ] Rate limiting: sliding window implementation
- [ ] Rate limiting: tier resolution (anon vs. free vs. pro)
- [ ] Rate limiting: HTTP headers + 429 responses
- [ ] Rate limiting: payload size enforcement

### Week 2: Anonymous Channels + Distribution

- [ ] Anonymous channel creation endpoint
- [ ] Ingest handler: `anon_` prefix recognition
- [ ] Relay gRPC: unauthenticated anonymous connections
- [ ] CLI: no-auth flow → anonymous channel → listen
- [ ] CLI: session banners, expiry warnings, upgrade nudges
- [ ] GoReleaser config + GitHub Actions release workflow
- [ ] Homebrew tap repo + first release
- [ ] Test full anonymous flow end-to-end

### Week 3: Landing Page + Launch

- [ ] Landing page: design, copy, pricing table
- [ ] Record terminal GIF with vhs
- [ ] README overhaul with GIF, install commands, badges
- [ ] Write Show HN post draft
- [ ] Write dev.to article draft
- [ ] **Launch day:** Show HN + Reddit + dev.to
- [ ] Monitor Redis memory, rate limit hits, error rates
- [ ] Respond to every GitHub issue within 24h
