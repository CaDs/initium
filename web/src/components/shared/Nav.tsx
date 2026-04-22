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
    <nav aria-label={t("brand")}>
      <LocaleSwitcher current={locale} />
      <ThemeSwitcher />
      {isLoggedIn ? (
        <form action="/api/auth/logout" method="POST">
          <button type="submit" aria-label={t("logout")}>
            {t("logout")}
          </button>
        </form>
      ) : (
        <Link href="/login">{t("signIn")}</Link>
      )}
    </nav>
  );
}
