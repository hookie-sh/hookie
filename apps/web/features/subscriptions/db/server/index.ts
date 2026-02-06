import { SupabaseClient } from "@supabase/supabase-js";
import { CreateSubscriptionInput, Subscription } from "../../types";

export async function createSubscription(
  supabase: SupabaseClient,
  data: CreateSubscriptionInput,
) {
  const { data: subscription, error } = await supabase
    .from("subscriptions")
    .insert(data)
    .select()
    .single();

  if (error) {
    throw error;
  }

  return subscription;
}

export async function getSubscriptionByOrgId(
  supabase: SupabaseClient,
  orgId: string | null,
): Promise<Subscription | null> {
  if (!orgId) {
    return null;
  }

  const { data: subscription, error } = await supabase
    .from("subscriptions")
    .select("*")
    .eq("org_id", orgId)
    .eq("subscribed", true)
    .maybeSingle();

  if (error) {
    throw error;
  }

  return subscription;
}
