'use client'

import useSWR from 'swr'
import { Separator } from '@hookie/ui/components/separator'
import { CreateTopicForm } from './create-topic-form'
import { TopicsList } from './topics-list'
import { fetcher } from '@/utils/api'

interface Topic {
  id: string
  name: string
  description?: string
  created_at: string
  updated_at: string
}

interface TopicsSectionProps {
  applicationId: string
}

export function TopicsSection({ applicationId }: TopicsSectionProps) {
  const {
    data: topics,
    error: topicsError,
    isLoading: topicsLoading,
  } = useSWR<Topic[]>(
    applicationId ? `/api/applications/${applicationId}/topics` : null,
    fetcher,
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: true,
      errorRetryCount: 3,
    }
  )

  return (
    <>
      <Separator className="my-8" />

      {/* Topics Section */}
      <div className="mb-6 flex justify-between items-center">
        <div>
          <h3 className="text-2xl font-bold mb-2">Topics</h3>
          <p className="text-muted-foreground">
            Create topics to receive webhooks at unique endpoints
          </p>
        </div>
        <CreateTopicForm applicationId={applicationId} />
      </div>

      <TopicsList
        applicationId={applicationId}
        topics={topics}
        isLoading={topicsLoading}
        error={topicsError}
      />
    </>
  )
}
