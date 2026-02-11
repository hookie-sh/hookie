"use client";

import { Button } from "@hookie/ui/components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import { Terminal } from "lucide-react";
import { useSearchParams } from "next/navigation";
import { useState } from "react";

export default function CLIPage() {
  const searchParams = useSearchParams();
  const redirectUrl = searchParams.get("redirect_url");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleAuthorize = async () => {
    if (!redirectUrl) {
      setError("Missing redirect_url parameter");
      return;
    }

    // Validate redirect URL is localhost only (security)
    try {
      const url = new URL(redirectUrl);
      if (url.hostname !== "localhost" && url.hostname !== "127.0.0.1") {
        setError("Invalid redirect URL. Only localhost is allowed.");
        return;
      }
    } catch {
      setError("Invalid redirect URL format");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await fetch("/api/cli", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ redirect_url: redirectUrl }),
      });

      if (!response.ok) {
        const data = await response
          .json()
          .catch(() => ({ error: "Failed to authorize CLI" }));
        throw new Error(data.error || "Failed to authorize CLI");
      }

      // Get the token from the response
      const data = await response.json();
      const { token } = data;

      if (!token) {
        throw new Error("No token received from server");
      }

      // Build the redirect URL with the token
      const finalUrl = new URL(redirectUrl);
      finalUrl.searchParams.set("token", token);

      // Redirect to the localhost callback URL
      window.location.href = finalUrl.toString();
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
      setLoading(false);
    }
  };

  if (!redirectUrl) {
    return (
      <main className="flex flex-col items-center justify-center h-screen mx-auto px-4 py-8">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>CLI Authorization</CardTitle>
            <CardDescription>Missing redirect URL parameter</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              This page should be accessed from the CLI with a redirect_url
              parameter.
            </p>
          </CardContent>
        </Card>
      </main>
    );
  }

  if (success) {
    return (
      <main className="flex flex-col items-center justify-center h-screen mx-auto px-4 py-8">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Terminal className="h-5 w-5" />
              Authorization Successful
            </CardTitle>
            <CardDescription>
              You can close this window and return to your terminal
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              The authentication token has been sent to your CLI. You can close
              this window now.
            </p>
          </CardContent>
        </Card>
      </main>
    );
  }

  return (
    <main className="flex flex-col items-center justify-center h-screen mx-auto px-4 py-8">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Terminal className="h-5 w-5" />
            Authorize CLI
          </CardTitle>
          <CardDescription>
            Click the button below to generate an authentication token for your
            CLI
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && (
            <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}
          <Button
            onClick={handleAuthorize}
            disabled={loading}
            className="w-full"
          >
            {loading ? "Authorizing..." : "Authorize CLI"}
          </Button>
          <p className="text-xs text-muted-foreground">
            This will generate a secure authentication token and send it to your
            local CLI server.
          </p>
        </CardContent>
      </Card>
    </main>
  );
}
