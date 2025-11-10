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
