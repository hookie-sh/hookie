import { SupabaseClient } from '@supabase/supabase-js'
import { CreateSubscriptionInput } from '../../types'

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
