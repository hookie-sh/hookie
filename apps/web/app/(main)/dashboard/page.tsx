import { ApplicationCard } from "@/features/applications/components/card";
import { CreateApplicationForm } from "@/features/applications/components/client/create-application-form";
import { getRecentApplicationsByUserId } from "@/features/applications/db/server";
import { getDashboardStats } from "@/features/dashboard/db/server";
import { auth } from "@clerk/nextjs/server";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import { Folder, TrendingUp, Webhook } from "lucide-react";

export default async function DashboardPage() {
  const { userId, orgId } = await auth();

  if (!userId) {
    return null;
  }

  const stats = await getDashboardStats(userId, orgId);
  const recentApplications = await getRecentApplicationsByUserId(
    userId,
    orgId,
    5
  );

  // Transform applications to match ApplicationCard props
  const applicationsWithTopicCount =
    recentApplications?.map((app) => {
      let topicCount = 0;
      if (app.topics && Array.isArray(app.topics) && app.topics.length > 0) {
        topicCount = (app.topics[0] as any)?.count || 0;
      }
      return {
        id: app.id,
        name: app.name,
        description: app.description,
        topicCount,
      };
    }) || [];

  return (
    <>
      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        <div className="mb-8 flex justify-between items-start">
          <div>
            <h2 className="text-3xl font-bold mb-2">Dashboard</h2>
            <p className="text-muted-foreground">
              Welcome back! Here's an overview of your webhook activity.
            </p>
          </div>
          <CreateApplicationForm />
        </div>

        {/* Stats Cards */}
        <div className="grid md:grid-cols-3 gap-6 mb-8">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Total Applications
              </CardTitle>
              <Folder className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {stats.totalApplications}
              </div>
              <p className="text-xs text-muted-foreground">
                {stats.totalApplications === 0
                  ? "Create your first application"
                  : `${stats.totalApplications} application${stats.totalApplications === 1 ? "" : "s"}`}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Active Topics
              </CardTitle>
              <Webhook className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalTopics}</div>
              <p className="text-xs text-muted-foreground">
                {stats.totalTopics === 0
                  ? "No active topics yet"
                  : `${stats.totalTopics} topic${stats.totalTopics === 1 ? "" : "s"}`}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Webhooks Today
              </CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.webhooksToday}</div>
              <p className="text-xs text-muted-foreground">
                {stats.webhooksToday === 0
                  ? "No activity today"
                  : "Active today"}
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Recent Applications */}
        <div className="mb-8">
          <div className="mb-4">
            <h3 className="text-xl font-semibold mb-1">Recent Applications</h3>
            <p className="text-sm text-muted-foreground">
              Your most recently created applications
            </p>
          </div>
          {applicationsWithTopicCount.length === 0 ? (
            <div className="border border-border rounded-lg p-8 text-center">
              <p className="text-sm text-muted-foreground">
                No applications yet. Create your first application to get
                started.
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
              {applicationsWithTopicCount.map((application) => (
                <ApplicationCard
                  key={application.id}
                  {...application}
                  href={`/applications/${application.id}`}
                />
              ))}
            </div>
          )}
        </div>

        {/* Recent Activity */}
        <div>
          <div className="mb-4">
            <h3 className="text-xl font-semibold mb-1">Recent Activity</h3>
            <p className="text-sm text-muted-foreground">
              Your latest webhook events will appear here
            </p>
          </div>
          <div className="border border-border rounded-lg p-8 text-center">
            <p className="text-sm text-muted-foreground">
              No recent activity. Create an application to start receiving
              webhooks.
            </p>
          </div>
        </div>
      </main>
    </>
  );
}
