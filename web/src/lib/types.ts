export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string;
  role: "user" | "admin";
  created_at: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
}

export type ApiError = {
  code: string;
  message: string;
  request_id?: string;
};

export type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: ApiError };
