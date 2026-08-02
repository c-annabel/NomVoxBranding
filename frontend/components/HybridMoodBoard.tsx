"use client";

import React from "react";

/**
 * HybridMoodBoard
 *
 * Combines deterministic brand-system panels (exact palette, real typeface,
 * real logo lockups) with AI-generated texture tiles.
 *
 * Rationale: image models cannot render an exact hex swatch, set type in a
 * specific typeface, or place the real logo. They are good at texture and
 * atmosphere. So the structural panels are CSS and the texture panels are AI.
 *
 * Drop into components/ and import from VisualIdentityPanel.tsx.
 */

type Palette = { bg: string; accent: string; text: string; accent2: string };

export interface HybridMoodBoardProps {
  brandName: string;
  tagline?: string;
  pal: Palette;
  /** AI texture tiles (data URIs). First two are used; missing ones fall back to CSS. */
  textures?: string[];
  /** Logo data URIs, any may be empty. */
  logoProfile?: string;
  logoApp?: string;
  logoBusiness?: string;
  /** Rendered when a logo URI is missing. */
  logoFallback?: (type: "profile" | "app" | "business") => React.ReactNode;
  personality?: string;
}

const PANEL_BORDER = "1px solid rgba(255,255,255,0.10)";
const LABEL_STYLE: React.CSSProperties = {
  fontSize: "0.6rem",
  fontWeight: 900,
  letterSpacing: "0.14em",
  textTransform: "uppercase",
};

function isUsable(uri?: string): boolean {
  return !!uri && uri.startsWith("data:") && uri.length > 200;
}

export default function HybridMoodBoard({
  brandName,
  tagline,
  pal,
  textures = [],
  logoProfile,
  logoApp,
  logoBusiness,
  logoFallback,
  personality,
}: HybridMoodBoardProps) {
  const initials = brandName.slice(0, 2).toUpperCase();
  // Only ever show two texture tiles, whatever the backend returned.
  const usableTextures = textures.filter(isUsable).slice(0, 2);

  return (
    <div
      className="rounded-xl overflow-hidden"
      style={{
        border: PANEL_BORDER,
        background: pal.bg,
        height: "100%",
        display: "grid",
        // Explicit rows. Flex basis let the image row grow to its content
        // and push the logo lockups out of the fixed-height container.
        gridTemplateRows: "minmax(0, 0.85fr) minmax(0, 1fr) 128px",
      }}
    >
      {/* ── Row 1: exact brand facts (deterministic) ──────────────────── */}
      <div className="grid grid-cols-2" style={{ minHeight: 0, overflow: "hidden", gap: 1, background: "rgba(255,255,255,0.08)" }}>
        {/* Colour world */}
        <div className="p-4 flex flex-col justify-between" style={{ background: pal.bg }}>
          <p style={{ ...LABEL_STYLE, color: pal.accent }}>Colour World</p>
          <div>
            <p className="font-black leading-none mb-1" style={{ color: pal.text, fontSize: "1.35rem" }}>
              {brandName}
            </p>
            {tagline && (
              <p className="italic" style={{ color: pal.accent, fontSize: "0.8rem" }}>
                {tagline}
              </p>
            )}
          </div>
          <div className="flex gap-2 items-center">
            {[pal.bg, pal.accent, pal.accent2, pal.text].map((c, i) => (
              <div
                key={i}
                title={c}
                style={{
                  width: 26,
                  height: 26,
                  borderRadius: "50%",
                  background: c,
                  border: "1px solid rgba(255,255,255,0.28)",
                }}
              />
            ))}
          </div>
        </div>

        {/* Typography scale */}
        <div className="p-4 flex flex-col justify-between" style={{ background: pal.bg }}>
          <p style={{ ...LABEL_STYLE, color: pal.accent }}>Typography Scale</p>
          <div>
            <p className="font-black leading-none" style={{ color: pal.text, fontSize: "2.6rem" }}>
              Aa
            </p>
            <p className="font-black mt-1" style={{ color: pal.text, fontSize: "0.68rem", letterSpacing: "0.06em" }}>
              SPACE GROTESK / WIDE / MODERN
            </p>
            <p style={{ color: "rgba(255,255,255,0.45)", fontSize: "0.62rem" }}>Regular · Hint text</p>
          </div>
          <button
            type="button"
            className="font-bold self-start"
            style={{
              background: pal.accent,
              color: "#fff",
              fontSize: "0.68rem",
              padding: "6px 14px",
              borderRadius: 6,
              border: "none",
            }}
          >
            CTA Button
          </button>
        </div>
      </div>

      {/* ── Row 2: AI texture studies ─────────────────────────────────── */}
      <div className="grid grid-cols-2" style={{ minHeight: 0, overflow: "hidden", gap: 1, background: "rgba(255,255,255,0.08)" }}>
        {[0, 1].map((i) => {
          const uri = usableTextures[i];
          const label = i === 0 ? "Texture Study" : "Atmosphere";
          return (
            <div key={i} style={{ position: "relative", overflow: "hidden", minHeight: 0, background: pal.bg }}>
              {isUsable(uri) ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={uri}
                  alt={`${brandName} ${label.toLowerCase()}`}
                  style={{ position: "absolute", inset: 0, width: "100%", height: "100%", objectFit: "cover", display: "block" }}
                />
              ) : (
                <div
                  style={{
                    position: "absolute",
                    inset: 0,
                    background:
                      i === 0
                        ? `radial-gradient(circle at 30% 35%, ${pal.accent}55, transparent 62%),
                           radial-gradient(circle at 72% 68%, ${pal.accent2}44, transparent 58%), ${pal.bg}`
                        : `linear-gradient(135deg, ${pal.bg} 0%, ${pal.accent}33 48%, ${pal.accent2}22 100%)`,
                  }}
                />
              )}
              <span
                className="absolute top-2 left-3"
                style={{ ...LABEL_STYLE, color: "rgba(255,255,255,0.82)", textShadow: "0 1px 4px rgba(0,0,0,0.85)" }}
              >
                {label}
              </span>
            </div>
          );
        })}
      </div>

      {/* ── Row 3: real logo lockups ──────────────────────────────────── */}
      <div
        className="flex items-center justify-around px-4"
        style={{ minHeight: 0, overflow: "hidden", background: pal.bg, borderTop: PANEL_BORDER }}
      >
        {([
          { key: "profile" as const, uri: logoProfile, label: "Social", radius: "50%" },
          { key: "app" as const, uri: logoApp, label: "App", radius: "14px" },
          { key: "business" as const, uri: logoBusiness, label: "Print", radius: "6px" },
        ]).map(({ key, uri, label, radius }) => (
          <div key={key} className="flex flex-col items-center gap-1.5">
            <div
              className="flex items-center justify-center overflow-hidden"
              style={{
                width: 76,
                height: 76,
                borderRadius: radius,
                background: "rgba(255,255,255,0.05)",
                border: "1px solid rgba(255,255,255,0.14)",
              }}
            >
              {isUsable(uri) ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={uri}
                  alt={`${brandName} ${label} logo`}
                  style={{ width: "100%", height: "100%", objectFit: "contain" }}
                />
              ) : logoFallback ? (
                logoFallback(key)
              ) : (
                <span className="font-black" style={{ color: pal.accent, fontSize: "1.05rem" }}>
                  {key === "profile" ? initials.charAt(0) : initials}
                </span>
              )}
            </div>
            <span style={{ ...LABEL_STYLE, color: "rgba(255,255,255,0.42)", fontSize: "0.55rem" }}>{label}</span>
          </div>
        ))}

        {personality && (
          <div className="max-w-[9rem]">
            <p style={{ ...LABEL_STYLE, color: pal.accent, marginBottom: 2 }}>Personality</p>
            <p style={{ color: "rgba(255,255,255,0.62)", fontSize: "0.66rem", lineHeight: 1.35 }}>{personality}</p>
          </div>
        )}
      </div>
    </div>
  );
}
