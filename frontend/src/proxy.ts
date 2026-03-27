import { NextResponse } from "next/server";

import type { NextRequest } from "next/server";

const protectedRoutes = ["/dashboard"];

const authRoutes = ["/sign-in", "/sign-up"];

export function proxy(request: NextRequest) {
  const path = request.nextUrl.pathname;

  const isAuth = request.cookies.has("session");

  const isProtectedRoute = protectedRoutes.some((route) =>
    path.startsWith(route),
  );
  if (isProtectedRoute && !isAuth) {
    return NextResponse.redirect(new URL("/sign-in", request.url));
  }

  const isAuthRoute = authRoutes.some((route) => path.startsWith(route));
  if (isAuthRoute && isAuth) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico).*)"],
};
