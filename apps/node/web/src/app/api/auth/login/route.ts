import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

import { backendFetch } from "@/lib/api/server";
import {
  TOKEN_COOKIE,
  TENANT_COOKIE,
  cookieOptions,
} from "@/lib/auth/constants";
import type { components } from "@/lib/api/generated";

type LoginResponse = components["schemas"]["LoginResponse"];
type MeResponse = components["schemas"]["MeResponse"];
type ErrorResponse = components["schemas"]["ErrorResponse"];

export async function POST(request: NextRequest) {
  const body = await request.json();

  const loginRes = await backendFetch("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(body),
  });

  if (!loginRes.ok) {
    const err: ErrorResponse = await loginRes.json();
    return NextResponse.json(err, { status: loginRes.status });
  }

  const loginData: LoginResponse = await loginRes.json();

  // Fetch /me to get default tenant
  const meRes = await backendFetch("/api/v1/auth/me", {
    headers: { Authorization: `Bearer ${loginData.token}` },
  });

  const jar = await cookies();

  jar.set(TOKEN_COOKIE, loginData.token, cookieOptions);

  if (meRes.ok) {
    const meData: MeResponse = await meRes.json();
    if (meData.tenants.length > 0) {
      jar.set(TENANT_COOKIE, meData.tenants[0].id, cookieOptions);
    }
    return NextResponse.json({
      user_id: meData.user_id,
      email: meData.email,
      display_name: meData.display_name,
    });
  }

  return NextResponse.json({ ok: true });
}
