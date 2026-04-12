"use client";

import { useRouter } from "next/navigation";
import { useTransition } from "react";

const localeLabels: Record<string, string> = {
  en: "EN",
  es: "ES",
  ja: "JA",
};

export default function LocaleSwitcher({ current }: { current: string }) {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();

  function switchLocale(locale: string) {
    document.cookie = `locale=${locale};path=/;max-age=31536000;samesite=lax`;
    startTransition(() => {
      router.refresh();
    });
  }

  return (
    <div className="flex items-center gap-1" role="radiogroup" aria-label="Language">
      {Object.entries(localeLabels).map(([locale, label]) => (
        <button
          key={locale}
          onClick={() => switchLocale(locale)}
          disabled={isPending}
          role="radio"
          aria-checked={locale === current}
          className={`px-2 py-1 text-xs rounded transition-colors ${
            locale === current
              ? "bg-neutral-900 text-white dark:bg-white dark:text-neutral-900"
              : "text-neutral-500 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-white"
          }`}
        >
          {label}
        </button>
      ))}
    </div>
  );
}
