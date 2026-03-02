import { NextResponse } from "next/server";

import { readApiErrorMessage } from "@/lib/api/error";
import { backendFetch } from "@/lib/api/server";

type GoogleStartResponse = {
  url: string;
};

type ErrorResponse = {
  error: string;
};

export async function GET() {
  const res = await backendFetch("/api/v1/auth/google/start", {
    method: "GET",
    redirect: "manual",
  });

  const setCookies =
    typeof res.headers.getSetCookie === "function"
      ? res.headers.getSetCookie()
      : [];

  const location = res.headers.get("location");
  if (location && res.status >= 300 && res.status < 400) {
    const out = NextResponse.json({ url: location } satisfies GoogleStartResponse);
    for (const c of setCookies) {
      out.headers.append("set-cookie", c);
    }
    return out;
  }

  if (!res.ok) {
    const message = await readApiErrorMessage(
      res,
      "google oauth is not available"
    );
    const out = NextResponse.json(
      { error: message } satisfies ErrorResponse,
      { status: res.status }
    );
    for (const c of setCookies) {
      out.headers.append("set-cookie", c);
    }
    return out;
  }

  const out = NextResponse.json(
    { error: "google oauth start did not return redirect url" } satisfies ErrorResponse,
    { status: 502 }
  );
  for (const c of setCookies) {
    out.headers.append("set-cookie", c);
  }
  return out;
}
