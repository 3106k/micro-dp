import type { components } from "@/lib/api/generated";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

type EventsSummaryResponse = components["schemas"]["EventsSummaryResponse"];

export function EventSummary({
  summary,
}: {
  summary: EventsSummaryResponse | null;
}) {
  if (!summary || summary.total === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Events</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            No events tracked yet.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-3">
      <h2 className="text-lg font-semibold tracking-tight">Events</h2>
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{summary.total}</p>
          </CardContent>
        </Card>
        {summary.counts.map((item) => (
          <Card key={item.event_name}>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {item.event_name}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold">{item.count}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
