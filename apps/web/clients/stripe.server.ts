// FAKE CARDS FOR TESTING PURPOSES => https://docs.stripe.com/testing#cards

import Stripe from 'stripe'

if (!process.env.STRIPE_SECRET_KEY) {
  throw new Error('STRIPE_SECRET_KEY is not set')
}

if (!process.env.NEXT_PUBLIC_APP_URL) {
  throw new Error('NEXT_PUBLIC_APP_URL is not set')
}

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY)

export async function getStripeProducts() {
  const products = await stripe.products.list()

  return products.data
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
