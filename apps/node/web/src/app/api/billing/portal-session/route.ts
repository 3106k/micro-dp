import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import type { components } from "@/lib/api/generated";

type ErrorResponse = components["schemas"]["ErrorResponse"];
type CreateBillingPortalSessionRequest =
  components["schemas"]["CreateBillingPortalSessionRequest"];

export async function POST(req: Request) {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantId = jar.get(TENANT_COOKIE)?.value;

  if (!token || !tenantId) {
    return NextResponse.json(
      { error: "not authenticated" } satisfies ErrorResponse,
      { status: 401 }
    );
  }

  let body: CreateBillingPortalSessionRequest = {};
  if (req.headers.get("content-length") && req.headers.get("content-length") !== "0") {
    try {
      body = await req.json();
    } catch {
      return NextResponse.json(
        { error: "invalid json body" } satisfies ErrorResponse,
        { status: 400 }
      );
    }
  }

  const res = await backendFetch("/api/v1/billing/portal-session", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
    body: JSON.stringify(body),
  });

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
