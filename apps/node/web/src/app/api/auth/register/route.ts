import { NextRequest, NextResponse } from "next/server";

import { backendFetch } from "@/lib/api/server";
import type { components } from "@/lib/api/generated";

type ErrorResponse = components["schemas"]["ErrorResponse"];

export async function POST(request: NextRequest) {
  const body = await request.json();

  const res = await backendFetch("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify(body),
  });

  const data: components["schemas"]["RegisterResponse"] | ErrorResponse =
    await res.json();

  return NextResponse.json(data, { status: res.status });
}
