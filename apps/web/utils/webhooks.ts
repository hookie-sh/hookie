export function generateWebhookUrl(applicationId: string, topicId: string, baseUrl: string = "https://hookie.app"): string {
  return `${baseUrl}/webhooks/${applicationId}/${topicId}`;
}

