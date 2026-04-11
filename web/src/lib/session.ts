import { cookies } from "next/headers";

const COOKIE_OPTIONS = {
  httpOnly: true,
  secure: process.env.NODE_ENV === "production",
  sameSite: "lax" as const,
  path: "/",
};

export async function setTokenCookies(accessToken: string, refreshToken: string) {
  const cookieStore = await cookies();

  cookieStore.set("access_token", accessToken, {
    ...COOKIE_OPTIONS,
    maxAge: 900, // 15 min
  });

  cookieStore.set("refresh_token", refreshToken, {
    ...COOKIE_OPTIONS,
    path: "/api/auth",
    maxAge: 604800, // 7 days
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
