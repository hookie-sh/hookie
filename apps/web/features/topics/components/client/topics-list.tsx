'use client'

import { mutate } from 'swr'
import { Card, CardContent } from '@hookie/ui/components/card'
import { TopicCard } from '../card'
import { generateWebhookUrl } from '@/utils/webhooks'

interface Topic {
  id: string
  name: string
  description?: string
  created_at: string
  updated_at: string
}

interface TopicsListProps {
  applicationId: string
  topics?: Topic[]
  isLoading?: boolean
  error?: Error | null
  onDelete?: (topicId: string) => void
}

export function TopicsList({
  applicationId,
  topics,
  isLoading,
  error,
  onDelete,
}: TopicsListProps) {
  const handleDeleteTopic = async (topicId: string) => {
    try {
      const response = await fetch(`/api/topics/${topicId}`, {
        method: 'DELETE',
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to delete topic')
      }

      // Optimistically update the cache, then revalidate
      mutate(
        `/api/applications/${applicationId}/topics`,
        (current: Topic[] | undefined) => {
          return current ? current.filter((t) => t.id !== topicId) : []
        },
        false
      )

      // Revalidate to confirm with server
      await mutate(`/api/applications/${applicationId}/topics`)

      // Also update the applications list to reflect new topic count
      await mutate('/api/applications')

      onDelete?.(topicId)
    } catch (err) {
      console.error('Failed to delete topic:', err)
      // Revalidate on error to get accurate state
      await mutate(`/api/applications/${applicationId}/topics`)
    }
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <p className="text-muted-foreground">Loading topics...</p>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <div className="mb-4 p-4 bg-destructive/10 text-destructive rounded-md">
        {error instanceof Error ? error.message : 'Failed to load topics'}
      </div>
    )
  }

  if (!topics || topics.length === 0) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <p className="text-muted-foreground mb-4">
            No topics yet. Create your first topic to start receiving webhooks.
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="grid gap-6">
      {topics.map((topic) => (
        <TopicCard
          key={topic.id}
          id={topic.id}
          name={topic.name}
          description={topic.description}
          webhookUrl={generateWebhookUrl(applicationId, topic.id)}
          onCopy={() => {
            // Toast notification could be added here
            console.log('Copied to clipboard')
          }}
          onDelete={() => handleDeleteTopic(topic.id)}
        />
      ))}
    </div>
  )
}
