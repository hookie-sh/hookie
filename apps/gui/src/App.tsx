import { useState, useEffect, useCallback } from "react";
import { Card, CardContent } from "@hookie/ui/components/card";
import { Badge } from "@hookie/ui/components/badge";
import { Button } from "@hookie/ui/components/button";
import { Input } from "@hookie/ui/components/input";
import { LogoWordmark } from "@hookie/ui/components/logo-wordmark";
import {
  Inbox,
  Search,
  X,
  Clock,
  ChevronRight,
  Heading,
  FileJson,
  Filter,
} from "lucide-react";
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

const METHOD_STYLES: Record<
  string,
  { variant: "default" | "secondary" | "outline" | "destructive"; className?: string }
> = {
  GET: { variant: "secondary", className: "font-semibold" },
  POST: { variant: "default", className: "font-semibold" },
  PUT: { variant: "secondary", className: "font-semibold" },
  PATCH: { variant: "secondary", className: "font-semibold" },
  DELETE: { variant: "destructive", className: "font-semibold" },
};

interface EventDetailDrawerProps {
  event: StoredEvent | null;
  onClose: () => void;
}

type DetailTab = "headers" | "query" | "body";

function EventDetailDrawer({ event, onClose }: EventDetailDrawerProps) {
  const [activeTab, setActiveTab] = useState<DetailTab>("body");
  if (!event) return null;

  const style = METHOD_STYLES[event.method] ?? { variant: "outline" as const };
  const hasHeaders = Object.keys(event.headers ?? {}).length > 0;
  const hasQuery = Object.keys(event.query ?? {}).length > 0;

  const tabs: { id: DetailTab; label: string; icon: React.ReactNode }[] = [
    { id: "body", label: "Body", icon: <FileJson className="h-3.5 w-3.5" /> },
    { id: "headers", label: "Headers", icon: <Heading className="h-3.5 w-3.5" /> },
    { id: "query", label: "Query", icon: <Filter className="h-3.5 w-3.5" /> },
  ].filter((t) => (t.id === "headers" ? hasHeaders : t.id === "query" ? hasQuery : true));

  return (
    <>
      <div
        className="fixed inset-0 bg-black/30 backdrop-blur-sm z-40 animate-in fade-in duration-200"
        onClick={onClose}
        aria-hidden="true"
      />
      <div
        className="fixed right-0 top-0 bottom-0 w-full max-w-xl bg-card border-l border-border shadow-2xl z-50 overflow-hidden flex flex-col animate-in slide-in-from-right duration-200"
        role="dialog"
        aria-labelledby="drawer-title"
      >
        <div className="flex items-center justify-between px-5 py-4 border-b border-border bg-card">
          <div className="flex items-center gap-3 min-w-0 flex-1">
            <Badge variant={style.variant} className={`font-mono px-2.5 py-1 rounded-md shrink-0 ${style.className ?? ""}`}>
              {event.method}
            </Badge>
            <span
              id="drawer-title"
              className="font-mono text-sm truncate text-foreground"
            >
              {event.path}
            </span>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} className="shrink-0">
            <X className="h-4 w-4" />
          </Button>
        </div>

        <div className="px-5 py-3 border-b border-border bg-muted/30">
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <span className="flex items-center gap-1.5">
              <Clock className="h-3.5 w-3.5" />
              {new Date(event.timestamp).toLocaleString()}
            </span>
            {(event.appId || event.topicId) && (
              <span>
                {event.appId && <span className="text-foreground/80">{event.appId}</span>}
                {event.appId && event.topicId && " · "}
                {event.topicId && <span className="text-foreground/80">{event.topicId}</span>}
              </span>
            )}
          </div>
        </div>

        <div className="flex gap-1 px-4 pt-3 border-b border-border bg-muted/20">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
              className={[
                "flex items-center gap-2 px-3 py-2 text-sm font-medium rounded-t-md transition-colors",
                activeTab === tab.id
                  ? "bg-background text-foreground border border-b-0 border-border -mb-px"
                  : "text-muted-foreground hover:text-foreground",
              ].join(" ")}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </div>

        <div className="flex-1 overflow-auto p-5">
          {activeTab === "headers" && (
            <div>
              <JsonViewer data={event.headers} />
            </div>
          )}
          {activeTab === "query" && (
            <div>
              <JsonViewer data={event.query} />
            </div>
          )}
          {activeTab === "body" && (
            <div>
              <JsonViewer data={event.body} />
            </div>
          )}
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
      <Card className="border-destructive/50 bg-destructive/5">
        <CardContent className="pt-6">
          <p className="text-destructive font-medium">Failed to load events</p>
          <p className="text-sm text-muted-foreground mt-1">{error.message}</p>
        </CardContent>
      </Card>
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center gap-3 text-muted-foreground py-12">
        <div className="h-5 w-5 animate-spin rounded-full border-2 border-muted-foreground/30 border-t-muted-foreground" />
        <span>Loading events...</span>
      </div>
    );
  }

  const filteredEvents = filterEvents(events, methodFilter, pathFilter);

  if (events.length === 0) {
    return (
      <Card className="border-dashed border-2">
        <CardContent className="flex flex-col items-center justify-center py-20">
          <div className="rounded-full bg-muted p-5 mb-5">
            <Inbox className="h-12 w-12 text-muted-foreground" />
          </div>
          <h3 className="text-lg font-semibold mb-1">No events yet</h3>
          <p className="text-muted-foreground text-center max-w-sm mb-6">
            Start the CLI with the GUI flag to capture webhook events in real-time.
          </p>
          <div className="rounded-lg bg-muted/50 px-4 py-3 font-mono text-sm">
            <code className="text-foreground">hookie listen --gui</code>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center gap-3">
          <span className="text-sm text-muted-foreground">
            <span className="font-semibold text-foreground">{events.length}</span>{" "}
            event{events.length !== 1 ? "s" : ""}
          </span>
          {live && (
            <span className="flex items-center gap-1.5 text-xs text-green-600 dark:text-green-400">
              <span className="relative flex h-2 w-2">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75" />
                <span className="relative inline-flex h-2 w-2 rounded-full bg-green-500" />
              </span>
              Live
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Method"
              value={methodFilter}
              onChange={(e) => setMethodFilter(e.target.value)}
              className="pl-9 w-28 h-9"
            />
          </div>
          <div className="relative">
            <Filter className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Path"
              value={pathFilter}
              onChange={(e) => setPathFilter(e.target.value)}
              className="pl-9 w-44 h-9"
            />
          </div>
        </div>
      </div>

      <div className="space-y-3">
        {filteredEvents.length === 0 ? (
          <div className="py-20 px-6 text-center text-muted-foreground text-sm rounded-xl bg-card/50 border border-dashed border-border">
            No events match your filters.
          </div>
        ) : (
          filteredEvents.map((event) => {
            const style = METHOD_STYLES[event.method] ?? { variant: "outline" as const };
            const isSelected = selectedEvent?.id === event.id;
            return (
              <button
                key={event.id}
                type="button"
                onClick={() => setSelectedEvent(event)}
                className={[
                  "w-full text-left flex items-center gap-4 px-5 py-4 rounded-xl transition-all duration-150",
                  "bg-card/60 hover:bg-card/80 border border-border/50 hover:border-border",
                  "focus:outline-none focus:ring-2 focus:ring-ring/50 focus:ring-offset-2 focus:ring-offset-background",
                  isSelected && "ring-2 ring-primary/30 border-primary/30 bg-card",
                ].join(" ")}
              >
                <Badge variant={style.variant} className={`font-mono text-xs px-2.5 py-1 rounded-md shrink-0 ${style.className ?? ""}`}>
                  {event.method}
                </Badge>
                <span className="font-mono text-sm truncate flex-1 min-w-0">{event.path}</span>
                <span className="text-muted-foreground text-xs tabular-nums shrink-0">
                  {new Date(event.timestamp).toLocaleTimeString()}
                </span>
                {event.appId && (
                  <span className="text-muted-foreground text-xs truncate max-w-24 shrink-0 hidden sm:block">
                    {event.appId}
                  </span>
                )}
                <ChevronRight className="h-4 w-4 text-muted-foreground shrink-0" />
              </button>
            );
          })
        )}
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

      <header className="relative border-b border-border/50 bg-background/80 backdrop-blur-sm">
        <div className="container mx-auto px-4 lg:px-6">
          <div className="flex h-14 items-center gap-6">
            <LogoWordmark className="h-6 text-foreground" />
            <span className="text-muted-foreground">|</span>
            <span className="text-sm text-muted-foreground">Event Listener</span>
          </div>
        </div>
      </header>

      <main className="relative container mx-auto px-4 lg:px-6 py-10 max-w-5xl">
        <div className="mb-10">
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
