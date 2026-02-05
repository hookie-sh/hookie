"use client";

import { createClient } from "@supabase/supabase-js";

// For client-side, we need NEXT_PUBLIC_ prefixed env vars
// These should be set to the same values as SUPABASE_URL and SUPABASE_PUBLISHABLE_KEY
const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!;
const supabasePublishableKey =
  process.env.NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY!;

if (!supabaseUrl || !supabasePublishableKey) {
  console.error(
    "Supabase client credentials not set. Please set NEXT_PUBLIC_SUPABASE_URL and NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY environment variables."
  );
}

export const supabaseClient = createClient(
  supabaseUrl,
  supabasePublishableKey,
  {
    realtime: {
      params: {
        eventsPerSecond: 10,
      },
    },
    auth: {
      persistSession: false,
      autoRefreshToken: false,
    },
  }
);
