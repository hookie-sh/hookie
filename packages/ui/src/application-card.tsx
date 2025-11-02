import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./components/card";
import { Badge } from "./components/badge";

export interface ApplicationCardProps {
  id: string;
  name: string;
  description?: string;
  topicCount?: number;
  href?: string;
}

export function ApplicationCard({
  id,
  name,
  description,
  topicCount = 0,
  href,
}: ApplicationCardProps) {
  const CardContent_ = (
    <Card className="hover:shadow-md transition-shadow cursor-pointer">
      <CardHeader>
        <CardTitle>{name}</CardTitle>
        {description && <CardDescription>{description}</CardDescription>}
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-2">
          <Badge variant="secondary">{topicCount} topics</Badge>
        </div>
      </CardContent>
    </Card>
  );

  if (href) {
    return (
      <a href={href} className="block">
        {CardContent_}
      </a>
    );
  }

  return CardContent_;
}
