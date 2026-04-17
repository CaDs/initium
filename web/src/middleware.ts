import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import {
  COOKIE_SECURE,
  ACCESS_TOKEN_MAX_AGE,
  REFRESH_TOKEN_MAX_AGE,
  REFRESH_TOKEN_PATH,
} from "./lib/cookie-config";
import { tokenPairSchema } from "./lib/schemas";
import { API_URL } from "./lib/env";
const PROTECTED_PATHS = ["/home"];
const DEV_BYPASS = process.env.DEV_BYPASS_AUTH === "true";

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  const isProtected = PROTECTED_PATHS.some((p) => pathname.startsWith(p));
  if (!isProtected) return NextResponse.next();

  if (DEV_BYPASS) return NextResponse.next();

  const accessToken = request.cookies.get("access_token")?.value;
  if (accessToken) return NextResponse.next();

  // No access token — attempt refresh if refresh token exists
  const refreshToken = request.cookies.get("refresh_token")?.value;
  if (!refreshToken) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  try {
    const refreshRes = await fetch(`${API_URL}/api/auth/refresh`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Cookie: `refresh_token=${refreshToken}`,
      },
    });

    if (!refreshRes.ok) {
      const response = NextResponse.redirect(new URL("/login", request.url));
      response.cookies.delete("access_token");
      response.cookies.delete({ name: "refresh_token", path: REFRESH_TOKEN_PATH });
      return response;
    }

    const parsed = tokenPairSchema.safeParse(await refreshRes.json());
    if (!parsed.success) {
      // Unexpected response shape — never write "undefined" cookies.
      const response = NextResponse.redirect(new URL("/login", request.url));
      response.cookies.delete("access_token");
      response.cookies.delete({ name: "refresh_token", path: REFRESH_TOKEN_PATH });
      return response;
    }
    const tokens = parsed.data;
    const response = NextResponse.next();

    response.cookies.set("access_token", tokens.access_token, {
      httpOnly: true,
      secure: COOKIE_SECURE,
      sameSite: "lax",
      path: "/",
      maxAge: ACCESS_TOKEN_MAX_AGE,
    });

    response.cookies.set("refresh_token", tokens.refresh_token, {
      httpOnly: true,
      secure: COOKIE_SECURE,
      sameSite: "lax",
      path: REFRESH_TOKEN_PATH,
      maxAge: REFRESH_TOKEN_MAX_AGE,
    });

    return response;
  } catch {
    return NextResponse.redirect(new URL("/login", request.url));
  }
}

export const config = {
  matcher: ["/home/:path*"],
};
