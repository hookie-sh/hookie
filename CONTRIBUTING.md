# Contributing to Hookie

We welcome contributions to the CLI, docs, and backend services (ingest/relay). The web app is maintained by the core team; we're not looking for community PRs there.

## How to contribute

1. Create a feature branch
2. Make your changes
3. Ensure all tests pass and code is linted
4. Submit a pull request

## Prerequisites

- **Node.js** >= 24
- **Go** 1.25 or later
- **pnpm** (package manager)
- **Redis** (for event streaming)
- **Supabase** — run locally for development (see below)
- **Clerk** — create a development application for local auth (contributors use their own dev Clerk app)
- **Stripe** (optional, for subscriptions)

## Repository setup

```bash
git clone git@github.com:hookie-sh/hookie.git
cd hookie
pnpm install
```

## Local development

1. **Start infrastructure** (Redis and Supabase):

```bash
pnpm infra:up
```

2. **Configure environment variables** — each service has its own requirements. See:

   - [apps/web/README.md](apps/web/README.md) — Web application setup
   - [backend/ingest/README.md](backend/ingest/README.md) — Ingest service setup
   - [backend/relay/README.md](backend/relay/README.md) — Relay service setup
   - [cli/README.md](cli/README.md) — CLI setup

3. **Run development servers:**

```bash
pnpm dev
```

This starts:

- Web app on `http://localhost:3000`
- Ingest service on `http://localhost:4000`
- Relay service on `localhost:50051` (gRPC)

**Other commands:**

- Build all: `pnpm build` (or `pnpm build --filter=web` for a specific package)
- Type check: `pnpm check-types`
- Lint: `pnpm lint`

## Project structure

```
hookie/
├── apps/
│   ├── web/              # Next.js web application (main app)
│   ├── playground/       # Development playground
│   ├── gui/              # Vite React app for UI development
│   └── docs/             # Fumadocs documentation site
├── backend/
│   ├── ingest/           # Webhook ingestion service (Go)
│   └── relay/            # gRPC relay service (Go)
├── cli/                  # CLI tool (Go)
├── config/
│   ├── eslint-config/    # ESLint configurations
│   └── typescript-config/ # TypeScript configurations
├── packages/
│   └── ui/               # Shared UI components
├── supabase/             # Database migrations and config
└── scripts/              # Utility scripts
```

The **web** app is the main user-facing application. **playground**, **gui**, and **docs** are for development and documentation.

## Environment setup

Each service has its own environment configuration. Copy the relevant `.env.example` files:

- [apps/web/.env.example](apps/web/.env.example)
- [backend/ingest/.env.example](backend/ingest/.env.example)
- [backend/relay/.env.example](backend/relay/.env.example)
- [apps/gui/.env.example](apps/gui/.env.example) — optional (only if you run the GUI)
- [scripts/.env.example](scripts/.env.example) — optional (only if you run scripts)

GUI and scripts are optional; copy their `.env.example` only if you run them. For variable details, see each service's README (linked in Local development above).
