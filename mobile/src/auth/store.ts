import { create } from 'zustand';

import { APIClient, UnauthorizedError } from '../api/client';
import {
  fetchMe,
  logout as logoutCall,
  requestMagicLink,
  verifyGoogleIDToken,
  verifyMagicLink,
} from '../api/endpoints';
import type { User } from '../api/models';
import { secureTokenStorage, type TokenStorage } from '../api/tokens';
import { config } from '../config';

export type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated' | 'error';

type AuthState = {
  status: AuthStatus;
  user: User | null;
  error: string | null;
  client: APIClient;
  tokens: TokenStorage;
  bootstrap(): Promise<void>;
  requestMagicLink(email: string): Promise<void>;
  verifyMagicLink(token: string): Promise<void>;
  verifyGoogle(idToken: string): Promise<void>;
  logout(): Promise<void>;
};

const DEV_USER: User = {
  id: 'dev-bypass',
  email: 'dev@initium.local',
  name: 'Dev User',
  role: 'user',
  avatar_url: null,
};

const buildClient = (tokens: TokenStorage, onUnauthorized: () => void): APIClient =>
  new APIClient({ baseURL: config.apiBaseURL, tokens, onUnauthorized });

export const useAuthStore = create<AuthState>((set, get) => {
  const tokens = secureTokenStorage;
  const onUnauthorized = () => {
    set({ status: 'unauthenticated', user: null, error: null });
  };
  const client = buildClient(tokens, onUnauthorized);

  return {
    status: 'loading',
    user: null,
    error: null,
    client,
    tokens,

    async bootstrap() {
      if (config.devBypassAuth) {
        set({ status: 'authenticated', user: DEV_USER });
        return;
      }
      const stored = await tokens.read();
      if (!stored) {
        set({ status: 'unauthenticated', user: null, error: null });
        return;
      }
      try {
        const user = await fetchMe(get().client);
        set({ status: 'authenticated', user, error: null });
      } catch (err) {
        if (err instanceof UnauthorizedError) {
          await tokens.clear();
          set({ status: 'unauthenticated', user: null, error: null });
        } else {
          set({ status: 'error', error: errorMessage(err) });
        }
      }
    },

    async requestMagicLink(email: string) {
      try {
        await requestMagicLink(get().client, email);
      } catch (err) {
        set({ status: 'error', error: errorMessage(err) });
        throw err;
      }
    },

    async verifyMagicLink(token: string) {
      try {
        const pair = await verifyMagicLink(get().client, token);
        await tokens.write(pair);
        const user = await fetchMe(get().client);
        set({ status: 'authenticated', user, error: null });
      } catch (err) {
        set({ status: 'error', error: errorMessage(err) });
        throw err;
      }
    },

    async verifyGoogle(idToken: string) {
      try {
        const pair = await verifyGoogleIDToken(get().client, idToken);
        await tokens.write(pair);
        const user = await fetchMe(get().client);
        set({ status: 'authenticated', user, error: null });
      } catch (err) {
        set({ status: 'error', error: errorMessage(err) });
        throw err;
      }
    },

    async logout() {
      try {
        await logoutCall(get().client);
      } catch {
        // best-effort: revoke server-side, but always clear local state
      }
      await tokens.clear();
      set({ status: 'unauthenticated', user: null, error: null });
    },
  };
});

const errorMessage = (err: unknown): string =>
  err instanceof Error ? err.message : String(err);
