import { SupabaseClient } from '@supabase/supabase-js'
import { CreateSubscriptionInput } from '../../types'
import { stripe } from '@/clients/stripe.server'

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

export async function createCheckoutSession(priceId: string) {
  const session = await stripe.checkout.sessions.create({
    payment_method_types: ['card'],
    line_items: [{ price: priceId, quantity: 1 }],
    mode: 'subscription',
    success_url: `${process.env.NEXT_PUBLIC_APP_URL}/dashboard`,
    cancel_url: `${process.env.NEXT_PUBLIC_APP_URL}`,
  })

  return session
}
