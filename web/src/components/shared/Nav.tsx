import Link from "next/link";
import { getAccessToken } from "@/lib/session";

export default async function Nav() {
  const isLoggedIn = !!(await getAccessToken());

  return (
    <nav className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
      <Link href="/" className="text-xl font-bold text-gray-900">
        Initium
      </Link>
      <div className="flex items-center gap-4">
        {isLoggedIn ? (
          <>
            <Link href="/home" className="text-gray-600 hover:text-gray-900">
              Dashboard
            </Link>
            <form action="/api/auth/logout" method="POST">
              <button
                type="submit"
                className="text-gray-600 hover:text-gray-900"
              >
                Logout
              </button>
            </form>
          </>
        ) : (
          <Link
            href="/login"
            className="bg-gray-900 text-white px-4 py-2 rounded-lg hover:bg-gray-800"
          >
            Sign In
          </Link>
        )}
      </div>
    </nav>
  );
}
