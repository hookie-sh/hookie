import { NextRequest, NextResponse } from "next/server";

const MAX_EVENTS = 100;

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

const events: StoredEvent[] = [];

function headersToObject(headers: Headers): Record<string, string> {
  const obj: Record<string, string> = {};
  headers.forEach((value, key) => {
    obj[key] = value;
  });
  return obj;
}

function searchParamsToObject(searchParams: URLSearchParams): Record<string, string> {
  const obj: Record<string, string> = {};
  searchParams.forEach((value, key) => {
    obj[key] = value;
  });
  return obj;
}

async function storeRequest(req: NextRequest): Promise<StoredEvent> {
  const url = req.nextUrl ?? new URL(req.url);
  let body: unknown;
  try {
    if (req.body) {
      body = await req.json();
    } else {
      body = null;
    }
  } catch {
    const text = await req.text();
    body = text || null;
  }

  const event: StoredEvent = {
    id: crypto.randomUUID(),
    method: req.method,
    path: url.pathname,
    url: url.toString(),
    query: searchParamsToObject(url.searchParams),
    headers: headersToObject(req.headers),
    body,
    timestamp: new Date().toISOString(),
  };

  events.unshift(event);
  if (events.length > MAX_EVENTS) {
    events.pop();
  }

  return event;
}

export async function GET() {
  return NextResponse.json({
    events: [...events],
  });
}

export async function POST(req: NextRequest) {
  await storeRequest(req);
  return NextResponse.json({ ok: true }, { status: 200 });
}

export async function PUT(req: NextRequest) {
  await storeRequest(req);
  return NextResponse.json({ ok: true }, { status: 200 });
}

export async function PATCH(req: NextRequest) {
  await storeRequest(req);
  return NextResponse.json({ ok: true }, { status: 200 });
}

export async function DELETE(req: NextRequest) {
  await storeRequest(req);
  return NextResponse.json({ ok: true }, { status: 200 });
}
