import { createCheckoutSession } from '@/services/stripe.server'
import { NextRequest, NextResponse } from 'next/server'

export async function POST(req: NextRequest) {
  const { priceId } = await req.json()

  console.log('priceId', priceId)

  // const session = await createCheckoutSession(priceId)
  return NextResponse.json({ message: 'Hello, world!' })
}
