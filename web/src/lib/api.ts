import { cookies } from "next/headers";
import type { ApiResult, ApiError } from "./types";

const API_URL = process.env.API_URL || "http://localhost:8000";

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<ApiResult<T>> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get("access_token")?.value;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...((options.headers as Record<string, string>) || {}),
  };

  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
    headers["Cookie"] = `access_token=${accessToken}`;
  }

  let res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
    cache: "no-store",
  });

  // Token refresh on 401
  if (res.status === 401) {
    const refreshToken = cookieStore.get("refresh_token")?.value;
    if (refreshToken) {
      const refreshRes = await fetch(`${API_URL}/api/auth/refresh`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Cookie: `refresh_token=${refreshToken}`,
        },
        cache: "no-store",
      });

      if (refreshRes.ok) {
        const tokens = await refreshRes.json();
        // Set new cookies for subsequent requests
        const newHeaders = { ...headers };
        newHeaders["Authorization"] = `Bearer ${tokens.access_token}`;
        newHeaders["Cookie"] = `access_token=${tokens.access_token}`;

        // Retry original request with new token
        res = await fetch(`${API_URL}${path}`, {
          ...options,
          headers: newHeaders,
          cache: "no-store",
        });
      }
    }
  }

  if (!res.ok) {
    const error: ApiError = await res.json().catch(() => ({
      code: "UNKNOWN",
      message: res.statusText,
    }));
    return { ok: false, error };
  }

  const data = await res.json();
  return { ok: true, data: data as T };
}
