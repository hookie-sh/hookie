"use client";

import { useParams } from "next/navigation";
import useSWR from "swr";
import Link from "next/link";
import { Button } from "@hookie/ui/components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import { Separator } from "@hookie/ui/components/separator";
import { ApplicationHeader } from "@/features/applications/components/client/application-header";
import { ApplicationStats } from "@/features/applications/components/client/application-stats";
import { TopicsSection } from "@/features/topics/components/client/topics-section";
import { fetcher } from "@/utils/api";

interface Application {
  id: string;
  name: string;
  description?: string;
}

export default function ApplicationDetailPage() {
  const params = useParams();
  const applicationId = params.id as string;

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
    },
  );

  const { data: topics } = useSWR<any[]>(
    applicationId ? `/api/applications/${applicationId}/topics` : null,
    fetcher,
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: true,
      errorRetryCount: 3,
    },
  );

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    );
  }

  if (error || !application) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <p className="text-destructive mb-4">
            {error instanceof Error ? error.message : "Application not found"}
          </p>
          <Link href="/applications">
            <Button variant="outline">Back to Applications</Button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <main className="container mx-auto px-4 py-8">
      <ApplicationHeader
        name={application.name}
        description={application.description}
      />

      <ApplicationStats topicCount={topics?.length || 0} />

      <TopicsSection applicationId={applicationId} />

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
  );
}
