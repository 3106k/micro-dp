import type { ResponseCookie } from "next/dist/compiled/@edge-runtime/cookies";

export const TOKEN_COOKIE = "micro-dp-token";
export const TENANT_COOKIE = "micro-dp-tenant-id";

const isProduction = process.env.NODE_ENV === "production";

export const cookieOptions: Partial<ResponseCookie> = {
  httpOnly: true,
  secure: isProduction,
  sameSite: "lax",
  path: "/",
};
