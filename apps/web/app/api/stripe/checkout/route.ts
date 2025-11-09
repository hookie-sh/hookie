import { createCheckoutSession } from '@/clients/stripe.server'
import { checkoutSessionSchema } from '@/data/stripe/validation'
import { NextRequest, NextResponse } from 'next/server'
import { ZodError } from 'zod'
import { createSubscription } from '@/data/db/subscriptions'
import { supabaseServiceClient } from '@/clients/supabase.service'
import { auth } from '@clerk/nextjs/server'

export async function POST(req: NextRequest) {
  try {
    const { userId, orgId } = await auth()

    if (!userId) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const body = await req.json()

    // Validate input
    const validatedData = checkoutSessionSchema.parse(body)
    const { priceId } = validatedData

    if (!priceId) {
      return NextResponse.json(
        { error: 'Price ID is required' },
        { status: 400 }
      )
    }

    // Create checkout session
    const session = await createCheckoutSession(priceId)

    if (!session.url) {
      return NextResponse.json(
        { error: 'Failed to create checkout session URL' },
        { status: 500 }
      )
    }

    // console.log(session)

    // if (!session.customer || !session.subscription) {
    //   return NextResponse.json(
    //     { error: 'Failed to create subscription' },
    //     { status: 500 }
    //   )
    // }

    // await createSubscription(supabaseServiceClient, {
    //   user_id: userId,
    //   org_id: orgId as string,
    //   stripe_customer_id: session.customer as string,
    //   stripe_subscription_id: session.subscription as string,
    //   subscribed: true,
    // })

    return NextResponse.json({ url: session.url })
  } catch (error) {
    // Handle Zod validation errors
    if (error instanceof ZodError) {
      return NextResponse.json(
        { error: 'Invalid input', details: error.issues },
        { status: 400 }
      )
    }

    // Handle Stripe errors
    if (error instanceof Error) {
      console.error('Error creating checkout session:', error)
      return NextResponse.json(
        { error: error.message || 'Failed to create checkout session' },
        { status: 500 }
      )
    }

    // Handle unknown errors
    console.error('Unknown error creating checkout session:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
