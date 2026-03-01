"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { FormError } from "@/components/ui/form-error";

type SessionResponse = { url: string };
type ErrorResponse = { error: string };

export function BillingActions() {
  const [priceId, setPriceId] = useState("");
  const [error, setError] = useState("");
  const [loadingCheckout, setLoadingCheckout] = useState(false);
  const [loadingPortal, setLoadingPortal] = useState(false);

  async function startCheckout() {
    setError("");
    setLoadingCheckout(true);
    try {
      const res = await fetch("/api/billing/checkout-session", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ price_id: priceId }),
      });
      const data = (await res.json()) as SessionResponse | ErrorResponse;
      if (!res.ok || !("url" in data)) {
        throw new Error(("error" in data && data.error) || "checkout failed");
      }
      window.location.href = data.url;
    } catch (e) {
      setError(e instanceof Error ? e.message : "checkout failed");
    } finally {
      setLoadingCheckout(false);
    }
  }

  async function openPortal() {
    setError("");
    setLoadingPortal(true);
    try {
      const res = await fetch("/api/billing/portal-session", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({}),
      });
      const data = (await res.json()) as SessionResponse | ErrorResponse;
      if (!res.ok || !("url" in data)) {
        throw new Error(("error" in data && data.error) || "portal failed");
      }
      window.location.href = data.url;
    } catch (e) {
      setError(e instanceof Error ? e.message : "portal failed");
    } finally {
      setLoadingPortal(false);
    }
  }

  return (
    <div className="space-y-4 rounded-lg border bg-card p-4">
      <FormError message={error} />
      <div className="space-y-2">
        <Label htmlFor="price-id">Stripe Price ID</Label>
        <Input
          id="price-id"
          value={priceId}
          onChange={(e) => setPriceId(e.target.value)}
          placeholder="price_..."
        />
      </div>
      <div className="flex flex-wrap gap-3">
        <Button onClick={startCheckout} disabled={loadingCheckout || !priceId}>
          {loadingCheckout ? "Redirecting..." : "Start Checkout"}
        </Button>
        <Button variant="outline" onClick={openPortal} disabled={loadingPortal}>
          {loadingPortal ? "Opening..." : "Open Billing Portal"}
        </Button>
      </div>
    </div>
  );
}
