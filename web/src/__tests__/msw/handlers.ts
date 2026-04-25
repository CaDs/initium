import { http, HttpResponse } from "msw";

// Default backend base URL — must match the API_URL fallback in src/lib/env.ts.
// Tests can override per-handler with absoluteUrl().
const API_URL = process.env.API_URL ?? "http://localhost:8000";

export const absoluteUrl = (path: string) => `${API_URL}${path}`;

// Default success-path handlers covering the 9 Huma backend endpoints.
// Each test can override with server.use(http.post(...)) for failure
// paths. Keeping defaults "happy" means a test that forgets to mock
// an endpoint surfaces as a missing-handler error instead of silently
// returning a 200 with stale fixture data.
export const defaultHandlers = [
  http.get(absoluteUrl("/api/landing"), () =>
    HttpResponse.json({
      name: "Initium",
      description: "Test landing payload",
      version: "0.0.0-test",
    }),
  ),

  http.post(absoluteUrl("/api/auth/magic-link"), () =>
    HttpResponse.json({ message: "magic link sent" }),
  ),

  http.post(absoluteUrl("/api/auth/refresh"), () =>
    HttpResponse.json({
      access_token: "test-access-token",
      refresh_token: "test-refresh-token",
    }),
  ),

  http.post(absoluteUrl("/api/auth/logout"), () =>
    HttpResponse.json({ message: "logged out" }),
  ),

  http.post(absoluteUrl("/api/auth/logout-all"), () =>
    HttpResponse.json({ message: "all sessions revoked" }),
  ),

  http.post(absoluteUrl("/api/auth/mobile/google"), () =>
    HttpResponse.json({
      access_token: "test-access-token",
      refresh_token: "test-refresh-token",
    }),
  ),

  http.post(absoluteUrl("/api/auth/mobile/verify"), () =>
    HttpResponse.json({
      access_token: "test-access-token",
      refresh_token: "test-refresh-token",
    }),
  ),

  http.get(absoluteUrl("/api/me"), () =>
    HttpResponse.json({
      id: "test-user-id",
      email: "user@example.com",
      name: "Test User",
      avatar_url: "",
      role: "user",
      created_at: "2026-01-01T00:00:00Z",
    }),
  ),

  http.patch(absoluteUrl("/api/me"), async ({ request }) => {
    const body = (await request.json()) as { name: string };
    return HttpResponse.json({
      id: "test-user-id",
      email: "user@example.com",
      name: body.name,
      avatar_url: "",
      role: "user",
      created_at: "2026-01-01T00:00:00Z",
    });
  }),
];

// errorResponse builds a backend-shaped error envelope for handler overrides.
// Mirrors the Huma APIError wire shape (code, message, request_id?).
export const errorResponse = (status: number, code: string, message: string) =>
  HttpResponse.json({ code, message }, { status });
