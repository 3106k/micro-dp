"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { track, flush } from "@/components/tracker-provider";

export function DashboardHeader({
  displayName,
  email,
  platformRole,
}: {
  displayName: string;
  email: string;
  platformRole: "user" | "superadmin";
}) {
  const router = useRouter();
  const pathname = usePathname();

  const navItems: Array<{ href: string; label: string }> = [
    { href: "/dashboard", label: "Dashboard" },
    { href: "/jobs", label: "Jobs" },
    { href: "/job-runs", label: "Job Runs" },
    { href: "/datasets", label: "Datasets" },
    { href: "/connections", label: "Connections" },
  ];
  if (platformRole === "superadmin") {
    navItems.push({ href: "/admin/tenants", label: "Admin Tenants" });
  }

  async function handleSignOut() {
    track("sign_out");
    flush({ useBeacon: true });
    await fetch("/api/auth/logout", { method: "POST" });
    router.push("/signin");
    router.refresh();
  }

  return (
    <header className="border-b">
      <div className="container flex min-h-14 flex-wrap items-center justify-between gap-3 py-2">
        <div className="flex items-center gap-6">
          <span className="font-semibold">micro-dp</span>
          <nav className="flex items-center gap-2 text-sm">
            {navItems.map((item) => {
              const active = pathname === item.href;
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={
                    active
                      ? "rounded-md bg-secondary px-2 py-1 font-medium"
                      : "rounded-md px-2 py-1 text-muted-foreground hover:bg-muted hover:text-foreground"
                  }
                >
                  {item.label}
                </Link>
              );
            })}
          </nav>
        </div>
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
