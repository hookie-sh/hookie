'use client'

import { TopicCard } from '@/features/topics/components/card'
import {
  createTopicSchema,
  type CreateTopicInput,
} from '@/features/topics/schemas/topic'
import { fetcher } from '@/utils/api'
import { generateWebhookUrl } from '@/utils/webhooks'
import { zodResolver } from '@hookform/resolvers/zod'
import { Button } from '@hookie/ui/components/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@hookie/ui/components/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@hookie/ui/components/dialog'
import { Input } from '@hookie/ui/components/input'
import { Label } from '@hookie/ui/components/label'
import { Separator } from '@hookie/ui/components/separator'
import { Activity, ArrowLeft, TrendingUp, Webhook } from 'lucide-react'
import Link from 'next/link'
import { useParams, useRouter } from 'next/navigation'
import { useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import useSWR, { mutate } from 'swr'

interface Application {
  id: string
  name: string
  description?: string
}

interface Topic {
  id: string
  name: string
  description?: string
  created_at: string
  updated_at: string
}

export default function ApplicationDetailPage() {
  const params = useParams()
  const router = useRouter()
  const applicationId = params.id as string
  const [isOpen, setIsOpen] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)

  const {
    data: application,
    error,
    isLoading,
  } = useSWR<Application>(
    applicationId ? `/api/applications/${applicationId}` : null,
    fetcher,
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: true,
      errorRetryCount: 3,
    }
  )

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

  const {
    control,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<CreateTopicInput>({
    resolver: zodResolver(createTopicSchema),
    defaultValues: {
      name: '',
      description: '',
    },
  })

  const onSubmit = async (data: CreateTopicInput) => {
    try {
      setSubmitError(null)
      const response = await fetch(
        `/api/applications/${applicationId}/topics`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(data),
        }
      )

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to create topic')
      }

      const newTopic = await response.json()

      // Optimistically update the cache, then revalidate
      mutate(
        `/api/applications/${applicationId}/topics`,
        (current: Topic[] | undefined) => {
          return current ? [newTopic, ...current] : [newTopic]
        },
        false
      )

      // Revalidate to confirm with server
      await mutate(`/api/applications/${applicationId}/topics`)

      // Also update the applications list to reflect new topic count
      await mutate('/api/applications')

      setIsOpen(false)
      reset()
      setSubmitError(null)
    } catch (err) {
      console.error('Failed to create topic:', err)
      setSubmitError(
        err instanceof Error ? err.message : 'Failed to create topic'
      )
    }
  }

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
    } catch (err) {
      console.error('Failed to delete topic:', err)
      // Revalidate on error to get accurate state
      await mutate(`/api/applications/${applicationId}/topics`)
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (error || !application) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <p className="text-destructive mb-4">
            {error instanceof Error ? error.message : 'Application not found'}
          </p>
          <Link href="/applications">
            <Button variant="outline">Back to Applications</Button>
          </Link>
        </div>
      </div>
    )
  }

  return (
    <>
      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        {/* Back Button */}
        <Link href="/applications">
          <Button variant="ghost" className="mb-6">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Applications
          </Button>
        </Link>

        {/* Application Overview */}
        <div className="mb-8">
          <h2 className="text-3xl font-bold mb-2">{application.name}</h2>
          {application.description && (
            <p className="text-muted-foreground">{application.description}</p>
          )}
        </div>

        {/* Stats Cards */}
        <div className="grid md:grid-cols-3 gap-6 mb-8">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Total Topics
              </CardTitle>
              <Webhook className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{topics?.length || 0}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Webhooks Today
              </CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">0</div>
              <p className="text-xs text-muted-foreground">No activity yet</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Success Rate
              </CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">-</div>
              <p className="text-xs text-muted-foreground">No data available</p>
            </CardContent>
          </Card>
        </div>

        <Separator className="my-8" />

        {/* Topics Section */}
        <div className="mb-6 flex justify-between items-center">
          <div>
            <h3 className="text-2xl font-bold mb-2">Topics</h3>
            <p className="text-muted-foreground">
              Create topics to receive webhooks at unique endpoints
            </p>
          </div>
          <Dialog open={isOpen} onOpenChange={setIsOpen}>
            <DialogTrigger asChild>
              <Button>Create Topic</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create New Topic</DialogTitle>
                <DialogDescription>
                  Create a new topic to receive webhooks at a unique endpoint
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSubmit(onSubmit)}>
                <div className="space-y-4 py-4">
                  <div className="space-y-2">
                    <Label htmlFor="topic-name">Name</Label>
                    <Controller
                      name="name"
                      control={control}
                      render={({ field }) => (
                        <Input
                          id="topic-name"
                          placeholder="payment-success"
                          {...field}
                        />
                      )}
                    />
                    {errors.name && (
                      <p className="text-sm text-destructive">
                        {errors.name.message}
                      </p>
                    )}
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="topic-description">
                      Description (optional)
                    </Label>
                    <Controller
                      name="description"
                      control={control}
                      render={({ field }) => (
                        <Input
                          id="topic-description"
                          placeholder="Receives payment success events"
                          {...field}
                        />
                      )}
                    />
                    {errors.description && (
                      <p className="text-sm text-destructive">
                        {errors.description.message}
                      </p>
                    )}
                  </div>
                </div>
                <DialogFooter>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setIsOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={isSubmitting}>
                    {isSubmitting ? 'Creating...' : 'Create'}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        {/* Error Messages */}
        {(topicsError || submitError) && (
          <div className="mb-4 p-4 bg-destructive/10 text-destructive rounded-md">
            {submitError ||
              (topicsError instanceof Error
                ? topicsError.message
                : 'Failed to load topics')}
          </div>
        )}

        {/* Topics List */}
        {topicsLoading ? (
          <Card>
            <CardContent className="py-12 text-center">
              <p className="text-muted-foreground">Loading topics...</p>
            </CardContent>
          </Card>
        ) : !topics || topics.length === 0 ? (
          <Card>
            <CardContent className="py-12 text-center">
              <p className="text-muted-foreground mb-4">
                No topics yet. Create your first topic to start receiving
                webhooks.
              </p>
            </CardContent>
          </Card>
        ) : (
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
        )}

        {/* Activity Section (Placeholder) */}
        <Separator className="my-8" />
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>
              Webhook events and activity will appear here
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground text-center py-8">
              No activity yet. Once you start receiving webhooks, events will be
              displayed here.
            </p>
          </CardContent>
        </Card>
      </main>
    </>
  )
}
