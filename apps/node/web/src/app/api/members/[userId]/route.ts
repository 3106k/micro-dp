import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

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

export async function PATCH(
  request: NextRequest,
  context: { params: Promise<{ userId: string }> }
) {
  const headers = await authHeaders();
  if (!headers) {
    return NextResponse.json(
      { error: "not authenticated" } satisfies ErrorResponse,
      { status: 401 }
    );
  }

  const { userId } = await context.params;
  const body = await request.json();
  const res = await backendFetch(
    `/api/v1/tenants/current/members/${userId}`,
    {
      method: "PATCH",
      headers,
      body: JSON.stringify(body),
    }
  );
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function DELETE(
  _request: NextRequest,
  context: { params: Promise<{ userId: string }> }
) {
  const headers = await authHeaders();
  if (!headers) {
    return NextResponse.json(
      { error: "not authenticated" } satisfies ErrorResponse,
      { status: 401 }
    );
  }

  const { userId } = await context.params;
  const res = await backendFetch(
    `/api/v1/tenants/current/members/${userId}`,
    {
      method: "DELETE",
      headers,
    }
  );
  if (res.status === 204) {
    return new NextResponse(null, { status: 204 });
  }
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
