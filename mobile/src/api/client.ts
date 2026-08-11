import {
  HttpError,
  TokenPairSchema,
  UnauthorizedError,
  requestJson,
  type ApiFailure,
  type FetchLike,
  type TokenPair,
  type TokenStorageError,
} from '@initium/core';
import { Effect, type Schema } from 'effect';

import type { TokenStorage } from './tokens';

type SendOptions = {
  method?: string;
  body?: unknown;
  skipAuth?: boolean;
};

export type MobileApiFailure = ApiFailure | UnauthorizedError | TokenStorageError;

export type APIClientOptions = {
  baseURL: string;
  tokens: TokenStorage;
  fetchImpl?: FetchLike;
};

export class APIClient {
  private readonly baseURL: string;
  private readonly tokens: TokenStorage;
  private readonly fetchImpl: FetchLike;
  private readonly refreshSemaphore = Effect.runSync(Effect.makeSemaphore(1));

  constructor(opts: APIClientOptions) {
    this.baseURL = opts.baseURL.replace(/\/$/, '');
    this.tokens = opts.tokens;
    // Defer the global lookup so tests that spy on `global.fetch` after
    // construction still intercept calls.
    this.fetchImpl = opts.fetchImpl ?? ((input, init) => fetch(input, init));
  }

  send<A, I>(
    path: string,
    schema: Schema.Schema<A, I>,
    opts: SendOptions = {},
  ): Effect.Effect<A, MobileApiFailure> {
    return this.tokens.read.pipe(
      Effect.flatMap((stored) => {
        const accessToken = stored?.access_token ?? null;
        return this.attempt(path, schema, opts, accessToken).pipe(
          Effect.catchAll((error): Effect.Effect<A, MobileApiFailure> => {
            if (!(error instanceof HttpError) || error.status !== 401 || opts.skipAuth) {
              return Effect.fail(error);
            }
            return this.refresh(accessToken).pipe(
              Effect.flatMap((pair) =>
                this.attempt(path, schema, opts, pair.access_token).pipe(
                  Effect.catchAll((retryError): Effect.Effect<A, MobileApiFailure> =>
                    retryError instanceof HttpError && retryError.status === 401
                      ? this.clearAndUnauthorized()
                      : Effect.fail(retryError),
                  ),
                ),
              ),
            );
          }),
        );
      }),
    );
  }

  private attempt<A, I>(
    path: string,
    schema: Schema.Schema<A, I>,
    opts: SendOptions,
    accessToken: string | null,
  ): Effect.Effect<A, ApiFailure> {
    const headers: Record<string, string> = { Accept: 'application/json' };
    if (opts.body !== undefined) headers['Content-Type'] = 'application/json';
    if (accessToken && !opts.skipAuth) headers.Authorization = `Bearer ${accessToken}`;

    return requestJson(
      this.fetchImpl,
      {
        url: `${this.baseURL}${path}`,
        method: opts.method,
        headers,
        body: opts.body,
      },
      schema,
    );
  }

  private refresh(failedAccessToken: string | null): Effect.Effect<TokenPair, MobileApiFailure> {
    return this.refreshSemaphore.withPermits(1)(
      this.tokens.read.pipe(
        Effect.flatMap((current) => {
          if (!current) return this.clearAndUnauthorized();
          if (current.access_token !== failedAccessToken) return Effect.succeed(current);

          return requestJson(
            this.fetchImpl,
            {
              url: `${this.baseURL}/api/auth/refresh`,
              method: 'POST',
              headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
              body: { refresh_token: current.refresh_token },
            },
            TokenPairSchema,
          ).pipe(
            Effect.flatMap((pair) => this.tokens.write(pair).pipe(Effect.as(pair))),
            Effect.catchAll(() => this.clearAndUnauthorized()),
          );
        }),
      ),
    );
  }

  private clearAndUnauthorized(): Effect.Effect<never, UnauthorizedError | TokenStorageError> {
    return this.tokens.clear.pipe(
      Effect.zipRight(Effect.fail(new UnauthorizedError({ message: 'Your session has expired.' }))),
    );
  }
}
