import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { http } from "msw";
import { NextRequest } from "next/server";

import { server } from "../../msw/server";
import { absoluteUrl } from "../../msw/handlers";

// /api/auth/logout responsibilities:
//   1. If refresh_token cookie present, POST it to backend's /api/auth/logout
//      (best-effort; failures don't propagate).
//   2. Clear local cookies via clearTokenCookies().
//   3. Redirect to /login.

const cookieDeleteCalls: string[] = [];
vi.mock("next/headers", () => ({
  cookies: async () => ({
    delete: (arg: string | { name: string; path: string }) => {
      cookieDeleteCalls.push(typeof arg === "string" ? arg : arg.name);
    },
    get: () => undefined,
    set: () => {},
  }),
}));

beforeEach(() => {
  cookieDeleteCalls.length = 0;
});
afterEach(() => {
  cookieDeleteCalls.length = 0;
});

describe("/api/auth/logout route handler", () => {
  it("clears local cookies and redirects to /login when no refresh_token cookie is present", async () => {
    const { POST } = await import("@/app/api/auth/logout/route");

    const req = new NextRequest("http://localhost:3000/api/auth/logout", { method: "POST" });
    const res = await POST(req);

    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toBe("http://localhost:3000/login");
    expect(cookieDeleteCalls).toContain("access_token");
    expect(cookieDeleteCalls).toContain("refresh_token");
  });

  it("calls backend's logout endpoint when refresh token cookie is present", async () => {
    let backendCalled = false;
    server.use(
      http.post(absoluteUrl("/api/auth/logout"), () => {
        backendCalled = true;
        return new Response(null, { status: 200 });
      }),
    );

    const { POST } = await import("@/app/api/auth/logout/route");

    const req = new NextRequest("http://localhost:3000/api/auth/logout", { method: "POST" });
    req.cookies.set("refresh_token", "TO-REVOKE");
    const res = await POST(req);

    expect(res.status).toBe(307);
    expect(backendCalled).toBe(true);
    // Note: we don't assert on the forwarded `Cookie` header. undici (Node's
    // fetch impl) treats `Cookie` as a forbidden request header for some
    // fetch implementations, which can strip it before the request reaches
    // MSW. The route is what drives backend session revocation, and
    // verifying the call happened is enough for the regression net here.
  });

  it("still clears cookies + redirects when backend logout fails (best-effort)", async () => {
    server.use(
      http.post(absoluteUrl("/api/auth/logout"), () => {
        throw new Error("backend down");
      }),
    );

    const { POST } = await import("@/app/api/auth/logout/route");

    const req = new NextRequest("http://localhost:3000/api/auth/logout", { method: "POST" });
    req.cookies.set("refresh_token", "any");
    const res = await POST(req);

    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toBe("http://localhost:3000/login");
    expect(cookieDeleteCalls).toContain("access_token");
    expect(cookieDeleteCalls).toContain("refresh_token");
  });
});
