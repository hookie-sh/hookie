import Stripe from 'stripe'
import { EnhancedProduct, ProductMetadata } from '../types'

export const productsMetadata: ProductMetadata[] = [
  {
    name: 'Free',
    displayName: 'Free',
    badge: {
      label: 'Individual',
      variant: 'secondary',
    },
    price: {
      display: '$0',
      webhookLimit: '10k webhooks/month',
    },
    features: [
      { text: 'Personal organization' },
      { text: '7-day retention' },
      { text: 'Basic webhook delivery' },
    ],
    cta: {
      label: 'Get Started',
      variant: 'outline',
    },
  },
  {
    name: 'Pro',
    displayName: 'Pro',
    badge: {
      label: 'Popular',
      variant: 'default',
    },
    price: {
      display: '$29',
      monthly: 'per month',
      webhookLimit: '1M webhooks included',
    },
    previousPlanName: 'Free',
    features: [
      { text: 'Unlimited apps & members' },
      { text: '30-day retention' },
      { text: 'Team collaboration' },
      { text: 'Advanced analytics' },
    ],
    cta: {
      label: 'Get Started',
      variant: 'default',
    },
    highlight: true,
  },
  {
    name: 'Scale',
    displayName: 'Scale',
    badge: {
      label: 'Mid-market',
      variant: 'secondary',
    },
    price: {
      display: '$99',
      monthly: 'per month',
      webhookLimit: '10M webhooks included',
    },
    previousPlanName: 'Pro',
    features: [
      { text: 'Priority support' },
      { text: 'Webhook signing' },
      { text: 'Custom integrations' },
      { text: 'Extended retention options' },
    ],
    cta: {
      label: 'Get Started',
      variant: 'outline',
    },
  },
  {
    name: 'Enterprise',
    displayName: 'Enterprise',
    badge: {
      label: 'Custom',
      variant: 'secondary',
    },
    price: {
      display: 'Custom',
      webhookLimit: 'Custom webhook limits',
      monthly: 'Volume pricing',
    },
    previousPlanName: 'Scale',
    features: [
      { text: 'Dedicated support' },
      { text: 'SLA guarantee' },
      { text: 'SSO integration' },
      { text: 'Audit logs' },
      { text: 'Custom contracts' },
    ],
    cta: {
      label: 'Contact Sales',
      variant: 'outline',
    },
  },
]

/**
 * Matches a Stripe product name to a plan name (case-insensitive, flexible matching)
 */
function matchProductName(stripeName: string, productName: string): boolean {
  const normalizedStripe = stripeName.toLowerCase().trim()
  const normalizedProduct = productName.toLowerCase().trim()

  // Exact match
  if (normalizedStripe === normalizedProduct) return true

  // Contains match (e.g., "Pro Plan" matches "Pro")
  if (
    normalizedStripe.includes(normalizedProduct) ||
    normalizedProduct.includes(normalizedStripe)
  ) {
    return true
  }

  return false
}

/**
 * Enhances Stripe products with plan metadata by matching product names
 */
export function enhanceStripeProducts(
  stripeProducts: Stripe.Product[]
): EnhancedProduct[] {
  const enhanced: EnhancedProduct[] = []

  for (const stripeProduct of stripeProducts) {
    // Find matching metadata by product name
    const metadata = productsMetadata.find((meta) =>
      matchProductName(stripeProduct.name, meta.name)
    )

    if (!metadata) {
      continue
    }

    // Extract price ID - handle both string and Price object
    const priceId =
      typeof stripeProduct.default_price === 'string'
        ? stripeProduct.default_price
        : stripeProduct.default_price?.id

    enhanced.push({
      ...metadata,
      stripeProductId: stripeProduct.id,
      stripePriceId: priceId,
    })
  }

  return enhanced
}
