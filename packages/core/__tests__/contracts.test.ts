import { Effect, Exit, Schema } from "effect";
import { describe, expect, it, vi } from "vitest";

import {
  HttpError,
  MessageResponseSchema,
  TokenPairSchema,
  TransportError,
  UserSchema,
  requestJson,
} from "../src";

const response = (status: number, body: unknown): Response =>
  ({
    ok: status >= 200 && status < 300,
    status,
    statusText: "STATUS",
    json: async () => body,
  }) as Response;

describe("shared contracts", () => {
  it("decodes the canonical user payload", async () => {
    const value = await Effect.runPromise(
      Schema.decodeUnknown(UserSchema)({
        id: "u1",
        email: "dev@initium.local",
        name: "Dev User",
        avatar_url: "",
        role: "admin",
        created_at: "2026-01-01T00:00:00Z",
      }),
    );

    expect(value.role).toBe("admin");
  });

  it("rejects malformed token pairs", async () => {
    const exit = await Effect.runPromiseExit(
      Schema.decodeUnknown(TokenPairSchema)({ access_token: "only-one" }),
    );

    expect(Exit.isFailure(exit)).toBe(true);
  });

  it("turns backend envelopes into typed HTTP failures", async () => {
    const fetchImpl = vi.fn(async () =>
      response(403, {
        code: "FORBIDDEN",
        message: "admin required",
        request_id: "r1",
      }),
    );

    const failure = await Effect.runPromise(
      Effect.flip(
        requestJson(
          fetchImpl,
          { url: "https://api.example.com/api/admin/ping" },
          UserSchema,
        ),
      ),
    );

    expect(failure).toBeInstanceOf(HttpError);
    expect(failure).toMatchObject({ code: "FORBIDDEN", requestId: "r1" });
    expect(fetchImpl).toHaveBeenCalledOnce();
  });

  it("constructs interruptible JSON requests and decodes their response", async () => {
    const fetchImpl = vi.fn(async () => response(200, { message: "sent" }));

    const value = await Effect.runPromise(
      requestJson(
        fetchImpl,
        {
          url: "https://api.example.com/api/auth/magic-link",
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: { email: "dev@initium.local" },
          cache: "no-store",
        },
        MessageResponseSchema,
      ),
    );

    expect(value.message).toBe("sent");
    expect(fetchImpl).toHaveBeenCalledWith(
      "https://api.example.com/api/auth/magic-link",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ email: "dev@initium.local" }),
        cache: "no-store",
        signal: expect.any(AbortSignal),
      }),
    );
  });

  it("turns rejected fetches into typed transport failures", async () => {
    const failure = await Effect.runPromise(
      Effect.flip(
        requestJson(
          async () => {
            throw new Error("offline");
          },
          { url: "https://api.example.com/healthz" },
          MessageResponseSchema,
        ),
      ),
    );

    expect(failure).toBeInstanceOf(TransportError);
  });

  it("exposes HttpError as a tagged data error", () => {
    const error = new HttpError({
      status: 500,
      code: "SERVER_ERROR",
      message: "boom",
    });
    expect(error._tag).toBe("HttpError");
  });
});
