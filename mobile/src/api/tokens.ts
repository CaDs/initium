import * as SecureStore from 'expo-secure-store';
import { TokenStorageError, type TokenPair } from '@initium/core';
import { Effect } from 'effect';

const ACCESS_KEY = 'initium.access_token';
const REFRESH_KEY = 'initium.refresh_token';

export type TokenStorage = {
  read: Effect.Effect<TokenPair | null, TokenStorageError>;
  write(pair: TokenPair): Effect.Effect<void, TokenStorageError>;
  clear: Effect.Effect<void, TokenStorageError>;
};

export const secureTokenStorage: TokenStorage = {
  read: Effect.tryPromise({
    try: async () => {
      const [access, refresh] = await Promise.all([
        SecureStore.getItemAsync(ACCESS_KEY),
        SecureStore.getItemAsync(REFRESH_KEY),
      ]);
      if (!access || !refresh) return null;
      return { access_token: access, refresh_token: refresh };
    },
    catch: (cause) => new TokenStorageError({ operation: 'read', cause }),
  }),
  write: (pair) =>
    Effect.tryPromise({
      try: async () => {
        await Promise.all([
          SecureStore.setItemAsync(ACCESS_KEY, pair.access_token),
          SecureStore.setItemAsync(REFRESH_KEY, pair.refresh_token),
        ]);
      },
      catch: (cause) => new TokenStorageError({ operation: 'write', cause }),
    }),
  clear: Effect.tryPromise({
    try: async () => {
      await Promise.all([
        SecureStore.deleteItemAsync(ACCESS_KEY),
        SecureStore.deleteItemAsync(REFRESH_KEY),
      ]);
    },
    catch: (cause) => new TokenStorageError({ operation: 'clear', cause }),
  }),
};
