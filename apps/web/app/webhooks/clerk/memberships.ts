import { NextResponse } from "next/server";
import { supabaseServiceClient } from "@/clients/supabase.service";
import type { DeletedObjectJSON } from "@clerk/nextjs/server";

interface OrganizationRef {
  id: string;
}

interface PublicUserData {
  user_id?: string;
  userId?: string;
}

interface OrganizationMembershipJSON {
  id: string;
  organization: OrganizationRef;
  role: string;
  public_user_data?: PublicUserData;
  publicUserData?: PublicUserData;
  created_at?: number;
  updated_at?: number;
}

function getUserId(data: OrganizationMembershipJSON): string | null {
  const pub = data.public_user_data ?? data.publicUserData;
  return (pub?.user_id ?? pub?.userId) ?? null;
}

export async function membershipCreatedOrUpdated(
  data: OrganizationMembershipJSON,
): Promise<NextResponse> {
  const { id, organization, role } = data;
  const userId = getUserId(data);

  if (!id || !organization?.id || !role) {
    return NextResponse.json(
      { error: "Membership id, organization.id and role are required" },
      { status: 400 },
    );
  }

  if (!userId) {
    return NextResponse.json(
      { error: "Membership public_user_data.user_id is required" },
      { status: 400 },
    );
  }

  const { data: existing } = await supabaseServiceClient
    .from("memberships")
    .select("id")
    .eq("id", id)
    .single();

  const row = {
    organization_id: organization.id,
    user_id: userId,
    role,
  };

  if (existing) {
    const { error } = await supabaseServiceClient
      .from("memberships")
      .update(row)
      .eq("id", id);

    if (error) {
      console.error("Error updating membership:", error);
      return NextResponse.json(
        { error: "Failed to update membership" },
        { status: 500 },
      );
    }
  } else {
    const { error } = await supabaseServiceClient.from("memberships").insert({
      id,
      ...row,
    });

    if (error) {
      console.error("Error creating membership:", error);
      return NextResponse.json(
        { error: "Failed to create membership" },
        { status: 500 },
      );
    }
  }

  return NextResponse.json({ success: true });
}

export async function membershipDeleted(
  data: DeletedObjectJSON,
): Promise<NextResponse> {
  const { id } = data;

  if (!id) {
    return NextResponse.json(
      { error: "Membership id is required" },
      { status: 400 },
    );
  }

  const { error } = await supabaseServiceClient
    .from("memberships")
    .delete()
    .eq("id", id);

  if (error) {
    console.error("Error deleting membership:", error);
    return NextResponse.json(
      { error: "Failed to delete membership" },
      { status: 500 },
    );
  }

  return NextResponse.json({ success: true });
}
