import { cookies } from "next/headers";

import { backendFetch } from "@/lib/api/server";
import { TOKEN_COOKIE } from "@/lib/auth/constants";
import type { components } from "@/lib/api/generated";
import { EventSummary } from "@/components/event-summary";

type EventsSummaryResponse = components["schemas"]["EventsSummaryResponse"];

export default async function AdminAnalyticsPage() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value!;

  let eventsSummary: EventsSummaryResponse | null = null;
  const summaryRes = await backendFetch("/api/v1/tracker/summary", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (summaryRes.ok) {
    eventsSummary = await summaryRes.json();
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Analytics</h1>
      <EventSummary summary={eventsSummary} />
    </div>
  );
}
