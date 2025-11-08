'use server'

import { createCheckoutSession } from '@/services/stripe.server'
import { redirect } from 'next/navigation'

export async function createCheckoutSessionAction(formData: FormData) {
  const priceId = formData.get('priceId') as string | undefined

  if (!priceId) {
    throw new Error('Price ID is required')
  }

  const session = await createCheckoutSession(priceId)

  redirect(session.url || '/')
}
