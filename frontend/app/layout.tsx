import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "NomVox Branding",
  description:
    "AI-powered brand identity platform. From void to voice — names, handles, logos, and landing pages synthesized in under 60 seconds.",
  icons: {
    icon: [
      { url: "/nomvox-icon.png", type: "image/png" },
    ],
    shortcut: "/nomvox-icon.png",
    apple: "/nomvox-icon.png",
  },
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className="h-full">
      {/*
        Next.js App Router manages <head> automatically via the metadata export above.
        Do NOT add a manual <head> block here — it causes hydration mismatch and blank page.
        The two <Image priority> tags in HomeClient already handle preloading the
        logo and background via the browser's native priority hints.
      */}
      <body className="min-h-full flex flex-col antialiased">{children}</body>
    </html>
  );
}
