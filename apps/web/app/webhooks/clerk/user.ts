import { NextResponse } from "next/server";
import { supabaseServiceClient } from "@/services/supabase.service";
import type { DeletedObjectJSON, UserJSON } from "@clerk/nextjs/server";

export async function userCreatedOrUpdated(
  userData: UserJSON
): Promise<NextResponse> {
  const { id, email_addresses, first_name, last_name, image_url } = userData;
  const email = email_addresses?.[0]?.email_address;

  if (!email) {
    return NextResponse.json(
      { error: "Email address is required" },
      { status: 400 }
    );
  }

  // Check if user exists
  const { data: existingUser } = await supabaseServiceClient
    .from("users")
    .select("id")
    .eq("id", id)
    .single();

  if (existingUser) {
    // Update existing user
    const { error } = await supabaseServiceClient
      .from("users")
      .update({
        email,
        first_name,
        last_name,
        image_url,
      })
      .eq("id", id);

    if (error) {
      console.error("Error updating user:", error);
      return NextResponse.json(
        { error: "Failed to update user" },
        { status: 500 }
      );
    }
  } else {
    // Insert new user
    const { error } = await supabaseServiceClient.from("users").insert({
      id,
      email,
      first_name,
      last_name,
      image_url,
    });

    if (error) {
      console.error("Error creating user:", error);
      return NextResponse.json(
        { error: "Failed to create user" },
        { status: 500 }
      );
    }
  }

  return NextResponse.json({ success: true });
}

export async function userDeleted(
  userData: DeletedObjectJSON
): Promise<NextResponse> {
  const { id } = userData;

  if (!id) {
    return NextResponse.json(
      { error: "User ID is required" },
      { status: 400 }
    );
  }

  const { error } = await supabaseServiceClient.from("users").delete().eq("id", id);

  if (error) {
    console.error("Error deleting user:", error);
    return NextResponse.json(
      { error: "Failed to delete user" },
      { status: 500 }
    );
  }

  return NextResponse.json({ success: true });
}

