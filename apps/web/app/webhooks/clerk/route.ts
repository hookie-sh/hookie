import {
  organizationCreatedOrUpdated,
  organizationDeleted,
} from "@/app/webhooks/clerk/organizations";
import {
  membershipCreatedOrUpdated,
  membershipDeleted,
} from "@/app/webhooks/clerk/memberships";
import { userCreatedOrUpdated, userDeleted } from "@/app/webhooks/clerk/user";
import { headers } from "next/headers";
import { NextRequest, NextResponse } from "next/server";
import { Webhook } from "svix";

interface ClerkWebhookEvent {
  type: string;
  data: Record<string, unknown>;
}

export async function POST(req: NextRequest) {
  const WEBHOOK_SECRET = process.env.CLERK_WEBHOOK_SECRET;

  if (!WEBHOOK_SECRET) {
    throw new Error(
      "Please add WEBHOOK_SECRET from Clerk Dashboard to .env or .env.local",
    );
  }

  const headerPayload = await headers();
  const svix_id = headerPayload.get("svix-id");
  const svix_timestamp = headerPayload.get("svix-timestamp");
  const svix_signature = headerPayload.get("svix-signature");

  if (!svix_id || !svix_timestamp || !svix_signature) {
    return NextResponse.json(
      { error: "Error occurred -- no svix headers" },
      { status: 400 },
    );
  }

  const payload = await req.json();
  const body = JSON.stringify(payload);
  const wh = new Webhook(WEBHOOK_SECRET);

  let evt: ClerkWebhookEvent;

  try {
    evt = wh.verify(body, {
      "svix-id": svix_id,
      "svix-timestamp": svix_timestamp,
      "svix-signature": svix_signature,
    }) as ClerkWebhookEvent;
  } catch (err) {
    console.error("Webhook verification failed:", err);
    return NextResponse.json(
      { error: "Error occurred -- webhook verification failed" },
      { status: 400 },
    );
  }

  const { id } = evt.data ?? {};
  const eventType = evt.type;

  switch (eventType) {
    case "user.created":
    case "user.updated":
      return userCreatedOrUpdated(evt.data as unknown as Parameters<typeof userCreatedOrUpdated>[0]);
    case "user.deleted":
      return userDeleted(evt.data as unknown as Parameters<typeof userDeleted>[0]);
    case "organization.created":
    case "organization.updated":
      return organizationCreatedOrUpdated(evt.data as unknown as Parameters<typeof organizationCreatedOrUpdated>[0]);
    case "organization.deleted":
      return organizationDeleted(evt.data as unknown as Parameters<typeof organizationDeleted>[0]);
    case "organizationMembership.created":
    case "organizationMembership.updated":
      return membershipCreatedOrUpdated(evt.data as unknown as Parameters<typeof membershipCreatedOrUpdated>[0]);
    case "organizationMembership.deleted":
      return membershipDeleted(evt.data as unknown as Parameters<typeof membershipDeleted>[0]);
    default:
      console.log("Unknown event type");
      console.log(`Webhook with and ID of ${id} and type of ${eventType}`);
      console.log("Webhook body:", body);
      return NextResponse.json(null, { status: 204 });
  }
}
