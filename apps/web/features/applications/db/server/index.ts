import { supabase } from "@/clients/supabase.server";
import { CreateApplicationInput } from "../../schemas/application";

export async function createApplication(input: CreateApplicationInput) {
  const { data, error } = await supabase
    .from("applications")
    .insert(input)
    .select("id, name, description, created_at, updated_at")
    .single();

  if (error) throw error;
  return data;
}

export async function getApplicationsWithTopicCountByUserId(
  userId: string,
  orgId?: string | null
) {
  // Query user-owned applications (user_id = userId AND org_id IS NULL)
  const { data: userApps, error: userError } = await supabase
    .from("applications")
    .select("*, topics(count)")
    .eq("user_id", userId)
    .is("org_id", null)
    .order("created_at", { ascending: false });

  if (userError) throw userError;

  // Query organization-owned applications if orgId is provided
  let orgApps: NonNullable<typeof userApps> | null = null;
  if (orgId) {
    const { data, error: orgError } = await supabase
      .from("applications")
      .select("*, topics(count)")
      .eq("org_id", orgId)
      .is("user_id", null)
      .order("created_at", { ascending: false });

    if (orgError) throw orgError;
    orgApps = data;
  }

  // Merge and deduplicate by ID, then sort by created_at
  const allApps = [...(userApps || []), ...(orgApps || [])];
  const uniqueApps = Array.from(
    new Map(allApps.map((app) => [app.id, app])).values()
  );
  return uniqueApps.sort(
    (a, b) =>
      new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  );
}

export async function getRecentApplicationsByUserId(
  userId: string,
  orgId?: string | null,
  limit: number = 5
) {
  // Query user-owned applications (user_id = userId AND org_id IS NULL)
  const { data: userApps, error: userError } = await supabase
    .from("applications")
    .select("*, topics(count)")
    .eq("user_id", userId)
    .is("org_id", null)
    .order("created_at", { ascending: false })
    .limit(limit);

  if (userError) throw userError;

  // Query organization-owned applications if orgId is provided
  let orgApps: NonNullable<typeof userApps> | null = null;
  if (orgId) {
    const { data, error: orgError } = await supabase
      .from("applications")
      .select("*, topics(count)")
      .eq("org_id", orgId)
      .is("user_id", null)
      .order("created_at", { ascending: false })
      .limit(limit);

    if (orgError) throw orgError;
    orgApps = data;
  }

  // Merge and deduplicate by ID, then sort by created_at and limit
  const allApps = [...(userApps || []), ...(orgApps || [])];
  const uniqueApps = Array.from(
    new Map(allApps.map((app) => [app.id, app])).values()
  );
  return uniqueApps
    .sort(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    )
    .slice(0, limit);
}

export async function getApplicationById(id: string) {
  const { data, error } = await supabase
    .from("applications")
    .select("id, name, description, user_id, org_id, created_at, updated_at")
    .eq("id", id)
    .single();

  if (error) throw error;
  return data;
}

export async function deleteApplication(id: string) {
  const { error } = await supabase.from("applications").delete().eq("id", id);

  if (error) throw error;
}
