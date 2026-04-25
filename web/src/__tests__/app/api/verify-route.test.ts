import { describe, expect, it } from "vitest";
import { NextRequest } from "next/server";

import { GET } from "@/app/api/auth/verify/route";

// /api/auth/verify is a thin redirect to the backend's verify endpoint.
// Two cases:
//   1. Missing token → /login?error=invalid_token
//   2. Token present → 307 to backend's /api/auth/verify?token=...

describe("/api/auth/verify route handler", () => {
  it("redirects to /login when token is missing", async () => {
    const req = new NextRequest("http://localhost:3000/api/auth/verify");
    const res = await GET(req);

    expect(res.status).toBe(307);
    const loc = res.headers.get("location");
    expect(loc).toContain("/login");
    expect(loc).toContain("error=invalid_token");
  });

  it("redirects to backend verify endpoint when token is present", async () => {
    const req = new NextRequest(
      "http://localhost:3000/api/auth/verify?token=magic-link-token",
    );
    const res = await GET(req);

    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toContain("/api/auth/verify?token=magic-link-token");
  });
});
