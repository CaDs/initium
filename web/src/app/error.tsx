"use client";

import { useTranslations } from "next-intl";

interface ErrorPageProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function ErrorPage({ reset }: ErrorPageProps) {
  const t = useTranslations("errors");

  return (
    <div className="flex flex-col items-center justify-center min-h-[50vh] px-6 py-12 text-center">
      <h1 className="text-2xl font-bold mb-2">{t("title")}</h1>
      <p className="text-muted mb-6">{t("body")}</p>
      <button
        onClick={reset}
        className="bg-accent text-accent-foreground rounded-lg px-5 py-2.5 font-medium hover:opacity-90 transition-opacity"
      >
        {t("retry")}
      </button>
    </div>
  );
}
