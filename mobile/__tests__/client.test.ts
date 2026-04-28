import { APIClient, APIError, UnauthorizedError } from '@/api/client';
import type { TokenStorage } from '@/api/tokens';

type Call = { url: string; init: RequestInit };

const memoryTokens = (initial: { access_token: string; refresh_token: string } | null = null): TokenStorage => {
  let pair = initial;
  return {
    async read() {
      return pair;
    },
    async write(next) {
      pair = next;
    },
    async clear() {
      pair = null;
    },
  };
};

const jsonResponse = (status: number, body: unknown): Response =>
  ({
    ok: status >= 200 && status < 300,
    status,
    statusText: 'STATUS',
    json: async () => body,
  }) as unknown as Response;

const noContentResponse = (): Response =>
  ({
    ok: true,
    status: 204,
    statusText: 'No Content',
    json: async () => undefined,
  }) as unknown as Response;

describe('APIClient.send', () => {
  it('attaches the Bearer header when a token is stored', async () => {
    const calls: Call[] = [];
    const fetchImpl: typeof fetch = async (url, init) => {
      calls.push({ url: String(url), init: init as RequestInit });
      return jsonResponse(200, { ok: true });
    };
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: memoryTokens({ access_token: 'AT', refresh_token: 'RT' }),
      fetchImpl,
    });

    await client.send('/api/me');

    const headers = calls[0]!.init.headers as Record<string, string>;
    expect(headers.Authorization).toBe('Bearer AT');
  });

  it('omits the Bearer header when skipAuth is set', async () => {
    const calls: Call[] = [];
    const fetchImpl: typeof fetch = async (url, init) => {
      calls.push({ url: String(url), init: init as RequestInit });
      return jsonResponse(200, { ok: true });
    };
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: memoryTokens({ access_token: 'AT', refresh_token: 'RT' }),
      fetchImpl,
    });

    await client.send('/api/landing', { skipAuth: true });

    const headers = calls[0]!.init.headers as Record<string, string>;
    expect(headers.Authorization).toBeUndefined();
  });

  it('refreshes once on 401 and retries the request with the new token', async () => {
    const tokens = memoryTokens({ access_token: 'OLD', refresh_token: 'RT' });
    const calls: Call[] = [];
    const fetchImpl: typeof fetch = async (url, init) => {
      const u = String(url);
      const headers = (init?.headers ?? {}) as Record<string, string>;
      calls.push({ url: u, init: init as RequestInit });
      if (u.endsWith('/api/auth/refresh')) {
        return jsonResponse(200, { access_token: 'NEW', refresh_token: 'RT2' });
      }
      if (headers.Authorization === 'Bearer OLD') {
        return jsonResponse(401, { code: 'UNAUTHORIZED', message: 'expired' });
      }
      return jsonResponse(200, { id: 'u', email: 'e', name: 'n', role: 'user' });
    };
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens,
      fetchImpl,
    });

    const me = await client.send<{ id: string }>('/api/me');

    expect(me.id).toBe('u');
    expect(calls.map((c) => c.url)).toEqual([
      'https://api.example.com/api/me',
      'https://api.example.com/api/auth/refresh',
      'https://api.example.com/api/me',
    ]);
    expect(await tokens.read()).toEqual({ access_token: 'NEW', refresh_token: 'RT2' });
  });

  it('coalesces concurrent 401s into a single refresh request (single-flight)', async () => {
    const tokens = memoryTokens({ access_token: 'OLD', refresh_token: 'RT' });
    const calls: Call[] = [];
    let resolveRefresh: (value: Response) => void;
    const refreshPromise = new Promise<Response>((resolve) => {
      resolveRefresh = resolve;
    });

    const fetchImpl: typeof fetch = async (url, init) => {
      const u = String(url);
      const headers = (init?.headers ?? {}) as Record<string, string>;
      calls.push({ url: u, init: init as RequestInit });
      if (u.endsWith('/api/auth/refresh')) {
        return refreshPromise;
      }
      if (headers.Authorization === 'Bearer OLD') {
        return jsonResponse(401, { code: 'UNAUTHORIZED', message: 'expired' });
      }
      return jsonResponse(200, { ok: true });
    };
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens,
      fetchImpl,
    });

    const r1 = client.send('/api/me');
    const r2 = client.send('/api/me');

    // Give both calls a chance to enqueue refresh
    await new Promise((r) => setImmediate(r));

    resolveRefresh!(jsonResponse(200, { access_token: 'NEW', refresh_token: 'RT2' }));

    await Promise.all([r1, r2]);

    const refreshCalls = calls.filter((c) => c.url.endsWith('/api/auth/refresh'));
    expect(refreshCalls).toHaveLength(1);
  });

  it('throws UnauthorizedError when refresh fails', async () => {
    const tokens = memoryTokens({ access_token: 'OLD', refresh_token: 'RT' });
    let unauthorizedFired = false;
    const fetchImpl: typeof fetch = async (url) => {
      const u = String(url);
      if (u.endsWith('/api/auth/refresh')) {
        return jsonResponse(401, { code: 'UNAUTHORIZED', message: 'revoked' });
      }
      return jsonResponse(401, { code: 'UNAUTHORIZED', message: 'expired' });
    };
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens,
      fetchImpl,
      onUnauthorized: () => {
        unauthorizedFired = true;
      },
    });

    await expect(client.send('/api/me')).rejects.toBeInstanceOf(UnauthorizedError);
    expect(unauthorizedFired).toBe(true);
    expect(await tokens.read()).toBeNull();
  });

  it('throws APIError on non-401 errors without retrying', async () => {
    const calls: Call[] = [];
    const fetchImpl: typeof fetch = async (url, init) => {
      calls.push({ url: String(url), init: init as RequestInit });
      return jsonResponse(500, { code: 'SERVER_ERROR', message: 'boom' });
    };
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: memoryTokens({ access_token: 'AT', refresh_token: 'RT' }),
      fetchImpl,
    });

    await expect(client.send('/api/me')).rejects.toBeInstanceOf(APIError);
    expect(calls).toHaveLength(1);
  });

  it('returns undefined for 204 responses', async () => {
    const fetchImpl: typeof fetch = async () => noContentResponse();
    const client = new APIClient({
      baseURL: 'https://api.example.com',
      tokens: memoryTokens({ access_token: 'AT', refresh_token: 'RT' }),
      fetchImpl,
    });

    const result = await client.send('/api/something');
    expect(result).toBeUndefined();
  });
});
