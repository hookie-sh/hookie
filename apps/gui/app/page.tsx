"use client";

import { useState } from "react";
import useSWR from "swr";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import { Badge } from "@hookie/ui/components/badge";
import { LogoWordmark } from "@hookie/ui/components/logo-wordmark";
import { ChevronDown, ChevronRight, Inbox } from "lucide-react";
import { JsonViewer } from "@/components/json-viewer";

const FORWARD_URL = "http://localhost:4840/api/events";

interface StoredEvent {
  id: string;
  method: string;
  path: string;
  url: string;
  query: Record<string, string>;
  headers: Record<string, string>;
  body: unknown;
  timestamp: string;
}

const fetcher = (url: string) => fetch(url).then((res) => res.json());

const METHOD_COLORS: Record<string, "default" | "secondary" | "outline" | "destructive"> = {
  GET: "secondary",
  POST: "default",
  PUT: "secondary",
  PATCH: "secondary",
  DELETE: "destructive",
};

function EventCard({ event }: { event: StoredEvent }) {
  const [expanded, setExpanded] = useState(false);
  const badgeVariant = METHOD_COLORS[event.method] ?? "outline";

  return (
    <Card className="transition-shadow hover:shadow-md">
      <CardHeader
        className="cursor-pointer select-none"
        onClick={() => setExpanded(!expanded)}
      >
        <CardTitle className="flex flex-wrap items-center gap-2 text-base font-medium">
          {expanded ? (
            <ChevronDown className="h-4 w-4 shrink-0 text-muted-foreground" />
          ) : (
            <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
          )}
          <Badge variant={badgeVariant} className="font-mono">
            {event.method}
          </Badge>
          <span className="font-mono break-all text-foreground">{event.path}</span>
          <span className="text-muted-foreground text-sm font-normal ml-auto">
            {new Date(event.timestamp).toLocaleString()}
          </span>
        </CardTitle>
      </CardHeader>
      {expanded && (
        <CardContent className="space-y-4 pt-0 border-t border-border">
          {Object.keys(event.query).length > 0 && (
            <div>
              <p className="text-sm font-medium text-muted-foreground mb-1">
                Query
              </p>
              <JsonViewer data={event.query} />
            </div>
          )}
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-1">
              Body
            </p>
            <JsonViewer data={event.body} />
          </div>
        </CardContent>
      )}
    </Card>
  );
}

function EventList() {
  const { data, error, isLoading } = useSWR<{ events: StoredEvent[] }>(
    "/api/events",
    fetcher,
    { refreshInterval: 2000 },
  );

  if (error) {
    return (
      <Card className="border-destructive/50">
        <CardContent className="pt-6">
          <p className="text-destructive">Failed to load events: {error.message}</p>
        </CardContent>
      </Card>
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-muted-foreground">
        <div className="h-4 w-4 animate-pulse rounded-full bg-muted" />
        <span>Loading events...</span>
      </div>
    );
  }

  const events = data?.events ?? [];

  if (events.length === 0) {
    return (
      <Card className="border-dashed">
        <CardContent className="flex flex-col items-center justify-center py-16">
          <div className="rounded-full bg-muted p-4 mb-4">
            <Inbox className="h-10 w-10 text-muted-foreground" />
          </div>
          <h3 className="text-lg font-medium mb-1">No events yet</h3>
          <p className="text-muted-foreground text-center max-w-sm mb-6">
            Forward webhook events from the Hookie CLI to inspect them here in real-time.
          </p>
          <div className="rounded-lg bg-muted/50 px-4 py-3 font-mono text-sm w-full max-w-lg">
            <p className="text-muted-foreground text-xs mb-1">Forward URL</p>
            <p className="break-all text-foreground">{FORWARD_URL}</p>
          </div>
          <p className="text-muted-foreground text-sm mt-4">
            Run{" "}
            <code className="bg-muted px-2 py-1 rounded text-foreground">
              hookie listen --forward {FORWARD_URL}
            </code>
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          <span className="font-medium text-foreground">{events.length}</span>{" "}
          event{events.length !== 1 ? "s" : ""} received
        </p>
        <p className="text-xs text-muted-foreground font-mono">{FORWARD_URL}</p>
      </div>
      <div className="space-y-3">
        {events.map((event) => (
          <EventCard key={event.id} event={event} />
        ))}
      </div>
    </div>
  );
}

export default function Home() {
  return (
    <div className="min-h-screen">
      <div className="absolute inset-0 bg-gradient-to-br from-background via-background to-muted/20" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,oklch(0.6171_0.1375_39.0427/0.08),transparent_50%)]" />
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#80808006_1px,transparent_1px),linear-gradient(to_bottom,#80808006_1px,transparent_1px)] bg-[size:24px_24px]" />

      <header className="relative border-b border-border bg-background/80 backdrop-blur-sm">
        <div className="container mx-auto px-4 py-4 flex items-center gap-4">
          <LogoWordmark className="h-7 text-foreground" />
          <span className="text-muted-foreground">|</span>
          <span className="text-sm text-muted-foreground">Event Listener</span>
        </div>
      </header>

      <main className="relative container mx-auto px-4 py-8 max-w-4xl">
        <div className="mb-8">
          <h1 className="text-2xl font-semibold mb-2">Inspect Webhook Events</h1>
          <p className="text-muted-foreground">
            Receive events forwarded from the CLI and inspect their payloads in real-time.
          </p>
        </div>
        <EventList />
      </main>
    </div>
  );
}
