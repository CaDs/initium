export const COOKIE_SECURE = process.env.NODE_ENV === "production";
export const ACCESS_TOKEN_MAX_AGE = 900; // 15 min
export const REFRESH_TOKEN_MAX_AGE = 604800; // 7 days
export const REFRESH_TOKEN_PATH = "/api/auth";
