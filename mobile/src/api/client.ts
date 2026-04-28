import type { ErrorEnvelope, TokenPair } from './models';
import type { TokenStorage } from './tokens';

export class APIError extends Error {
  readonly status: number;
  readonly code: string;
  readonly requestId?: string;

  constructor(status: number, envelope: ErrorEnvelope) {
    super(envelope.message);
    this.name = 'APIError';
    this.status = status;
    this.code = envelope.code;
    this.requestId = envelope.request_id;
  }
}

export class UnauthorizedError extends Error {
  constructor() {
    super('unauthorized');
    this.name = 'UnauthorizedError';
  }
}

type FetchLike = typeof fetch;

type SendOptions = {
  method?: string;
  body?: unknown;
  skipAuth?: boolean;
};

export type APIClientOptions = {
  baseURL: string;
  tokens: TokenStorage;
  onUnauthorized?: () => void;
  fetchImpl?: FetchLike;
};

export class APIClient {
  private readonly baseURL: string;
  private readonly tokens: TokenStorage;
  private readonly fetchImpl: FetchLike;
  private readonly onUnauthorized?: () => void;
  private refreshInFlight: Promise<TokenPair | null> | null = null;

  constructor(opts: APIClientOptions) {
    this.baseURL = opts.baseURL.replace(/\/$/, '');
    this.tokens = opts.tokens;
    this.onUnauthorized = opts.onUnauthorized;
    // Defer the global lookup so tests that spy on `global.fetch` after
    // construction still intercept calls.
    this.fetchImpl = opts.fetchImpl ?? ((input, init) => fetch(input, init));
  }

  async send<T>(path: string, opts: SendOptions = {}): Promise<T> {
    const stored = await this.tokens.read();
    const accessToken = stored?.access_token ?? null;

    const first = await this.attempt<T>(path, opts, accessToken);
    if (first.kind === 'ok') return first.value;
    if (first.status !== 401 || opts.skipAuth) {
      throw first.error;
    }

    const refreshed = await this.runRefresh();
    if (!refreshed) {
      this.onUnauthorized?.();
      throw new UnauthorizedError();
    }

    const retried = await this.attempt<T>(path, opts, refreshed.access_token);
    if (retried.kind === 'ok') return retried.value;
    if (retried.status === 401) {
      this.onUnauthorized?.();
      throw new UnauthorizedError();
    }
    throw retried.error;
  }

  private async attempt<T>(
    path: string,
    opts: SendOptions,
    accessToken: string | null,
  ): Promise<{ kind: 'ok'; value: T } | { kind: 'err'; status: number; error: Error }> {
    const headers: Record<string, string> = { Accept: 'application/json' };
    if (opts.body !== undefined) headers['Content-Type'] = 'application/json';
    if (accessToken && !opts.skipAuth) headers.Authorization = `Bearer ${accessToken}`;

    const response = await this.fetchImpl(`${this.baseURL}${path}`, {
      method: opts.method ?? 'GET',
      headers,
      body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
    });

    if (response.status === 204) {
      return { kind: 'ok', value: undefined as T };
    }

    if (response.ok) {
      const value = (await response.json()) as T;
      return { kind: 'ok', value };
    }

    let envelope: ErrorEnvelope = { code: 'UNKNOWN', message: response.statusText };
    try {
      envelope = (await response.json()) as ErrorEnvelope;
    } catch {
      // body was not JSON; keep the synthetic envelope
    }

    return {
      kind: 'err',
      status: response.status,
      error: new APIError(response.status, envelope),
    };
  }

  private runRefresh(): Promise<TokenPair | null> {
    if (this.refreshInFlight) return this.refreshInFlight;

    this.refreshInFlight = (async (): Promise<TokenPair | null> => {
      const current = await this.tokens.read();
      if (!current) return null;
      try {
        const response = await this.fetchImpl(`${this.baseURL}/api/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
          body: JSON.stringify({ refresh_token: current.refresh_token }),
        });
        if (!response.ok) {
          await this.tokens.clear();
          return null;
        }
        const pair = (await response.json()) as TokenPair;
        await this.tokens.write(pair);
        return pair;
      } catch {
        return null;
      }
    })();

    const settled = this.refreshInFlight;
    settled.finally(() => {
      if (this.refreshInFlight === settled) this.refreshInFlight = null;
    });
    return settled;
  }
}
