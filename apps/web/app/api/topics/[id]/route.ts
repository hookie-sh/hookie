import { NextRequest, NextResponse } from "next/server";
import { auth } from "@clerk/nextjs/server";
import { deleteTopic } from "@/features/topics/db/server";

interface RouteContext {
  params: Promise<{ id: string }>;
}

export async function DELETE(req: NextRequest, context: RouteContext) {
  try {
    const { userId } = await auth();
    const { id: topicId } = await context.params;

    if (!userId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // RLS policies will automatically verify user has access to the topic and parent application
    await deleteTopic(topicId);

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error("Error in DELETE /api/topics/[id]:", error);
    // RLS might return a permission error - treat as not found for security
    return NextResponse.json({ error: "Topic not found" }, { status: 404 });
  }
}
