import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import { Button } from "@hookie/ui/components/button";
import { Copy, Trash2 } from "lucide-react";

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
  const handleCopy = () => {
    navigator.clipboard.writeText(webhookUrl);
    onCopy?.();
  };

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
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={handleCopy}
              >
                <Copy className="h-4 w-4" />
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
