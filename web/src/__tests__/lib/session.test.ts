import { beforeEach, describe, expect, it, vi } from "vitest";

// session.ts is the single point of cookie management for auth state.
// Bugs here are silent — login appears to succeed but the cookie is
// scoped wrong, or logout looks fine but the refresh token survives.
// These tests verify the contract by inspecting the mocked cookies()
// call sequence.

type CookieOpts = {
  httpOnly?: boolean;
  secure?: boolean;
  sameSite?: string;
  path?: string;
  maxAge?: number;
};

type CookieCall =
  | { kind: "set"; name: string; value: string; opts: CookieOpts }
  | { kind: "delete-name"; name: string }
  | { kind: "delete-with-opts"; name: string; path: string };

const calls: CookieCall[] = [];
const store = new Map<string, string>();

vi.mock("next/headers", () => ({
  cookies: async () => ({
    set: (name: string, value: string, opts: CookieOpts = {}) => {
      calls.push({ kind: "set", name, value, opts });
      store.set(name, value);
    },
    get: (name: string) => (store.has(name) ? { value: store.get(name) } : undefined),
    delete: (arg: string | { name: string; path: string }) => {
      if (typeof arg === "string") {
        calls.push({ kind: "delete-name", name: arg });
      } else {
        calls.push({ kind: "delete-with-opts", name: arg.name, path: arg.path });
      }
      store.delete(typeof arg === "string" ? arg : arg.name);
    },
  }),
}));

beforeEach(() => {
  calls.length = 0;
  store.clear();
});

describe("setTokenCookies", () => {
  it("writes access_token (path /) + refresh_token (path /api/auth) with httpOnly + Lax", async () => {
    const { setTokenCookies } = await import("@/lib/session");
    await setTokenCookies("ACCESS", "REFRESH");

    expect(calls).toHaveLength(2);

    const setCalls = calls.filter((c): c is Extract<CookieCall, { kind: "set" }> =>
      c.kind === "set",
    );

    const access = setCalls.find((c) => c.name === "access_token");
    const refresh = setCalls.find((c) => c.name === "refresh_token");

    expect(access).toBeDefined();
    expect(access?.value).toBe("ACCESS");
    expect(access?.opts.httpOnly).toBe(true);
    expect(access?.opts.sameSite).toBe("lax");
    expect(access?.opts.path).toBe("/");
    expect(access?.opts.maxAge).toBe(900);

    expect(refresh).toBeDefined();
    expect(refresh?.value).toBe("REFRESH");
    expect(refresh?.opts.path).toBe("/api/auth");
    expect(refresh?.opts.maxAge).toBe(604800);
  });
});

describe("clearTokenCookies", () => {
  it("deletes both tokens, scoping refresh_token by the same path it was set with", async () => {
    const { clearTokenCookies } = await import("@/lib/session");
    await clearTokenCookies();

    expect(calls).toContainEqual({ kind: "delete-name", name: "access_token" });
    expect(calls).toContainEqual({
      kind: "delete-with-opts",
      name: "refresh_token",
      path: "/api/auth",
    });
  });
});

describe("getAccessToken / hasSession", () => {
  it("getAccessToken returns the stored value", async () => {
    store.set("access_token", "VAL");
    const { getAccessToken } = await import("@/lib/session");
    expect(await getAccessToken()).toBe("VAL");
  });

  it("getAccessToken returns undefined when no cookie exists", async () => {
    const { getAccessToken } = await import("@/lib/session");
    expect(await getAccessToken()).toBeUndefined();
  });

  it("hasSession returns true only when access_token is present", async () => {
    const { hasSession } = await import("@/lib/session");
    expect(await hasSession()).toBe(false);

    store.set("access_token", "VAL");
    expect(await hasSession()).toBe(true);
  });
});
