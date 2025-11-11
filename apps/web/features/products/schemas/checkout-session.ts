import { z } from 'zod'

export const checkoutSessionSchema = z.object({
  priceId: z.string().min(1, 'Price ID is required'),
})

export type CheckoutSessionInput = z.infer<typeof checkoutSessionSchema>
