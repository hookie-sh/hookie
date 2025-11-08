import Stripe from 'stripe'

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!)

export async function getStripeProducts() {
  const products = await stripe.products.list()

  return products.data
}
