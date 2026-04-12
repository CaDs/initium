import Link from "next/link";
import { useTranslations } from "next-intl";

export default function LandingPage() {
  const t = useTranslations("landing");

  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] px-6 text-center">
      <h1 className="text-5xl font-bold mb-4">
        {t("title")}
      </h1>
      <p className="text-xl text-muted max-w-lg mb-8">
        {t("subtitle")}
      </p>
      <Link
        href="/login"
        className="bg-accent text-accent-foreground px-8 py-3 rounded-lg text-lg font-medium hover:opacity-90 transition-opacity"
      >
        {t("cta")}
      </Link>
    </div>
  );
}
