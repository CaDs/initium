jest.mock('expo-secure-store', () => {
  const memory = new Map<string, string>();
  return {
    setItemAsync: jest.fn(async (k: string, v: string) => {
      memory.set(k, v);
    }),
    getItemAsync: jest.fn(async (k: string) => memory.get(k) ?? null),
    deleteItemAsync: jest.fn(async (k: string) => {
      memory.delete(k);
    }),
    __reset: () => memory.clear(),
  };
});

jest.mock('expo-constants', () => ({
  expoConfig: { extra: {} },
}));

jest.mock('expo-linking', () => ({
  createURL: (path: string) => `initium://${path}`,
  parse: (url: string) => {
    const u = new URL(url.replace('initium://', 'http://x/'));
    const queryParams: Record<string, string> = {};
    u.searchParams.forEach((v, k) => {
      queryParams[k] = v;
    });
    return { path: u.pathname.replace(/^\//, ''), queryParams };
  },
}));
