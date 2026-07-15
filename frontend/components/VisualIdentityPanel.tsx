"use client";

import { useState, useMemo } from "react";
import type { VisualsResponse, BrandPersona, NameCard, IntakePayload, ExportRequest } from "@/lib/types";

type VisualStep = "logos" | "moodboard" | "mockup";

interface VisualIdentityPanelProps {
  brandName: string;
  card: NameCard;
  intake: IntakePayload;
  sessionId: string;
  visuals: VisualsResponse;
  onBack: () => void;
  onStartOver: () => void;
  onRegenerateLogos: () => void;
  onRegenerateMoodboard: (logoKey: string, logoStyle: string) => void;
  generatingLogos?: boolean;
  generatingMoodboard?: boolean;
}

const BACKEND = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const LOGO_STYLE_LABELS: Record<string, string> = {
  profile:  "flat geometric Bauhaus mark — bold shapes, strong negative space, minimal palette",
  app:      "vibrant gradient icon — deep colour depth, glassmorphism, neon glow accent",
  business: "horizontal wordmark lockup — clean white background, corporate typography, single accent colour",
};

export default function VisualIdentityPanel({
  brandName, card, intake, sessionId, visuals,
  onBack, onStartOver, onRegenerateLogos, onRegenerateMoodboard,
  generatingLogos, generatingMoodboard,
}: VisualIdentityPanelProps) {
  const [visualStep, setVisualStep] = useState<VisualStep>("logos");
  const [selectedLogoType, setSelectedLogoType] = useState<string>("");
  const [exporting, setExporting]   = useState(false);
  const [exportError, setExportError] = useState("");

  async function handleExport() {
    setExporting(true);
    setExportError("");
    try {
      const payload: ExportRequest = {
        session_id: sessionId, brand_name: brandName, card, intake,
        mood_board: visuals.mood_board ?? [],
        logo_profile: visuals.logo_profile ?? "",
        logo_app: visuals.logo_app ?? "",
        logo_business: visuals.logo_business ?? "",
        selected_logo_key: selectedLogoType,
        mockup_html: visuals.mockup_html ?? "",
        persona: visuals.persona,
      };
      const resp = await fetch(BACKEND + "/api/export", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
      const blob = await resp.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${brandName.toLowerCase()}-nomvox-pack.zip`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setExportError(err instanceof Error ? err.message : String(err));
    } finally {
      setExporting(false);
    }
  }

  // CSS concept placeholders — each with a distinct visual style
  // Derive initials for the CSS logo concepts (max 2 chars)
  const initials = useMemo(() => {
    const parts = brandName.trim().split(/\s+/);
    if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
    return brandName.slice(0, 2).toUpperCase();
  }, [brandName]);

  const logoItems = useMemo(() => [
    {
      key: "profile",
      label: "Profile / Social",
      uri: visuals.logo_profile,
      hint: "1:1 square — Twitter, Instagram, LinkedIn",
      placeholder: {
        styleName: "Geometric Mark",
        render: () => (
          /* Bauhaus-style geometric mark — bold hexagonal shape + initial */
          <div className="w-full aspect-square flex flex-col items-center justify-center relative overflow-hidden"
            style={{ background: "linear-gradient(160deg, #0f0f1e 0%, #1c0a3a 50%, #0a0f1f 100%)" }}>
            {/* Outer ring */}
            <div className="absolute" style={{
              width: "58%", height: "58%",
              border: "2px solid rgba(139,92,246,0.45)",
              borderRadius: "50%",
              background: "rgba(139,92,246,0.07)",
            }} />
            {/* Hexagonal accent shape */}
            <div className="absolute" style={{
              width: "42%", height: "42%",
              background: "linear-gradient(135deg, rgba(139,92,246,0.6) 0%, rgba(34,211,238,0.35) 100%)",
              clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
            }} />
            {/* Brand initial */}
            <span className="relative z-10 font-black select-none"
              style={{ fontSize: "clamp(1.5rem,6vw,2.2rem)", color: "#fff", letterSpacing: "-0.04em", textShadow: "0 0 20px rgba(139,92,246,0.8)" }}>
              {initials.slice(0, 1)}
            </span>
            {/* Corner accent dots */}
            {[0, 90, 180, 270].map(deg => (
              <div key={deg} className="absolute" style={{
                width: 6, height: 6, borderRadius: "50%",
                background: "#22d3ee",
                transform: `rotate(${deg}deg) translateY(-36%)`,
                opacity: 0.7,
              }} />
            ))}
          </div>
        ),
      },
    },
    {
      key: "app",
      label: "App Icon",
      uri: visuals.logo_app,
      hint: "Rounded square — app stores",
      placeholder: {
        styleName: "Gradient Icon",
        render: () => (
          /* Vibrant app store icon — glassmorphic gradient + bold letterform */
          <div className="w-full aspect-square flex items-center justify-center relative overflow-hidden"
            style={{
              background: "linear-gradient(145deg, #0d1b3e 0%, #1a0554 40%, #0a2240 100%)",
              borderRadius: "0", /* parent handles radius */
            }}>
            {/* Ambient glow blobs */}
            <div className="absolute" style={{
              width: "80%", height: "80%", borderRadius: "50%",
              background: "radial-gradient(circle, rgba(139,92,246,0.30) 0%, transparent 70%)",
              top: "10%", left: "10%",
            }} />
            <div className="absolute" style={{
              width: "50%", height: "50%", borderRadius: "50%",
              background: "radial-gradient(circle, rgba(34,211,238,0.20) 0%, transparent 70%)",
              bottom: "5%", right: "5%",
            }} />
            {/* Glassmorphic card */}
            <div className="relative flex flex-col items-center justify-center"
              style={{
                width: "62%", height: "62%",
                background: "rgba(255,255,255,0.08)",
                borderRadius: "22%",
                border: "1.5px solid rgba(255,255,255,0.18)",
                backdropFilter: "blur(8px)",
                boxShadow: "0 8px 32px rgba(139,92,246,0.30)",
              }}>
              <span className="font-black select-none"
                style={{ fontSize: "clamp(1.2rem,5vw,1.8rem)", color: "#fff", letterSpacing: "-0.03em" }}>
                {initials}
              </span>
            </div>
          </div>
        ),
      },
    },
    {
      key: "business",
      label: "Business / Print",
      uri: visuals.logo_business,
      hint: "16:9 horizontal — business cards, letterhead",
      placeholder: {
        styleName: "Wordmark Lockup",
        render: () => (
          /* Clean horizontal lockup on light — corporate print quality */
          <div className="w-full aspect-square flex flex-col items-center justify-center gap-2 px-6 relative overflow-hidden"
            style={{ background: "linear-gradient(160deg, #faf9f7 0%, #f0ecf8 100%)" }}>
            {/* Subtle brand mark behind text */}
            <div className="absolute" style={{
              width: "55%", height: "55%", borderRadius: "50%",
              background: "radial-gradient(circle, rgba(139,92,246,0.06) 0%, transparent 70%)",
            }} />
            {/* Icon + wordmark row */}
            <div className="flex items-center gap-2.5 relative">
              {/* Small geometric icon */}
              <div style={{
                width: 28, height: 28,
                background: "linear-gradient(135deg, #7c3aed 0%, #2563eb 100%)",
                clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
                flexShrink: 0,
              }} />
              <span className="font-black tracking-tight select-none"
                style={{ fontSize: "clamp(1rem,4vw,1.4rem)", color: "#1a0a2e", letterSpacing: "-0.02em" }}>
                {brandName}
              </span>
            </div>
            {/* Divider rule */}
            <div style={{ width: "70%", height: 1, background: "linear-gradient(90deg, transparent, rgba(139,92,246,0.40), transparent)" }} />
            {/* Tagline stub */}
            <span className="text-xs uppercase tracking-widest select-none"
              style={{ color: "rgba(107,70,193,0.65)", fontSize: "0.6rem", letterSpacing: "0.15em" }}>
              Brand Identity
            </span>
          </div>
        ),
      },
    },
  // eslint-disable-next-line react-hooks/exhaustive-deps
  ], [visuals.logo_profile, visuals.logo_app, visuals.logo_business, brandName, initials]);

  // ── Step tabs ──────────────────────────────────────────────────────────────
  function StepTabs() {
    const steps: { id: VisualStep; label: string; locked: boolean }[] = [
      { id: "logos",    label: "① Logo Concepts",    locked: false },
      { id: "moodboard",label: "② Mood Board",       locked: !selectedLogoType },
      { id: "mockup",   label: "③ Landing Page",     locked: !selectedLogoType },
    ];
    return (
      <div className="flex border-b" style={{ borderColor: "rgba(255,255,255,0.07)" }}>
        {steps.map(s => (
          <button key={s.id} type="button"
            disabled={s.locked}
            onClick={() => !s.locked && setVisualStep(s.id)}
            className="flex-1 py-3 text-sm font-bold transition-all disabled:opacity-40"
            style={{
              background: visualStep === s.id ? "rgba(139,92,246,0.12)" : "transparent",
              borderBottom: visualStep === s.id ? "2px solid var(--color-pulse)" : "2px solid transparent",
              color: visualStep === s.id ? "#c4b5fd" : "var(--color-text-hint)",
            }}>
            {s.label}
          </button>
        ))}
      </div>
    );
  }

  // ══════════════════════════════════════════════════════════════════════════
  // STEP 1 — LOGO CONCEPTS
  // ══════════════════════════════════════════════════════════════════════════
  if (visualStep === "logos") {
    return (
      <div className="w-full max-w-2xl lg:max-w-4xl xl:max-w-5xl mx-auto">
        <div className="rounded-2xl border overflow-hidden"
          style={{ background: "rgba(18,22,42,0.97)", borderColor: "rgba(99,210,255,0.22)", boxShadow: "0 24px 64px rgba(5,7,15,0.7)" }}>

          {/* Header — Names + Start Over only (no Export here) */}
          <div className="flex items-center justify-between px-6 py-4 border-b"
            style={{ borderColor: "rgba(255,255,255,0.07)", background: "rgba(14,18,32,0.60)" }}>
            <div>
              <p className="text-xs font-black uppercase tracking-widest mb-0.5"
                style={{ color: "var(--color-pulse)" }}>Visual Identity — Step 1/3</p>
              <p className="text-xl font-black tracking-tight leading-none"
                style={{ color: "var(--color-text-primary)" }}>{brandName}</p>
            </div>
            <div className="flex items-center gap-2">
              <button type="button" onClick={onBack}
                className="px-4 py-2 rounded-lg text-sm font-bold transition-all"
                style={{ border: "1px solid rgba(99,210,255,0.30)", color: "var(--color-signal)", background: "transparent" }}
                onMouseEnter={e => { e.currentTarget.style.background = "rgba(34,211,238,0.10)"; }}
                onMouseLeave={e => { e.currentTarget.style.background = "transparent"; }}>
                ← Names
              </button>
              <button type="button" onClick={onStartOver}
                className="px-4 py-2 rounded-lg text-sm font-bold transition-all"
                style={{ border: "1px solid rgba(157,48,96,0.55)", color: "#e040a0", background: "transparent" }}
                onMouseEnter={e => { e.currentTarget.style.background = "rgba(224,64,160,0.10)"; e.currentTarget.style.borderColor = "#e040a0"; }}
                onMouseLeave={e => { e.currentTarget.style.background = "transparent"; e.currentTarget.style.borderColor = "rgba(157,48,96,0.55)"; }}>
                ⟳ Start Over
              </button>
            </div>
          </div>

          <StepTabs />

          <div className="p-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <p className="text-lg font-black" style={{ color: "var(--color-text-primary)" }}>
                  Choose your logo direction
                </p>
                <p className="text-sm mt-0.5" style={{ color: "var(--color-text-secondary)" }}>
                  Select one to continue. Regenerate to get fresh concepts.
                </p>
              </div>
              <button type="button" onClick={onRegenerateLogos} disabled={generatingLogos}
                className="flex items-center gap-2 px-5 py-2 rounded-lg text-sm font-black transition-all disabled:opacity-50"
                style={{
                  background: "linear-gradient(135deg,rgba(109,40,217,0.85) 0%,rgba(30,90,180,0.80) 100%)",
                  border: "1px solid rgba(139,92,246,0.40)", color: "#fff",
                }}>
                {generatingLogos ? (
                  <>Generating<span className="nv-dot1">.</span><span className="nv-dot2">.</span><span className="nv-dot3">.</span></>
                ) : "↻ New Logos"}
              </button>
            </div>

            {/* 3 logo columns */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-5">
              {logoItems.map(({ key, label, uri, hint, placeholder }) => {
                const isSelected = selectedLogoType === key;
                return (
                  <div key={key}
                    className="rounded-xl overflow-hidden flex flex-col cursor-pointer transition-all"
                    style={{
                      border: isSelected ? "2px solid #8B5CF6" : "1px solid rgba(255,255,255,0.12)",
                      background: isSelected ? "rgba(139,92,246,0.10)" : "rgba(255,255,255,0.02)",
                      boxShadow: isSelected ? "0 0 28px rgba(139,92,246,0.30)" : "none",
                    }}
                    onClick={() => setSelectedLogoType(key)}>

                    {uri ? (
                      /* eslint-disable-next-line @next/next/no-img-element */
                      <img src={uri} alt={label} className="w-full aspect-square object-cover" />
                    ) : generatingLogos ? (
                      <div className="w-full aspect-square flex items-center justify-center"
                        style={{ background: "rgba(139,92,246,0.06)" }}>
                        <p className="text-sm font-bold" style={{ color: "var(--color-pulse)" }}>
                          Generating<span className="nv-dot1">.</span><span className="nv-dot2">.</span><span className="nv-dot3">.</span>
                        </p>
                      </div>
                    ) : (
                      /* Rich CSS concept placeholder — branded, selectable, design-quality */
                      placeholder.render()
                    )}

                    <div className="px-4 py-3 flex items-center justify-between">
                      <div>
                        <p className="text-sm font-black" style={{ color: "var(--color-text-primary)" }}>{label}</p>
                        <p className="text-xs" style={{ color: "var(--color-text-hint)" }}>{hint}</p>
                      </div>
                      {isSelected && (
                        <span className="text-xs font-black px-2.5 py-1 rounded-full"
                          style={{ background: "rgba(139,92,246,0.22)", color: "#c4b5fd", border: "1px solid rgba(139,92,246,0.40)" }}>
                          ✓ Selected
                        </span>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* CTA — only when logo selected */}
            {selectedLogoType && (
              <div className="mt-8 flex justify-end">
                <button type="button" onClick={() => {
                    onRegenerateMoodboard(selectedLogoType, LOGO_STYLE_LABELS[selectedLogoType] ?? "");
                    setVisualStep("moodboard");
                  }}
                  className="flex items-center gap-2 px-8 py-3 rounded-xl text-base font-black transition-all"
                  style={{
                    background: "linear-gradient(135deg,rgba(109,40,217,0.90) 0%,rgba(30,90,180,0.85) 100%)",
                    border: "1px solid rgba(139,92,246,0.50)", color: "#fff",
                    boxShadow: "0 0 24px rgba(109,40,217,0.35)",
                  }}>
                  Next: Mood Board →
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }

  // ══════════════════════════════════════════════════════════════════════════
  // STEP 2 — MOOD BOARD (single presentation image)
  // ══════════════════════════════════════════════════════════════════════════
  if (visualStep === "moodboard") {
    return (
      <div className="w-full max-w-2xl lg:max-w-4xl xl:max-w-5xl mx-auto">
        <div className="rounded-2xl border overflow-hidden"
          style={{ background: "rgba(18,22,42,0.97)", borderColor: "rgba(99,210,255,0.22)", boxShadow: "0 24px 64px rgba(5,7,15,0.7)" }}>

          {/* Header — Names + Start Over only (no Export) */}
          <div className="flex items-center justify-between px-6 py-4 border-b"
            style={{ borderColor: "rgba(255,255,255,0.07)", background: "rgba(14,18,32,0.60)" }}>
            <div>
              <p className="text-xs font-black uppercase tracking-widest mb-0.5"
                style={{ color: "var(--color-pulse)" }}>Visual Identity — Step 2/3</p>
              <p className="text-xl font-black tracking-tight leading-none"
                style={{ color: "var(--color-text-primary)" }}>{brandName}</p>
            </div>
            <div className="flex items-center gap-2">
              <button type="button" onClick={onBack}
                className="px-4 py-2 rounded-lg text-sm font-bold transition-all"
                style={{ border: "1px solid rgba(99,210,255,0.30)", color: "var(--color-signal)", background: "transparent" }}
                onMouseEnter={e => { e.currentTarget.style.background = "rgba(34,211,238,0.10)"; }}
                onMouseLeave={e => { e.currentTarget.style.background = "transparent"; }}>
                ← Names
              </button>
              <button type="button" onClick={onStartOver}
                className="px-4 py-2 rounded-lg text-sm font-bold transition-all"
                style={{ border: "1px solid rgba(157,48,96,0.55)", color: "#e040a0", background: "transparent" }}
                onMouseEnter={e => { e.currentTarget.style.background = "rgba(224,64,160,0.10)"; e.currentTarget.style.borderColor = "#e040a0"; }}
                onMouseLeave={e => { e.currentTarget.style.background = "transparent"; e.currentTarget.style.borderColor = "rgba(157,48,96,0.55)"; }}>
                ⟳ Start Over
              </button>
            </div>
          </div>

          <StepTabs />

          <div className="p-6">
            <p className="text-lg font-black mb-1" style={{ color: "var(--color-text-primary)" }}>Mood Board</p>
            <p className="text-sm mb-4" style={{ color: "var(--color-text-secondary)" }}>
              Visual world of <strong>{brandName}</strong> — how your selected logo lives across platforms and contexts.
            </p>

            {/* ── Selected logo showcase — always shown above mood board ── */}
            {(() => {
              const logoURI = selectedLogoType === "profile" ? visuals.logo_profile
                : selectedLogoType === "app" ? visuals.logo_app
                : selectedLogoType === "business" ? visuals.logo_business
                : (visuals.logo_profile || visuals.logo_app || visuals.logo_business || "");
              return logoURI ? (
                <div className="flex flex-col items-center mb-5 gap-2">
                  <p className="text-xs font-black uppercase tracking-widest"
                    style={{ color: "var(--color-pulse)" }}>
                    Selected Logo — {selectedLogoType || "Brand Mark"}
                  </p>
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img src={logoURI} alt={`${brandName} selected logo`}
                    className="rounded-xl"
                    style={{
                      maxHeight: 120, maxWidth: "60%", objectFit: "contain",
                      background: "rgba(255,255,255,0.04)",
                      border: "1px solid rgba(139,92,246,0.30)",
                      padding: "12px",
                    }} />
                  <p className="text-xs italic" style={{ color: "var(--color-text-hint)" }}>
                    {card.tagline}
                  </p>
                </div>
              ) : null;
            })()}

            {/* ── Mood board: 2×2 SVG tile grid ─────────────────────────── */}
            <div style={{ height: 420, position: "relative" }}>
            {generatingMoodboard ? (
              <div className="rounded-xl h-full flex items-center justify-center"
                style={{ border: "1px solid rgba(139,92,246,0.18)", background: "rgba(139,92,246,0.06)" }}>
                <p className="text-sm font-bold" style={{ color: "var(--color-pulse)" }}>
                  Generating mood board<span className="nv-dot1">.</span><span className="nv-dot2">.</span><span className="nv-dot3">.</span>
                </p>
              </div>
            ) : visuals.mood_board && visuals.mood_board.length > 0 ? (
              /* Show AI-generated SVG/PNG tiles in a 2×2 grid */
              <div className="rounded-xl overflow-hidden h-full"
                style={{ border: "1px solid rgba(255,255,255,0.10)" }}>
                <div className="grid grid-cols-2 h-full" style={{ gap: 2, background: "rgba(255,255,255,0.04)" }}>
                  {visuals.mood_board.slice(0, 4).map((tile, i) => (
                    <div key={i} className="overflow-hidden rounded-sm" style={{ background: "#0b0f19" }}>
                      {/* eslint-disable-next-line @next/next/no-img-element */}
                      <img
                        src={tile}
                        alt={`Mood board tile ${i + 1}`}
                        className="w-full h-full object-cover"
                        style={{ display: "block" }}
                      />
                    </div>
                  ))}
                  {/* Fill remaining slots if fewer than 4 tiles */}
                  {Array.from({ length: Math.max(0, 4 - (visuals.mood_board?.length ?? 0)) }).map((_, i) => (
                    <div key={`empty-${i}`} className="flex items-center justify-center"
                      style={{ background: "rgba(139,92,246,0.06)", border: "1px solid rgba(139,92,246,0.12)" }}>
                      <p className="text-xs" style={{ color: "rgba(139,92,246,0.40)" }}>Generating…</p>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              /* Fallback when mood board hasn't been generated yet */
              <div className="rounded-xl h-full flex flex-col items-center justify-center gap-3"
                style={{ border: "1px dashed rgba(139,92,246,0.25)", background: "rgba(139,92,246,0.04)" }}>
                <p className="text-sm font-bold" style={{ color: "var(--color-pulse)" }}>
                  Mood board generating…
                </p>
                <p className="text-xs text-center px-8" style={{ color: "var(--color-text-hint)" }}>
                  AI is creating 4 abstract brand tiles in your colour palette.
                </p>
              </div>
            )}
            </div>

            {/* ── Disclaimer — improved visibility ─────────────────────── */}
            <div className="mt-3 px-4 py-2 rounded-lg flex items-start gap-2"
              style={{ background: "rgba(139,92,246,0.08)", border: "1px solid rgba(139,92,246,0.20)" }}>
              <span style={{ color: "var(--color-pulse)", fontSize: "0.85rem", lineHeight: 1 }}>✦</span>
              <p className="text-xs leading-relaxed" style={{ color: "rgba(230,232,240,0.65)" }}>
                <strong style={{ color: "#c4b5fd" }}>Free-tier AI generation</strong> — SVG vector artwork
                generated by IBM watsonx AI using your brand palette and description.
                Photo-quality raster images require Gemini image credits (planned for a future release).
              </p>
            </div>

            {/* Brand persona card */}
            {visuals.persona && (
              <div className="mt-6">
                <p className="text-base font-black mb-1" style={{ color: "var(--color-text-primary)" }}>Brand Persona</p>
                <p className="text-sm mb-3" style={{ color: "var(--color-text-secondary)" }}>
                  {brandName} as a person
                </p>
                <PersonaCard persona={visuals.persona} />
              </div>
            )}

            <div className="mt-6 flex justify-between">
              <button type="button" onClick={() => setVisualStep("logos")}
                className="px-6 py-2 rounded-lg text-sm font-bold transition-all"
                style={{ border: "1px solid rgba(99,210,255,0.30)", color: "var(--color-signal)", background: "transparent" }}>
                ← Back to Logo Concepts
              </button>
              <button type="button" onClick={() => setVisualStep("mockup")}
                className="flex items-center gap-2 px-8 py-3 rounded-xl text-base font-black transition-all"
                style={{
                  background: "linear-gradient(135deg,rgba(109,40,217,0.90) 0%,rgba(30,90,180,0.85) 100%)",
                  border: "1px solid rgba(139,92,246,0.50)", color: "#fff",
                  boxShadow: "0 0 24px rgba(109,40,217,0.35)",
                }}>
                Next: Landing Page →
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // ══════════════════════════════════════════════════════════════════════════
  // STEP 3 — LANDING PAGE MOCKUP
  // ══════════════════════════════════════════════════════════════════════════
  return (
    <div className="w-full max-w-2xl lg:max-w-4xl xl:max-w-5xl mx-auto">
      <div className="rounded-2xl border overflow-hidden"
        style={{ background: "rgba(18,22,42,0.97)", borderColor: "rgba(99,210,255,0.22)", boxShadow: "0 24px 64px rgba(5,7,15,0.7)" }}>

        {/* Header — ← Back to Mood Board left, Names + Start Over + Export right */}
        <div className="flex items-center justify-between px-6 py-4 border-b"
          style={{ borderColor: "rgba(255,255,255,0.07)", background: "rgba(14,18,32,0.60)" }}>
          {/* Left: back nav + step label + name */}
          <div className="flex items-center gap-3">
            <button type="button" onClick={() => setVisualStep("moodboard")}
              className="px-4 py-2 rounded-lg text-sm font-bold transition-all"
              style={{ border: "1px solid rgba(99,210,255,0.30)", color: "var(--color-signal)", background: "transparent" }}
              onMouseEnter={e => { e.currentTarget.style.background = "rgba(34,211,238,0.10)"; }}
              onMouseLeave={e => { e.currentTarget.style.background = "transparent"; }}>
              ← Mood Board
            </button>
            <div>
              <p className="text-xs font-black uppercase tracking-widest mb-0.5"
                style={{ color: "var(--color-pulse)" }}>Visual Identity — Step 3/3</p>
              <p className="text-xl font-black tracking-tight leading-none"
                style={{ color: "var(--color-text-primary)" }}>{brandName}</p>
            </div>
          </div>
          {/* Right: Names + Start Over + Export */}
          <div className="flex items-center gap-2">
            <button type="button" onClick={onBack}
              className="px-4 py-2 rounded-lg text-sm font-bold transition-all"
              style={{ border: "1px solid rgba(99,210,255,0.30)", color: "var(--color-signal)", background: "transparent" }}
              onMouseEnter={e => { e.currentTarget.style.background = "rgba(34,211,238,0.10)"; }}
              onMouseLeave={e => { e.currentTarget.style.background = "transparent"; }}>
              ← Names
            </button>
            <button type="button" onClick={onStartOver}
              className="px-4 py-2 rounded-lg text-sm font-bold transition-all"
              style={{ border: "1px solid rgba(157,48,96,0.55)", color: "#e040a0", background: "transparent" }}
              onMouseEnter={e => { e.currentTarget.style.background = "rgba(224,64,160,0.10)"; e.currentTarget.style.borderColor = "#e040a0"; }}
              onMouseLeave={e => { e.currentTarget.style.background = "transparent"; e.currentTarget.style.borderColor = "rgba(157,48,96,0.55)"; }}>
              ⟳ Start Over
            </button>
            <button type="button" onClick={handleExport} disabled={exporting}
              className="flex items-center gap-2 px-5 py-2 rounded-lg text-sm font-black transition-all disabled:opacity-50"
              style={{
                background: "linear-gradient(135deg,rgba(109,40,217,0.85) 0%,rgba(30,90,180,0.80) 100%)",
                border: "1px solid rgba(139,92,246,0.40)", color: "#fff",
              }}>
              {exporting ? "⏳ Exporting…" : "⬇ Export Pack"}
            </button>
          </div>
        </div>

        <StepTabs />

        {exportError && (
          <div className="px-6 py-2 text-xs" style={{ background: "rgba(248,113,113,0.07)", color: "#fca5a5" }}>
            Export failed: {exportError}
          </div>
        )}

        <div className="p-6">
          <p className="text-lg font-black mb-1" style={{ color: "var(--color-text-primary)" }}>Landing Page Mockup</p>
          <p className="text-sm mb-4" style={{ color: "var(--color-text-secondary)" }}>
            AI-generated hero section for <strong>{brandName}</strong> — brand colours, selected logo, and personality applied.
          </p>

          {/* Fixed height 420px — matches mood board */}
          <div style={{ height: 420, position: "relative" }}>
          {visuals.mockup_html ? (
            <div className="rounded-xl overflow-hidden h-full"
              style={{ border: "1px solid rgba(255,255,255,0.12)" }}>
              <iframe
                srcDoc={visuals.mockup_html}
                title={`${brandName} landing page mockup`}
                className="w-full border-0"
                style={{
                  height: "133%", width: "133%",
                  transform: "scale(0.75)", transformOrigin: "top left",
                }}
                sandbox="allow-same-origin"
              />
            </div>
          ) : (
            <div className="rounded-xl p-12 text-center h-full flex items-center justify-center"
              style={{ border: "1px solid rgba(255,255,255,0.07)", background: "rgba(255,255,255,0.02)" }}>
              <p className="text-sm" style={{ color: "var(--color-text-hint)" }}>
                Landing page mockup was not generated — check WATSONX_API_KEY.
              </p>
            </div>
          )}
          </div>

          {/* ── Free-tier disclaimer — card style matching mood board ──────── */}
          <div className="mt-3 px-4 py-2 rounded-lg flex items-start gap-2"
            style={{ background: "rgba(139,92,246,0.08)", border: "1px solid rgba(139,92,246,0.20)" }}>
            <span style={{ color: "var(--color-pulse)", fontSize: "0.85rem", lineHeight: 1 }}>✦</span>
            <p className="text-xs leading-relaxed" style={{ color: "rgba(230,232,240,0.65)" }}>
              <strong style={{ color: "#c4b5fd" }}>Free-tier mockup</strong> — HTML/CSS hero section
              generated by IBM watsonx AI, styled with your brand palette and selected logo.
              High-fidelity Figma-quality exports are planned for a future release.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

// ── Persona sub-component ──────────────────────────────────────────────────
function PersonaCard({ persona }: { persona: BrandPersona }) {
  return (
    <div className="rounded-xl p-5"
      style={{ background: "rgba(139,92,246,0.07)", border: "1px solid rgba(139,92,246,0.22)" }}>
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div><Label>Age</Label><Value>{persona.age}</Value></div>
        <div><Label>Occupation</Label><Value>{persona.occupation}</Value></div>
        <div className="sm:col-span-2"><Label>Brand Voice</Label><Value>{persona.voice}</Value></div>
        {persona.core_values?.length > 0 && (
          <div>
            <Label>Core Values</Label>
            <div className="flex flex-wrap gap-1 mt-1">
              {persona.core_values.map(v => (
                <span key={v} className="text-xs font-bold px-2 py-0.5 rounded-full"
                  style={{ background: "rgba(139,92,246,0.20)", color: "#c4b5fd", border: "1px solid rgba(139,92,246,0.35)" }}>
                  {v}
                </span>
              ))}
            </div>
          </div>
        )}
        {persona.reads?.length > 0 && (
          <div><Label>Reads</Label><Value>{persona.reads.join(", ")}</Value></div>
        )}
        {persona.never_says?.length > 0 && (
          <div className="sm:col-span-2">
            <Label>Would Never Say</Label>
            <div className="flex flex-wrap gap-1 mt-1">
              {persona.never_says.map(w => (
                <span key={w} className="text-xs font-bold px-2 py-0.5 rounded-full"
                  style={{ background: "rgba(248,113,113,0.12)", color: "#fca5a5", border: "1px solid rgba(248,113,113,0.25)" }}>
                  {w}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function Label({ children }: { children: React.ReactNode }) {
  return <p className="text-[10px] font-bold uppercase tracking-widest mb-0.5" style={{ color: "var(--color-text-hint)" }}>{children}</p>;
}
function Value({ children }: { children: React.ReactNode }) {
  return <p className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>{children}</p>;
}
