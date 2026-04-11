import Link from "next/link";

export default function LandingPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] px-6 text-center">
      <h1 className="text-5xl font-bold text-gray-900 mb-4">
        Welcome to Initium
      </h1>
      <p className="text-xl text-gray-600 max-w-lg mb-8">
        Your next great idea starts here. A modern, full-stack starter template
        ready for rapid prototyping.
      </p>
      <Link
        href="/login"
        className="bg-gray-900 text-white px-8 py-3 rounded-lg text-lg font-medium hover:bg-gray-800 transition-colors"
      >
        Get Started
      </Link>
    </div>
  );
}
