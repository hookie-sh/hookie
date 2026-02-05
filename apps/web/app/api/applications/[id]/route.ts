import { NextRequest, NextResponse } from "next/server";
import { auth } from "@clerk/nextjs/server";
import {
  getApplicationById,
  deleteApplication,
} from "@/features/applications/db/server";

interface RouteContext {
  params: Promise<{ id: string }>;
}

export async function GET(req: NextRequest, context: RouteContext) {
  try {
    const { userId } = await auth();
    const { id } = await context.params;

    if (!userId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // RLS policies will automatically filter based on user/org context
    const application = await getApplicationById(id);

    if (!application) {
      return NextResponse.json(
        { error: "Application not found" },
        { status: 404 },
      );
    }

    return NextResponse.json(application);
  } catch (error) {
    console.error("Error in GET /api/applications/[id]:", error);
    // RLS might return a permission error - treat as not found for security
    return NextResponse.json(
      { error: "Application not found" },
      { status: 404 },
    );
  }
}

export async function DELETE(req: NextRequest, context: RouteContext) {
  try {
    const { userId } = await auth();
    const { id } = await context.params;

    if (!userId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // RLS policies will automatically verify user has permission to delete
    await deleteApplication(id);

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error("Error in DELETE /api/applications/[id]:", error);
    // RLS might return a permission error - treat as not found for security
    return NextResponse.json(
      { error: "Application not found" },
      { status: 404 },
    );
  }
}
