import { NextRequest, NextResponse } from "next/server";

import { TOKEN_COOKIE } from "@/lib/auth/constants";

export function middleware(request: NextRequest) {
  const token = request.cookies.get(TOKEN_COOKIE)?.value;
  const { pathname } = request.nextUrl;

  // Protect authenticated routes
  if (
    (pathname.startsWith("/dashboard") ||
      pathname.startsWith("/jobs") ||
      pathname.startsWith("/job-runs") ||
      pathname.startsWith("/datasets") ||
      pathname.startsWith("/connections") ||
      pathname.startsWith("/admin")) &&
    !token
  ) {
    const url = request.nextUrl.clone();
    url.pathname = "/signin";
    url.searchParams.set("callbackUrl", pathname);
    return NextResponse.redirect(url);
  }

  // Redirect authenticated users away from auth pages
  if ((pathname === "/signin" || pathname === "/signup") && token) {
    const url = request.nextUrl.clone();
    url.pathname = "/dashboard";
    url.search = "";
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/dashboard/:path*",
    "/jobs/:path*",
    "/job-runs/:path*",
    "/datasets/:path*",
    "/connections/:path*",
    "/admin/:path*",
    "/signin",
    "/signup",
  ],
};
