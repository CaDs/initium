import { parseDeepLink } from '@/api/deeplink';

describe('parseDeepLink', () => {
  it('extracts the token from initium://auth/verify?token=...', () => {
    const result = parseDeepLink('initium://auth/verify?token=abc123');
    expect(result).toEqual({ kind: 'verify', token: 'abc123' });
  });

  it('returns unknown when the path is not /auth/verify', () => {
    expect(parseDeepLink('initium://settings')).toEqual({ kind: 'unknown' });
  });

  it('returns unknown when the token query param is missing', () => {
    expect(parseDeepLink('initium://auth/verify')).toEqual({ kind: 'unknown' });
  });

  it('returns unknown for malformed input', () => {
    expect(parseDeepLink('not-a-url')).toEqual({ kind: 'unknown' });
  });
});
