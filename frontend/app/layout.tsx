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

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className="h-full">
      {/*
        Next.js App Router manages <head> automatically via the metadata export above.
        favicon.ico in public/ is the primary tab icon (browsers request /favicon.ico first).
        nomvox-icon.png is the high-res fallback for Apple/PWA.
        Do NOT add a manual <head> block — causes hydration mismatch.
      */}
      <body className="min-h-full flex flex-col antialiased">{children}<Analytics /></body>
    </html>
  );
}
