export const COOKIE_SECURE = process.env.NODE_ENV === "production";
// 23h — one hour SHORTER than the 24h access JWT, so the cookie always dies
// first and expiry routes through middleware.ts (which can persist the
// rotated pair). A render-context refresh cannot set cookies and would burn
// the single-use refresh token.
export const ACCESS_TOKEN_MAX_AGE = 82800;
export const REFRESH_TOKEN_MAX_AGE = 604800; // 7 days
// Path "/" so the middleware sees the refresh cookie on page routes; scoped
// to /api/auth the browser never sends it there and sessions die at access
// expiry.
export const REFRESH_TOKEN_PATH = "/";
