import type { Metadata } from "next";
import Script from "next/script";
import localFont from "next/font/local";
import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages } from "next-intl/server";
import { Toaster } from "sonner";
import "./globals.css";
import Nav from "@/components/shared/Nav";
import DevModeBanner from "@/components/shared/DevModeBanner";

const geistSans = localFont({
  src: "../../public/fonts/geist-sans-latin.woff2",
  variable: "--font-geist-sans",
  display: "swap",
});

const geistMono = localFont({
  src: "../../public/fonts/geist-mono-latin.woff2",
  variable: "--font-geist-mono",
  display: "swap",
});

export const metadata: Metadata = {
  title: "Initium",
  description: "Your next great idea starts here.",
};

// Theme bootstrap script — runs before React hydrates, so the correct
// CSS classes are on <html> before first paint (prevents FOUC). Using
// next/script with `beforeInteractive` is the Next 15+ / React 19
// approved way to ship inline JS that must run before the page is
// interactive. Writing a raw `<script dangerouslySetInnerHTML>` now
// triggers a React warning ("Scripts inside React components are never
// executed when rendering on the client") because React trees re-render
// without re-executing scripts.
const themeBootstrap = `
  (function() {
    var theme = localStorage.getItem('theme') || 'system';
    var prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    if (theme === 'dark' || (theme === 'system' && prefersDark)) {
      document.documentElement.classList.add('dark');
    } else if (theme === 'light') {
      document.documentElement.setAttribute('data-theme', 'light');
    }
  })();
`;

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const locale = await getLocale();
  const messages = await getMessages();

  return (
    <html
      lang={locale}
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
      suppressHydrationWarning
    >
      <body className="min-h-full flex flex-col bg-background text-foreground">
        <Script id="theme-bootstrap" strategy="beforeInteractive">
          {themeBootstrap}
        </Script>
        <NextIntlClientProvider messages={messages}>
          <a href="#main-content" className="skip-link">
            Skip to main content
          </a>
          <DevModeBanner />
          <Nav />
          <main id="main-content" className="flex-1">
            {children}
          </main>
          <Toaster richColors position="top-right" theme="system" />
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
