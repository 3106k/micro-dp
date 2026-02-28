import { NextRequest, NextResponse } from "next/server";

import { TOKEN_COOKIE } from "@/lib/auth/constants";

export function middleware(request: NextRequest) {
  const token = request.cookies.get(TOKEN_COOKIE)?.value;
  const { pathname } = request.nextUrl;

  // Protect /dashboard routes
  if (pathname.startsWith("/dashboard") && !token) {
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
  matcher: ["/dashboard/:path*", "/signin", "/signup"],
};
