import { createApplicationSchema } from '@/features/applications/schemas/application'
import {
  createApplication,
  getApplicationsWithTopicCountByUserId,
} from '@/features/applications/db/server'
import { auth } from '@clerk/nextjs/server'
import { NextRequest, NextResponse } from 'next/server'

export async function GET(_: NextRequest) {
  try {
    const { userId } = await auth()

    if (!userId) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const applicationsWithTopicCount =
      await getApplicationsWithTopicCountByUserId(userId)

    return NextResponse.json(applicationsWithTopicCount)
  } catch (error) {
    console.error('Error in GET /api/applications:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}

export async function POST(req: NextRequest) {
  try {
    const { userId, orgId } = await auth()

    if (!userId) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const body = await req.json()
    const validatedData = createApplicationSchema.parse(body)

    const applicationData: {
      name: string
      description?: string
      user_id?: string
      org_id?: string
    } = {
      name: validatedData.name,
      description: validatedData.description ?? undefined,
    }

    if (orgId) {
      applicationData.org_id = orgId
    } else {
      applicationData.user_id = userId
    }

    const application = await createApplication(applicationData)

    return NextResponse.json(
      {
        application,
      },
      { status: 201 }
    )
  } catch (error) {
    if (error instanceof Error && error.name === 'ZodError') {
      return NextResponse.json(
        { error: 'Invalid input', details: error.message },
        { status: 400 }
      )
    }
    console.error('Error in POST /api/applications:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
