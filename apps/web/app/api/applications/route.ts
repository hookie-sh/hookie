import { createApplicationSchema } from "@/data/apps/validation";
import { supabase } from "@/services/supabase.server";
import { auth } from "@clerk/nextjs/server";
import { NextRequest, NextResponse } from "next/server";

export async function GET(req: NextRequest) {
  try {
    const { userId } = await auth();

    if (!userId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    // RLS policies will automatically filter applications based on user/org context
    const { data, error } = await supabase
      .from("applications")
      .select("id, name, description, created_at, updated_at")
      .order("created_at", { ascending: false });

    if (error) {
      console.error("Error fetching applications:", error);
      return NextResponse.json(
        { error: "Failed to fetch applications" },
        { status: 500 }
      );
    }

    // Count topics for each application
    const applicationsWithCount = await Promise.all(
      (data || []).map(async (app) => {
        const { count, error: countError } = await supabase
          .from("topics")
          .select("*", { count: "exact", head: true })
          .eq("application_id", app.id);

        if (countError) {
          console.error(
            `Error counting topics for application ${app.id}:`,
            countError
          );
        }

        return {
          ...app,
          topicCount: count || 0,
        };
      })
    );

    return NextResponse.json(applicationsWithCount);
  } catch (error) {
    console.error("Error in GET /api/applications:", error);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 }
    );
  }
}

export async function POST(req: NextRequest) {
  try {
    const { userId, orgId } = await auth();

    if (!userId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const body = await req.json();
    const validatedData = createApplicationSchema.parse(body);

    // Determine owner: organization if orgId exists, otherwise user
    // RLS policies will verify the user has permission to create with these values
    const applicationData: {
      name: string;
      description?: string;
      user_id?: string;
      org_id?: string;
    } = {
      name: validatedData.name,
      description: validatedData.description ?? undefined,
    };

    if (orgId) {
      applicationData.org_id = orgId;
    } else {
      applicationData.user_id = userId;
    }

    const { data, error } = await supabase
      .from("applications")
      .insert(applicationData)
      .select("id, name, description, created_at, updated_at")
      .single();

    if (error) {
      console.error("Error creating application:", error);
      return NextResponse.json(
        { error: "Failed to create application" },
        { status: 500 }
      );
    }

    // Count topics for the new application (will be 0 initially)
    const { count } = await supabase
      .from("topics")
      .select("*", { count: "exact", head: true })
      .eq("application_id", data.id);

    return NextResponse.json(
      {
        ...data,
        topicCount: count || 0,
      },
      { status: 201 }
    );
  } catch (error) {
    if (error instanceof Error && error.name === "ZodError") {
      return NextResponse.json(
        { error: "Invalid input", details: error.message },
        { status: 400 }
      );
    }
    console.error("Error in POST /api/applications:", error);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 }
    );
  }
}
