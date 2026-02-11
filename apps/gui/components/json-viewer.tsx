export function JsonViewer({ data }: { data: unknown }) {
  const str =
    data === null || data === undefined
      ? "null"
      : typeof data === "string"
        ? data
        : JSON.stringify(data, null, 2);
  return (
    <pre className="font-mono text-xs overflow-auto max-h-64 rounded bg-muted p-3">
      {str}
    </pre>
  );
}
