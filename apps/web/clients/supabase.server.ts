import "server-only";

import { auth } from "@clerk/nextjs/server";
import { createClient, type SupabaseClient } from "@supabase/supabase-js";

const supabaseUrl = process.env.SUPABASE_URL;
const supabasePublishableKey = process.env.SUPABASE_PUBLISHABLE_KEY;

if (!supabaseUrl || !supabasePublishableKey) {
  throw new Error("Supabase credentials are not set");
}

export function createSupabaseServerClient(): SupabaseClient {
  return createClient(supabaseUrl as string, supabasePublishableKey as string, {
    accessToken: async () => {
      const { getToken } = await auth();
      const clerkToken = await getToken();
      return clerkToken || null;
    },
  });
}
