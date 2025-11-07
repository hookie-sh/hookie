import { createClerkClient } from "@clerk/backend";
import { auth } from "@clerk/nextjs/server";
import { NextRequest, NextResponse } from "next/server";

export async function POST(req: NextRequest) {
  try {
    const { userId } = await auth();

    if (!userId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const body = await req.json();
    const { redirect_url } = body;

    if (!redirect_url) {
      return NextResponse.json(
        { error: "Missing redirect_url parameter" },
        { status: 400 }
      );
    }

    // Security: Validate redirect URL is localhost only
    let redirectUrl: URL;
    try {
      redirectUrl = new URL(redirect_url);
    } catch {
      return NextResponse.json(
        { error: "Invalid redirect_url format" },
        { status: 400 }
      );
    }

    if (
      redirectUrl.hostname !== "localhost" &&
      redirectUrl.hostname !== "127.0.0.1"
    ) {
      return NextResponse.json(
        { error: "Invalid redirect_url. Only localhost is allowed." },
        { status: 400 }
      );
    }

    // Generate sign-in token using Clerk backend SDK
    const clerkClient = createClerkClient({
      secretKey: process.env.CLERK_SECRET_KEY!,
    });

    const signInTokenResponse =
      await clerkClient.signInTokens.createSignInToken({
        userId,
        expiresInSeconds: 60, // Token expires in 60 seconds (short-lived for security)
      });

    if (!signInTokenResponse || !signInTokenResponse.token) {
      return NextResponse.json(
        { error: "Failed to create sign-in token" },
        { status: 500 }
      );
    }

    // Return the token to the frontend so it can redirect
    return NextResponse.json({
      token: signInTokenResponse.token,
      redirect_url: redirect_url,
    });
  } catch (error) {
    console.error("Error in POST /api/cli:", error);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 }
    );
  }
}
