This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/create-next-app).

## Environment Variables

Create a `.env.local` file in the root of this directory with the following variables:

### Required Variables

- `CLERK_WEBHOOK_SECRET` - Webhook signing secret from Clerk dashboard (used to verify webhook requests)
- `NEXT_PUBLIC_SUPABASE_URL` - Your Supabase project URL
- `SUPABASE_SECRET_KEY` - Supabase service role key (used for server-side operations that bypass RLS)

### Clerk Setup

To set up the Clerk webhook:

1. Go to your Clerk Dashboard → Webhooks → Add Endpoint
2. Set the endpoint URL to: `https://your-domain.com/webhooks/clerk`
3. Select the following events to subscribe to:
   - `user.created`
   - `user.updated`
   - `user.deleted`
4. Copy the webhook signing secret and add it to your `.env.local` as `CLERK_WEBHOOK_SECRET`

## Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load Inter, a custom Google Font.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
