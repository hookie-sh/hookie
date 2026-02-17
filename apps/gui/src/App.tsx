import { useState, useEffect, useCallback } from "react";
import { Card, CardContent } from "@hookie/ui/components/card";
import { Badge } from "@hookie/ui/components/badge";
import { Button } from "@hookie/ui/components/button";
import { Input } from "@hookie/ui/components/input";
import { LogoWordmark } from "@hookie/ui/components/logo-wordmark";
import {
  Inbox,
  Search,
  Clock,
  ChevronDown,
  ChevronRight,
  Filter,
  MessageSquare,
  Trash2,
} from "lucide-react";
import { JsonViewer } from "@/components/json-viewer.js";

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

const METHOD_STYLES: Record<
  string,
  {
    variant: "default" | "secondary" | "outline" | "destructive";
    className?: string;
  }
> = {
  GET: { variant: "secondary", className: "font-semibold" },
  POST: { variant: "default", className: "font-semibold" },
  PUT: { variant: "secondary", className: "font-semibold" },
  PATCH: { variant: "secondary", className: "font-semibold" },
  DELETE: { variant: "destructive", className: "font-semibold" },
};

interface CollapsibleSectionProps {
  label: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
}

function CollapsibleSection({
  label,
  children,
  defaultOpen = false,
}: CollapsibleSectionProps) {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <div className="border-b border-border last:border-b-0">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="w-full flex items-center gap-2 px-4 py-3 text-sm font-medium text-foreground hover:bg-muted/30 transition-colors text-left"
      >
        {open ? (
          <ChevronDown className="h-4 w-4 text-muted-foreground shrink-0" />
        ) : (
          <ChevronRight className="h-4 w-4 text-muted-foreground shrink-0" />
        )}
        {label}
      </button>
      {open && <div className="px-4 pb-4">{children}</div>}
    </div>
  );
}

interface EventDetailPanelProps {
  event: StoredEvent | null;
}

function EventDetailPanel({ event }: EventDetailPanelProps) {
  if (!event) {
    return (
      <div className="flex-1 flex items-center justify-center p-12">
        <div className="text-center text-muted-foreground">
          <MessageSquare className="h-12 w-12 mx-auto mb-4 opacity-50" />
          <p className="font-medium">Select an event to view details</p>
          <p className="text-sm mt-1">Click an event in the sidebar</p>
        </div>
      </div>
    );
  }

  const style = METHOD_STYLES[event.method] ?? { variant: "outline" as const };
  const hasHeaders = Object.keys(event.headers ?? {}).length > 0;
  const hasQuery = Object.keys(event.query ?? {}).length > 0;

  return (
    <div className="flex-1 overflow-auto p-6">
      <div className="mb-6">
        <div className="flex items-center gap-3 mb-2">
          <Badge
            variant={style.variant}
            className={`font-mono px-2.5 py-1 rounded-md ${style.className ?? ""}`}
          >
            {event.method}
          </Badge>
          <span className="font-mono text-sm text-foreground break-all">
            {event.path}
          </span>
        </div>
        <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
          <span className="flex items-center gap-1.5">
            <Clock className="h-3.5 w-3.5" />
            {new Date(event.timestamp).toLocaleString()}
          </span>
          {event.appId && <span>{event.appId}</span>}
          {event.topicId && <span>{event.topicId}</span>}
        </div>
      </div>

      <div className="rounded-lg border border-border bg-card/80 overflow-hidden">
        <CollapsibleSection label="Body" defaultOpen>
          <JsonViewer data={event.body} />
        </CollapsibleSection>
        {hasHeaders && (
          <CollapsibleSection label="Headers">
            <JsonViewer data={event.headers} />
          </CollapsibleSection>
        )}
        {hasQuery && (
          <CollapsibleSection label="Query">
            <JsonViewer data={event.query} />
          </CollapsibleSection>
        )}
      </div>
    </div>
  );
}

function filterEvents(
  events: StoredEvent[],
  methodFilter: string,
  pathFilter: string
): StoredEvent[] {
  return events.filter((e) => {
    if (methodFilter) {
      const m = e.method.toUpperCase();
      const f = methodFilter.toUpperCase();
      if (!m.includes(f)) return false;
    }
    if (
      pathFilter &&
      !e.path.toLowerCase().includes(pathFilter.toLowerCase())
    ) {
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
  const [live, setLive] = useState(false);

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
    es.onopen = () => setLive(true);
    es.onerror = () => setLive(false);
    es.addEventListener("event", (e) => {
      try {
        const event = JSON.parse(e.data) as StoredEvent;
        setEvents((prev) => [event, ...prev]);
      } catch {
        // ignore
      }
    });
    return () => {
      es.close();
      setLive(false);
    };
  }, []);

  if (error) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <Card className="border-destructive/50 bg-destructive/5 max-w-md">
          <CardContent className="pt-6">
            <p className="text-destructive font-medium">
              Failed to load events
            </p>
            <p className="text-sm text-muted-foreground mt-1">
              {error.message}
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="flex items-center gap-3 text-muted-foreground">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-muted-foreground/30 border-t-muted-foreground" />
          <span>Loading events...</span>
        </div>
      </div>
    );
  }

  if (events.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <Card className="border-dashed border-2 max-w-md p-4">
          <CardContent className="flex flex-col items-center justify-center py-16">
            <div className="rounded-full bg-muted p-5 mb-4">
              <Inbox className="h-12 w-12 text-muted-foreground" />
            </div>
            <h3 className="text-lg font-semibold mb-1">No events yet</h3>
            <p className="text-muted-foreground text-center text-sm mb-4">
              Start the CLI to capture webhook events. UI starts by default.
            </p>
            <code className="bg-muted px-3 py-2 rounded text-sm font-mono">
              hookie listen
            </code>
          </CardContent>
        </Card>
      </div>
    );
  }

  const filteredEvents = filterEvents(events, methodFilter, pathFilter);

  return (
    <div className="flex flex-1 min-h-0">
      <aside className="w-80 min-w-[280px] flex flex-col border-r border-border bg-card/50 shrink-0">
        <div className="p-4 border-b border-border">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Events
            </span>
            <div className="flex items-center gap-2">
              {live && (
                <span className="flex items-center gap-1.5 text-xs text-green-600 dark:text-green-400">
                  <span className="relative flex h-1.5 w-1.5">
                    <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75" />
                    <span className="relative h-1.5 w-1.5 rounded-full bg-green-500" />
                  </span>
                  Live
                </span>
              )}
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 text-muted-foreground hover:text-foreground"
                onClick={async () => {
                  await fetch("/api/events/clear", { method: "POST" });
                  fetchEvents();
                  setSelectedEvent(null);
                }}
                title="Clear events"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>
          <div className="flex gap-2">
            <div className="relative flex-1">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
              <Input
                placeholder="Method"
                value={methodFilter}
                onChange={(e) => setMethodFilter(e.target.value)}
                className="pl-8 h-8 text-sm"
              />
            </div>
            <div className="relative flex-1">
              <Filter className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
              <Input
                placeholder="Path"
                value={pathFilter}
                onChange={(e) => setPathFilter(e.target.value)}
                className="pl-8 h-8 text-sm"
              />
            </div>
          </div>
        </div>
        <div className="flex-1 overflow-auto">
          {filteredEvents.length === 0 ? (
            <div className="p-4 text-center text-sm text-muted-foreground">
              No events match your filters.
            </div>
          ) : (
            <div className="py-2">
              {filteredEvents.map((event) => {
                const style = METHOD_STYLES[event.method] ?? {
                  variant: "outline" as const,
                };
                const isSelected = selectedEvent?.id === event.id;
                return (
                  <button
                    key={event.id}
                    type="button"
                    onClick={() => setSelectedEvent(event)}
                    className={[
                      "w-full text-left px-4 py-3 flex flex-col gap-1.5 transition-colors",
                      "border-l-2 -mt-px first:mt-0",
                      isSelected
                        ? "bg-muted/30 border-l-primary"
                        : "border-l-transparent hover:bg-muted/20",
                    ].join(" ")}
                  >
                    <span className="font-mono text-sm text-foreground truncate">
                      {event.path}
                    </span>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <Badge
                        variant={style.variant}
                        className={`font-mono px-2 py-0.5 text-[10px] rounded ${style.className ?? ""}`}
                      >
                        {event.method}
                      </Badge>
                      <span className="tabular-nums">
                        {new Date(event.timestamp).toLocaleTimeString()}
                      </span>
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </div>
      </aside>

      <main className="flex-1 min-w-0 flex flex-col overflow-hidden">
        <EventDetailPanel event={selectedEvent} />
      </main>
    </div>
  );
}

export function App() {
  return (
    <div className="min-h-screen flex flex-col">
      <div className="absolute inset-0 bg-gradient-to-br from-background via-background to-muted/20" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,oklch(0.6171_0.1375_39.0427/0.08),transparent_50%)]" />
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#80808006_1px,transparent_1px),linear-gradient(to_bottom,#80808006_1px,transparent_1px)] bg-[size:24px_24px]" />

      <header className="relative border-b border-border/50 bg-background/80 backdrop-blur-sm">
        <div className="px-4 lg:px-6">
          <div className="flex h-14 items-center gap-6">
            <LogoWordmark className="h-6 text-foreground" />
            <span className="text-muted-foreground">|</span>
            <span className="text-sm text-muted-foreground">
              Event Listener
            </span>
          </div>
        </div>
      </header>

      <div className="relative flex-1 flex flex-col min-h-0">
        <div className="px-4 lg:px-6 py-6 border-b border-border/50">
          <h1 className="text-xl font-semibold">Webhook Events</h1>
          <p className="text-muted-foreground text-sm mt-1">
            Inspect webhook payloads captured from the CLI in real-time.
          </p>
        </div>
        <EventList />
      </div>
    </div>
  );
}
