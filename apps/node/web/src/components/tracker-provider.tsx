"use client";

import { useEffect, useRef } from "react";
import { usePathname, useSearchParams } from "next/navigation";
import { init, identify, page, track, flush } from "@micro-dp/sdk-tracker";

const TRACKER_ENABLED =
  process.env.NEXT_PUBLIC_TRACKER_ENABLED !== "false";
const TRACKER_DEBUG =
  process.env.NEXT_PUBLIC_TRACKER_DEBUG === "true";
const TRACKER_ENDPOINT =
  process.env.NEXT_PUBLIC_TRACKER_ENDPOINT ?? "/api/events";

export function TrackerProvider({
  tenantId,
  userId,
  children,
}: {
  tenantId: string;
  userId: string;
  children: React.ReactNode;
}) {
  const initialized = useRef(false);
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const prevPathname = useRef<string | null>(null);

  // Initialize tracker once
  useEffect(() => {
    if (initialized.current) return;
    initialized.current = true;

    init({
      endpoint: TRACKER_ENDPOINT,
      tenantId,
      userId,
      enabled: TRACKER_ENABLED,
      debug: TRACKER_DEBUG,
    });

    identify(userId);
  }, [tenantId, userId]);

  // Track page views on pathname change
  useEffect(() => {
    if (!initialized.current) return;
    if (prevPathname.current === pathname) return;
    prevPathname.current = pathname;

    page("page_view", { path: pathname });
  }, [pathname]);

  // Detect login_success query param
  useEffect(() => {
    if (!initialized.current) return;
    const event = searchParams.get("event");
    if (event === "login_success") {
      track("login_success");
    }
  }, [searchParams]);

  return <>{children}</>;
}

export { track, flush };
