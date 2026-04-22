import { redirect } from "next/navigation";
import { useTranslations } from "next-intl";
import MagicLinkForm from "@/components/auth/MagicLinkForm";
import { hasSession } from "@/lib/session";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

export default async function LoginPage() {
  if (await hasSession()) {
    redirect("/home");
  }
  return <LoginContent />;
}

function LoginContent() {
  const t = useTranslations("login");

  return (
    <div className="mx-auto w-full max-w-sm px-6 py-12 space-y-6">
      <h1 className="text-2xl font-semibold">{t("title")}</h1>
      <div className="rounded-lg border border-border bg-card p-6 space-y-4">
        <a
          href={`${API_URL}/api/auth/google`}
          className="inline-flex items-center justify-center w-full rounded-md border border-border px-4 py-2 text-sm font-medium hover:bg-foreground/5 transition-colors"
        >
          {t("google")}
        </a>
        <MagicLinkForm />
      </div>
    </div>
  );
}
