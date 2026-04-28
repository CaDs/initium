import { useAuthStore } from '@/auth/store';

const json = (status: number, body: unknown): Response =>
  ({
    ok: status >= 200 && status < 300,
    status,
    statusText: 'STATUS',
    json: async () => body,
  }) as unknown as Response;

describe('auth store', () => {
  beforeEach(async () => {
    await useAuthStore.getState().tokens.clear();
    useAuthStore.setState({ status: 'loading', user: null, error: null });
  });

  it('bootstrap with no stored token transitions to unauthenticated', async () => {
    await useAuthStore.getState().bootstrap();
    expect(useAuthStore.getState().status).toBe('unauthenticated');
    expect(useAuthStore.getState().user).toBeNull();
  });

  it('bootstrap with stored token + valid /api/me transitions to authenticated', async () => {
    await useAuthStore.getState().tokens.write({
      access_token: 'AT',
      refresh_token: 'RT',
    });
    const fetchSpy = jest.spyOn(global, 'fetch').mockImplementation(async () =>
      json(200, { id: 'u1', email: 'a@b.c', name: 'A', role: 'user' }),
    );

    await useAuthStore.getState().bootstrap();

    expect(useAuthStore.getState().status).toBe('authenticated');
    expect(useAuthStore.getState().user?.id).toBe('u1');
    fetchSpy.mockRestore();
  });

  it('logout clears tokens and transitions to unauthenticated', async () => {
    await useAuthStore.getState().tokens.write({
      access_token: 'AT',
      refresh_token: 'RT',
    });
    useAuthStore.setState({
      status: 'authenticated',
      user: { id: 'u1', email: 'a@b.c', name: 'A', role: 'user' },
    });
    const fetchSpy = jest.spyOn(global, 'fetch').mockImplementation(async () =>
      json(200, { message: 'ok' }),
    );

    await useAuthStore.getState().logout();

    expect(useAuthStore.getState().status).toBe('unauthenticated');
    expect(useAuthStore.getState().user).toBeNull();
    expect(await useAuthStore.getState().tokens.read()).toBeNull();
    fetchSpy.mockRestore();
  });

  it('verifyMagicLink stores tokens and fetches the user', async () => {
    const fetchSpy = jest.spyOn(global, 'fetch').mockImplementation(async (url) => {
      if (String(url).endsWith('/api/auth/mobile/verify')) {
        return json(200, { access_token: 'AT', refresh_token: 'RT' });
      }
      return json(200, { id: 'u1', email: 'a@b.c', name: 'A', role: 'user' });
    });

    await useAuthStore.getState().verifyMagicLink('magic-token');

    expect(useAuthStore.getState().status).toBe('authenticated');
    expect(await useAuthStore.getState().tokens.read()).toEqual({
      access_token: 'AT',
      refresh_token: 'RT',
    });
    fetchSpy.mockRestore();
  });
});
