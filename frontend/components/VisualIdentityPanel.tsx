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

// ── Colour extraction from intake ─────────────────────────────────────────────
// Parses the user's colour_mood field and returns CSS-usable values for fallback renders.
function extractPalette(colorMood: string): { bg: string; accent: string; text: string; accent2: string } {
  const lower = (colorMood || "").toLowerCase();
  let bg = "#0d1b2a";
  let accent = "#22d3ee";
  let accent2 = "#8B5CF6";
  const text = "#ffffff";

  if (lower.includes("dark") || lower.includes("black") || lower.includes("void")) bg = "#050810";
  else if (lower.includes("navy") || lower.includes("dark blue")) bg = "#0a1628";
  else if (lower.includes("midnight")) bg = "#0c0c1e";

  if (lower.includes("neon yellow") || lower.includes("yellow")) accent = "#facc15";
  else if (lower.includes("electric blue") || lower.includes("blue")) accent = "#3b82f6";
  else if (lower.includes("teal") || lower.includes("cyan")) accent = "#22d3ee";
  else if (lower.includes("gold") || lower.includes("amber")) accent = "#f59e0b";
  else if (lower.includes("coral") || lower.includes("orange")) accent = "#f97316";
  else if (lower.includes("green") || lower.includes("emerald")) accent = "#10b981";
  else if (lower.includes("purple") || lower.includes("violet")) accent = "#8B5CF6";
  else if (lower.includes("red") || lower.includes("crimson")) accent = "#ef4444";
  else if (lower.includes("pink") || lower.includes("rose")) accent = "#ec4899";

  // Second accent — look for 2nd colour keyword
  const words = lower.split(/[,\s]+/);
  let foundFirst = false;
  for (const w of words) {
    const hit =
      w.includes("yellow") ? "#facc15" :
      w.includes("blue") ? "#3b82f6" :
      w.includes("teal") || w.includes("cyan") ? "#22d3ee" :
      w.includes("gold") || w.includes("amber") ? "#f59e0b" :
      w.includes("coral") || w.includes("orange") ? "#f97316" :
      w.includes("green") || w.includes("emerald") ? "#10b981" :
      w.includes("purple") || w.includes("violet") ? "#8B5CF6" :
      w.includes("red") || w.includes("crimson") ? "#ef4444" :
      w.includes("pink") || w.includes("rose") ? "#ec4899" : null;
    if (hit) {
      if (!foundFirst) { foundFirst = true; }
      else { accent2 = hit; break; }
    }
  }

  return { bg, accent, text, accent2 };
}

export default function VisualIdentityPanel({
  brandName, card, intake, sessionId, visuals,
  onBack, onStartOver, onRegenerateLogos, onRegenerateMoodboard,
  generatingLogos, generatingMoodboard,
}: VisualIdentityPanelProps) {
  const [visualStep, setVisualStep] = useState<VisualStep>("logos");
  const [selectedLogoType, setSelectedLogoType] = useState<string>("");
  const [exporting, setExporting]   = useState(false);
  const [exportError, setExportError] = useState("");

  // Derived user palette for CSS fallbacks
  const pal = useMemo(() => extractPalette(intake.color_mood ?? ""), [intake.color_mood]);

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

  // Derive initials for CSS logo placeholders (max 2 chars)
  const initials = useMemo(() => {
    const parts = brandName.trim().split(/\s+/);
    if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
    return brandName.slice(0, 2).toUpperCase();
  }, [brandName]);

  // ── CSS logo placeholders — adapt to user's brand palette ───────────────────
  const logoItems = useMemo(() => [
    {
      key: "profile",
      label: "Profile / Social",
      uri: visuals.logo_profile,
      hint: "1:1 square — Twitter, Instagram, LinkedIn",
      placeholder: {
        render: () => (
          <div className="w-full aspect-square flex flex-col items-center justify-center relative overflow-hidden"
            style={{ background: `linear-gradient(160deg, ${pal.bg} 0%, ${pal.bg}cc 50%, ${pal.bg}ee 100%)` }}>
            {/* Outer ring in accent colour */}
            <div className="absolute" style={{
              width: "58%", height: "58%",
              border: `2px solid ${pal.accent}66`,
              borderRadius: "50%",
              background: `${pal.accent}11`,
            }} />
            {/* Hexagonal accent shape */}
            <div className="absolute" style={{
              width: "42%", height: "42%",
              background: `linear-gradient(135deg, ${pal.accent}99 0%, ${pal.accent2}55 100%)`,
              clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
            }} />
            {/* Brand initial */}
            <span className="relative z-10 font-black select-none"
              style={{ fontSize: "clamp(1.5rem,6vw,2.2rem)", color: pal.text, letterSpacing: "-0.04em", textShadow: `0 0 20px ${pal.accent}cc` }}>
              {initials.slice(0, 1)}
            </span>
            {/* Corner accent dots */}
            {[0, 90, 180, 270].map(deg => (
              <div key={deg} className="absolute" style={{
                width: 6, height: 6, borderRadius: "50%",
                background: pal.accent,
                transform: `rotate(${deg}deg) translateY(-36%)`,
                opacity: 0.8,
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
        render: () => (
          <div className="w-full aspect-square flex items-center justify-center relative overflow-hidden"
            style={{ background: `linear-gradient(145deg, ${pal.bg} 0%, ${pal.bg}dd 40%, ${pal.bg}bb 100%)` }}>
            {/* Ambient glow blobs */}
            <div className="absolute" style={{
              width: "80%", height: "80%", borderRadius: "50%",
              background: `radial-gradient(circle, ${pal.accent}44 0%, transparent 70%)`,
              top: "10%", left: "10%",
            }} />
            <div className="absolute" style={{
              width: "50%", height: "50%", borderRadius: "50%",
              background: `radial-gradient(circle, ${pal.accent2}33 0%, transparent 70%)`,
              bottom: "5%", right: "5%",
            }} />
            {/* Glassmorphic card */}
            <div className="relative flex flex-col items-center justify-center"
              style={{
                width: "62%", height: "62%",
                background: "rgba(255,255,255,0.08)",
                borderRadius: "22%",
                border: `1.5px solid ${pal.accent}44`,
                backdropFilter: "blur(8px)",
                boxShadow: `0 8px 32px ${pal.accent}44`,
              }}>
              <span className="font-black select-none"
                style={{ fontSize: "clamp(1.2rem,5vw,1.8rem)", color: pal.text, letterSpacing: "-0.03em" }}>
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
        render: () => (
          /* Horizontal wordmark on brand bg (not generic white) */
          <div className="w-full aspect-square flex flex-col items-center justify-center gap-2 px-6 relative overflow-hidden"
            style={{ background: `linear-gradient(160deg, ${pal.bg} 0%, ${pal.bg}ee 100%)` }}>
            <div className="absolute" style={{
              width: "55%", height: "55%", borderRadius: "50%",
              background: `radial-gradient(circle, ${pal.accent}18 0%, transparent 70%)`,
            }} />
            {/* Icon + wordmark row */}
            <div className="flex items-center gap-2.5 relative">
              <div style={{
                width: 28, height: 28,
                background: `linear-gradient(135deg, ${pal.accent} 0%, ${pal.accent2} 100%)`,
                clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
                flexShrink: 0,
              }} />
              <span className="font-black tracking-tight select-none"
                style={{ fontSize: "clamp(1rem,4vw,1.4rem)", color: pal.text, letterSpacing: "-0.02em" }}>
                {brandName}
              </span>
            </div>
            {/* Divider */}
            <div style={{ width: "70%", height: 1, background: `linear-gradient(90deg, transparent, ${pal.accent}66, transparent)` }} />
            {/* Tagline */}
            <span className="text-xs uppercase tracking-widest select-none text-center"
              style={{ color: `${pal.accent}aa`, fontSize: "0.6rem", letterSpacing: "0.15em", maxWidth: "80%" }}>
              {card.tagline || "Brand Identity"}
            </span>
          </div>
        ),
      },
    },
  // eslint-disable-next-line react-hooks/exhaustive-deps
  ], [visuals.logo_profile, visuals.logo_app, visuals.logo_business, brandName, initials, pal]);

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

  // ── CSS Mood Board Fallback — branded, uses user's palette ─────────────────
  function CSSMoodBoardFallback() {
    const name = brandName;
    const tagline = card.tagline || "";
    const bg2 = pal.bg + "ee";
    return (
      <div className="rounded-xl overflow-hidden h-full" style={{ border: `1px solid ${pal.accent}33` }}>
        <div className="grid grid-cols-2 h-full" style={{ gap: 2, background: "rgba(0,0,0,0.3)" }}>
          {/* Panel 1: Colour atmosphere */}
          <div className="flex flex-col items-center justify-center relative overflow-hidden"
            style={{ background: `linear-gradient(135deg, ${pal.bg} 0%, ${pal.accent}22 100%)` }}>
            <div className="absolute inset-0" style={{
              background: `radial-gradient(ellipse at 40% 40%, ${pal.accent}44 0%, transparent 60%)`,
            }} />
            <div className="absolute" style={{
              width: 80, height: 80, borderRadius: "50%",
              background: `radial-gradient(circle, ${pal.accent}66 0%, transparent 70%)`,
              top: "20%", left: "25%",
            }} />
            <div className="absolute" style={{
              width: 50, height: 50, borderRadius: "50%",
              background: `radial-gradient(circle, ${pal.accent2}55 0%, transparent 70%)`,
              bottom: "25%", right: "20%",
            }} />
            <p className="relative z-10 text-xs font-black uppercase tracking-widest text-center px-3"
              style={{ color: pal.accent }}>Colour Atmosphere</p>
          </div>

          {/* Panel 2: Brand typography */}
          <div className="flex flex-col items-center justify-center px-4 text-center relative overflow-hidden"
            style={{ background: bg2 }}>
            <div className="absolute inset-0" style={{
              background: `linear-gradient(to bottom right, ${pal.accent2}22, transparent)`,
            }} />
            <div className="relative">
              {/* Geometric mark */}
              <div className="mx-auto mb-3" style={{
                width: 40, height: 40,
                background: `linear-gradient(135deg, ${pal.accent}, ${pal.accent2})`,
                clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
              }} />
              <p className="text-lg font-black leading-none" style={{ color: pal.text }}>{name}</p>
              <p className="mt-1 text-xs italic" style={{ color: pal.accent }}>&ldquo;{tagline}&rdquo;</p>
            </div>
          </div>

          {/* Panel 3: Geometric pattern */}
          <div className="relative overflow-hidden" style={{ background: pal.bg }}>
            {/* Repeating diamond/hex grid */}
            {Array.from({ length: 12 }).map((_, i) => (
              <div key={i} className="absolute" style={{
                width: 36, height: 36,
                background: `${pal.accent}${i % 3 === 0 ? "44" : "22"}`,
                clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
                left: `${(i % 4) * 26}%`,
                top: `${Math.floor(i / 4) * 32 + 10}%`,
                transform: `rotate(${i * 15}deg)`,
              }} />
            ))}
            <div className="absolute inset-0 flex items-end justify-start p-3">
              <p className="text-xs font-black uppercase tracking-widest" style={{ color: `${pal.accent}bb` }}>
                Geometric DNA
              </p>
            </div>
          </div>

          {/* Panel 4: Motion/energy */}
          <div className="flex flex-col justify-center px-4 relative overflow-hidden"
            style={{ background: `linear-gradient(to bottom, ${pal.bg}, ${pal.accent2}22)` }}>
            {/* Diagonal lines */}
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="absolute" style={{
                height: 1,
                width: "140%",
                background: `${i % 2 === 0 ? pal.accent : pal.accent2}${i % 3 === 0 ? "55" : "33"}`,
                top: `${10 + i * 11}%`,
                left: "-20%",
                transform: "rotate(-12deg)",
              }} />
            ))}
            <div className="relative">
              <p className="text-xs font-black uppercase tracking-widest mb-1" style={{ color: pal.accent }}>
                Brand Energy
              </p>
              <p className="text-xs" style={{ color: `${pal.text}88` }}>
                {intake.personality || "modern, distinctive"}
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // ── CSS Landing Page Fallback — full hero section using brand palette ───────
  function CSSLandingPageFallback() {
    const name = brandName;
    const tagline = card.tagline || "";
    const desc = card.short_desc || "The brand built for the future.";
    const industry = intake.industry || "Creative";
    return (
      <div className="rounded-xl overflow-hidden h-full" style={{ border: `1px solid ${pal.accent}33` }}>
        <div className="w-full h-full flex flex-col" style={{ background: `linear-gradient(135deg, ${pal.bg} 0%, ${pal.bg}ee 40%, ${pal.accent}18 100%)` }}>
          {/* Nav bar */}
          <div className="flex items-center justify-between px-5 py-3 border-b" style={{ borderColor: `${pal.accent}33` }}>
            <div className="flex items-center gap-2">
              {/* Small logo mark */}
              <div style={{
                width: 22, height: 22,
                background: `linear-gradient(135deg, ${pal.accent}, ${pal.accent2})`,
                clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
                flexShrink: 0,
              }} />
              <span className="font-black text-sm tracking-tight" style={{ color: pal.text }}>{name}</span>
            </div>
            <div className="flex gap-3">
              <span className="text-xs font-bold" style={{ color: `${pal.accent}cc` }}>About</span>
              <span className="text-xs font-bold px-2 py-0.5 rounded"
                style={{ background: pal.accent, color: pal.bg, fontSize: "0.65rem" }}>
                Get Started
              </span>
            </div>
          </div>

          {/* Hero section */}
          <div className="flex-1 flex flex-col items-start justify-center px-8 py-4 relative">
            {/* Decorative circle */}
            <div className="absolute right-8 top-1/2 -translate-y-1/2 opacity-20" style={{
              width: 120, height: 120, borderRadius: "50%",
              background: `radial-gradient(circle, ${pal.accent} 0%, transparent 70%)`,
            }} />
            <div className="absolute right-12 top-[40%] opacity-10" style={{
              width: 70, height: 70,
              background: `${pal.accent2}`,
              clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
            }} />

            <p className="text-[10px] font-black uppercase tracking-widest mb-2" style={{ color: pal.accent }}>
              {industry} · Brand Identity
            </p>
            <h1 className="text-3xl font-black leading-none mb-2 tracking-tight"
              style={{ color: pal.text, textShadow: `0 0 30px ${pal.accent}55` }}>
              {name}
            </h1>
            <p className="text-sm italic mb-3" style={{ color: pal.accent }}>
              &ldquo;{tagline}&rdquo;
            </p>
            <p className="text-xs mb-5 max-w-[55%] leading-relaxed" style={{ color: `${pal.text}aa` }}>
              {desc}
            </p>
            <button type="button" className="px-4 py-2 text-xs font-black rounded-lg"
              style={{ background: pal.accent, color: pal.bg, boxShadow: `0 4px 16px ${pal.accent}55` }}>
              {name} — Get Started →
            </button>
          </div>

          {/* Footer strip */}
          <div className="px-5 py-2 border-t text-center" style={{ borderColor: `${pal.accent}22` }}>
            <p className="text-[9px]" style={{ color: `${pal.accent}66` }}>
              AI-generated hero concept · Powered by IBM watsonx
            </p>
          </div>
        </div>
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

          {/* Header */}
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
                      /* Rich CSS concept placeholder — uses user's brand palette */
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

            {/* Note about business logo blank */}
            {!visuals.logo_business && !generatingLogos && (
              <div className="mt-3 px-4 py-2 rounded-lg flex items-start gap-2"
                style={{ background: "rgba(245,158,11,0.07)", border: "1px solid rgba(245,158,11,0.22)" }}>
                <span style={{ color: "#f59e0b", fontSize: "0.85rem" }}>ℹ</span>
                <p className="text-xs leading-relaxed" style={{ color: "rgba(245,158,11,0.80)" }}>
                  <strong>Business / Print</strong> — AI logo not yet generated. The concept above is a
                  branded CSS preview using your colour palette. Click <em>New Logos</em> to request
                  AI-generated SVG logos for all three styles.
                </p>
              </div>
            )}

            {/* Disclaimer — no Gemini/Figma mentions */}
            <div className="mt-3 px-4 py-2 rounded-lg flex items-start gap-2"
              style={{ background: "rgba(139,92,246,0.08)", border: "1px solid rgba(139,92,246,0.20)" }}>
              <span style={{ color: "var(--color-pulse)", fontSize: "0.85rem", lineHeight: 1 }}>✦</span>
              <p className="text-xs leading-relaxed" style={{ color: "rgba(230,232,240,0.65)" }}>
                <strong style={{ color: "#c4b5fd" }}>AI-generated brand visuals</strong> — SVG logos generated
                by IBM watsonx AI using your brand palette and personality.
                Branded CSS concepts shown when AI visuals are processing.
                Advanced AI tools will be adapted for a future release.
              </p>
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
  // STEP 2 — MOOD BOARD
  // ══════════════════════════════════════════════════════════════════════════
  if (visualStep === "moodboard") {
    const hasMoodBoard = visuals.mood_board && visuals.mood_board.length > 0;

    return (
      <div className="w-full max-w-2xl lg:max-w-4xl xl:max-w-5xl mx-auto">
        <div className="rounded-2xl border overflow-hidden"
          style={{ background: "rgba(18,22,42,0.97)", borderColor: "rgba(99,210,255,0.22)", boxShadow: "0 24px 64px rgba(5,7,15,0.7)" }}>

          {/* Header */}
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
              Visual world of <strong>{brandName}</strong> — brand colours, logo DNA, and personality expressed as design panels.
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

            {/* ── Mood board area ─────────────────────────────────────────── */}
            <div style={{ height: 420, position: "relative" }}>
            {generatingMoodboard ? (
              /* Full animated loading state */
              <div className="rounded-xl h-full flex flex-col items-center justify-center gap-4"
                style={{ border: "1px solid rgba(139,92,246,0.18)", background: "rgba(139,92,246,0.06)" }}>
                <div className="flex gap-3">
                  {[pal.accent, pal.accent2, pal.accent, pal.accent2].map((c, i) => (
                    <div key={i} className="rounded-sm" style={{
                      width: 14, height: 14,
                      background: c,
                      opacity: 0.7,
                      animation: `bounce 1.2s ease-in-out ${i * 0.15}s infinite`,
                    }} />
                  ))}
                </div>
                <p className="text-sm font-black" style={{ color: "var(--color-pulse)" }}>
                  Generating mood board<span className="nv-dot1">.</span><span className="nv-dot2">.</span><span className="nv-dot3">.</span>
                </p>
                <p className="text-xs text-center px-8" style={{ color: "var(--color-text-hint)" }}>
                  IBM watsonx AI is crafting 4 brand tiles using your palette: <em>{intake.color_mood || "brand colours"}</em>
                </p>
              </div>
            ) : hasMoodBoard ? (
              /* AI-generated SVG/PNG tiles in a 2×2 grid */
              <div className="rounded-xl overflow-hidden h-full"
                style={{ border: "1px solid rgba(255,255,255,0.10)" }}>
                <div className="grid grid-cols-2 h-full" style={{ gap: 2, background: "rgba(255,255,255,0.04)" }}>
                  {visuals.mood_board!.slice(0, 4).map((tile, i) => (
                    <div key={i} className="overflow-hidden rounded-sm" style={{ background: pal.bg }}>
                      {/* eslint-disable-next-line @next/next/no-img-element */}
                      <img
                        src={tile}
                        alt={`Mood board tile ${i + 1}`}
                        className="w-full h-full object-cover"
                        style={{ display: "block" }}
                      />
                    </div>
                  ))}
                  {Array.from({ length: Math.max(0, 4 - (visuals.mood_board?.length ?? 0)) }).map((_, i) => (
                    <div key={`empty-${i}`} className="flex items-center justify-center"
                      style={{ background: "rgba(139,92,246,0.06)", border: "1px solid rgba(139,92,246,0.12)" }}>
                      <p className="text-xs" style={{ color: "rgba(139,92,246,0.40)" }}>Loading…</p>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              /* CSS fallback — always shows a branded mood board when AI result is empty */
              <CSSMoodBoardFallback />
            )}
            </div>

            {/* Disclaimer — no Gemini/Figma */}
            <div className="mt-3 px-4 py-2 rounded-lg flex items-start gap-2"
              style={{ background: "rgba(139,92,246,0.08)", border: "1px solid rgba(139,92,246,0.20)" }}>
              <span style={{ color: "var(--color-pulse)", fontSize: "0.85rem", lineHeight: 1 }}>✦</span>
              <p className="text-xs leading-relaxed" style={{ color: "rgba(230,232,240,0.65)" }}>
                {hasMoodBoard
                  ? <><strong style={{ color: "#c4b5fd" }}>AI-generated mood board</strong> — SVG brand tiles created by IBM watsonx AI using your colour palette ({intake.color_mood || "brand palette"}) and personality.</>
                  : <><strong style={{ color: "#c4b5fd" }}>Branded CSS concept</strong> — mood board built from your palette, tagline, and brand DNA. Advanced AI image generation will be adapted for a future release.</>
                }
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
  const hasMockup = Boolean(visuals.mockup_html && visuals.mockup_html.trim().length > 20);

  return (
    <div className="w-full max-w-2xl lg:max-w-4xl xl:max-w-5xl mx-auto">
      <div className="rounded-2xl border overflow-hidden"
        style={{ background: "rgba(18,22,42,0.97)", borderColor: "rgba(99,210,255,0.22)", boxShadow: "0 24px 64px rgba(5,7,15,0.7)" }}>

        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b"
          style={{ borderColor: "rgba(255,255,255,0.07)", background: "rgba(14,18,32,0.60)" }}>
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

          <div style={{ height: 420, position: "relative" }}>
          {hasMockup ? (
            <div className="rounded-xl overflow-hidden h-full"
              style={{ border: "1px solid rgba(255,255,255,0.12)" }}>
              <iframe
                srcDoc={visuals.mockup_html!}
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
            /* CSS fallback hero — always shown when AI mockup is empty */
            <CSSLandingPageFallback />
          )}
          </div>

          {/* Disclaimer — no Figma mention */}
          <div className="mt-3 px-4 py-2 rounded-lg flex items-start gap-2"
            style={{ background: "rgba(139,92,246,0.08)", border: "1px solid rgba(139,92,246,0.20)" }}>
            <span style={{ color: "var(--color-pulse)", fontSize: "0.85rem", lineHeight: 1 }}>✦</span>
            <p className="text-xs leading-relaxed" style={{ color: "rgba(230,232,240,0.65)" }}>
              {hasMockup
                ? <><strong style={{ color: "#c4b5fd" }}>AI-generated mockup</strong> — HTML/CSS hero section built by IBM watsonx AI with your brand palette and selected logo.</>
                : <><strong style={{ color: "#c4b5fd" }}>Branded CSS concept</strong> — hero layout built from your brand inputs. Advanced AI tools will be adapted for a future release.</>
              }
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
