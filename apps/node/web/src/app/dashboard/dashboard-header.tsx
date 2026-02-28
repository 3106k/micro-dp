"use client";

import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";

export function DashboardHeader({
  displayName,
  email,
}: {
  displayName: string;
  email: string;
}) {
  const router = useRouter();

  async function handleSignOut() {
    await fetch("/api/auth/logout", { method: "POST" });
    router.push("/signin");
    router.refresh();
  }

  return (
    <header className="border-b">
      <div className="container flex h-14 items-center justify-between">
        <span className="font-semibold">micro-dp</span>
        <div className="flex items-center gap-4">
          <span className="text-sm text-muted-foreground">
            {displayName || email}
          </span>
          <Button variant="outline" size="sm" onClick={handleSignOut}>
            Sign out
          </Button>
        </div>
      </div>
    </header>
  );
}
