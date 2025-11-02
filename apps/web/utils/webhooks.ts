export function generateWebhookUrl(
  applicationId: string,
  topicId: string
): string {
  const url = process.env.NEXT_PUBLIC_INGEST_BASE_URL;
  return `${url}/webhooks/${applicationId}/${topicId}`;
}
