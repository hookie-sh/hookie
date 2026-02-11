export function JsonViewer({ data }: { data: unknown }) {
  const str =
    data === null || data === undefined
      ? "null"
      : typeof data === "string"
        ? data
        : JSON.stringify(data, null, 2);
  return (
    <pre className="font-mono text-xs overflow-auto max-h-64 rounded-md border border-border bg-muted/50 p-4">
      {str}
    </pre>
  );
}
