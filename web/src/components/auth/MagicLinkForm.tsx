"use client";

import { useState } from "react";
import { requestMagicLink } from "@/actions/auth";

export default function MagicLinkForm() {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<"idle" | "loading" | "sent" | "error">("idle");
  const [message, setMessage] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setStatus("loading");

    const result = await requestMagicLink(email);

    if (result.ok) {
      setStatus("sent");
      setMessage("Check your email for the magic link!");
    } else {
      setStatus("error");
      setMessage(result.message);
    }
  }

  if (status === "sent") {
    return (
      <div className="text-center p-4 bg-green-50 rounded-lg">
        <p className="text-green-700 font-medium">{message}</p>
        <p className="text-green-600 text-sm mt-1">Check your inbox (or Mailpit at localhost:8025)</p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Enter your email"
        required
        className="w-full border border-gray-300 rounded-lg px-4 py-3 text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-gray-900"
      />
      <button
        type="submit"
        disabled={status === "loading"}
        className="w-full bg-gray-900 text-white rounded-lg px-4 py-3 font-medium hover:bg-gray-800 transition-colors disabled:opacity-50"
      >
        {status === "loading" ? "Sending..." : "Send Magic Link"}
      </button>
      {status === "error" && (
        <p className="text-red-600 text-sm">{message}</p>
      )}
    </form>
  );
}
