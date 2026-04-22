import Link from "next/link";
import { getLocale, getTranslations } from "next-intl/server";
import { getAccessToken } from "@/lib/session";
import LocaleSwitcher from "./LocaleSwitcher";
import ThemeSwitcher from "./ThemeSwitcher";

export default async function Nav() {
  const isLoggedIn = !!(await getAccessToken());
  const t = await getTranslations("nav");
  const locale = await getLocale();

  return (
    <nav
      aria-label={t("brand")}
      className="flex items-center justify-end gap-3 px-6 py-3 border-b border-border"
    >
      <LocaleSwitcher current={locale} />
      <ThemeSwitcher />
      {isLoggedIn ? (
        <form action="/api/auth/logout" method="POST">
          <button
            type="submit"
            aria-label={t("logout")}
            className="inline-flex items-center justify-center rounded-md border border-border px-4 py-2 text-sm font-medium hover:bg-foreground/5 transition-colors"
          >
            {t("logout")}
          </button>
        </form>
      ) : (
        <Link
          href="/login"
          className="inline-flex items-center justify-center rounded-md bg-accent text-accent-foreground px-4 py-2 text-sm font-medium hover:opacity-90 transition-opacity"
        >
          {t("signIn")}
        </Link>
      )}
    </nav>
  );
}
