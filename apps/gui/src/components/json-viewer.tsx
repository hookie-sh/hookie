import { useMemo } from "react";

interface JsonViewerProps {
  data: unknown;
  className?: string;
}

function formatJson(data: unknown): string {
  if (data === null || data === undefined) return "null";
  if (typeof data === "string") return JSON.stringify(data);
  return JSON.stringify(data, null, 2);
}

export function JsonViewer({ data, className = "" }: JsonViewerProps) {
  const str = useMemo(() => formatJson(data), [data]);

  return (
    <pre
      className={[
        "font-mono text-[13px] overflow-auto max-h-72 rounded-lg border border-border bg-muted/20 p-4 leading-relaxed",
        "text-foreground/90 selection:bg-primary/20",
        className,
      ].join(" ")}
    >
      <code>{str}</code>
    </pre>
  );
}
