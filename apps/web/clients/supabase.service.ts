import { createClient } from "@supabase/supabase-js";

const supabaseUrl = process.env.SUPABASE_URL;
const supabaseSecretKey = process.env.SUPABASE_SECRET_KEY;

if (!supabaseUrl || !supabaseSecretKey) {
  throw new Error("Supabase credentials are not set");
}

export const supabaseServiceClient = createClient(
  supabaseUrl,
  supabaseSecretKey,
);
