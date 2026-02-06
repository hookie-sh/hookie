import { stripe } from "@/clients/stripe.server";
import { enhanceStripeProducts } from "../../adapters/product";

export async function listProducts() {
  try {
    const { data: products } = await stripe.products.list({
      active: true,
    });

    return enhanceStripeProducts(products);
  } catch (error) {
    console.error("Error listing products:", error);
    throw error;
  }
}

export async function getProductByName(planName: string) {
  try {
    const products = await listProducts();
    const normalizedPlanName = planName.toLowerCase();
    return (
      products.find(
        (product) => product.name.toLowerCase() === normalizedPlanName,
      ) || null
    );
  } catch (error) {
    console.error("Error getting product by name:", error);
    throw error;
  }
}

export async function createCheckoutSession(
  priceId: string,
  cancelUrl?: string,
) {
  const baseUrl = process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000";
  const cancelUrlFinal = cancelUrl
    ? `${baseUrl}${cancelUrl}`
    : `${baseUrl}/paywall`;

  const session = await stripe.checkout.sessions.create({
    payment_method_types: ["card"],
    line_items: [{ price: priceId, quantity: 1 }],
    mode: "subscription",
    success_url: `${baseUrl}/dashboard`,
    cancel_url: cancelUrlFinal,
  });

  return session;
}
