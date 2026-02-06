import Stripe from "stripe";
import { EnhancedProduct, ProductMetadata } from "../types";

export const productsMetadata: ProductMetadata[] = [
  {
    name: "Free",
    displayName: "Free",
    badge: {
      label: "Individual",
      variant: "secondary",
    },
    price: {
      display: "$0",
      webhookLimit: "10k webhooks/month",
    },
    features: [
      { text: "Single organization" },
      { text: "7-day webhook history" },
      { text: "Reliable delivery with retries" },
    ],
    cta: {
      label: "Get Started",
      variant: "outline",
    },
  },
  {
    name: "Pro",
    displayName: "Pro",
    badge: {
      label: "Popular",
      variant: "default",
    },
    price: {
      display: "$29",
      monthly: "per month",
      webhookLimit: "1M webhooks included",
    },
    previousPlanName: "Free",
    features: [
      { text: "Unlimited applications" },
      { text: "Unlimited team members" },
      { text: "30-day webhook history" },
      { text: "Real-time analytics & insights" },
    ],
    cta: {
      label: "Get Started",
      variant: "default",
    },
    highlight: true,
  },
  {
    name: "Scale",
    displayName: "Scale",
    badge: {
      label: "Mid-market",
      variant: "secondary",
    },
    price: {
      display: "$99",
      monthly: "per month",
      webhookLimit: "10M webhooks included",
    },
    previousPlanName: "Pro",
    features: [
      { text: "Priority support & SLA" },
      { text: "Webhook signing & verification" },
      { text: "Custom webhook transformations" },
      { text: "90-day retention available" },
    ],
    cta: {
      label: "Get Started",
      variant: "outline",
    },
  },
  {
    name: "Enterprise",
    displayName: "Enterprise",
    badge: {
      label: "Custom",
      variant: "secondary",
    },
    price: {
      display: "Custom",
      webhookLimit: "Custom webhook limits",
      monthly: "Volume pricing",
    },
    previousPlanName: "Scale",
    features: [
      { text: "Dedicated account manager" },
      { text: "99.9% uptime SLA" },
      { text: "SSO & advanced security" },
      { text: "Complete audit trail" },
      { text: "Custom contracts & terms" },
    ],
    cta: {
      label: "Contact Sales",
      variant: "outline",
    },
  },
];

/**
 * Matches a Stripe product name to a plan name (case-insensitive, flexible matching)
 */
function matchProductName(stripeName: string, productName: string): boolean {
  const normalizedStripe = stripeName.toLowerCase().trim();
  const normalizedProduct = productName.toLowerCase().trim();

  // Exact match
  if (normalizedStripe === normalizedProduct) return true;

  // Contains match (e.g., "Pro Plan" matches "Pro")
  if (
    normalizedStripe.includes(normalizedProduct) ||
    normalizedProduct.includes(normalizedStripe)
  ) {
    return true;
  }

  return false;
}

/**
 * Enhances Stripe products with plan metadata by matching product names
 */
export function enhanceStripeProducts(
  stripeProducts: Stripe.Product[],
): EnhancedProduct[] {
  const enhanced: EnhancedProduct[] = [];

  for (const stripeProduct of stripeProducts) {
    // Find matching metadata by product name
    const metadata = productsMetadata.find((meta) =>
      matchProductName(stripeProduct.name, meta.name),
    );

    if (!metadata) {
      continue;
    }

    // Extract price ID - handle both string and Price object
    const priceId =
      typeof stripeProduct.default_price === "string"
        ? stripeProduct.default_price
        : stripeProduct.default_price?.id;

    enhanced.push({
      ...metadata,
      stripeProductId: stripeProduct.id,
      stripePriceId: priceId,
    });
  }

  // Sort products to match the order in productsMetadata: Free, Pro, Scale, Enterprise
  return enhanced.sort((a, b) => {
    const orderA = productsMetadata.findIndex((meta) => meta.name === a.name);
    const orderB = productsMetadata.findIndex((meta) => meta.name === b.name);
    return orderA - orderB;
  });
}
