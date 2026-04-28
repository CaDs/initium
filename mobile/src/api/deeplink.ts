export type VerifyDeepLink = { kind: 'verify'; token: string };

export type DeepLink = VerifyDeepLink | { kind: 'unknown' };

export const parseDeepLink = (url: string): DeepLink => {
  try {
    const normalized = url.replace(/^initium:\/\//, 'http://x/');
    const u = new URL(normalized);
    if (u.pathname === '/auth/verify') {
      const token = u.searchParams.get('token');
      if (token) return { kind: 'verify', token };
    }
  } catch {
    // fall through
  }
  return { kind: 'unknown' };
};
