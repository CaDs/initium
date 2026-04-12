import { describe, it, expect } from "vitest";
import {
  userSchema,
  tokenPairSchema,
  errorSchema,
  messageSchema,
} from "../lib/schemas";

describe("userSchema", () => {
  it("parses a valid user", () => {
    const result = userSchema.safeParse({
      id: "550e8400-e29b-41d4-a716-446655440000",
      email: "test@example.com",
      name: "Test User",
      avatar_url: "https://example.com/avatar.png",
      created_at: "2025-01-01T00:00:00Z",
    });
    expect(result.success).toBe(true);
  });

  it("rejects missing email", () => {
    const result = userSchema.safeParse({
      id: "123",
      name: "Test",
      avatar_url: "",
      created_at: "",
    });
    expect(result.success).toBe(false);
  });

  it("rejects invalid email format", () => {
    const result = userSchema.safeParse({
      id: "123",
      email: "not-an-email",
      name: "Test",
      avatar_url: "",
      created_at: "",
    });
    expect(result.success).toBe(false);
  });
});

describe("tokenPairSchema", () => {
  it("parses valid token pair", () => {
    const result = tokenPairSchema.safeParse({
      access_token: "eyJhbGciOiJSUzI1NiJ9...",
      refresh_token: "abc123",
    });
    expect(result.success).toBe(true);
  });

  it("rejects missing refresh_token", () => {
    const result = tokenPairSchema.safeParse({
      access_token: "abc",
    });
    expect(result.success).toBe(false);
  });
});

describe("errorSchema", () => {
  it("parses error with optional request_id", () => {
    const result = errorSchema.safeParse({
      code: "TOKEN_INVALID",
      message: "token is invalid",
      request_id: "req-123",
    });
    expect(result.success).toBe(true);
  });

  it("parses error without request_id", () => {
    const result = errorSchema.safeParse({
      code: "INTERNAL_ERROR",
      message: "something went wrong",
    });
    expect(result.success).toBe(true);
  });
});

describe("messageSchema", () => {
  it("parses message response", () => {
    const result = messageSchema.safeParse({ message: "magic link sent" });
    expect(result.success).toBe(true);
  });
});
