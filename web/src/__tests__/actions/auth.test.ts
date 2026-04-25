import { beforeEach, describe, expect, it, vi } from "vitest";
import { http } from "msw";

import { server } from "../msw/server";
import { absoluteUrl, errorResponse } from "../msw/handlers";

// Server Action contract for requestMagicLink:
//   - Empty / missing email → early return { ok: false, "Email is required." }
//   - Backend error → { ok: false, message: error.message }
//   - Backend success → { ok: true, message: backend.message }

vi.mock("next/headers", () => ({
  cookies: async () => ({ get: () => undefined }),
}));

const formData = (entries: Record<string, string>): FormData => {
  const fd = new FormData();
  for (const [k, v] of Object.entries(entries)) fd.append(k, v);
  return fd;
};

const initialState = { ok: false, message: "" };

beforeEach(() => {
  // Each test starts from a clean cookie + module state.
  vi.resetModules();
});

describe("requestMagicLink Server Action", () => {
  it("returns an early validation error when email is missing", async () => {
    const { requestMagicLink } = await import("@/actions/auth");
    const result = await requestMagicLink(initialState, formData({}));

    expect(result).toEqual({ ok: false, message: "Email is required." });
  });

  it("returns an early validation error for empty string email", async () => {
    const { requestMagicLink } = await import("@/actions/auth");
    const result = await requestMagicLink(initialState, formData({ email: "" }));

    expect(result).toEqual({ ok: false, message: "Email is required." });
  });

  it("returns the backend success message on 2xx", async () => {
    const { requestMagicLink } = await import("@/actions/auth");
    const result = await requestMagicLink(
      initialState,
      formData({ email: "user@example.com" }),
    );

    expect(result).toEqual({ ok: true, message: "magic link sent" });
  });

  it("propagates the backend error message on a 4xx envelope", async () => {
    server.use(
      http.post(absoluteUrl("/api/auth/magic-link"), () =>
        errorResponse(400, "EMAIL_REQUIRED", "email is required"),
      ),
    );

    const { requestMagicLink } = await import("@/actions/auth");
    const result = await requestMagicLink(
      initialState,
      formData({ email: "x@y.com" }),
    );

    expect(result).toEqual({ ok: false, message: "email is required" });
  });

  it("falls back to the generic UNKNOWN error message on non-JSON failure", async () => {
    server.use(
      http.post(absoluteUrl("/api/auth/magic-link"), () =>
        new Response("boom", { status: 502, statusText: "Bad Gateway" }),
      ),
    );

    const { requestMagicLink } = await import("@/actions/auth");
    const result = await requestMagicLink(
      initialState,
      formData({ email: "x@y.com" }),
    );

    expect(result.ok).toBe(false);
    expect(result.message).toBe("Bad Gateway");
  });
});
