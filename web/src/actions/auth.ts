"use server";

import { apiFetch } from "@/lib/api";

export async function requestMagicLink(email: string): Promise<{ ok: boolean; message: string }> {
  const result = await apiFetch<{ message: string }>("/api/auth/magic-link", {
    method: "POST",
    body: JSON.stringify({ email }),
  });

  if (!result.ok) {
    return { ok: false, message: result.error.message };
  }

  return { ok: true, message: result.data.message };
}
