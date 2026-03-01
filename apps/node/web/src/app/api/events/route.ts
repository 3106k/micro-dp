import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

import { backendFetch } from "@/lib/api/server";
import { TOKEN_COOKIE, TENANT_COOKIE } from "@/lib/auth/constants";

const BATCH_LIMIT = 100;

interface TrackerEvent {
  event_id: string;
  tenant_id?: string;
  user_id?: string;
  anonymous_id?: string;
  session_id: string;
  event_name: string;
  properties: Record<string, unknown>;
  event_time: string;
  sent_at: string;
}

export async function POST(request: NextRequest) {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantId = jar.get(TENANT_COOKIE)?.value;

  if (!token || !tenantId) {
    return NextResponse.json({ error: "not authenticated" }, { status: 401 });
  }

  let body: { events?: TrackerEvent[] };
  try {
    body = await request.json();
  } catch {
    return NextResponse.json(
      { error: "invalid request body" },
      { status: 400 }
    );
  }

  const events = body.events;
  if (!Array.isArray(events) || events.length === 0) {
    return NextResponse.json(
      { error: "events array is required" },
      { status: 400 }
    );
  }

  if (events.length > BATCH_LIMIT) {
    return NextResponse.json(
      { error: `batch limit is ${BATCH_LIMIT} events` },
      { status: 400 }
    );
  }

  const results: { event_id: string; status: string; error?: string }[] = [];
  let hasFailure = false;

  for (const event of events) {
    try {
      const res = await backendFetch("/api/v1/events", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "X-Tenant-ID": tenantId,
        },
        body: JSON.stringify({
          event_id: event.event_id,
          event_name: event.event_name,
          properties: event.properties,
          event_time: event.event_time,
        }),
      });

      if (res.ok) {
        results.push({ event_id: event.event_id, status: "accepted" });
      } else {
        hasFailure = true;
        const data = await res.json().catch(() => ({ error: "unknown" }));
        results.push({
          event_id: event.event_id,
          status: "failed",
          error: data.error,
        });
      }
    } catch {
      hasFailure = true;
      results.push({
        event_id: event.event_id,
        status: "failed",
        error: "internal error",
      });
    }
  }

  const status = hasFailure ? 207 : 202;
  return NextResponse.json({ results }, { status });
}
