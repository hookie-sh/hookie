import { SupabaseClient } from '@supabase/supabase-js'

export interface Subscription {
  id: string
  user_id: string
  org_id: string
  stripe_customer_id: string
  stripe_subscription_id: string
  subscribed: boolean
  created_at: string
  updated_at: string
}

export interface CreateSubscriptionInput {
  user_id: string
  org_id: string
  stripe_customer_id: string
  stripe_subscription_id: string
  subscribed: boolean
}

export async function createSubscription(
  supabase: SupabaseClient,
  data: CreateSubscriptionInput
) {
  const { data: subscription, error } = await supabase
    .from('subscriptions')
    .insert(data)
    .select()
    .single()

  if (error) {
    throw error
  }

  return subscription
}
