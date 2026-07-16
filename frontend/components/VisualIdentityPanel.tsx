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
// Parses the user's colour_mood field (e.g. "bright orange, sky blue, white")
// Returns CSS hex values for: bg (always dark), accent, accent2, text (#fff always).
function extractPalette(colorMood: string): { bg: string; accent: string; text: string; accent2: string } {
  const lower = (colorMood || "").toLowerCase();
  const text = "#ffffff"; // text is always white on dark bg

  // ── Background — always dark even if user said "white" ────────────────────
  // "white" = light text, not bg. We keep bg dark for readability.
  let bg = "#0a1628"; // default dark navy
  if (lower.includes("midnight") || lower.includes("void") || lower.includes("black")) bg = "#050810";
  else if (lower.includes("dark blue") || lower.includes("navy")) bg = "#0a1628";
  else if (lower.includes("dark")) bg = "#0d0d1a";

  // ── Map each word → hex colour ────────────────────────────────────────────
  function colourOf(w: string): string | null {
    if (w.includes("orange") || w.includes("coral")) return "#f97316";
    if (w.includes("yellow") || w.includes("gold") || w.includes("amber")) return "#facc15";
    if (w.includes("sky") || w.includes("azure") || w.includes("cerulean")) return "#38bdf8";
    if (w.includes("electric") || w.includes("neon")) return "#3b82f6";
    if (w.includes("blue") && !w.includes("dark")) return "#3b82f6";
    if (w.includes("teal") || w.includes("cyan") || w.includes("aqua")) return "#22d3ee";
    if (w.includes("green") || w.includes("emerald") || w.includes("lime")) return "#10b981";
    if (w.includes("purple") || w.includes("violet") || w.includes("indigo")) return "#8B5CF6";
    if (w.includes("pink") || w.includes("rose") || w.includes("magenta")) return "#ec4899";
    if (w.includes("red") || w.includes("crimson") || w.includes("scarlet")) return "#ef4444";
    return null;
  }

  // Collect all colour tokens from the full string, in order
  const tokens = lower.split(/[\s,/+&]+/).filter(Boolean);
  const colours: string[] = [];
  // Also try bigrams for "bright orange", "sky blue", "electric blue"
  for (let i = 0; i < tokens.length; i++) {
    const bigram = (tokens[i] + " " + (tokens[i+1] || "")).trim();
    const c = colourOf(bigram) || colourOf(tokens[i]);
    if (c && !colours.includes(c)) colours.push(c);
  }
  if (colours.length === 0) colours.push("#f97316", "#38bdf8"); // fallback

  const accent  = colours[0] ?? "#f97316";
  const accent2 = colours[1] ?? "#38bdf8";

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

  // ── CSS Mood Board — matches reference layout:
  //   TOP ROW (60% height): [Colour World panel] | [Typography Scale panel]
  //   BOTTOM ROW (40% height): [Social preview] | [App preview] | [Print preview]
  // All colours driven by pal (user's actual palette).
  function CSSMoodBoardFallback() {
    const name = brandName;
    const tagline = card.tagline || "";
    // Derive a readable font style label from intake.style or intake.personality
    const styleRaw = (intake.style || intake.personality || "").toLowerCase();
    const fontLabel =
      styleRaw.includes("minimal") ? "INTER / CLEAN / MINIMAL" :
      styleRaw.includes("bold") || styleRaw.includes("edgy") ? "IMPACT / BOLD / STRONG" :
      styleRaw.includes("playful") || styleRaw.includes("fun") ? "NUNITO / ROUND / PLAYFUL" :
      styleRaw.includes("elegant") || styleRaw.includes("luxury") ? "CORMORANT / SERIF / ELEGANT" :
      styleRaw.includes("tech") ? "SPACE GROTESK / WIDE / TECHNICAL" :
      "SPACE GROTESK / WIDE / MODERN";
    const fontWeight = styleRaw.includes("minimal") ? "300" : styleRaw.includes("bold") || styleRaw.includes("edgy") ? "900" : "700";

    // Selected logo URI for the showcase in the mood board
    const selectedLogoURI =
      selectedLogoType === "profile" ? visuals.logo_profile :
      selectedLogoType === "app" ? visuals.logo_app :
      selectedLogoType === "business" ? visuals.logo_business :
      (visuals.logo_profile || visuals.logo_app || visuals.logo_business || "");

    return (
      <div className="rounded-xl overflow-hidden h-full" style={{ border: `1px solid ${pal.accent}44`, display: "flex", flexDirection: "column", gap: 2, background: `${pal.bg}` }}>

        {/* ── TOP ROW: 60% height — 2 large panels ─────────────────── */}
        <div style={{ flex: "0 0 60%", display: "grid", gridTemplateColumns: "1fr 1fr", gap: 2 }}>

          {/* LEFT: Colour World — shows name, tagline, palette swatches */}
          <div className="flex flex-col justify-between p-4 relative overflow-hidden"
            style={{ background: `linear-gradient(135deg, ${pal.bg} 0%, ${pal.accent}28 100%)` }}>
            {/* Ambient glow */}
            <div className="absolute inset-0" style={{
              background: `radial-gradient(ellipse at 20% 30%, ${pal.accent}33 0%, transparent 55%)`,
            }} />
            <div>
              <p className="text-[9px] font-black uppercase tracking-widest mb-2 relative"
                style={{ color: `${pal.accent}cc` }}>Colour World</p>
              {/* Logo mark if available */}
              {selectedLogoURI && selectedLogoURI.length > 50 ? (
                /* eslint-disable-next-line @next/next/no-img-element */
                <img src={selectedLogoURI} alt={name}
                  style={{ width: 40, height: 40, objectFit: "contain", marginBottom: 8, borderRadius: 6 }} />
              ) : (
                <div className="mb-2" style={{
                  width: 36, height: 36,
                  background: `linear-gradient(135deg, ${pal.accent}, ${pal.accent2})`,
                  clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
                }} />
              )}
              <p className="font-black text-base leading-none relative" style={{ color: pal.text, fontWeight }}>
                {name}
              </p>
              <p className="text-[10px] italic mt-1 relative" style={{ color: pal.accent }}>
                {tagline}
              </p>
            </div>
            {/* Colour swatches row */}
            <div className="flex gap-1.5 mt-2 relative">
              {[pal.bg, pal.accent, pal.accent2, pal.text].map((c, i) => (
                <div key={i} title={c} style={{
                  width: 20, height: 20, borderRadius: "50%",
                  background: c,
                  border: "1.5px solid rgba(255,255,255,0.18)",
                  flexShrink: 0,
                }} />
              ))}
            </div>
          </div>

          {/* RIGHT: Typography Scale */}
          <div className="flex flex-col justify-between p-4 relative overflow-hidden"
            style={{ background: `${pal.bg}dd` }}>
            <div className="absolute inset-0" style={{
              background: `linear-gradient(to bottom right, ${pal.accent2}18, transparent)`,
            }} />
            <div className="relative">
              <p className="text-[9px] font-black uppercase tracking-widest mb-2"
                style={{ color: `${pal.accent}cc` }}>Typography Scale</p>
              <p className="font-black leading-none" style={{ fontSize: "1.8rem", color: pal.text, fontWeight: "900" }}>Aa</p>
              <p className="text-xs font-black mt-1" style={{ color: pal.text }}>{fontLabel}</p>
              <p className="text-[9px] mt-0.5" style={{ color: `${pal.text}66` }}>Regular · Hint text</p>
            </div>
            {/* CTA button preview */}
            <div className="relative mt-2">
              <div className="inline-flex items-center px-3 py-1 rounded text-[10px] font-black"
                style={{ background: pal.accent, color: pal.bg }}>
                CTA Button
              </div>
            </div>
          </div>
        </div>

        {/* ── BOTTOM ROW: 40% height — 3 logo preview panels ────────── */}
        <div style={{ flex: "0 0 40%", display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 2 }}>

          {/* Social / Profile preview */}
          <div className="flex flex-col items-center justify-center gap-2 relative overflow-hidden"
            style={{ background: `${pal.bg}cc` }}>
            {visuals.logo_profile && visuals.logo_profile.length > 200 ? (
              /* eslint-disable-next-line @next/next/no-img-element */
              <img src={visuals.logo_profile} alt="profile" style={{ width: 64, height: 64, objectFit: "contain", borderRadius: "50%" }} />
            ) : (
              <div className="flex items-center justify-center rounded-full"
                style={{ width: 64, height: 64, background: pal.accent, flexShrink: 0 }}>
                <span style={{ fontWeight: 900, fontSize: "1.6rem", color: pal.bg }}>{initials.slice(0,1)}</span>
              </div>
            )}
            <p style={{ fontSize: "0.65rem", fontWeight: 900, textTransform: "uppercase", letterSpacing: "0.1em", color: `${pal.accent}cc` }}>Social</p>
          </div>

          {/* App Icon preview */}
          <div className="flex flex-col items-center justify-center gap-2 relative overflow-hidden"
            style={{ background: `${pal.bg}cc` }}>
            {visuals.logo_app && visuals.logo_app.length > 200 ? (
              /* eslint-disable-next-line @next/next/no-img-element */
              <img src={visuals.logo_app} alt="app" style={{ width: 64, height: 64, objectFit: "contain", borderRadius: 14 }} />
            ) : (
              <div className="flex items-center justify-center"
                style={{ width: 64, height: 64, background: `linear-gradient(135deg,${pal.accent},${pal.accent2})`, borderRadius: 14, flexShrink: 0 }}>
                <span style={{ fontWeight: 900, fontSize: "1.4rem", color: pal.bg }}>{initials}</span>
              </div>
            )}
            <p style={{ fontSize: "0.65rem", fontWeight: 900, textTransform: "uppercase", letterSpacing: "0.1em", color: `${pal.accent}cc` }}>App</p>
          </div>

          {/* Print / Business preview */}
          <div className="flex flex-col items-center justify-center gap-2 relative overflow-hidden px-3"
            style={{ background: `${pal.bg}cc` }}>
            {visuals.logo_business && visuals.logo_business.length > 200 ? (
              /* eslint-disable-next-line @next/next/no-img-element */
              <img src={visuals.logo_business} alt="business" style={{ width: "80%", maxWidth: 110, height: 48, objectFit: "contain", borderRadius: 4 }} />
            ) : (
              <div className="flex items-center gap-2 px-3 py-2 rounded-lg w-full justify-center"
                style={{ background: "rgba(255,255,255,0.07)", border: `1px solid ${pal.accent}44` }}>
                <div style={{
                  width: 16, height: 16,
                  background: pal.accent,
                  clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
                  flexShrink: 0,
                }} />
                <span style={{ fontWeight: 900, fontSize: "0.85rem", color: pal.text, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{name}</span>
              </div>
            )}
            <p style={{ fontSize: "0.65rem", fontWeight: 900, textTransform: "uppercase", letterSpacing: "0.1em", color: `${pal.accent}cc` }}>Print</p>
          </div>
        </div>
      </div>
    );
  }

  // ── CSS Landing Page — rich, personality-driven, no generic circles ─────────
  function CSSLandingPageFallback() {
    const name = brandName;
    const tagline = card.tagline || "";
    const desc = card.short_desc || "";
    const industry = (intake.industry || "Creative").toUpperCase();
    const personality = (intake.personality || "").toLowerCase();

    // Derive button radius from personality
    const btnRadius = personality.includes("playful") || personality.includes("fun") ? "9999px"
      : personality.includes("minimal") || personality.includes("premium") ? "4px"
      : personality.includes("edgy") || personality.includes("dark") ? "2px"
      : "8px";

    // Derive heading letter-spacing
    const headingLS = personality.includes("minimal") ? "0.06em"
      : personality.includes("bold") || personality.includes("edgy") ? "-0.03em" : "-0.01em";

    // Derive background treatment — diagonal split, no circles
    const heroBg = personality.includes("minimal") || personality.includes("clean")
      ? `linear-gradient(160deg, ${pal.bg} 60%, ${pal.accent}18 100%)`
      : personality.includes("bold") || personality.includes("edgy")
        ? `linear-gradient(135deg, ${pal.bg} 0%, ${pal.bg} 50%, ${pal.accent}22 100%)`
        : `linear-gradient(150deg, ${pal.bg} 0%, ${pal.bg}ee 55%, ${pal.accent2}18 100%)`;

    // Selected logo URI
    const logoURI =
      selectedLogoType === "profile" ? visuals.logo_profile :
      selectedLogoType === "app" ? visuals.logo_app :
      selectedLogoType === "business" ? visuals.logo_business :
      (visuals.logo_profile || visuals.logo_app || visuals.logo_business || "");
    const hasLogo = logoURI && logoURI.startsWith("data:") && logoURI.length > 200;

    return (
      <div className="rounded-xl overflow-hidden h-full" style={{ border: `1px solid ${pal.accent}33` }}>
        <div className="w-full h-full flex flex-col" style={{ background: heroBg }}>

          {/* ── Nav bar ─────────────────────────────────────── */}
          <div className="flex items-center justify-between px-5 py-2.5 border-b"
            style={{ borderColor: `${pal.accent}25`, background: `${pal.bg}cc` }}>
            <div className="flex items-center gap-2">
              {hasLogo ? (
                /* eslint-disable-next-line @next/next/no-img-element */
                <img src={logoURI!} alt={name}
                  style={{ height: 22, width: "auto", objectFit: "contain" }}
                  onError={e => { (e.currentTarget as HTMLImageElement).style.display = "none"; }} />
              ) : (
                <div style={{
                  width: 18, height: 18, flexShrink: 0,
                  background: `linear-gradient(135deg, ${pal.accent}, ${pal.accent2})`,
                  clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
                }} />
              )}
              <span style={{ fontWeight: 900, fontSize: "0.8rem", color: pal.text, letterSpacing: "-0.02em" }}>{name}</span>
            </div>
            <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
              <span style={{ fontSize: "0.6rem", fontWeight: 700, color: `${pal.accent}bb` }}>About</span>
              <span style={{ fontSize: "0.6rem", fontWeight: 700, color: `${pal.accent}bb` }}>Features</span>
              <span style={{
                fontSize: "0.6rem", fontWeight: 900, color: pal.bg,
                background: pal.accent, padding: "3px 10px", borderRadius: btnRadius,
              }}>Get Started</span>
            </div>
          </div>

          {/* ── Hero: left text + right visual accent ───────── */}
          <div className="flex-1 flex" style={{ overflow: "hidden" }}>
            {/* Left column: headline + tagline + description + CTA */}
            <div className="flex flex-col justify-center px-7 py-4" style={{ flex: "0 0 60%" }}>
              {/* Industry eyebrow */}
              <p style={{ fontSize: "0.7rem", fontWeight: 900, textTransform: "uppercase",
                letterSpacing: "0.14em", color: pal.accent, marginBottom: 7 }}>
                {industry}
              </p>
              {/* Brand name — large hero headline */}
              <h1 style={{ fontSize: "2.4rem", fontWeight: 900, lineHeight: 0.95,
                color: "#ffffff", letterSpacing: headingLS, marginBottom: 9 }}>
                {name}
              </h1>
              {/* Accent rule */}
              <div style={{ width: 40, height: 3, background: pal.accent,
                borderRadius: 2, marginBottom: 9 }} />
              {/* Tagline */}
              <p style={{ fontSize: "0.88rem", fontStyle: "italic", color: pal.accent,
                marginBottom: 9, fontWeight: 600 }}>
                &ldquo;{tagline}&rdquo;
              </p>
              {/* Short description */}
              {desc && (
                <p style={{ fontSize: "0.75rem", color: `${pal.text}88`, lineHeight: 1.6,
                  marginBottom: 14, maxWidth: "90%" }}>
                  {desc.slice(0, 90)}{desc.length > 90 ? "…" : ""}
                </p>
              )}
              {/* CTA */}
              <div style={{ display: "inline-flex", gap: 10 }}>
                <span style={{
                  fontSize: "0.78rem", fontWeight: 900, color: pal.bg,
                  background: pal.accent, padding: "7px 16px", borderRadius: btnRadius,
                  boxShadow: `0 3px 12px ${pal.accent}55`,
                }}>
                  Start with {name} →
                </span>
                <span style={{
                  fontSize: "0.78rem", fontWeight: 700, color: pal.accent,
                  border: `1px solid ${pal.accent}55`, padding: "7px 14px",
                  borderRadius: btnRadius,
                }}>
                  Learn More
                </span>
              </div>
            </div>

            {/* Right column: abstract brand mark — geometric, NOT a circle */}
            <div className="flex flex-col items-center justify-center" style={{ flex: "0 0 40%", position: "relative", overflow: "hidden" }}>
              {/* Diagonal stripe accent */}
              <div style={{ position: "absolute", inset: 0,
                background: `linear-gradient(135deg, transparent 40%, ${pal.accent}12 100%)`,
              }} />
              {/* Large geometric mark */}
              <div style={{
                width: 90, height: 90,
                background: `linear-gradient(135deg, ${pal.accent}cc, ${pal.accent2}88)`,
                clipPath: "polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)",
                marginBottom: 12, position: "relative",
              }} />
              {/* Brand initial watermark */}
              <p style={{ fontSize: "3rem", fontWeight: 900, color: `${pal.accent}22`,
                position: "absolute", top: "28%", letterSpacing: "-0.05em" }}>
                {name.slice(0, 2).toUpperCase()}
              </p>
              {/* Personality tag — larger */}
              <p style={{ fontSize: "0.65rem", fontWeight: 900, textTransform: "uppercase",
                letterSpacing: "0.12em", color: `${pal.accent2}bb`, position: "relative", textAlign: "center", paddingInline: 8 }}>
                {intake.personality || "brand identity"}
              </p>
            </div>
          </div>

          {/* ── Feature strip ─────────────────────────────── */}
          <div style={{
            display: "flex", gap: 0, borderTop: `1px solid ${pal.accent}22`,
            background: `${pal.bg}bb`,
          }}>
            {[industry, intake.personality || "Creative", intake.target_audience || "Everyone"].slice(0,3).map((feat, i) => (
              <div key={i} style={{
                flex: 1, padding: "6px 10px", textAlign: "center",
                borderRight: i < 2 ? `1px solid ${pal.accent}22` : "none",
              }}>
                <p style={{ fontSize: "0.65rem", fontWeight: 900, textTransform: "uppercase",
                  letterSpacing: "0.08em", color: `${pal.accent}cc` }}>
                  {feat.slice(0, 20)}
                </p>
              </div>
            ))}
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

                    {/* Always render CSS placeholder first (z-index below) */}
                    <div className="w-full aspect-square relative" style={{ position: "relative" }}>
                      {/* CSS placeholder — always visible */}
                      <div className="w-full h-full absolute inset-0" style={{ zIndex: 0 }}>
                        {placeholder.render()}
                      </div>
                      {/* AI image — overlaid on top; only shown when URI is a valid data URI > 200 chars */}
                      {uri && uri.startsWith("data:") && uri.length > 200 ? (
                        /* eslint-disable-next-line @next/next/no-img-element */
                        <img src={uri} alt={label}
                          className="w-full h-full absolute inset-0 object-cover"
                          style={{ zIndex: 1 }}
                          onError={e => { (e.currentTarget as HTMLImageElement).style.display = "none"; }}
                        />
                      ) : generatingLogos ? (
                        <div className="w-full h-full absolute inset-0 flex items-center justify-center"
                          style={{ background: "rgba(11,15,25,0.85)", zIndex: 1 }}>
                          <p className="text-sm font-bold" style={{ color: "var(--color-pulse)" }}>
                            Generating<span className="nv-dot1">.</span><span className="nv-dot2">.</span><span className="nv-dot3">.</span>
                          </p>
                        </div>
                      ) : null}
                    </div>

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
              /* AI tiles — 2×2 grid; per-tile CSS fallback for any missing/broken tiles.
                 We accept any non-empty mood_board array (even 1 tile) — CSS covers the rest. */
              <div className="rounded-xl overflow-hidden h-full"
                style={{ border: "1px solid rgba(255,255,255,0.10)" }}>
                <div className="grid grid-cols-2 h-full" style={{ gap: 2, background: `${pal.bg}` }}>
                  {/* Always render exactly 4 slots; pad with empty string if fewer tiles returned */}
                  {Array.from({ length: 4 }, (_, i) => visuals.mood_board![i] ?? "").map((tile, i) => {
                    const isValid = tile && tile.startsWith("data:") && tile.length > 200;
                    // Per-tile CSS fallbacks — each has a clearly visible distinct theme.
                    // All backgrounds use solid/strongly-contrasting colours so tiles are
                    // never invisible against the dark grid background.
                    const tileFallbacks = [
                      // Tile 0: Colour World — accent gradient bg, always vivid
                      <div key={`f0`} className="w-full h-full flex flex-col items-center justify-center relative overflow-hidden"
                        style={{ background: `linear-gradient(135deg, ${pal.accent}cc 0%, ${pal.accent2}99 100%)` }}>
                        <div style={{ position:"absolute",inset:0,background:`radial-gradient(ellipse at 30% 35%,rgba(255,255,255,0.18) 0%,transparent 60%)` }} />
                        {/* Palette swatches */}
                        <div style={{ display:"flex",gap:5,marginBottom:10 }}>
                          {[pal.bg, pal.accent, pal.accent2, "#ffffff"].map((c,j)=>(
                            <div key={j} style={{ width:16,height:16,borderRadius:"50%",background:c,border:"2px solid rgba(255,255,255,0.35)",boxShadow:"0 2px 6px rgba(0,0,0,0.4)" }} />
                          ))}
                        </div>
                        <p style={{ fontSize:"0.65rem",fontWeight:900,textTransform:"uppercase",letterSpacing:"0.14em",color:"#ffffff",textShadow:"0 1px 4px rgba(0,0,0,0.6)" }}>Colour World</p>
                        <p style={{ fontSize:"0.5rem",fontWeight:700,color:"rgba(255,255,255,0.75)",marginTop:3,letterSpacing:"0.08em" }}>{pal.accent} · {pal.accent2}</p>
                      </div>,
                      // Tile 1: Brand Identity — semi-dark bg with strong accent overlay so text is legible
                      <div key={`f1`} className="w-full h-full flex flex-col items-center justify-center px-4 relative overflow-hidden"
                        style={{ background:`linear-gradient(160deg, ${pal.bg}f0 0%, ${pal.accent}55 100%)` }}>
                        <div style={{ position:"absolute",inset:0,background:`radial-gradient(ellipse at 50% 60%,${pal.accent2}44,transparent 70%)` }} />
                        <div style={{ width:36,height:36,background:`linear-gradient(135deg,${pal.accent},${pal.accent2})`,clipPath:"polygon(50% 0%,100% 25%,100% 75%,50% 100%,0% 75%,0% 25%)",marginBottom:10,boxShadow:`0 0 16px ${pal.accent}88`,flexShrink:0 }} />
                        <p style={{ fontSize:"1.05rem",fontWeight:900,color:"#ffffff",textAlign:"center",lineHeight:1,textShadow:"0 1px 6px rgba(0,0,0,0.7)" }}>{brandName}</p>
                        <div style={{ width:32,height:2,background:pal.accent,borderRadius:2,margin:"6px 0" }} />
                        <p style={{ fontSize:"0.6rem",color:"rgba(255,255,255,0.85)",fontStyle:"italic",textAlign:"center",lineHeight:1.4 }}>&ldquo;{card.tagline}&rdquo;</p>
                        <p style={{ fontSize:"0.52rem",color:"rgba(255,255,255,0.55)",marginTop:5,textAlign:"center",textTransform:"uppercase",letterSpacing:"0.1em" }}>{intake.personality||"modern brand"}</p>
                      </div>,
                      // Tile 2: Pattern DNA — accent2 tinted bg so hexagons are clearly visible
                      <div key={`f2`} className="w-full h-full relative overflow-hidden"
                        style={{ background:`linear-gradient(160deg,${pal.bg}f0 0%,${pal.accent2}55 100%)` }}>
                        {/* Grid of hexagonal shapes with high-contrast colours */}
                        {Array.from({length:16}).map((_,j)=>(
                          <div key={j} style={{
                            position:"absolute",
                            width:22,height:22,
                            background:j%2===0?`${pal.accent}cc`:`${pal.accent2}99`,
                            clipPath:"polygon(50% 0%,100% 25%,100% 75%,50% 100%,0% 75%,0% 25%)",
                            left:`${(j%4)*23+3}%`,
                            top:`${Math.floor(j/4)*22+6}%`,
                            transform:`rotate(${j*22}deg)`,
                            boxShadow:`0 0 8px ${j%2===0?pal.accent:pal.accent2}66`,
                          }} />
                        ))}
                        <div style={{ position:"absolute",bottom:8,left:10,right:10 }}>
                          <div style={{ height:1,background:`${pal.accent}66`,marginBottom:5 }} />
                          <p style={{ fontSize:"0.55rem",fontWeight:900,textTransform:"uppercase",letterSpacing:"0.14em",color:"rgba(255,255,255,0.85)",textShadow:"0 1px 4px rgba(0,0,0,0.8)" }}>Pattern DNA</p>
                        </div>
                      </div>,
                      // Tile 3: Typography + CTA — accent tinted bg for maximum contrast
                      <div key={`f3`} className="w-full h-full flex flex-col justify-between p-3 relative overflow-hidden"
                        style={{ background:`linear-gradient(145deg,${pal.bg}f0 0%,${pal.accent}44 100%)` }}>
                        <div>
                          <p style={{ fontSize:"0.55rem",fontWeight:900,textTransform:"uppercase",letterSpacing:"0.14em",color:"rgba(255,255,255,0.7)",marginBottom:5 }}>Typography</p>
                          <p style={{ fontSize:"2rem",fontWeight:900,color:"#ffffff",lineHeight:1,textShadow:`0 0 20px ${pal.accent}88` }}>Aa</p>
                          <p style={{ fontSize:"0.55rem",fontWeight:700,color:"rgba(255,255,255,0.80)",marginTop:4,letterSpacing:"0.06em" }}>{(intake.style||"SPACE GROTESK").toUpperCase()}</p>
                          <p style={{ fontSize:"0.5rem",color:"rgba(255,255,255,0.45)",marginTop:2 }}>Regular · Bold · Black</p>
                        </div>
                        {/* CTA pill — solid accent bg, very visible */}
                        <div style={{ display:"inline-flex",alignItems:"center",padding:"5px 12px",background:pal.accent,borderRadius:6,boxShadow:`0 2px 10px ${pal.accent}66`,width:"fit-content" }}>
                          <p style={{ fontSize:"0.6rem",fontWeight:900,color:pal.bg,whiteSpace:"nowrap" }}>CTA Button</p>
                        </div>
                      </div>,
                    ];
                    return (
                      <div key={i} className="overflow-hidden" style={{ background: pal.bg, position:"relative" }}>
                        {isValid ? (
                          /* eslint-disable-next-line @next/next/no-img-element */
                          <img src={tile} alt={`Tile ${i+1}`}
                            className="w-full h-full object-cover"
                            style={{ display:"block", position:"relative", zIndex:1 }}
                            onError={e => { (e.currentTarget as HTMLImageElement).style.display = "none"; }}
                          />
                        ) : (
                          tileFallbacks[i]
                        )}
                      </div>
                    );
                  })}
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
  // Smart mockup validation: reject AI HTML that is too short, has no H1/heading,
  // or is missing the brand name (means the AI produced a generic non-brand page).
  const hasMockup = (() => {
    const html = visuals.mockup_html;
    if (!html || html.trim().length < 200) return false;
    const lower = html.toLowerCase();
    // Must contain the brand name (case-insensitive) and at least one heading tag
    const hasBrandName = lower.includes(brandName.toLowerCase());
    const hasHeading = lower.includes("<h1") || lower.includes("<h2");
    return hasBrandName && hasHeading;
  })();

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

          <div style={{ height: 420, position: "relative", overflow: "hidden" }}>
          {hasMockup ? (() => {
            // Strip 100vh / min-height:100vh from AI HTML to prevent iframe scrollbar.
            // The scaled iframe (0.75×) already handles viewport fitting; any 100vh
            // in the AI HTML creates overflow inside the 420px container.
            const cleanMockupHtml = (visuals.mockup_html ?? "")
              .replace(/min-height\s*:\s*100vh/gi, "min-height: 100%")
              .replace(/(?<![a-z-])height\s*:\s*100vh/gi, "height: 100%")
              // Also inject overflow:hidden on html+body as a safety net
              .replace(/<html([^>]*)>/i, "<html$1 style=\"overflow:hidden\">")
              .replace(/<body([^>]*)>/i, "<body$1 style=\"overflow:hidden;margin:0\">");
            return (
            <div className="rounded-xl overflow-hidden h-full"
              style={{ border: "1px solid rgba(255,255,255,0.12)" }}>
              <iframe
                srcDoc={cleanMockupHtml}
                title={`${brandName} landing page mockup`}
                className="w-full border-0"
                style={{
                  height: "133%", width: "133%",
                  transform: "scale(0.75)", transformOrigin: "top left",
                  overflow: "hidden",
                }}
                sandbox="allow-same-origin"
              />
            </div>
            );
          })()
          : (
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
