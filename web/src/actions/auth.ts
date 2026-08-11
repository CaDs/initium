"use server";

import { MagicLinkInputSchema, MessageResponseSchema } from "@initium/core";
import { Either, Schema } from "effect";

import { apiFetch } from "@/lib/api";

export type MagicLinkState = { ok: boolean; message: string };

export async function requestMagicLink(
  _prev: MagicLinkState,
  formData: FormData,
): Promise<MagicLinkState> {
  const email = formData.get("email");

  if (typeof email !== "string" || !email) {
    return { ok: false, message: "Email is required." };
  }

  const input = Schema.decodeUnknownEither(MagicLinkInputSchema)({ email });
  if (Either.isLeft(input)) {
    return { ok: false, message: "Enter a valid email address." };
  }

  const result = await apiFetch(
    "/api/auth/magic-link",
    { method: "POST", body: input.right },
    MessageResponseSchema,
  );

  if (!result.ok) {
    return { ok: false, message: result.error.message };
  }

  return { ok: true, message: result.data.message };
}
