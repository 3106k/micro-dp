import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { TOKEN_COOKIE, TENANT_COOKIE } from "@/lib/auth/constants";

export async function POST() {
  const jar = await cookies();
  jar.delete(TOKEN_COOKIE);
  jar.delete(TENANT_COOKIE);
  return NextResponse.json({ ok: true });
}
