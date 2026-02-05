import { getApplicationsWithTopicCountByUserId } from "@/features/applications/db/server";

export interface DashboardStats {
  totalApplications: number;
  totalTopics: number;
  webhooksToday: number;
}

export async function getDashboardStats(
  userId: string
): Promise<DashboardStats> {
  try {
    const applications = await getApplicationsWithTopicCountByUserId(userId);

    const totalApplications = applications?.length || 0;

    // Sum up topic counts from all applications
    // Supabase returns topics(count) as an array: topics: [{ count: number }]
    const totalTopics =
      applications?.reduce((sum, app) => {
        let topicCount = 0;
        if (app.topics && Array.isArray(app.topics) && app.topics.length > 0) {
          // Extract count from the first element of the topics array
          topicCount = (app.topics[0] as any)?.count || 0;
        }
        return sum + topicCount;
      }, 0) || 0;

    // Webhooks today: placeholder - webhooks are stored in Redis, not DB
    // This requires future implementation (Redis query or event storage table)
    const webhooksToday = 0;

    return {
      totalApplications,
      totalTopics,
      webhooksToday,
    };
  } catch (error) {
    console.error("Error fetching dashboard stats:", error);
    // Return zeros on error to prevent page crash
    return {
      totalApplications: 0,
      totalTopics: 0,
      webhooksToday: 0,
    };
  }
}
