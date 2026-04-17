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
      <div
        className="text-center p-4 bg-success-bg rounded-lg"
        role="status"
        aria-live="polite"
      >
        <p className="text-success font-medium">{t("sent")}</p>
        <p className="text-success/80 text-sm mt-1">{t("sentDetail")}</p>
      </div>
    );
  }

  return (
    <form action={action} className="space-y-3">
      <div>
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
          className="w-full border border-border bg-card rounded-lg px-4 py-3 placeholder-muted focus:outline-none focus:ring-2 focus:ring-accent"
        />
      </div>
      <button
        type="submit"
        disabled={isPending}
        aria-busy={isPending}
        className="w-full bg-accent text-accent-foreground rounded-lg px-4 py-3 font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
      >
        {isPending ? t("sending") : t("submit")}
      </button>
      {hasError && (
        <p id="magic-link-error" className="text-error text-sm" role="alert" aria-live="polite">
          {state.message}
        </p>
      )}
    </form>
  );
}
