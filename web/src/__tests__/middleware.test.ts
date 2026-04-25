import { describe, expect, it, beforeEach, vi } from "vitest";
import { http } from "msw";
import { NextRequest } from "next/server";

import { server } from "./msw/server";
import { absoluteUrl, errorResponse } from "./msw/handlers";

// middleware.ts is the cookie-driven token-refresh hot path. Six branches:
//   1. Unprotected path → next()
//   2. DEV_BYPASS_AUTH=true → next()
//   3. Has access_token → next()
//   4. No access_token, no refresh_token → /login redirect
//   5. Has refresh_token, refresh succeeds → next() with new cookies
//   6. Has refresh_token, refresh fails → /login redirect with cookies cleared
//
// Tests construct NextRequest directly (no Next.js runtime needed) and
// assert on the NextResponse — status, redirect Location, Set-Cookie
// headers. MSW intercepts the /api/auth/refresh call.

const PROTECTED = "http://localhost:3000/home/dashboard";
const PUBLIC = "http://localhost:3000/login";

// makeRequest builds a NextRequest with the given cookies. Cookies are set
// via .cookies.set() rather than the constructor's `cookie` header because
// happy-dom's Headers + NextRequest cookie parsing don't compose reliably
// — set() is the documented test path.
const makeRequest = (url: string, cookies: Record<string, string> = {}) => {
  const req = new NextRequest(url);
  for (const [k, v] of Object.entries(cookies)) {
    req.cookies.set(k, v);
  }
  return req;
};

// loadMiddleware re-imports the middleware module after env var changes so
// the `DEV_BYPASS` constant is recomputed. Without this, the module-level
// `const DEV_BYPASS = ...` snapshots the env at first-load time.
const loadMiddleware = async () => {
  const mod = await import("@/middleware");
  return mod.middleware;
};

describe("middleware", () => {
  // Reset module cache + env between tests so each test gets a freshly
  // imported middleware module that snapshots its env vars at load time.
  beforeEach(() => {
    delete process.env.DEV_BYPASS_AUTH;
    vi.resetModules();
  });

  it("calls next() for unprotected paths", async () => {
    const middleware = await loadMiddleware();
    const res = await middleware(makeRequest(PUBLIC));

    expect(res.status).toBe(200);
    // No redirect — the response is a pass-through, not a NextResponse.redirect.
    expect(res.headers.get("location")).toBeNull();
  });

  it("calls next() when DEV_BYPASS_AUTH is true even for protected paths", async () => {
    process.env.DEV_BYPASS_AUTH = "true";
    // beforeEach already reset modules; loadMiddleware now performs a fresh
    // import that captures DEV_BYPASS=true from the env above.
    const middleware = await loadMiddleware();

    const res = await middleware(makeRequest(PROTECTED));

    expect(res.status).toBe(200);
    expect(res.headers.get("location")).toBeNull();
  });

  it("calls next() when access_token cookie is present", async () => {
    const middleware = await loadMiddleware();
    const res = await middleware(
      makeRequest(PROTECTED, { access_token: "valid-access" }),
    );

    expect(res.status).toBe(200);
    expect(res.headers.get("location")).toBeNull();
  });

  it("redirects to /login when neither access_token nor refresh_token exists", async () => {
    const middleware = await loadMiddleware();
    const res = await middleware(makeRequest(PROTECTED));

    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toBe("http://localhost:3000/login");
  });

  it("refreshes tokens and proceeds when only refresh_token exists and refresh succeeds", async () => {
    const middleware = await loadMiddleware();
    const res = await middleware(
      makeRequest(PROTECTED, { refresh_token: "valid-refresh" }),
    );

    expect(res.status).toBe(200);
    expect(res.headers.get("location")).toBeNull();

    // Both new tokens must be set as cookies (NextResponse stores them on
    // its `.cookies` API, which the test inspects directly).
    const accessCookie = res.cookies.get("access_token");
    const refreshCookie = res.cookies.get("refresh_token");
    expect(accessCookie?.value).toBe("test-access-token");
    expect(refreshCookie?.value).toBe("test-refresh-token");
  });

  it("redirects to /login and clears cookies when refresh returns non-2xx", async () => {
    server.use(
      http.post(absoluteUrl("/api/auth/refresh"), () =>
        errorResponse(401, "SESSION_EXPIRED", "session expired"),
      ),
    );

    const middleware = await loadMiddleware();
    const res = await middleware(
      makeRequest(PROTECTED, { refresh_token: "expired-refresh" }),
    );

    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toBe("http://localhost:3000/login");

    // Stale cookies must be deleted. NextResponse marks deleted cookies
    // with an empty value (the browser sees Max-Age=0).
    expect(res.cookies.get("access_token")?.value).toBe("");
    expect(res.cookies.get("refresh_token")?.value).toBe("");
  });

  it("redirects to /login and clears cookies when refresh response shape is invalid", async () => {
    server.use(
      http.post(absoluteUrl("/api/auth/refresh"), () =>
        // Wrong shape — Zod rejects, middleware must not write `undefined`
        // cookies (regression test for a real failure mode).
        Response.json({ wrong: "shape" }),
      ),
    );

    const middleware = await loadMiddleware();
    const res = await middleware(
      makeRequest(PROTECTED, { refresh_token: "garbage-response" }),
    );

    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toBe("http://localhost:3000/login");
    expect(res.cookies.get("access_token")?.value).toBe("");
    expect(res.cookies.get("refresh_token")?.value).toBe("");
  });

  it("redirects to /login when refresh request throws (network error)", async () => {
    server.use(
      http.post(absoluteUrl("/api/auth/refresh"), () => {
        throw new Error("network down");
      }),
    );

    const middleware = await loadMiddleware();
    const res = await middleware(
      makeRequest(PROTECTED, { refresh_token: "any-refresh" }),
    );

    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toBe("http://localhost:3000/login");
  });
});
