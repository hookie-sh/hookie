import { z } from "zod";

export const checkoutSessionSchema = z.object({
  priceId: z.string().min(1, "Price ID is required"),
  returnUrl: z.string().optional(),
});

export type CheckoutSessionInput = z.infer<typeof checkoutSessionSchema>;
