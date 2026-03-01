import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";

type ErrorResponse = components["schemas"]["ErrorResponse"];

async function authHeaders() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantId = jar.get(TENANT_COOKIE)?.value;
  if (!token || !tenantId) {
    return null;
  }
  return {
    Authorization: `Bearer ${token}`,
    "X-Tenant-ID": tenantId,
  };
}

export async function GET() {
  const headers = await authHeaders();
  if (!headers) {
    return NextResponse.json(
      { error: "not authenticated" } satisfies ErrorResponse,
      { status: 401 }
    );
  }

  const res = await backendFetch("/api/v1/tenants/current/members", {
    headers,
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
