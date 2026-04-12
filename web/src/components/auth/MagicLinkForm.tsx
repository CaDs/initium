"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { requestMagicLink } from "@/actions/auth";

export default function MagicLinkForm() {
  const t = useTranslations("login.magicLink");
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<"idle" | "loading" | "sent" | "error">("idle");
  const [message, setMessage] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setStatus("loading");

    const result = await requestMagicLink(email);

    if (result.ok) {
      setStatus("sent");
      setMessage(t("sent"));
    } else {
      setStatus("error");
      setMessage(result.message);
    }
  }

  if (status === "sent") {
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
    <form onSubmit={handleSubmit} className="space-y-3">
      <div>
        <label htmlFor="magic-link-email" className="sr-only">
          {t("placeholder")}
        </label>
        <input
          id="magic-link-email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder={t("placeholder")}
          required
          aria-describedby={status === "error" ? "magic-link-error" : undefined}
          aria-invalid={status === "error"}
          className="w-full border border-border bg-card rounded-lg px-4 py-3 placeholder-muted focus:outline-none focus:ring-2 focus:ring-accent"
        />
      </div>
      <button
        type="submit"
        disabled={status === "loading"}
        aria-busy={status === "loading"}
        className="w-full bg-accent text-accent-foreground rounded-lg px-4 py-3 font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
      >
        {status === "loading" ? t("sending") : t("submit")}
      </button>
      {status === "error" && (
        <p id="magic-link-error" className="text-error text-sm" role="alert">
          {message}
        </p>
      )}
    </form>
  );
}
