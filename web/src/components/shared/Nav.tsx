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
      className="flex items-center justify-between px-6 py-4 border-b border-border"
      aria-label={t("brand")}
    >
      <Link href="/" className="text-xl font-bold">
        {t("brand")}
      </Link>
      <div className="flex items-center gap-3">
        <LocaleSwitcher current={locale} />
        <ThemeSwitcher />
        {isLoggedIn ? (
          <>
            <Link href="/home" className="text-muted hover:text-foreground transition-colors">
              {t("dashboard")}
            </Link>
            <form action="/api/auth/logout" method="POST">
              <button
                type="submit"
                className="text-muted hover:text-foreground transition-colors"
                aria-label={t("logout")}
              >
                {t("logout")}
              </button>
            </form>
          </>
        ) : (
          <Link
            href="/login"
            className="bg-accent text-accent-foreground px-4 py-2 rounded-lg hover:opacity-90 transition-opacity"
          >
            {t("signIn")}
          </Link>
        )}
      </div>
    </nav>
  );
}
