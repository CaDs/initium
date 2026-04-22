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
      <p
        role="status"
        aria-live="polite"
        className="text-sm text-foreground"
      >
        {t("sent")}
      </p>
    );
  }

  return (
    <form action={action} className="space-y-3">
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
        className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
      />
      <button
        type="submit"
        disabled={isPending}
        aria-busy={isPending}
        className="inline-flex items-center justify-center w-full rounded-md bg-accent text-accent-foreground px-4 py-2 text-sm font-medium hover:opacity-90 disabled:opacity-50 transition-opacity"
      >
        {isPending ? t("sending") : t("submit")}
      </button>
      {hasError && (
        <p
          id="magic-link-error"
          role="alert"
          aria-live="polite"
          className="text-sm text-error"
        >
          {state.message}
        </p>
      )}
    </form>
  );
}
