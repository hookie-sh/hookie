import { getDashboardStats } from "@/features/dashboard/db/server";
import { auth } from "@clerk/nextjs/server";
import { Button } from "@hookie/ui/components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import { Folder, TrendingUp, Webhook } from "lucide-react";
import Link from "next/link";

export default async function DashboardPage() {
  const { userId } = await auth();

  if (!userId) {
    return null;
  }

  const stats = await getDashboardStats(userId);

  return (
    <>
      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h2 className="text-3xl font-bold mb-2">Dashboard</h2>
          <p className="text-muted-foreground">
            Welcome back! Here's an overview of your webhook activity.
          </p>
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

        {/* Quick Actions */}
        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>
              Get started with Hookie by creating your first application
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/applications">
              <Button>Create Application</Button>
            </Link>
          </CardContent>
        </Card>

        {/* Recent Activity */}
        <Card className="mt-6">
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>
              Your latest webhook events will appear here
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground text-center py-8">
              No recent activity. Create an application to start receiving
              webhooks.
            </p>
          </CardContent>
        </Card>
      </main>
    </>
  );
}
