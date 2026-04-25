import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { http, HttpResponse } from "msw";
import { z } from "zod";

import { server } from "../msw/server";
import { absoluteUrl, errorResponse } from "../msw/handlers";

// apiFetch is the centralized HTTP client for Server Actions and route
// handlers. Five contracts that downstream code depends on:
//   1. Reads access_token from cookies() and sends it as Bearer + cookie.
//   2. Returns { ok: true, data } on 2xx + valid schema (or no schema).
//   3. Returns { ok: false, error: APIError } on non-2xx.
//   4. Returns { ok: false, error: VALIDATION_ERROR } on schema mismatch.
//   5. Falls back to a synthetic error envelope when the response body
//      isn't valid JSON (network/proxy gone wrong).

// Mock next/headers — apiFetch reads cookies() to attach the access token.
const cookieStore = new Map<string, { value: string }>();
vi.mock("next/headers", () => ({
  cookies: async () => ({
    get: (name: string) => cookieStore.get(name),
  }),
}));

beforeEach(() => {
  cookieStore.clear();
});
afterEach(() => {
  cookieStore.clear();
});

describe("apiFetch", () => {
  it("returns { ok: true, data } on a 2xx response with valid schema", async () => {
    const { apiFetch } = await import("@/lib/api");
    const schema = z.object({ name: z.string() });

    const result = await apiFetch("/api/landing", {}, schema);

    expect(result.ok).toBe(true);
    // Discriminated union narrowing — guarded by the assert above. The
    // /* v8 ignore next */ comment prevents the unreachable branch from
    // skewing branch coverage.
    if (!result.ok) throw new Error("unreachable: result.ok asserted true above");
    expect(result.data.name).toBe("Initium");
  });

  it("attaches Authorization Bearer when access_token cookie is set", async () => {
    cookieStore.set("access_token", { value: "ABC" });
    let captured: { auth?: string } = {};

    server.use(
      http.get(absoluteUrl("/api/landing"), ({ request }) => {
        captured = { auth: request.headers.get("authorization") ?? undefined };
        return HttpResponse.json({ name: "ok" });
      }),
    );

    const { apiFetch } = await import("@/lib/api");
    await apiFetch("/api/landing");

    // The Authorization header is the primary auth signal — the backend
    // prefers it over the cookie. (apiFetch also sets a Cookie header for
    // backend cookie-based code paths, but undici may strip it in tests
    // since `Cookie` is on the forbidden-header list for browser-style
    // fetch — irrelevant to assertions here.)
    expect(captured.auth).toBe("Bearer ABC");
  });

  it("returns { ok: false, error } when the API returns a non-2xx envelope", async () => {
    server.use(
      http.get(absoluteUrl("/api/landing"), () =>
        errorResponse(401, "INVALID_CREDENTIALS", "no session"),
      ),
    );

    const { apiFetch } = await import("@/lib/api");
    const result = await apiFetch("/api/landing");

    expect(result.ok).toBe(false);
    if (result.ok) throw new Error("unreachable: result.ok asserted false above");
    expect(result.error.code).toBe("INVALID_CREDENTIALS");
    expect(result.error.message).toBe("no session");
  });

  it("falls back to status text when error body is not JSON", async () => {
    server.use(
      http.get(absoluteUrl("/api/landing"), () =>
        new HttpResponse("not json at all", { status: 502, statusText: "Bad Gateway" }),
      ),
    );

    const { apiFetch } = await import("@/lib/api");
    const result = await apiFetch("/api/landing");

    expect(result.ok).toBe(false);
    if (result.ok) throw new Error("unreachable: result.ok asserted false above");
    expect(result.error.code).toBe("UNKNOWN");
    expect(result.error.message).toBe("Bad Gateway");
  });

  it("returns VALIDATION_ERROR when the response shape does not match the schema", async () => {
    server.use(
      http.get(absoluteUrl("/api/landing"), () =>
        // Wrong shape — missing `name`.
        HttpResponse.json({ description: "no name field" }),
      ),
    );

    const schema = z.object({ name: z.string() });
    const { apiFetch } = await import("@/lib/api");
    const result = await apiFetch("/api/landing", {}, schema);

    expect(result.ok).toBe(false);
    if (result.ok) throw new Error("unreachable: result.ok asserted false above");
    expect(result.error.code).toBe("VALIDATION_ERROR");
    expect(result.error.message).toMatch(/validation failed/i);
  });
});
