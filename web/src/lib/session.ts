import { cookies } from "next/headers";
import {
  COOKIE_SECURE,
  ACCESS_TOKEN_MAX_AGE,
  REFRESH_TOKEN_MAX_AGE,
  REFRESH_TOKEN_PATH,
} from "./cookie-config";

const COOKIE_OPTIONS = {
  httpOnly: true,
  secure: COOKIE_SECURE,
  sameSite: "lax" as const,
  path: "/",
};

export async function setTokenCookies(accessToken: string, refreshToken: string) {
  const cookieStore = await cookies();

  cookieStore.set("access_token", accessToken, {
    ...COOKIE_OPTIONS,
    maxAge: ACCESS_TOKEN_MAX_AGE,
  });

  cookieStore.set("refresh_token", refreshToken, {
    ...COOKIE_OPTIONS,
    path: REFRESH_TOKEN_PATH,
    maxAge: REFRESH_TOKEN_MAX_AGE,
  });
}

export async function clearTokenCookies() {
  const cookieStore = await cookies();
  cookieStore.delete("access_token");
  cookieStore.delete("refresh_token");
}

export async function getAccessToken(): Promise<string | undefined> {
  const cookieStore = await cookies();
  return cookieStore.get("access_token")?.value;
}

export async function hasSession(): Promise<boolean> {
  const token = await getAccessToken();
  return !!token;
}
