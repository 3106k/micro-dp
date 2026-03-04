"use client";

import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import type { components } from "@/lib/api/generated";

type Credential = components["schemas"]["Credential"];

export function IntegrationsManager({
  initialCredentials,
}: {
  initialCredentials: Credential[];
}) {
  const [credentials, setCredentials] = useState(initialCredentials);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);

  const searchParams = useSearchParams();

  useEffect(() => {
    if (searchParams.get("linked") === "true") {
      setMessage("Google account linked successfully.");
      refreshCredentials();
    }
    if (searchParams.get("error")) {
      setMessage(`Error: ${searchParams.get("error")}`);
    }
  }, [searchParams]);

  async function refreshCredentials() {
    const res = await fetch("/api/credentials", { cache: "no-store" });
    if (res.ok) {
      const data: { items: Credential[] } = await res.json();
      setCredentials(data.items ?? []);
    }
  }

  async function handleDisconnect(id: string) {
    if (!confirm("Remove this integration?")) return;
    setLoading(true);
    setMessage("");
    try {
      const res = await fetch(`/api/credentials/${id}`, { method: "DELETE" });
      if (res.status === 204 || res.ok) {
        setMessage("Integration removed.");
        await refreshCredentials();
      } else {
        setMessage("Failed to remove integration.");
      }
    } finally {
      setLoading(false);
    }
  }

  async function handleConnect() {
    setLoading(true);
    setMessage("");
    try {
      const res = await fetch("/api/credentials/google/start");
      const data = (await res.json()) as { url?: string; error?: string };
      if (!res.ok || !data.url) {
        setMessage(`Error: ${data.error || "failed to start google oauth"}`);
        return;
      }
      window.location.href = data.url;
    } catch {
      setMessage("Error: network error");
    } finally {
      setLoading(false);
    }
  }

  const googleCredentials = credentials.filter((c) => c.provider === "google");

  return (
    <div className="space-y-6">
      {message && (
        <p className="text-sm text-muted-foreground">{message}</p>
      )}

      <section className="space-y-4">
        <h2 className="text-lg font-medium">Google</h2>
        <p className="text-sm text-muted-foreground">
          Connect your Google account to access Google Sheets as a data source.
        </p>

        {googleCredentials.length > 0 ? (
          <div className="space-y-2">
            {googleCredentials.map((cred) => (
              <div
                key={cred.id}
                className="flex items-center justify-between rounded-lg border p-4"
              >
                <div>
                  <p className="font-medium">
                    {cred.provider_label || "Google Account"}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Scopes: {cred.scopes || "N/A"}
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleDisconnect(cred.id)}
                  disabled={loading}
                >
                  Disconnect
                </Button>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            No Google account connected.
          </p>
        )}

        <Button onClick={handleConnect} disabled={loading}>
          {googleCredentials.length > 0
            ? "Reconnect Google Account"
            : "Connect Google Account"}
        </Button>
      </section>
    </div>
  );
}
