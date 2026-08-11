import {
  APIErrorSchema,
  MessageResponseSchema,
  TokenPairSchema,
  UserSchema,
} from "@initium/core";
import { Either, Schema } from "effect";
import { describe, expect, it } from "vitest";

const parses = <A, I>(schema: Schema.Schema<A, I>, input: unknown) =>
  Either.isRight(Schema.decodeUnknownEither(schema)(input));

describe("userSchema", () => {
  it("parses a valid user", () => {
    const result = parses(UserSchema, {
      id: "550e8400-e29b-41d4-a716-446655440000",
      email: "test@example.com",
      name: "Test User",
      avatar_url: "https://example.com/avatar.png",
      role: "user",
      created_at: "2025-01-01T00:00:00Z",
    });
    expect(result).toBe(true);
  });

  it("rejects missing email", () => {
    const result = parses(UserSchema, {
      id: "123",
      name: "Test",
      avatar_url: "",
      role: "user",
      created_at: "",
    });
    expect(result).toBe(false);
  });

  it("rejects invalid email format", () => {
    const result = parses(UserSchema, {
      id: "123",
      email: "not-an-email",
      name: "Test",
      avatar_url: "",
      role: "user",
      created_at: "",
    });
    expect(result).toBe(false);
  });
});

describe("tokenPairSchema", () => {
  it("parses valid token pair", () => {
    const result = parses(TokenPairSchema, {
      access_token: "eyJhbGciOiJSUzI1NiJ9...",
      refresh_token: "abc123",
    });
    expect(result).toBe(true);
  });

  it("rejects missing refresh_token", () => {
    const result = parses(TokenPairSchema, {
      access_token: "abc",
    });
    expect(result).toBe(false);
  });
});

describe("errorSchema", () => {
  it("parses error with optional request_id", () => {
    const result = parses(APIErrorSchema, {
      code: "TOKEN_INVALID",
      message: "token is invalid",
      request_id: "req-123",
    });
    expect(result).toBe(true);
  });

  it("parses error without request_id", () => {
    const result = parses(APIErrorSchema, {
      code: "INTERNAL_ERROR",
      message: "something went wrong",
    });
    expect(result).toBe(true);
  });
});

describe("messageSchema", () => {
  it("parses message response", () => {
    const result = parses(MessageResponseSchema, {
      message: "magic link sent",
    });
    expect(result).toBe(true);
  });
});
