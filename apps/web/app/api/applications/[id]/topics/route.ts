import { NextRequest, NextResponse } from "next/server";
import { auth } from "@clerk/nextjs/server";
import { supabase } from "@/services/supabase.server";
import { createTopicSchema } from "@/data/topics/validation";

interface RouteContext {
  params: Promise<{ id: string }>;
}

export async function GET(req: NextRequest, context: RouteContext) {
  try {
    const { userId } = await auth();
    const { id: applicationId } = await context.params;

    if (!userId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // RLS policies will automatically verify user has access to the application and its topics
    const { data, error } = await supabase
      .from("topics")
      .select("id, name, description, created_at, updated_at")
      .eq("application_id", applicationId)
      .order("created_at", { ascending: false });

    if (error) {
      console.error("Error fetching topics:", error);
      return NextResponse.json(
        { error: "Failed to fetch topics" },
        { status: 500 }
      );
    }

    return NextResponse.json(data || []);
  } catch (error) {
    console.error("Error in GET /api/applications/[id]/topics:", error);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 }
    );
  }
}

export async function POST(req: NextRequest, context: RouteContext) {
  try {
    const { userId } = await auth();
    const { id: applicationId } = await context.params;

    if (!userId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const body = await req.json();
    const validatedData = createTopicSchema.parse(body);

    // RLS policies will automatically verify user has access to the application
    const { data, error } = await supabase
      .from("topics")
      .insert({
        name: validatedData.name,
        description: validatedData.description || null,
        application_id: applicationId,
      })
      .select("id, name, description, created_at, updated_at")
      .single();

    if (error) {
      console.error("Error creating topic:", error);
      // RLS might return a permission error - treat as not found for security
      return NextResponse.json(
        { error: "Application not found" },
        { status: 404 }
      );
    }

    return NextResponse.json(data, { status: 201 });
  } catch (error) {
    if (error instanceof Error && error.name === "ZodError") {
      return NextResponse.json(
        { error: "Invalid input", details: error.message },
        { status: 400 }
      );
    }
    console.error("Error in POST /api/applications/[id]/topics:", error);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 }
    );
  }
}

