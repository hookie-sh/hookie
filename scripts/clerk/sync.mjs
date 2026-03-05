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

const clerk = createClerkClient({ secretKey: CLERK_SECRET_KEY });
const supabase = createClient(SUPABASE_URL, SUPABASE_SECRET_KEY);

const LIMIT = 100;

async function fetchAllUsers() {
  const users = [];
  let offset = 0;
  let hasMore = true;
  while (hasMore) {
    const { data } = await clerk.users.getUserList({ limit: LIMIT, offset });
    users.push(...data);
    hasMore = data.length === LIMIT;
    offset += LIMIT;
  }
  return users;
}

async function fetchAllOrganizations() {
  const orgs = [];
  let offset = 0;
  let hasMore = true;
  while (hasMore) {
    const { data } = await clerk.organizations.getOrganizationList({
      limit: LIMIT,
      offset,
    });
    orgs.push(...data);
    hasMore = data.length === LIMIT;
    offset += LIMIT;
  }
  return orgs;
}

async function fetchAllMemberships(organizationId) {
  const memberships = [];
  let offset = 0;
  let hasMore = true;
  while (hasMore) {
    const { data } =
      await clerk.organizations.getOrganizationMembershipList({
        organizationId,
        limit: LIMIT,
        offset,
      });
    memberships.push(...data);
    hasMore = data.length === LIMIT;
    offset += LIMIT;
  }
  return memberships;
}

console.log("Syncing users from Clerk…");
const users = await fetchAllUsers();
for (const user of users) {
  const email = user.emailAddresses?.[0]?.emailAddress;
  if (!email) {
    console.warn(`Skipping user ${user.id}: no primary email`);
    continue;
  }
  const { error } = await supabase.from("users").upsert(
    {
      id: user.id,
      first_name: user.firstName ?? null,
      last_name: user.lastName ?? null,
      email,
      image_url: user.imageUrl ?? null,
      created_at: new Date(user.createdAt).toISOString(),
      updated_at: new Date(user.updatedAt).toISOString(),
      last_active_at: new Date(user.lastActiveAt ?? user.updatedAt).toISOString(),
    },
    { onConflict: "id" }
  );
  if (error) console.error("User upsert error:", user.id, error);
}
console.log(`Synced ${users.length} users.`);

console.log("Syncing organizations from Clerk…");
const organizations = await fetchAllOrganizations();
for (const org of organizations) {
  const { error } = await supabase.from("organizations").upsert(
    {
      id: org.id,
      name: org.name,
      image_url: org.imageUrl ?? null,
      created_at: new Date(org.createdAt).toISOString(),
      updated_at: new Date(org.updatedAt).toISOString(),
    },
    { onConflict: "id" }
  );
  if (error) console.error("Organization upsert error:", org.id, error);
}
console.log(`Synced ${organizations.length} organizations.`);

console.log("Syncing memberships from Clerk…");
let membershipCount = 0;
for (const org of organizations) {
  const memberships = await fetchAllMemberships(org.id);
  for (const membership of memberships) {
    const userId = membership.publicUserData?.userId ?? membership.userId;
    if (!userId) {
      console.warn(`Skipping membership ${membership.id}: no user id`);
      continue;
    }
    const { error } = await supabase.from("memberships").upsert(
      {
        id: membership.id,
        organization_id: org.id,
        user_id: userId,
        role: membership.role,
        created_at: new Date(membership.createdAt).toISOString(),
        updated_at: new Date(membership.updatedAt).toISOString(),
      },
      { onConflict: "id" }
    );
    if (error) console.error("Membership upsert error:", membership.id, error);
    else membershipCount++;
  }
}
console.log(`Synced ${membershipCount} memberships.`);

console.log("Sync complete.");
