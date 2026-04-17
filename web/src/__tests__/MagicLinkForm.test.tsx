import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import MagicLinkForm from "@/components/auth/MagicLinkForm";

// Mock next-intl
vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => {
    const map: Record<string, string> = {
      placeholder: "Enter your email",
      submit: "Send Magic Link",
      sending: "Sending...",
      sent: "Check your email!",
      sentDetail: "A magic link has been sent to your inbox.",
    };
    return map[key] ?? key;
  },
}));

// Mock the server action
vi.mock("@/actions/auth", () => ({
  requestMagicLink: vi.fn(),
}));

describe("MagicLinkForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the submit button with correct label", () => {
    render(<MagicLinkForm />);
    expect(
      screen.getByRole("button", { name: "Send Magic Link" })
    ).toBeInTheDocument();
  });

  it("renders the email input", () => {
    render(<MagicLinkForm />);
    expect(screen.getByRole("textbox")).toBeInTheDocument();
  });

  it("submit button is enabled by default", () => {
    render(<MagicLinkForm />);
    expect(
      screen.getByRole("button", { name: "Send Magic Link" })
    ).not.toBeDisabled();
  });
});
