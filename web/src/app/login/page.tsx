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
    <div>
      <h1>{t("title")}</h1>
      <a href={`${API_URL}/api/auth/google`}>{t("google")}</a>
      <MagicLinkForm />
    </div>
  );
}
