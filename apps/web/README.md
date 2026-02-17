# Web app

Next.js application for Hookie: dashboard, webhook management, authentication.

## Environment

Copy [.env.example](.env.example) to `.env.local`. Required:

- `CLERK_WEBHOOK_SECRET` — Clerk webhook signing secret
- `SUPABASE_URL`, `SUPABASE_SECRET_KEY`, `SUPABASE_PUBLISHABLE_KEY`

**Clerk webhook:** Dashboard → Webhooks → Add Endpoint → `https://your-domain.com/webhooks/clerk`. Subscribe to `user.created`, `user.updated`, `user.deleted`; set the signing secret in `.env.local`.

## Run

```bash
pnpm dev
```

Open http://localhost:3000.
