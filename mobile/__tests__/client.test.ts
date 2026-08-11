import { HttpError, UnauthorizedError, UserSchema, type TokenPair } from '@initium/core';
import { Effect, Schema } from 'effect';

import { APIClient } from '@/api/client';
import type { TokenStorage } from '@/api/tokens';

type Call = { url: string; init: RequestInit };

const memoryTokens = (initial: TokenPair | null = null) => {
  let pair = initial;
  const storage: TokenStorage = {
    read: Effect.sync(() => pair),
    write: (next) => Effect.sync(() => void (pair = next)),
    clear: Effect.sync(() => void (pair = null)),
  };
  return { storage, current: () => pair };
};

const jsonResponse = (status: number, body: unknown): Response =>
  ({
    ok: status >= 200 && status < 300,
    status,
    statusText: 'STATUS',
    json: async () => body,
  }) as Response;

const user = {
  id: 'u',
  email: 'user@example.com',
  name: 'User',
  role: 'user' as const,
  avatar_url: '',
  created_at: '2026-01-01T00:00:00Z',
};

describe('Effect APIClient', () => {
  it('attaches the Bearer header and decodes the response', async () => {
    const calls: Call[] = [];
    const tokens = memoryTokens({ access_token: 'AT', refresh_token: 'RT' });
    const fetchImpl: typeof fetch = async (url, init) => {
      calls.push({ url: String(url), init: init as RequestInit });
      return jsonResponse(200, user);
    };
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: tokens.storage,
      fetchImpl,
    });

    const result = await Effect.runPromise(client.send('/api/me', UserSchema));

    expect(result.id).toBe('u');
    expect((calls[0]!.init.headers as Record<string, string>).Authorization).toBe('Bearer AT');
  });

  it('omits auth for public requests', async () => {
    const calls: Call[] = [];
    const tokens = memoryTokens({ access_token: 'AT', refresh_token: 'RT' });
    const fetchImpl: typeof fetch = async (url, init) => {
      calls.push({ url: String(url), init: init as RequestInit });
      return jsonResponse(200, { ok: true });
    };
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: tokens.storage,
      fetchImpl,
    });

    await Effect.runPromise(
      client.send('/api/landing', Schema.Struct({ ok: Schema.Boolean }), { skipAuth: true }),
    );

    expect((calls[0]!.init.headers as Record<string, string>).Authorization).toBeUndefined();
  });

  it('coalesces concurrent 401 responses into one refresh and retries both', async () => {
    const tokens = memoryTokens({ access_token: 'OLD', refresh_token: 'RT' });
    const calls: Call[] = [];
    let resolveRefresh!: (response: Response) => void;
    const refreshResponse = new Promise<Response>((resolve) => {
      resolveRefresh = resolve;
    });
    const fetchImpl: typeof fetch = async (url, init) => {
      const requestUrl = String(url);
      const headers = (init?.headers ?? {}) as Record<string, string>;
      calls.push({ url: requestUrl, init: init as RequestInit });
      if (requestUrl.endsWith('/api/auth/refresh')) return refreshResponse;
      if (headers.Authorization === 'Bearer OLD') {
        return jsonResponse(401, { code: 'UNAUTHORIZED', message: 'expired' });
      }
      return jsonResponse(200, user);
    };
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: tokens.storage,
      fetchImpl,
    });

    const first = Effect.runPromise(client.send('/api/me', UserSchema));
    const second = Effect.runPromise(client.send('/api/me', UserSchema));
    await new Promise((resolve) => setImmediate(resolve));
    resolveRefresh(jsonResponse(200, { access_token: 'NEW', refresh_token: 'RT2' }));
    await Promise.all([first, second]);

    expect(calls.filter((call) => call.url.endsWith('/api/auth/refresh'))).toHaveLength(1);
    expect(tokens.current()).toEqual({ access_token: 'NEW', refresh_token: 'RT2' });
  });

  it('clears tokens and fails with UnauthorizedError when refresh fails', async () => {
    const tokens = memoryTokens({ access_token: 'OLD', refresh_token: 'RT' });
    const fetchImpl: typeof fetch = async () =>
      jsonResponse(401, { code: 'UNAUTHORIZED', message: 'expired' });
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: tokens.storage,
      fetchImpl,
    });

    const failure = await Effect.runPromise(Effect.flip(client.send('/api/me', UserSchema)));
    expect(failure).toBeInstanceOf(UnauthorizedError);
    expect(tokens.current()).toBeNull();
  });

  it('does not retry non-401 HTTP failures', async () => {
    const tokens = memoryTokens({ access_token: 'AT', refresh_token: 'RT' });
    const fetchImpl = jest.fn(async () =>
      jsonResponse(500, { code: 'SERVER_ERROR', message: 'boom' }),
    );
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: tokens.storage,
      fetchImpl,
    });

    const failure = await Effect.runPromise(Effect.flip(client.send('/api/me', UserSchema)));
    expect(failure).toBeInstanceOf(HttpError);
    expect(fetchImpl).toHaveBeenCalledTimes(1);
  });
});
