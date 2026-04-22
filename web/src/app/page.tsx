import { redirect } from "next/navigation";
import { hasSession } from "@/lib/session";

export default async function RootPage() {
  redirect((await hasSession()) ? "/home" : "/login");
}
