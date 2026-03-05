import { cookies } from "next/headers";
import { type NextRequest, NextResponse } from "next/server";

import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";

/**
 * Proxy credential OAuth callback through Next.js so that
 * PKCE cookies (set on localhost:3900) are available.
 *
 * Flow: Provider → localhost:3900/api/credentials/oauth/{provider}/callback?code=...&state=...
 *       → this route reads cookies → forwards to Go API → follows redirect
 */
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ provider: string }> }
) {
  const { provider } = await params;
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantId = jar.get(TENANT_COOKIE)?.value;

  // Forward the full query string (code, state, scope, etc.) to Go backend
  const { searchParams } = request.nextUrl;
  const qs = searchParams.toString();

  // Read PKCE verifier cookie and forward it to the backend
  const verifierCookie = jar.get("micro-dp-cred-oauth-verifier")?.value;

  const headers: Record<string, string> = {};
  if (token) headers["Authorization"] = `Bearer ${token}`;
  if (tenantId) headers["X-Tenant-ID"] = tenantId;

  // Forward the PKCE cookies as Cookie header to Go backend
  const cookiePairs: string[] = [];
  if (verifierCookie)
    cookiePairs.push(`micro-dp-cred-oauth-verifier=${verifierCookie}`);
  const stateCookie = jar.get("micro-dp-cred-oauth-state")?.value;
  if (stateCookie)
    cookiePairs.push(`micro-dp-cred-oauth-state=${stateCookie}`);
  if (cookiePairs.length > 0) {
    headers["Cookie"] = cookiePairs.join("; ");
  }

  const res = await backendFetch(
    `/api/v1/credentials/${provider}/callback?${qs}`,
    {
      headers,
      redirect: "manual",
    }
  );

  // Backend returns 302 redirect to integrations page on success/failure
  const location = res.headers.get("location");
  if (location && res.status >= 300 && res.status < 400) {
    const out = NextResponse.redirect(location);
    // Forward any Set-Cookie from backend (e.g., cookie cleanup)
    const setCookies =
      typeof res.headers.getSetCookie === "function"
        ? res.headers.getSetCookie()
        : [];
    for (const c of setCookies) {
      out.headers.append("set-cookie", c);
    }
    return out;
  }

  // Fallback: return backend response as-is
  const body = await res.text();
  return new NextResponse(body, {
    status: res.status,
    headers: { "content-type": res.headers.get("content-type") || "text/plain" },
  });
}
