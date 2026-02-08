import { createSupabaseServerClient } from "@/clients/supabase.server";
import { CreateTopicInput } from "../../schemas/topic";

export async function getTopicCountByApplicationId(applicationId: string) {
  const supabase = createSupabaseServerClient();
  const { count, error } = await supabase
    .from("topics")
    .select("*", { count: "exact", head: true })
    .eq("application_id", applicationId);

  if (error) throw error;
  return count;
}

export async function createTopic(input: CreateTopicInput) {
  const supabase = createSupabaseServerClient();
  const { data, error } = await supabase
    .from("topics")
    .insert(input)
    .select("id, name, description, created_at, updated_at")
    .single();

  if (error) throw error;
  return data;
}

export async function getTopicsByApplicationId(applicationId: string) {
  const supabase = createSupabaseServerClient();
  const { data, error } = await supabase
    .from("topics")
    .select("id, name, description, created_at, updated_at")
    .eq("application_id", applicationId)
    .order("created_at", { ascending: false });

  if (error) throw error;
  return data || [];
}

export async function createTopicForApplication(
  applicationId: string,
  input: CreateTopicInput
) {
  const supabase = createSupabaseServerClient();
  const { data, error } = await supabase
    .from("topics")
    .insert({
      name: input.name,
      description: input.description || null,
      application_id: applicationId,
    })
    .select("id, name, description, created_at, updated_at")
    .single();

  if (error) throw error;
  return data;
}

export async function deleteTopic(id: string) {
  const supabase = createSupabaseServerClient();
  const { error } = await supabase.from("topics").delete().eq("id", id);

  if (error) throw error;
}
