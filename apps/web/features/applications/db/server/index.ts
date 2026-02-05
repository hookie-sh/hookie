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
  // If orgId is provided, only return organization-owned applications
  if (orgId) {
    const { data: orgApps, error: orgError } = await supabase
      .from("applications")
      .select("*, topics(count)")
      .eq("org_id", orgId)
      .is("user_id", null)
      .order("created_at", { ascending: false });

    if (orgError) throw orgError;

    return (orgApps || []).sort(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    );
  }

  // Otherwise, only return user-owned applications
  const { data: userApps, error: userError } = await supabase
    .from("applications")
    .select("*, topics(count)")
    .eq("user_id", userId)
    .is("org_id", null)
    .order("created_at", { ascending: false });

  if (userError) throw userError;

  return (userApps || []).sort(
    (a, b) =>
      new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  );
}

export async function getRecentApplicationsByUserId(
  userId: string,
  orgId?: string | null,
  limit: number = 5
) {
  // If orgId is provided, only return organization-owned applications
  if (orgId) {
    const { data: orgApps, error: orgError } = await supabase
      .from("applications")
      .select("*, topics(count)")
      .eq("org_id", orgId)
      .is("user_id", null)
      .order("created_at", { ascending: false })
      .limit(limit);

    if (orgError) throw orgError;

    return (orgApps || [])
      .sort(
        (a, b) =>
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      )
      .slice(0, limit);
  }

  // Otherwise, only return user-owned applications
  const { data: userApps, error: userError } = await supabase
    .from("applications")
    .select("*, topics(count)")
    .eq("user_id", userId)
    .is("org_id", null)
    .order("created_at", { ascending: false })
    .limit(limit);

  if (userError) throw userError;

  return (userApps || [])
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
