"use client";

import { useRouter } from "next/navigation";
import { useTransition } from "react";

const locales = [
  { code: "en", label: "English" },
  { code: "es", label: "Español" },
  { code: "ja", label: "日本語" },
];

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
    <select
      value={current}
      onChange={(e) => switchLocale(e.target.value)}
      disabled={isPending}
      aria-label="Language"
      className="bg-transparent border border-border rounded px-2 py-1 text-sm text-foreground cursor-pointer focus:outline-none focus:ring-2 focus:ring-accent"
    >
      {locales.map(({ code, label }) => (
        <option key={code} value={code}>
          {label}
        </option>
      ))}
    </select>
  );
}
