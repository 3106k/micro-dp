import { NextRequest, NextResponse } from "next/server";

import { readApiErrorMessage } from "@/lib/api/error";
import { backendFetch } from "@/lib/api/server";

export async function POST(request: NextRequest) {
  const body = await request.json();

  const res = await backendFetch("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    return NextResponse.json(
      { error: await readApiErrorMessage(res, "registration failed") },
      { status: res.status }
    );
  }

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
