"use client";

import Image from "next/image";
import { useState } from "react";
import IntakeForm from "@/components/IntakeForm";
import type { IntakePayload } from "@/lib/types";

export default function HomeClient() {
  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState<IntakePayload | null>(null);

  async function handleSubmit(payload: IntakePayload) {
    setLoading(true);
    // ST-03 will stream name cards here.
    console.log("Intake submitted:", payload);
    await new Promise((r) => setTimeout(r, 800));
    setSubmitted(payload);
    setLoading(false);
  }

  return (
    <>
      {/* ── Full-bleed star-field background ────────────────────── */}
      <div className="fixed inset-0 -z-10 overflow-hidden">
        <Image
          src="/nomvox-bg.png"
          alt=""
          fill
          priority
          quality={90}
          className="object-cover object-center"
          style={{ opacity: 0.85 }}
        />
        {/* Radial vignette — deepens centre focus */}
        <div
          className="absolute inset-0"
          style={{
            background:
              "radial-gradient(ellipse at 50% 0%, rgba(139,92,246,0.07) 0%, transparent 65%), " +
              "radial-gradient(ellipse at 50% 100%, rgba(5,7,15,0.7) 0%, transparent 70%)",
          }}
        />
      </div>

      {/* ── Page — compact top, generous sides ──────────────────── */}
      <main className="flex min-h-screen flex-col items-center px-3 sm:px-6 pt-8 pb-16">

        {/* ── Logo + tagline hero — tight & centred ────────────── */}
        <div className="mb-6 flex flex-col items-center text-center">

          {/* Logo — includes tagline in the image itself */}
          <Image
            src="/nomvox-logo2.png"
            alt="NomVox — Born from the Void"
            width={600}
            height={200}
            priority
            className="w-[300px] sm:w-[300px] h-auto
              drop-shadow-[0_0_32px_rgba(139,92,246,0.55)]"
          />

          {/* Sub-copy */}
          <p className="mt-3 text-sm whitespace-nowrap"
            style={{ color: "var(--color-text-secondary)" }}>
            Name your brand in 60 seconds — handles, logos &amp; landing pages included.
          </p>
        </div>

        {/* ── Intake form panel ────────────────────────────────── */}
        <div
          className="w-full max-w-2xl lg:max-w-3xl xl:max-w-4xl
            rounded-2xl border p-5 sm:p-7
            backdrop-blur-sm"
          style={{
            background: "rgba(11,15,28,0.88)",
            borderColor: "rgba(139,92,246,0.28)",
            boxShadow:
              "0 0 0 1px rgba(139,92,246,0.10), " +
              "0 20px 50px rgba(5,7,15,0.65), " +
              "0 0 60px rgba(139,92,246,0.07)",
          }}
        >
          <IntakeForm onSubmit={handleSubmit} loading={loading} />
        </div>

        {/* ── Temporary confirmation (removed in ST-03) ─────────── */}
        {submitted && (
          <div
            className="mt-5 w-full max-w-2xl lg:max-w-3xl xl:max-w-4xl
              rounded-xl border px-5 py-4 text-sm backdrop-blur-sm"
            style={{
              background: "rgba(11,15,28,0.85)",
              borderColor: "rgba(139,92,246,0.3)",
            }}
          >
            <p className="font-bold mb-1" style={{ color: "var(--color-pulse)" }}>
              ✓ Intake received
            </p>
            <p style={{ color: "var(--color-text-secondary)" }}>
              Core idea:{" "}
              <span style={{ color: "var(--color-text-primary)" }}>
                {submitted.core_idea}
              </span>
            </p>
            <p className="text-xs mt-2" style={{ color: "var(--color-text-hint)" }}>
              Name cards will stream here after ST-03 is wired up.
            </p>
          </div>
        )}
      </main>
    </>
  );
}
