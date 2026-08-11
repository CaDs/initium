import { Effect, Schema } from "effect";

import { APIErrorSchema } from "./contracts";
import {
  DecodeError,
  HttpError,
  TransportError,
  type ApiFailure,
} from "./errors";

export type FetchLike = typeof fetch;

export type JsonRequest = {
  readonly url: string;
  readonly method?: string;
  readonly headers?: Readonly<Record<string, string>>;
  readonly body?: unknown;
  readonly cache?: RequestCache;
};

const decode = <A, I>(schema: Schema.Schema<A, I>, value: unknown) =>
  Schema.decodeUnknown(schema)(value).pipe(
    Effect.mapError(
      (cause) =>
        new DecodeError({ message: "API response validation failed", cause }),
    ),
  );

const readJson = (response: Response) =>
  Effect.tryPromise({
    try: () => response.json() as Promise<unknown>,
    catch: (cause) =>
      new DecodeError({ message: "API response was not valid JSON", cause }),
  });

const httpFailure = (response: Response) =>
  readJson(response).pipe(
    Effect.flatMap((body) => decode(APIErrorSchema, body)),
    Effect.map(
      (envelope) =>
        new HttpError({
          status: response.status,
          code: envelope.code,
          message: envelope.message,
          ...(envelope.request_id === undefined
            ? {}
            : { requestId: envelope.request_id }),
        }),
    ),
    Effect.catchAll(() =>
      Effect.succeed(
        new HttpError({
          status: response.status,
          code: "UNKNOWN",
          message: response.statusText || `HTTP ${response.status}`,
        }),
      ),
    ),
  );

export const requestJson = <A, I>(
  fetchImpl: FetchLike,
  request: JsonRequest,
  schema: Schema.Schema<A, I>,
): Effect.Effect<A, ApiFailure> =>
  Effect.tryPromise({
    try: (signal) =>
      fetchImpl(request.url, {
        method: request.method ?? "GET",
        headers: request.headers,
        body:
          request.body === undefined ? undefined : JSON.stringify(request.body),
        cache: request.cache,
        signal,
      }),
    catch: (cause) => new TransportError({ cause }),
  }).pipe(
    Effect.flatMap((response): Effect.Effect<A, ApiFailure> => {
      if (response.ok) {
        return readJson(response).pipe(
          Effect.flatMap((body) => decode(schema, body)),
        );
      }
      return httpFailure(response).pipe(
        Effect.flatMap((error): Effect.Effect<never, HttpError> =>
          Effect.fail(error),
        ),
      );
    }),
  );
