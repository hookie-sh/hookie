#! /usr/bin/env zx
import { createClerkClient } from "@clerk/backend";
import { createClient } from "@supabase/supabase-js";
import dotenv from "dotenv";
import { resolve } from "node:path";

dotenv.config({ path: resolve(__dirname, "../.env") });

const CLERK_SECRET_KEY = process.env.CLERK_SECRET_KEY;
const SUPABASE_URL = process.env.SUPABASE_URL;
const SUPABASE_SECRET_KEY = process.env.SUPABASE_SECRET_KEY;

if (!CLERK_SECRET_KEY) {
  throw new Error("CLERK_SECRET_KEY is not set");
}

if (!SUPABASE_URL) {
  throw new Error("SUPABASE_URL is not set");
}

if (!SUPABASE_SECRET_KEY) {
  throw new Error("SUPABASE_SECRET_KEY is not set");
}

console.log("Loading client users from Clerk", CLERK_SECRET_KEY);

const clerk = createClerkClient({
  secretKey: CLERK_SECRET_KEY,
});
const supabase = createClient(SUPABASE_URL, SUPABASE_SECRET_KEY);

const { data: users } = await clerk.users.getUserList();

users.forEach(async (user) => {
  const { error } = await supabase.from("users").upsert(
    {
      id: user.id,
      first_name: user.firstName,
      last_name: user.lastName,
      email: user.emailAddresses[0].emailAddress,
      image_url: user.imageUrl,
      created_at: new Date(user.createdAt),
      updated_at: new Date(user.updatedAt),
      last_active_at: new Date(user.lastActiveAt),
    },
    { onConflict: "id" }
  );

  if (error) {
    console.error(error);
  }
});
