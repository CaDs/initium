import { cookies } from "next/headers";
import {
  DecodeError,
  HttpError,
  TransportError,
  requestJson,
  type APIError,
} from "@initium/core";
import { Effect, Either, type Schema } from "effect";

import { API_URL } from "./env";

export type ApiResult<T> =
  { ok: true; data: T } | { ok: false; error: APIError };

type ApiFetchOptions = Omit<RequestInit, "body" | "headers"> & {
  readonly body?: unknown;
  readonly headers?: Readonly<Record<string, string>>;
};

export async function apiFetch<A, I>(
  path: string,
  options: ApiFetchOptions,
  schema: Schema.Schema<A, I>,
): Promise<ApiResult<A>> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get("access_token")?.value;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };

  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
    headers["Cookie"] = `access_token=${accessToken}`;
  }

  const result = await Effect.runPromise(
    Effect.either(
      requestJson(
        fetch,
        {
          url: `${API_URL}${path}`,
          method: options.method,
          headers,
          body: options.body,
          cache: "no-store",
        },
        schema,
      ),
    ),
  );

  if (Either.isRight(result)) return { ok: true, data: result.right };

  const failure = result.left;
  if (failure instanceof HttpError) {
    return {
      ok: false,
      error: {
        code: failure.code,
        message: failure.message,
        ...(failure.requestId === undefined
          ? {}
          : { request_id: failure.requestId }),
      },
    };
  }
  if (failure instanceof DecodeError) {
    return {
      ok: false,
      error: { code: "VALIDATION_ERROR", message: failure.message },
    };
  }
  if (failure instanceof TransportError) {
    return {
      ok: false,
      error: { code: "UNKNOWN", message: "API request failed" },
    };
  }
  return {
    ok: false,
    error: { code: "UNKNOWN", message: "API request failed" },
  };
}
