import { NextResponse } from "next/server";
import { supabaseServiceClient } from "@/clients/supabase.service";
import type { DeletedObjectJSON } from "@clerk/nextjs/server";

interface OrganizationJSON {
  id: string;
  name: string;
  image_url?: string | null;
  imageUrl?: string | null;
  created_at?: number;
  updated_at?: number;
}

export async function organizationCreatedOrUpdated(
  data: OrganizationJSON,
): Promise<NextResponse> {
  const { id, name, image_url, imageUrl } = data;
  const image = image_url ?? imageUrl ?? null;

  if (!id || !name) {
    return NextResponse.json(
      { error: "Organization id and name are required" },
      { status: 400 },
    );
  }

  const { data: existing } = await supabaseServiceClient
    .from("organizations")
    .select("id")
    .eq("id", id)
    .single();

  if (existing) {
    const { error } = await supabaseServiceClient
      .from("organizations")
      .update({ name, image_url: image })
      .eq("id", id);

    if (error) {
      console.error("Error updating organization:", error);
      return NextResponse.json(
        { error: "Failed to update organization" },
        { status: 500 },
      );
    }
  } else {
    const { error } = await supabaseServiceClient.from("organizations").insert({
      id,
      name,
      image_url: image,
    });

    if (error) {
      console.error("Error creating organization:", error);
      return NextResponse.json(
        { error: "Failed to create organization" },
        { status: 500 },
      );
    }
  }

  return NextResponse.json({ success: true });
}

export async function organizationDeleted(
  data: DeletedObjectJSON,
): Promise<NextResponse> {
  const { id } = data;

  if (!id) {
    return NextResponse.json(
      { error: "Organization id is required" },
      { status: 400 },
    );
  }

  const { error } = await supabaseServiceClient
    .from("organizations")
    .delete()
    .eq("id", id);

  if (error) {
    console.error("Error deleting organization:", error);
    return NextResponse.json(
      { error: "Failed to delete organization" },
      { status: 500 },
    );
  }

  return NextResponse.json({ success: true });
}
