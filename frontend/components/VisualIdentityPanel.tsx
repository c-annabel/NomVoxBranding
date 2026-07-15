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
  /** Called when the user picks a logo and advances — triggers targeted mood board regen */
  onRegenerateMoodboard: (logoKey: string, logoStyle: string) => void;
  generatingLogos?: boolean;
  generatingMoodboard?: boolean;
}

const BACKEND = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// Map logo key → human-readable style string passed to mood board prompt
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

  // CSS concept placeholders shown when Imagen 3 is unavailable (no GOOGLE_AI_API_KEY).
  // Each uses a distinct visual style so user can genuinely pick a direction.
  const logoItems = useMemo(() => [
    {
      key: "profile",
      label: "Profile / Social",
      uri: visuals.logo_profile,
      hint: "1:1 square — Twitter, Instagram, LinkedIn",
      placeholder: {
        styleName: "Minimal Wordmark",
        style: { background: "linear-gradient(135deg, #0f0f1a 0%, #1a0a2e 100%)" } as React.CSSProperties,
        textColor: "#c4b5fd",
      },
    },
    {
      key: "app",
      label: "App Icon",
      uri: visuals.logo_app,
      hint: "Rounded square — app stores",
      placeholder: {
        styleName: "Bold Icon",
        style: { background: "linear-gradient(135deg, #0a1628 0%, #0d2040 100%)" } as React.CSSProperties,
        textColor: "#67e8f9",
      },
    },
    {
      key: "business",
      label: "Business / Print",
      uri: visuals.logo_business,
      hint: "16:9 horizontal — business cards, letterhead",
      placeholder: {
        styleName: "Clean Lockup",
        style: { background: "linear-gradient(135deg, #f5f0e8 0%, #ede4d3 100%)" } as React.CSSProperties,
        textColor: "#1a0a2e",
      },
    },
  ], [visuals.logo_profile, visuals.logo_app, visuals.logo_business, brandName]); // eslint-disable-line react-hooks/exhaustive-deps

  // ── Shared header (no logo-as-button; step label is large and prominent) ────
  function Header({ step }: { step: string }) {
    return (
      <div className="flex items-center justify-between px-6 py-4 border-b"
        style={{ borderColor: "rgba(255,255,255,0.07)", background: "rgba(14,18,32,0.60)" }}>
        <div>
          <p className="text-xs font-black uppercase tracking-widest mb-0.5"
            style={{ color: "var(--color-pulse)" }}>{step}</p>
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
    );
  }

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
          <Header step="Visual Identity — Step 1/3" />
          <StepTabs />

          {exportError && (
            <div className="px-6 py-2 text-xs" style={{ background: "rgba(248,113,113,0.07)", color: "#fca5a5" }}>
              Export failed: {exportError}
            </div>
          )}

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

            {/* 3 logo columns — always selectable; CSS placeholder when Imagen unavailable */}
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
                      // CSS concept placeholder — always shows, always selectable
                      <div className="w-full aspect-square flex flex-col items-center justify-center gap-3 p-6"
                        style={placeholder.style}>
                        <div className="font-black text-3xl tracking-tight" style={{ color: placeholder.textColor }}>
                          {brandName}
                        </div>
                        <div className="text-xs font-bold uppercase tracking-widest opacity-70"
                          style={{ color: placeholder.textColor }}>
                          {placeholder.styleName}
                        </div>
                        <div className="w-12 h-0.5 rounded-full opacity-40"
                          style={{ background: placeholder.textColor }} />
                        <div className="text-[10px] opacity-50 text-center" style={{ color: placeholder.textColor }}>
                          Concept direction
                        </div>
                      </div>
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

            {/* CTA to next step — triggers targeted mood board regen with logo context */}
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
    return (
      <div className="w-full max-w-2xl lg:max-w-4xl xl:max-w-5xl mx-auto">
        <div className="rounded-2xl border overflow-hidden"
          style={{ background: "rgba(18,22,42,0.97)", borderColor: "rgba(99,210,255,0.22)", boxShadow: "0 24px 64px rgba(5,7,15,0.7)" }}>
          <Header step="Visual Identity — Step 2/3" />
          <StepTabs />

          <div className="p-6">
            <p className="text-lg font-black mb-1" style={{ color: "var(--color-text-primary)" }}>Mood Board</p>
            <p className="text-sm mb-6" style={{ color: "var(--color-text-secondary)" }}>
              Visual world of <strong>{brandName}</strong> — colour palette, textures, and atmosphere.
            </p>

            {generatingMoodboard ? (
              <div className="grid grid-cols-2 gap-3">
                {[1,2,3,4].map(i => (
                  <div key={i} className="rounded-xl aspect-square flex items-center justify-center"
                    style={{ border: "1px solid rgba(139,92,246,0.18)", background: "rgba(139,92,246,0.06)" }}>
                    <p className="text-sm font-bold" style={{ color: "var(--color-pulse)" }}>
                      Generating<span className="nv-dot1">.</span><span className="nv-dot2">.</span><span className="nv-dot3">.</span>
                    </p>
                  </div>
                ))}
              </div>
            ) : visuals.mood_board && visuals.mood_board.length > 0 ? (
              <div className="grid grid-cols-2 gap-3">
                {visuals.mood_board.map((uri, i) => (
                  <div key={i} className="rounded-xl overflow-hidden aspect-square"
                    style={{ border: "1px solid rgba(255,255,255,0.08)" }}>
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={uri} alt={`Mood board panel ${i + 1}`} className="w-full h-full object-cover" />
                  </div>
                ))}
              </div>
            ) : (
              <div className="rounded-xl p-12 text-center"
                style={{ border: "1px solid rgba(255,255,255,0.07)", background: "rgba(255,255,255,0.02)" }}>
                <p className="text-sm" style={{ color: "var(--color-text-hint)" }}>
                  Mood board not generated yet — select a logo on step 1 and click "Next: Mood Board".
                </p>
              </div>
            )}

            {/* Brand persona card */}
            {visuals.persona && (
              <div className="mt-8">
                <p className="text-base font-black mb-1" style={{ color: "var(--color-text-primary)" }}>Brand Persona</p>
                <p className="text-sm mb-3" style={{ color: "var(--color-text-secondary)" }}>
                  {brandName} as a person
                </p>
                <PersonaCard persona={visuals.persona} />
              </div>
            )}

            <div className="mt-8 flex justify-between">
              <button type="button" onClick={() => setVisualStep("logos")}
                className="px-6 py-2 rounded-lg text-sm font-bold transition-all"
                style={{ border: "1px solid rgba(99,210,255,0.30)", color: "var(--color-signal)", background: "transparent" }}>
                ← Back to Logos
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
  // STEP 3 — LANDING PAGE MOCKUP (always visible, no collapse)
  // ══════════════════════════════════════════════════════════════════════════
  return (
    <div className="w-full max-w-2xl lg:max-w-4xl xl:max-w-5xl mx-auto">
      <div className="rounded-2xl border overflow-hidden"
        style={{ background: "rgba(18,22,42,0.97)", borderColor: "rgba(99,210,255,0.22)", boxShadow: "0 24px 64px rgba(5,7,15,0.7)" }}>
        <Header step="Visual Identity — Step 3/3" />
        <StepTabs />

        <div className="p-6">
          <p className="text-lg font-black mb-1" style={{ color: "var(--color-text-primary)" }}>Landing Page Mockup</p>
          <p className="text-sm mb-4" style={{ color: "var(--color-text-secondary)" }}>
            AI-generated hero section styled for <strong>{brandName}</strong> — based on your brand inputs.
          </p>

          {visuals.mockup_html ? (
            <div className="rounded-xl overflow-hidden"
              style={{ border: "1px solid rgba(255,255,255,0.12)", height: 640 }}>
              <iframe
                srcDoc={visuals.mockup_html}
                title={`${brandName} landing page mockup`}
                className="w-full h-full border-0"
                sandbox="allow-same-origin"
              />
            </div>
          ) : (
            <div className="rounded-xl p-12 text-center"
              style={{ border: "1px solid rgba(255,255,255,0.07)", background: "rgba(255,255,255,0.02)" }}>
              <p className="text-sm" style={{ color: "var(--color-text-hint)" }}>
                Landing page mockup was not generated. Check WATSONX_API_KEY is set.
              </p>
            </div>
          )}

          <div className="mt-6 flex justify-between">
            <button type="button" onClick={() => setVisualStep("moodboard")}
              className="px-6 py-2 rounded-lg text-sm font-bold transition-all"
              style={{ border: "1px solid rgba(99,210,255,0.30)", color: "var(--color-signal)", background: "transparent" }}>
              ← Back to Mood Board
            </button>
            <div className="flex gap-3">
              <button type="button" onClick={onStartOver}
                className="px-5 py-2 rounded-lg text-sm font-bold transition-all"
                style={{ border: "1px solid rgba(157,48,96,0.55)", color: "#e040a0", background: "transparent" }}>
                ⟳ Start Over
              </button>
              <button type="button" onClick={handleExport} disabled={exporting}
                className="flex items-center gap-2 px-6 py-2 rounded-lg text-sm font-black disabled:opacity-50"
                style={{
                  background: "linear-gradient(135deg,rgba(109,40,217,0.90) 0%,rgba(30,90,180,0.85) 100%)",
                  border: "1px solid rgba(139,92,246,0.50)", color: "#fff",
                }}>
                {exporting ? "⏳ Exporting…" : "⬇ Export Full Pack"}
              </button>
            </div>
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
        <div>
          <Label>Age</Label><Value>{persona.age}</Value>
        </div>
        <div>
          <Label>Occupation</Label><Value>{persona.occupation}</Value>
        </div>
        <div className="sm:col-span-2">
          <Label>Brand Voice</Label><Value>{persona.voice}</Value>
        </div>
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
          <div>
            <Label>Reads</Label><Value>{persona.reads.join(", ")}</Value>
          </div>
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
