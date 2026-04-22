import { redirect } from "next/navigation";
import { apiFetch } from "@/lib/api";
import { userSchema } from "@/lib/schemas";
import type { User } from "@/lib/types";

export default async function HomePage() {
  const result = await apiFetch<User>("/api/me", {}, userSchema);

  if (!result.ok) {
    redirect("/login");
  }

  return (
    <div>
      <h1>Home</h1>
      <pre>{JSON.stringify(result.data, null, 2)}</pre>
    </div>
  );
}
