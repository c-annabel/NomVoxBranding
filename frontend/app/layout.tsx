import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "NomVox — Signal Sent. Brand Received.",
  description:
    "AI-powered brand identity platform. From void to voice — names, handles, logos, and landing pages synthesized in under 60 seconds.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className="h-full">
      <body className="min-h-full flex flex-col antialiased">{children}</body>
    </html>
  );
}
