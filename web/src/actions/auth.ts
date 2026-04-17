"use server";

import { apiFetch } from "@/lib/api";

export type MagicLinkState = { ok: boolean; message: string };

export async function requestMagicLink(
  _prev: MagicLinkState,
  formData: FormData
): Promise<MagicLinkState> {
  const email = formData.get("email");

  if (typeof email !== "string" || !email) {
    return { ok: false, message: "Email is required." };
  }

  const result = await apiFetch<{ message: string }>("/api/auth/magic-link", {
    method: "POST",
    body: JSON.stringify({ email }),
  });

  if (!result.ok) {
    return { ok: false, message: result.error.message };
  }

  return { ok: true, message: result.data.message };
}
