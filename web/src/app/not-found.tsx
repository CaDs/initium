import { useTranslations } from "next-intl";
import Link from "next/link";

export default function NotFoundPage() {
  const t = useTranslations("notFound");

  return (
    <div className="flex flex-col items-center justify-center min-h-[50vh] px-6 py-12 text-center">
      <h1 className="text-2xl font-bold mb-2">{t("title")}</h1>
      <p className="text-muted mb-6">{t("body")}</p>
      <Link
        href="/"
        className="bg-accent text-accent-foreground rounded-lg px-5 py-2.5 font-medium hover:opacity-90 transition-opacity"
      >
        {t("home")}
      </Link>
    </div>
  );
}
