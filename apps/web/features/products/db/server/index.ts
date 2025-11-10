import { stripe } from '@/clients/stripe.server'
import { enhanceStripeProducts } from '../../adapters/product'

export async function listProducts() {
  try {
    const { data: products } = await stripe.products.list({
      active: true,
    })

    const enhancedProducts = enhanceStripeProducts(products)
    return enhancedProducts
  } catch (error) {
    console.error('Error listing products:', error)
    throw error
  }
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
