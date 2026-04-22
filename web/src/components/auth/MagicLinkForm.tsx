"use client";

import { useActionState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { requestMagicLink, type MagicLinkState } from "@/actions/auth";

const initialState: MagicLinkState = { ok: false, message: "" };

export default function MagicLinkForm() {
  const t = useTranslations("login.magicLink");
  const [state, action, isPending] = useActionState(requestMagicLink, initialState);

  useEffect(() => {
    if (!state.message) return;
    if (state.ok) {
      toast.success(t("sentTitle"), { description: t("sentBody") });
    } else {
      toast.error(state.message);
    }
  }, [state, t]);

  const hasError = !state.ok && !!state.message;

  if (state.ok) {
    return (
      <p role="status" aria-live="polite">
        {t("sent")}
      </p>
    );
  }

  return (
    <form action={action}>
      <label htmlFor="magic-link-email" className="sr-only">
        {t("placeholder")}
      </label>
      <input
        id="magic-link-email"
        name="email"
        type="email"
        placeholder={t("placeholder")}
        required
        aria-describedby={hasError ? "magic-link-error" : undefined}
        aria-invalid={hasError}
      />
      <button type="submit" disabled={isPending} aria-busy={isPending}>
        {isPending ? t("sending") : t("submit")}
      </button>
      {hasError && (
        <p id="magic-link-error" role="alert" aria-live="polite">
          {state.message}
        </p>
      )}
    </form>
  );
}
