export interface ProductFeature {
  text: string;
}

export interface ProductMetadata {
  name: string;
  displayName: string;
  badge?: {
    label: string;
    variant: "default" | "secondary";
  };
  price: {
    display: string;
    monthly?: string;
    webhookLimit: string;
  };
  previousPlanName?: string;
  features: ProductFeature[];
  cta: {
    label: string;
    variant?: "default" | "outline";
  };
  highlight?: boolean;
}

export interface EnhancedProduct extends ProductMetadata {
  stripeProductId?: string;
  stripePriceId?: string;
  stripePrice?: number;
}
