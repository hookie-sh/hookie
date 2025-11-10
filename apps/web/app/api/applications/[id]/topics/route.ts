import { NextRequest, NextResponse } from 'next/server'
import { auth } from '@clerk/nextjs/server'
import { createTopicSchema } from '@/features/topics/schemas/topic'
import {
  getTopicsByApplicationId,
  createTopicForApplication,
} from '@/features/topics/db/server'

interface RouteContext {
  params: Promise<{ id: string }>
}

export async function GET(req: NextRequest, context: RouteContext) {
  try {
    const { userId } = await auth()
    const { id: applicationId } = await context.params

    if (!userId) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    // RLS policies will automatically verify user has access to the application and its topics
    const topics = await getTopicsByApplicationId(applicationId)

    return NextResponse.json(topics)
  } catch (error) {
    console.error('Error in GET /api/applications/[id]/topics:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}

export async function POST(req: NextRequest, context: RouteContext) {
  try {
    const { userId } = await auth()
    const { id: applicationId } = await context.params

    if (!userId) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const body = await req.json()
    const validatedData = createTopicSchema.parse(body)

    // RLS policies will automatically verify user has access to the application
    const topic = await createTopicForApplication(applicationId, validatedData)

    return NextResponse.json(topic, { status: 201 })
  } catch (error) {
    if (error instanceof Error && error.name === 'ZodError') {
      return NextResponse.json(
        { error: 'Invalid input', details: error.message },
        { status: 400 }
      )
    }
    console.error('Error in POST /api/applications/[id]/topics:', error)
    // RLS might return a permission error - treat as not found for security
    return NextResponse.json(
      { error: 'Application not found' },
      { status: 404 }
    )
  }
}
