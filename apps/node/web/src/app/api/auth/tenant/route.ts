import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

import { readApiErrorMessage } from "@/lib/api/error";
import { backendFetch } from "@/lib/api/server";
import {
  TOKEN_COOKIE,
  TENANT_COOKIE,
  cookieOptions,
} from "@/lib/auth/constants";
import type { components } from "@/lib/api/generated";

type MeResponse = components["schemas"]["MeResponse"];
type ErrorResponse = components["schemas"]["ErrorResponse"];

export async function POST(request: NextRequest) {
  const body = (await request.json()) as { tenant_id?: string };
  const tenantId = body.tenant_id;
  if (!tenantId) {
    return NextResponse.json(
      { error: "tenant_id is required" } satisfies ErrorResponse,
      { status: 400 }
    );
  }

  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  if (!token) {
    return NextResponse.json(
      { error: "not authenticated" } satisfies ErrorResponse,
      { status: 401 }
    );
  }

  const meRes = await backendFetch("/api/v1/auth/me", {
    headers: { Authorization: `Bearer ${token}` },
  });

  if (!meRes.ok) {
    return NextResponse.json(
      { error: await readApiErrorMessage(meRes, "failed to fetch user") },
      { status: meRes.status }
    );
  }

  const meData: MeResponse = await meRes.json();
  const allowed = meData.tenants.some((tenant) => tenant.id === tenantId);
  if (!allowed) {
    return NextResponse.json(
      { error: "tenant is not available for this user" } satisfies ErrorResponse,
      { status: 403 }
    );
  }

  jar.set(TENANT_COOKIE, tenantId, cookieOptions);
  return NextResponse.json({ ok: true, tenant_id: tenantId });
}
