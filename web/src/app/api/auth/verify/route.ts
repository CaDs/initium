import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// Magic link verification — the backend handles token verification and
// sets cookies directly via redirects. This route exists as a fallback.
export async function GET(request: NextRequest) {
  const token = request.nextUrl.searchParams.get("token");
  if (!token) {
    return NextResponse.redirect(new URL("/login?error=invalid_token", request.url));
  }

  // Redirect to backend for verification
  const apiUrl = process.env.API_URL || "http://localhost:8000";
  return NextResponse.redirect(`${apiUrl}/api/auth/verify?token=${token}`);
}
