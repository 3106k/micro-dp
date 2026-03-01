import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import type { components } from "@/lib/api/generated";

type ErrorResponse = components["schemas"]["ErrorResponse"];

export async function GET(
  _request: NextRequest,
  context: { params: Promise<{ id: string; versionId: string }> }
) {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantId = jar.get(TENANT_COOKIE)?.value;
  if (!token || !tenantId) {
    return NextResponse.json(
      { error: "not authenticated" } satisfies ErrorResponse,
      { status: 401 }
    );
  }

  const { id, versionId } = await context.params;
  const res = await backendFetch(`/api/v1/jobs/${id}/versions/${versionId}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
