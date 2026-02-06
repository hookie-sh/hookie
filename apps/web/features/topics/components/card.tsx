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
import { Check, Copy, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";

export interface TopicCardProps {
  id: string;
  name: string;
  description?: string;
  webhookUrl: string;
  onCopy?: () => void;
  onDelete?: () => void;
}

export function TopicCard({
  id,
  name,
  description,
  webhookUrl,
  onCopy,
  onDelete,
}: TopicCardProps) {
  const [webhookCopied, setWebhookCopied] = useState(false);
  const [commandCopied, setCommandCopied] = useState(false);
  const [showForwardExample, setShowForwardExample] = useState(true);

  const listenCommand = `hookie listen --topic-id ${id}`;
  const listenCommandWithForward = `hookie listen --topic-id ${id} --forward ${webhookUrl}`;

  const handleCopyWebhook = () => {
    navigator.clipboard.writeText(webhookUrl);
    setWebhookCopied(true);
    onCopy?.();
  };

  const handleCopyCommand = () => {
    const commandToCopy = showForwardExample ? listenCommandWithForward : listenCommand;
    navigator.clipboard.writeText(commandToCopy);
    setCommandCopied(true);
    onCopy?.();
  };

  useEffect(() => {
    if (webhookCopied) {
      const timer = setTimeout(() => {
        setWebhookCopied(false);
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, [webhookCopied]);

  useEffect(() => {
    if (commandCopied) {
      const timer = setTimeout(() => {
        setCommandCopied(false);
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, [commandCopied]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>{name}</CardTitle>
        {description && <CardDescription>{description}</CardDescription>}
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div>
            <label className="text-sm font-medium text-muted-foreground">
              Webhook URL
            </label>
            <div className="flex items-center gap-2 mt-1">
              <code className="flex-1 px-3 py-2 bg-muted rounded-md text-sm font-mono">
                {webhookUrl}
              </code>
              <TooltipProvider>
                <Tooltip open={webhookCopied}>
                  <TooltipTrigger asChild>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      onClick={handleCopyWebhook}
                    >
                      {webhookCopied ? (
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
          </div>
          <div>
            <label className="text-sm font-medium text-muted-foreground">
              Listen Command
            </label>
            <div className="space-y-2 mt-1">
              <div className="flex items-center gap-2">
                <code className="flex-1 px-3 py-2 bg-muted rounded-md text-sm font-mono">
                  {showForwardExample ? listenCommandWithForward : listenCommand}
                </code>
                <TooltipProvider>
                  <Tooltip open={commandCopied}>
                    <TooltipTrigger asChild>
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        onClick={handleCopyCommand}
                      >
                        {commandCopied ? (
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
                className="text-xs h-7"
              >
                {showForwardExample
                  ? "Show basic command"
                  : "Show with --forward flag"}
              </Button>
            </div>
          </div>
          {onDelete && (
            <Button
              type="button"
              variant="destructive"
              size="sm"
              onClick={onDelete}
            >
              <Trash2 className="h-4 w-4 mr-2" />
              Delete Topic
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
