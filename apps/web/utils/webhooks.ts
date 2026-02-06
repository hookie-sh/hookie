export function generateWebhookUrl(topicId: string): string {
  const url = process.env.NEXT_PUBLIC_INGEST_BASE_URL;
  return `${url}/topics/${topicId}`;
}
