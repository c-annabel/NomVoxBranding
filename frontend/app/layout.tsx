import type { Metadata } from "next";
import "./globals.css";
import { Analytics } from '@vercel/analytics/next';

export const metadata: Metadata = {
  title: "NomVox Branding",
  description:
    "AI-powered brand identity platform. From void to voice — names, handles, logos, and landing pages synthesized in under 60 seconds.",
  icons: {
    icon: [
      { url: "/favicon.ico",      sizes: "any" },
      { url: "/nomvox-icon.png",  type: "image/png", sizes: "512x512" },
    ],
    shortcut: "/favicon.ico",
    apple:    "/nomvox-icon.png",
  },
};

// Next.js App Router manages <head> automatically via the metadata export above.
// favicon.ico in public/ is the primary tab icon (browsers request /favicon.ico first).
// nomvox-icon.png is the high-res fallback for Apple/PWA.
// Do NOT add a manual <head> block — it causes a hydration mismatch.
// Do NOT put a JSX comment between <html> and <body> — Next 16 / Turbopack's root
// layout validator then reports "Missing <html> and <body> tags in the root layout."
export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className="h-full">
      <body className="min-h-full flex flex-col antialiased">{children}<Analytics /></body>
    </html>
  );
}
