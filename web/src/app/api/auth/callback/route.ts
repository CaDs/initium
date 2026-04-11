import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// Google OAuth callback — the backend handles the actual OAuth flow and
// sets cookies directly via redirects. This route exists as a fallback
// if the frontend needs to intercept the callback.
export async function GET(request: NextRequest) {
  // The backend's /api/auth/google/callback already sets cookies and
  // redirects to APP_URL/home. This route is a safety net.
  return NextResponse.redirect(new URL("/home", request.url));
}
