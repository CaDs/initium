import { Effect, Schema } from 'effect';

import { APIClient } from '@/api/client';
import type { TokenStorage } from '@/api/tokens';
import { bootstrapAuth, logoutAuth, verifyMagicLinkAuth } from '@/auth/store';
import type { TokenPair } from '@initium/core';

const json = (status: number, body: unknown): Response =>
  ({
    ok: status >= 200 && status < 300,
    status,
    statusText: 'STATUS',
    json: async () => body,
  }) as Response;

const memoryTokens = (initial: TokenPair | null = null) => {
  let pair = initial;
  const storage: TokenStorage = {
    read: Effect.sync(() => pair),
    write: (next) => Effect.sync(() => void (pair = next)),
    clear: Effect.sync(() => void (pair = null)),
  };
  return { storage, current: () => pair };
};

const user = {
  id: 'u1',
  email: 'user@example.com',
  name: 'User',
  role: 'user',
  avatar_url: '',
  created_at: '2026-01-01T00:00:00Z',
};

describe('Effect auth workflows', () => {
  it('bootstraps to unauthenticated without stored tokens', async () => {
    const tokens = memoryTokens();
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: tokens.storage,
      fetchImpl: jest.fn(),
    });

    const state = await Effect.runPromise(bootstrapAuth(client, tokens.storage, false));

    expect(state).toEqual({ _tag: 'Unauthenticated' });
  });

  it('bootstraps to authenticated after decoding /api/me', async () => {
    const tokens = memoryTokens({ access_token: 'AT', refresh_token: 'RT' });
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: tokens.storage,
      fetchImpl: async () => json(200, user),
    });

    const state = await Effect.runPromise(bootstrapAuth(client, tokens.storage, false));

    expect(state._tag).toBe('Authenticated');
  });

  it('verifies a magic link, stores tokens, and fetches the user', async () => {
    const tokens = memoryTokens();
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: tokens.storage,
      fetchImpl: async (url) =>
        String(url).endsWith('/api/auth/mobile/verify')
          ? json(200, { access_token: 'AT', refresh_token: 'RT' })
          : json(200, user),
    });

    const state = await Effect.runPromise(
      verifyMagicLinkAuth('magic-token', client, tokens.storage),
    );

    expect(state._tag).toBe('Authenticated');
    expect(tokens.current()).toEqual({ access_token: 'AT', refresh_token: 'RT' });
  });

  it('clears local tokens even when server logout fails', async () => {
    const tokens = memoryTokens({ access_token: 'AT', refresh_token: 'RT' });
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: tokens.storage,
      fetchImpl: async () => json(500, { code: 'SERVER_ERROR', message: 'boom' }),
    });

    const state = await Effect.runPromise(logoutAuth(client, tokens.storage));

    expect(state).toEqual({ _tag: 'Unauthenticated' });
    expect(tokens.current()).toBeNull();
  });

  it('keeps schema decoding active in workflow tests', async () => {
    await expect(
      Effect.runPromise(Schema.decodeUnknown(Schema.Struct({ id: Schema.String }))({})),
    ).rejects.toBeDefined();
  });
});
