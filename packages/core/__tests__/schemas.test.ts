import { Either, Schema } from "effect";
import { describe, expect, it } from "vitest";

import {
  APIErrorSchema,
  MessageResponseSchema,
  TokenPairSchema,
  UserSchema,
} from "../src";

const parses = <A, I>(schema: Schema.Schema<A, I>, input: unknown) =>
  Either.isRight(Schema.decodeUnknownEither(schema)(input));

describe("UserSchema", () => {
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

describe("TokenPairSchema", () => {
  it("parses a valid token pair", () => {
    const result = parses(TokenPairSchema, {
      access_token: "eyJhbGciOiJSUzI1NiJ9...",
      refresh_token: "abc123",
    });
    expect(result).toBe(true);
  });
});

describe("APIErrorSchema", () => {
  it("parses an error with request_id", () => {
    const result = parses(APIErrorSchema, {
      code: "TOKEN_INVALID",
      message: "token is invalid",
      request_id: "req-123",
    });
    expect(result).toBe(true);
  });

  it("parses an error without request_id", () => {
    const result = parses(APIErrorSchema, {
      code: "INTERNAL_ERROR",
      message: "something went wrong",
    });
    expect(result).toBe(true);
  });
});

describe("MessageResponseSchema", () => {
  it("parses a message response", () => {
    const result = parses(MessageResponseSchema, {
      message: "magic link sent",
    });
    expect(result).toBe(true);
  });
});
