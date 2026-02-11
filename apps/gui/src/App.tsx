import { useState, useEffect, useCallback } from "react";
import {
  Card,
  CardContent,
} from "@hookie/ui/components/card";
import { Badge } from "@hookie/ui/components/badge";
import { Button } from "@hookie/ui/components/button";
import { Input } from "@hookie/ui/components/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@hookie/ui/components/table";
import { LogoWordmark } from "@hookie/ui/components/logo-wordmark";
import { Inbox, Search, X } from "lucide-react";
import { JsonViewer } from "@/components/json-viewer";

export interface StoredEvent {
  id: string;
  method: string;
  path: string;
  query: Record<string, string>;
  headers: Record<string, string>;
  body: unknown;
  timestamp: string;
  appId?: string;
  topicId?: string;
}

const METHOD_COLORS: Record<
  string,
  "default" | "secondary" | "outline" | "destructive"
> = {
  GET: "secondary",
  POST: "default",
  PUT: "secondary",
  PATCH: "secondary",
  DELETE: "destructive",
};

interface EventDetailDrawerProps {
  event: StoredEvent | null;
  onClose: () => void;
}

function EventDetailDrawer({ event, onClose }: EventDetailDrawerProps) {
  if (!event) return null;

  const badgeVariant = METHOD_COLORS[event.method] ?? "outline";

  return (
    <>
      <div
        className="fixed inset-0 bg-black/20 z-40"
        onClick={onClose}
        aria-hidden="true"
      />
      <div className="fixed right-0 top-0 bottom-0 w-full max-w-xl bg-background border-l border-border shadow-xl z-50 overflow-hidden flex flex-col">
        <div className="flex items-center justify-between p-4 border-b border-border">
          <div className="flex items-center gap-2 min-w-0">
            <Badge variant={badgeVariant} className="font-mono shrink-0">
              {event.method}
            </Badge>
            <span className="font-mono text-sm truncate">{event.path}</span>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>
        <div className="flex-1 overflow-auto p-4 space-y-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-1">
              Timestamp
            </p>
            <p className="text-sm">{new Date(event.timestamp).toLocaleString()}</p>
          </div>
          {(event.appId || event.topicId) && (
            <div>
              <p className="text-sm font-medium text-muted-foreground mb-1">
                Context
              </p>
              <p className="text-sm">
                {event.appId && <span>App: {event.appId}</span>}
                {event.appId && event.topicId && " · "}
                {event.topicId && <span>Topic: {event.topicId}</span>}
              </p>
            </div>
          )}
          {Object.keys(event.headers ?? {}).length > 0 && (
            <div>
              <p className="text-sm font-medium text-muted-foreground mb-1">
                Headers
              </p>
              <JsonViewer data={event.headers} />
            </div>
          )}
          {Object.keys(event.query ?? {}).length > 0 && (
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
        </div>
      </div>
    </>
  );
}

function filterEvents(
  events: StoredEvent[],
  methodFilter: string,
  pathFilter: string,
): StoredEvent[] {
  return events.filter((e) => {
    if (methodFilter) {
      const m = e.method.toUpperCase();
      const f = methodFilter.toUpperCase();
      if (!m.includes(f)) return false;
    }
    if (pathFilter && !e.path.toLowerCase().includes(pathFilter.toLowerCase())) {
      return false;
    }
    return true;
  });
}

function EventList() {
  const [events, setEvents] = useState<StoredEvent[]>([]);
  const [error, setError] = useState<Error | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedEvent, setSelectedEvent] = useState<StoredEvent | null>(null);
  const [methodFilter, setMethodFilter] = useState("");
  const [pathFilter, setPathFilter] = useState("");

  const fetchEvents = useCallback(() => {
    fetch("/api/events")
      .then((res) => res.json())
      .then((json) => {
        setEvents(json.events ?? []);
        setError(null);
      })
      .catch((err) => {
        setError(err);
        setEvents([]);
      })
      .finally(() => setIsLoading(false));
  }, []);

  useEffect(() => {
    fetchEvents();
  }, [fetchEvents]);

  useEffect(() => {
    const es = new EventSource("/api/stream");
    es.addEventListener("event", (e) => {
      try {
        const event = JSON.parse(e.data) as StoredEvent;
        setEvents((prev) => [event, ...prev]);
      } catch {
        // ignore parse errors
      }
    });
    return () => es.close();
  }, []);

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

  const filteredEvents = filterEvents(events, methodFilter, pathFilter);

  if (events.length === 0) {
    return (
      <Card className="border-dashed">
        <CardContent className="flex flex-col items-center justify-center py-16">
          <div className="rounded-full bg-muted p-4 mb-4">
            <Inbox className="h-10 w-10 text-muted-foreground" />
          </div>
          <h3 className="text-lg font-medium mb-1">No events yet</h3>
          <p className="text-muted-foreground text-center max-w-sm mb-6">
            Run{" "}
            <code className="bg-muted px-2 py-1 rounded text-foreground">
              hookie listen --gui
            </code>{" "}
            to capture webhook events.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4 flex-wrap">
        <p className="text-sm text-muted-foreground">
          <span className="font-medium text-foreground">{events.length}</span>{" "}
          event{events.length !== 1 ? "s" : ""} received
        </p>
        <div className="flex items-center gap-2 flex-1 min-w-0">
          <Search className="h-4 w-4 text-muted-foreground shrink-0" />
          <Input
            placeholder="Filter by method (e.g. POST)"
            value={methodFilter}
            onChange={(e) => setMethodFilter(e.target.value)}
            className="max-w-40"
          />
          <Input
            placeholder="Filter by path"
            value={pathFilter}
            onChange={(e) => setPathFilter(e.target.value)}
            className="max-w-48"
          />
        </div>
      </div>
      <div className="rounded-md border border-border overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-20">Method</TableHead>
              <TableHead>Path</TableHead>
              <TableHead className="w-44">Time</TableHead>
              <TableHead className="w-28">App</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredEvents.map((event) => {
              const variant = METHOD_COLORS[event.method] ?? "outline";
              return (
                <TableRow
                  key={event.id}
                  className="cursor-pointer"
                  onClick={() => setSelectedEvent(event)}
                >
                  <TableCell>
                    <Badge variant={variant} className="font-mono">
                      {event.method}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-mono text-sm truncate max-w-md">
                    {event.path}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">
                    {new Date(event.timestamp).toLocaleString()}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm truncate max-w-24">
                    {event.appId || "—"}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>
      <EventDetailDrawer
        event={selectedEvent}
        onClose={() => setSelectedEvent(null)}
      />
    </div>
  );
}

export function App() {
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

      <main className="relative container mx-auto px-4 py-8 max-w-5xl">
        <div className="mb-8">
          <h1 className="text-2xl font-semibold mb-2">Inspect Webhook Events</h1>
          <p className="text-muted-foreground">
            Receive events forwarded from the CLI and inspect their payloads in
            real-time.
          </p>
        </div>
        <EventList />
      </main>
    </div>
  );
}
