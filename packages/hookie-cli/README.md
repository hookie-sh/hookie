# Hookie CLI

**Hookie** is a webhook ingestion and relay platform. It lets you receive, inspect, and stream webhook events in real time—from the [web app](https://hookie.sh) or directly in your terminal with this CLI.

This package is the **npm distribution** of the Hookie CLI. The correct native binary for your platform (macOS, Linux, Windows) is installed automatically via optional dependencies, so you can run `hookie` without building from source.

## Install

```bash
npm install -g @hookie-sh/hookie
```

Or run once without installing:

```bash
npx @hookie-sh/hookie listen
```

**Requirements:** Node.js 18+. Supported platforms: macOS (amd64, arm64), Linux (amd64, arm64), Windows (amd64).

## Quick start

1. **Sign in** (opens browser):

   ```bash
   hookie login
   ```

2. **List your applications:**

   ```bash
   hookie apps list
   ```

3. **Stream webhook events** for an app (or pick interactively):

   ```bash
   hookie listen --app-id <app-id>
   ```

4. **Optional:** Pin an app to the current repo so you can run `hookie listen` without flags:

   ```bash
   hookie init
   ```

   This creates a `hookie.yml` in the current directory. The CLI discovers it when you run commands from this repo (or any subdirectory).

## Commands

| Command | Description |
|--------|-------------|
| `hookie login` | Authenticate with Hookie (Clerk; opens browser) |
| `hookie logout` | Clear local credentials |
| `hookie apps list` | List your Hookie applications |
| `hookie topics <app-id>` | List topics for an application |
| `hookie listen` | Stream webhook events (interactive or with flags) |
| `hookie init` | Create `hookie.yml` in the current directory |

### Listen options

- `--app-id`, `-a` — Subscribe to all topics of an application.
- `--topic-id`, `-t` — Subscribe to a single topic.
- `--forward`, `-f` — Forward each event as a request to a URL (e.g. your local server).
- `--ui` — Show the local event UI when using `--forward`.
- `--org-id` — Organization ID (if you use multiple orgs).

Without `--app-id` or `--topic-id`, the CLI prompts you to choose an app or topic. If you’re not logged in, `hookie listen` runs in anonymous mode with an ephemeral channel (handy for quick tests).

## Repository config: `hookie.yml`

In a repo root you can add a `hookie.yml` to set defaults for that project:

```yaml
app_id: your-app-id
forward: https://your-server.local/webhooks
topics:
  topic-id-1: https://service-a.local/events
  topic-id-2: https://service-b.local/events
```

- **app_id** / **topic_id** — Default app (and optionally topic) for `hookie listen`.
- **forward** — Default URL to forward events to.
- **topics** — Per-topic override URLs for forwarding.

CLI flags override these values. The CLI looks for `hookie.yml` from the current directory upward.

## Links

- [Web app](https://hookie.sh) — Create apps and webhooks.
- [Documentation](https://docs.hookie.sh) — Full guides and API details.
- [GitHub](https://github.com/hookie-sh/hookie) — Source and releases.

## License

Apache-2.0
