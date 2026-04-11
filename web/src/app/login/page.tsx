import { redirect } from "next/navigation";
import GoogleSignInButton from "@/components/auth/GoogleSignInButton";
import MagicLinkForm from "@/components/auth/MagicLinkForm";
import { hasSession } from "@/lib/session";

export default async function LoginPage() {
  if (await hasSession()) {
    redirect("/home");
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] px-6">
      <div className="w-full max-w-sm space-y-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900">Sign In</h1>
          <p className="text-gray-600 mt-1">No passwords needed.</p>
        </div>

        <GoogleSignInButton />

        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-gray-200" />
          </div>
          <div className="relative flex justify-center text-sm">
            <span className="bg-white px-4 text-gray-500">or</span>
          </div>
        </div>

        <MagicLinkForm />
      </div>
    </div>
  );
}
