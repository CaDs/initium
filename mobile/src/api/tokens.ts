import * as SecureStore from 'expo-secure-store';

import type { TokenPair } from './models';

const ACCESS_KEY = 'initium.access_token';
const REFRESH_KEY = 'initium.refresh_token';

export type TokenStorage = {
  read(): Promise<TokenPair | null>;
  write(pair: TokenPair): Promise<void>;
  clear(): Promise<void>;
};

export const secureTokenStorage: TokenStorage = {
  async read() {
    const [access, refresh] = await Promise.all([
      SecureStore.getItemAsync(ACCESS_KEY),
      SecureStore.getItemAsync(REFRESH_KEY),
    ]);
    if (!access || !refresh) return null;
    return { access_token: access, refresh_token: refresh };
  },
  async write(pair) {
    await Promise.all([
      SecureStore.setItemAsync(ACCESS_KEY, pair.access_token),
      SecureStore.setItemAsync(REFRESH_KEY, pair.refresh_token),
    ]);
  },
  async clear() {
    await Promise.all([
      SecureStore.deleteItemAsync(ACCESS_KEY),
      SecureStore.deleteItemAsync(REFRESH_KEY),
    ]);
  },
};
