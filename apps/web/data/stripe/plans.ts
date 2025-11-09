export interface PlanFeature {
  text: string
}

export interface PlanMetadata {
  name: string
  displayName: string
  badge?: {
    label: string
    variant: 'default' | 'secondary'
  }
  price: {
    display: string
    monthly?: string
    webhookLimit: string
  }
  previousPlanName?: string
  features: PlanFeature[]
  cta: {
    label: string
    variant?: 'default' | 'outline'
  }
  highlight?: boolean
}

export const plansMetadata: PlanMetadata[] = [
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

export interface EnhancedPlan extends PlanMetadata {
  stripeProductId?: string
  stripePriceId?: string
  stripePrice?: number
}

/**
 * Matches a Stripe product name to a plan name (case-insensitive, flexible matching)
 */
function matchPlanName(stripeName: string, planName: string): boolean {
  const normalizedStripe = stripeName.toLowerCase().trim()
  const normalizedPlan = planName.toLowerCase().trim()

  // Exact match
  if (normalizedStripe === normalizedPlan) return true

  // Contains match (e.g., "Pro Plan" matches "Pro")
  if (
    normalizedStripe.includes(normalizedPlan) ||
    normalizedPlan.includes(normalizedStripe)
  ) {
    return true
  }

  return false
}

/**
 * Enhances Stripe products with plan metadata by matching product names
 */
export function enhancePlansWithStripe(
  stripeProducts: Array<{
    id: string
    name: string
    description?: string | null
    default_price?: string | { id: string } | null
  }>
): EnhancedPlan[] {
  return plansMetadata.map((plan) => {
    // Find matching Stripe product
    const stripeProduct = stripeProducts.find((product) =>
      matchPlanName(product.name, plan.name)
    )

    if (!stripeProduct) {
      return plan
    }

    // Extract price ID - handle both string and Price object
    const priceId =
      typeof stripeProduct.default_price === 'string'
        ? stripeProduct.default_price
        : stripeProduct.default_price?.id

    return {
      ...plan,
      stripeProductId: stripeProduct.id,
      stripePriceId: priceId,
    }
  })
}
