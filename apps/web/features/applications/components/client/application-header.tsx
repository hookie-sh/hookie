"use client";

import { Button } from "@hookie/ui/components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@hookie/ui/components/tooltip";
import { ArrowLeft, Check, Copy } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

interface ApplicationHeaderProps {
  name: string;
  description?: string;
  applicationId: string;
}

export function ApplicationHeader({
  name,
  description,
  applicationId,
}: ApplicationHeaderProps) {
  const listenCommand = `hookie listen --app-id ${applicationId}`;
  const listenCommandWithForward = `hookie listen --app-id ${applicationId} --forward http://localhost:3001/webhooks`;
  const [copied, setCopied] = useState(false);
  const [showForwardExample, setShowForwardExample] = useState(true);

  const handleCopy = () => {
    const commandToCopy = showForwardExample ? listenCommandWithForward : listenCommand;
    navigator.clipboard.writeText(commandToCopy);
    setCopied(true);
  };

  useEffect(() => {
    if (copied) {
      const timer = setTimeout(() => {
        setCopied(false);
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, [copied]);

  return (
    <>
      {/* Back Button */}
      <Link href="/applications">
        <Button variant="ghost" className="mb-6">
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Applications
        </Button>
      </Link>

      {/* Application Overview */}
      <div className="mb-8">
        <h2 className="text-3xl font-bold mb-2">{name}</h2>
        {description && <p className="text-muted-foreground">{description}</p>}
      </div>

      {/* Listen Command */}
      <Card className="mb-8">
        <CardHeader>
          <CardTitle>Listen to Application</CardTitle>
          <CardDescription>
            Use this command to listen to all webhook events for this
            application. Optionally forward events to a local endpoint using the{" "}
            <code className="text-xs">--forward</code> flag.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <code className="flex-1 px-3 py-2 bg-muted rounded-md text-sm font-mono">
                {showForwardExample ? listenCommandWithForward : listenCommand}
              </code>
              <TooltipProvider>
                <Tooltip open={copied}>
                  <TooltipTrigger asChild>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      onClick={handleCopy}
                    >
                      {copied ? (
                        <Check className="h-4 w-4 text-green-600" />
                      ) : (
                        <Copy className="h-4 w-4" />
                      )}
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Copied to clipboard</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setShowForwardExample(!showForwardExample)}
              className="text-xs"
            >
              {showForwardExample
                ? "Show basic command"
                : "Show with --forward flag example"}
            </Button>
          </div>
        </CardContent>
      </Card>
    </>
  );
}
