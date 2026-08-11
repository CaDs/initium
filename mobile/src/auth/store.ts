import * as Atom from '@effect-atom/atom/Atom';
import {
  UnauthorizedError,
  failureMessage,
  type TokenPair,
  type TokenStorageError,
  type User,
} from '@initium/core';
import { Effect } from 'effect';

import { APIClient, type MobileApiFailure } from '../api/client';
import {
  fetchMe,
  logout as logoutCall,
  requestMagicLink,
  verifyGoogleIDToken,
  verifyMagicLink,
} from '../api/endpoints';
import { secureTokenStorage, type TokenStorage } from '../api/tokens';
import { config } from '../config';

export type AuthFailure = MobileApiFailure | TokenStorageError;

export type AuthState =
  | { readonly _tag: 'Loading' }
  | { readonly _tag: 'Authenticated'; readonly user: User }
  | { readonly _tag: 'Unauthenticated' }
  | { readonly _tag: 'Failed'; readonly error: AuthFailure };

const DEV_USER: User = {
  id: 'dev-bypass',
  email: 'dev@initium.local',
  name: 'Dev User',
  role: 'admin',
  avatar_url: '',
  created_at: new Date(0).toISOString(),
};

export const authStateAtom = Atom.make<AuthState>({ _tag: 'Loading' }).pipe(Atom.keepAlive);

const tokens = secureTokenStorage;
const client = new APIClient({ baseURL: config.apiBaseURL, tokens });

export const bootstrapAuth = (
  apiClient: APIClient = client,
  tokenStorage: TokenStorage = tokens,
  devBypass = config.devBypassAuth,
): Effect.Effect<AuthState, never> => {
  if (devBypass) return Effect.succeed({ _tag: 'Authenticated', user: DEV_USER });

  return tokenStorage.read.pipe(
    Effect.flatMap((stored) =>
      stored
        ? fetchMe(apiClient).pipe(
            Effect.map((user): AuthState => ({ _tag: 'Authenticated', user })),
          )
        : Effect.succeed<AuthState>({ _tag: 'Unauthenticated' }),
    ),
    Effect.catchAll((error) => {
      if (error instanceof UnauthorizedError) {
        return tokenStorage.clear.pipe(
          Effect.as<AuthState>({ _tag: 'Unauthenticated' }),
          Effect.catchAll((clearError) =>
            Effect.succeed<AuthState>({ _tag: 'Failed', error: clearError }),
          ),
        );
      }
      return Effect.succeed<AuthState>({ _tag: 'Failed', error });
    }),
  );
};

const authenticate = (
  exchange: Effect.Effect<TokenPair, MobileApiFailure>,
  apiClient: APIClient = client,
  tokenStorage: TokenStorage = tokens,
): Effect.Effect<AuthState, AuthFailure> =>
  exchange.pipe(
    Effect.flatMap((pair) => tokenStorage.write(pair)),
    Effect.zipRight(fetchMe(apiClient)),
    Effect.map((user): AuthState => ({ _tag: 'Authenticated', user })),
  );

export const verifyMagicLinkAuth = (
  token: string,
  apiClient: APIClient = client,
  tokenStorage: TokenStorage = tokens,
) => authenticate(verifyMagicLink(apiClient, token), apiClient, tokenStorage);

export const verifyGoogleAuth = (
  idToken: string,
  apiClient: APIClient = client,
  tokenStorage: TokenStorage = tokens,
) => authenticate(verifyGoogleIDToken(apiClient, idToken), apiClient, tokenStorage);

export const logoutAuth = (
  apiClient: APIClient = client,
  tokenStorage: TokenStorage = tokens,
): Effect.Effect<AuthState, TokenStorageError> =>
  logoutCall(apiClient).pipe(
    Effect.ignore,
    Effect.zipRight(tokenStorage.clear),
    Effect.as<AuthState>({ _tag: 'Unauthenticated' }),
  );

const commitState = (get: Atom.FnContext, state: AuthState) =>
  Effect.sync(() => get.set(authStateAtom, state));

const commitFailure = (get: Atom.FnContext, error: AuthFailure) =>
  commitState(get, { _tag: 'Failed', error });

export const bootstrapAuthAtom = Atom.fn<void>()((_arg, get) =>
  bootstrapAuth().pipe(Effect.tap((state) => commitState(get, state))),
);

export const requestMagicLinkAtom = Atom.fn<string>()((email, get) =>
  requestMagicLink(client, email).pipe(
    Effect.tapError((error) => commitFailure(get, error)),
    Effect.asVoid,
  ),
);

export const verifyMagicLinkAtom = Atom.fn<string>()((token, get) =>
  verifyMagicLinkAuth(token).pipe(
    Effect.tap((state) => commitState(get, state)),
    Effect.tapError((error) => commitFailure(get, error)),
  ),
);

export const verifyGoogleAtom = Atom.fn<string>()((idToken, get) =>
  verifyGoogleAuth(idToken).pipe(
    Effect.tap((state) => commitState(get, state)),
    Effect.tapError((error) => commitFailure(get, error)),
  ),
);

export const logoutAtom = Atom.fn<void>()((_arg, get) =>
  logoutAuth().pipe(
    Effect.tap((state) => commitState(get, state)),
    Effect.tapError((error) => commitFailure(get, error)),
  ),
);

export const authStatus = (state: AuthState) => {
  switch (state._tag) {
    case 'Loading':
      return 'loading' as const;
    case 'Authenticated':
      return 'authenticated' as const;
    case 'Unauthenticated':
      return 'unauthenticated' as const;
    case 'Failed':
      return 'error' as const;
  }
};

export const authError = (state: AuthState): string | null =>
  state._tag === 'Failed' ? failureMessage(state.error) : null;
