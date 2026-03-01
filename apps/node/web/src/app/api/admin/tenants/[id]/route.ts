import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TOKEN_COOKIE } from "@/lib/auth/constants";

type ErrorResponse = components["schemas"]["ErrorResponse"];
type MeResponse = components["schemas"]["MeResponse"];

type AdminAuthResult =
  | { headers: { Authorization: string } }
  | { error: ErrorResponse; status: 401 | 403 };

async function adminAuthHeader(): Promise<AdminAuthResult> {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  if (!token) {
    return { error: { error: "not authenticated" }, status: 401 };
  }

  const meRes = await backendFetch("/api/v1/auth/me", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!meRes.ok) {
    return { error: { error: "not authenticated" }, status: 401 };
  }
  const me: MeResponse = await meRes.json();
  if (me.platform_role !== "superadmin") {
    return { error: { error: "forbidden" }, status: 403 };
  }

  return { headers: { Authorization: `Bearer ${token}` } };
}

export async function PATCH(
  request: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  const auth = await adminAuthHeader();
  if ("error" in auth) {
    return NextResponse.json(auth.error, { status: auth.status });
  }

  const { id } = await context.params;
  const body = await request.json();
  const res = await backendFetch(`/api/v1/admin/tenants/${id}`, {
    method: "PATCH",
    headers: auth.headers,
    body: JSON.stringify(body),
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
