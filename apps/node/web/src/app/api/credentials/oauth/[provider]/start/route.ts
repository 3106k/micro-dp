import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";

type ErrorResponse = components["schemas"]["ErrorResponse"];

/**
 * Proxy credential OAuth start. Returns JSON { url } + Set-Cookie
 * so the browser stores PKCE cookies on localhost:3900 (same origin as callback).
 */
export async function GET(
  _request: Request,
  { params }: { params: Promise<{ provider: string }> }
) {
  const { provider } = await params;
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantId = jar.get(TENANT_COOKIE)?.value;
  if (!token || !tenantId) {
    return NextResponse.json(
      { error: "not authenticated" } satisfies ErrorResponse,
      { status: 401 }
    );
  }

  const res = await backendFetch(`/api/v1/credentials/${provider}/start`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
    redirect: "manual",
  });

  const setCookies =
    typeof res.headers.getSetCookie === "function"
      ? res.headers.getSetCookie()
      : [];

  const location = res.headers.get("location");
  if (location && res.status >= 300 && res.status < 400) {
    const out = NextResponse.json({ url: location });
    for (const c of setCookies) {
      out.headers.append("set-cookie", c);
    }
    return out;
  }

  const data = await res.json().catch(() => ({ error: "unexpected response" }));
  const out = NextResponse.json(data, { status: res.status });
  for (const c of setCookies) {
    out.headers.append("set-cookie", c);
  }
  return out;
}
