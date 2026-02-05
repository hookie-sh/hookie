import { getConnectedClientsByUserId } from "@/features/connected-clients/db/server";
import { auth } from "@clerk/nextjs/server";
import { NextRequest, NextResponse } from "next/server";

export async function GET(_: NextRequest) {
  try {
    const { userId, orgId } = await auth();

    if (!userId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // Pass orgId (can be null for personal, or string for org)
    const clients = await getConnectedClientsByUserId(userId, orgId || null);

    return NextResponse.json(clients);
  } catch (error) {
    console.error("Error in GET /api/connected-clients:", error);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 }
    );
  }
}
