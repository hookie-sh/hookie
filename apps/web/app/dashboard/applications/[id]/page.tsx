"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@hookie/ui/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@hookie/ui/components/dialog";
import { Input } from "@hookie/ui/components/input";
import { Label } from "@hookie/ui/components/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@hookie/ui/components/card";
import { Separator } from "@hookie/ui/components/separator";
import { TopicCard } from "@/components/topic-card";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { createTopicSchema, type CreateTopicInput } from "@/data/topics/validation";
import { generateWebhookUrl } from "@/utils/webhooks";
import { ArrowLeft, Webhook, Activity, TrendingUp } from "lucide-react";

export default function ApplicationDetailPage() {
  const params = useParams();
  const router = useRouter();
  const applicationId = params.id as string;
  const [isOpen, setIsOpen] = useState(false);
  const [application, setApplication] = useState<{
    id: string;
    name: string;
    description?: string;
  } | null>(null);
  const [topics, setTopics] = useState<
    Array<{ id: string; name: string; description?: string }>
  >([]);

  useEffect(() => {
    // TODO: Replace with actual API call to fetch application
    setApplication({
      id: applicationId,
      name: "Sample Application",
      description: "A sample application",
    });
    // TODO: Replace with actual API call to fetch topics
    setTopics([]);
  }, [applicationId]);

  const {
    control,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<CreateTopicInput>({
    resolver: zodResolver(createTopicSchema),
    defaultValues: {
      name: "",
      description: "",
    },
  });

  const onSubmit = async (data: CreateTopicInput) => {
    try {
      // TODO: Replace with actual API call
      const newTopic = {
        id: Math.random().toString(36).substring(7),
        name: data.name,
        description: data.description,
      };
      setTopics([...topics, newTopic]);
      setIsOpen(false);
      reset();
    } catch (error) {
      console.error("Failed to create topic:", error);
    }
  };

  const handleDeleteTopic = (topicId: string) => {
    // TODO: Replace with actual API call
    setTopics(topics.filter((t) => t.id !== topicId));
  };

  if (!application) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b">
        <div className="container mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold">Hookie</h1>
          <div className="flex items-center gap-4">
            <Link href="/dashboard">
              <Button variant="ghost">Dashboard</Button>
            </Link>
            <Link href="/dashboard/applications">
              <Button variant="ghost">Applications</Button>
            </Link>
            <Button variant="ghost">Settings</Button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        {/* Back Button */}
        <Link href="/dashboard/applications">
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
              <CardTitle className="text-sm font-medium">Total Topics</CardTitle>
              <Webhook className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{topics.length}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Webhooks Today</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">0</div>
              <p className="text-xs text-muted-foreground">No activity yet</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
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
                    {isSubmitting ? "Creating..." : "Create"}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        {/* Topics List */}
        {topics.length === 0 ? (
          <Card>
            <CardContent className="py-12 text-center">
              <p className="text-muted-foreground mb-4">
                No topics yet. Create your first topic to start receiving webhooks.
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
                  console.log("Copied to clipboard");
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
    </div>
  );
}
