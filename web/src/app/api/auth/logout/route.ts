import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { clearTokenCookies } from "@/lib/session";
import { API_URL } from "@/lib/env";

export async function POST(request: NextRequest) {
  const refreshToken = request.cookies.get("refresh_token")?.value;

  if (refreshToken) {
    await fetch(`${API_URL}/api/auth/logout`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Cookie: `refresh_token=${refreshToken}`,
      },
    }).catch(() => {}); // best-effort backend logout
  }

  await clearTokenCookies();
  return NextResponse.redirect(new URL("/login", request.url));
}
