import "server-only";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "./constants";
import type { components } from "@/lib/api/generated";

type MeResponse = components["schemas"]["MeResponse"];

export type AuthContext = {
  me: MeResponse;
  token: string;
  currentTenantId: string;
};

export async function getAuthContext(): Promise<AuthContext> {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;

  if (!token) {
    redirect("/signin");
  }

  const res = await backendFetch("/api/v1/auth/me", {
    headers: { Authorization: `Bearer ${token}` },
  });

  if (!res.ok) {
    redirect("/signin");
  }

  const me: MeResponse = await res.json();
  const tenantCookie = jar.get(TENANT_COOKIE)?.value;
  const tenantIds = new Set(me.tenants.map((tenant) => tenant.id));
  const currentTenantId =
    tenantCookie && tenantIds.has(tenantCookie)
      ? tenantCookie
      : me.tenants[0]?.id ?? "";

  return { me, token, currentTenantId };
}
