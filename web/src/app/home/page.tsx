import { redirect } from "next/navigation";
import { apiFetch } from "@/lib/api";
import type { User } from "@/lib/types";

export default async function HomePage() {
  const result = await apiFetch<User>("/api/me");

  if (!result.ok) {
    redirect("/login");
  }

  const user = result.data;

  return (
    <div className="max-w-2xl mx-auto px-6 py-12">
      <h1 className="text-3xl font-bold text-gray-900 mb-2">
        Welcome back{user.name ? `, ${user.name}` : ""}!
      </h1>
      <p className="text-gray-600 mb-8">
        This is your authenticated home screen. Customize it for your POC.
      </p>

      <div className="bg-gray-50 rounded-lg p-6 space-y-3">
        <h2 className="text-lg font-semibold text-gray-900">Your Profile</h2>
        <div className="grid grid-cols-2 gap-2 text-sm">
          <span className="text-gray-500">Email</span>
          <span className="text-gray-900">{user.email}</span>
          <span className="text-gray-500">Name</span>
          <span className="text-gray-900">{user.name || "—"}</span>
          <span className="text-gray-500">User ID</span>
          <span className="text-gray-900 font-mono text-xs">{user.id}</span>
        </div>
      </div>
    </div>
  );
}
