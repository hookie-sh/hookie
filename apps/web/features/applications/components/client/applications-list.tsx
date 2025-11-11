'use client'

import { useAuth } from '@clerk/nextjs'
import useSWR from 'swr'
import { ApplicationCard } from '../card'
import { fetcher } from '@/utils/api'

interface Application {
  id: string
  name: string
  description?: string
  topicCount: number
}

interface ApplicationsListProps {
  error?: Error | null
}

export function ApplicationsList({
  error: externalError,
}: ApplicationsListProps) {
  const { userId } = useAuth()

  const {
    data: applications,
    error,
    isLoading,
  } = useSWR<Application[]>(userId ? '/api/applications' : null, fetcher, {
    revalidateOnFocus: false,
    revalidateOnReconnect: true,
  })

  const displayError = externalError || error

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Loading applications...</p>
      </div>
    )
  }

  if (displayError) {
    return (
      <div className="mb-4 p-4 bg-destructive/10 text-destructive rounded-md">
        {displayError instanceof Error
          ? displayError.message
          : 'Failed to load applications'}
      </div>
    )
  }

  if (!applications || applications.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground mb-4">
          No applications yet. Create your first application to get started.
        </p>
      </div>
    )
  }

  return (
    <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
      {applications.map((application) => (
        <ApplicationCard
          key={application.id}
          {...application}
          href={`/applications/${application.id}`}
        />
      ))}
    </div>
  )
}
