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

function EventCard({ event }: { event: StoredEvent }) {
  const [expanded, setExpanded] = useState(false);
  return (
    <Card>
      <CardHeader
        className="cursor-pointer select-none"
        onClick={() => setExpanded(!expanded)}
      >
        <CardTitle className="flex flex-wrap items-center gap-2 text-base">
          <Badge variant="secondary">{event.method}</Badge>
          <span className="font-mono break-all">{event.path}</span>
          <span className="text-muted-foreground text-sm font-normal">
            {new Date(event.timestamp).toLocaleString()}
          </span>
        </CardTitle>
      </CardHeader>
      {expanded && (
        <CardContent className="space-y-4 pt-0">
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
      <p className="text-destructive">Failed to load events: {error.message}</p>
    );
  }

  if (isLoading) {
    return <p className="text-muted-foreground">Loading...</p>;
  }

  const events = data?.events ?? [];

  if (events.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground space-y-4">
        <p>No events received yet.</p>
        <p className="text-sm">
          Forward events to:{" "}
          <a
            href={FORWARD_URL}
            className="text-primary underline hover:no-underline"
          >
            {FORWARD_URL}
          </a>
        </p>
        <p className="text-sm">
          Run <code className="bg-muted px-1 rounded">hookie listen --forward{" "}
          {FORWARD_URL}</code>
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        Forward events to:{" "}
        <a
          href={FORWARD_URL}
          className="text-primary underline hover:no-underline"
        >
          {FORWARD_URL}
        </a>
      </p>
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
    <main className="min-h-screen p-8 max-w-4xl mx-auto">
      <h1 className="text-2xl font-semibold mb-6">Event Listener</h1>
      <EventList />
    </main>
  );
}
