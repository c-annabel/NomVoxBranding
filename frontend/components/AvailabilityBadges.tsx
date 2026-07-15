"use client";

import type { AvailabilityResult } from "@/lib/types";

interface Props {
  availability?: AvailabilityResult;
  loading?: boolean;
}

const DOT_PLATFORMS = [
  { key: "domain",    label: ".com",       weight: 2.0 },
  { key: "instagram", label: "IG",          weight: 1.0 },
  { key: "x",         label: "𝕏",           weight: 1.0 },
  { key: "tiktok",    label: "TikTok",      weight: 1.0 },
  { key: "threads",   label: "Threads",     weight: 0.5 },
  { key: "youtube",   label: "YT",          weight: 0.5 },
] as const;

type PlatformKey = (typeof DOT_PLATFORMS)[number]["key"];

export default function AvailabilityBadges({ availability, loading }: Props) {
  if (loading) {
    return (
      <div className="flex flex-wrap gap-1.5 mt-2">
        {DOT_PLATFORMS.map(p => (
          <span key={p.key}
            className="animate-pulse rounded-full px-2 py-0.5 text-[10px] font-bold"
            style={{ background: "rgba(255,255,255,0.05)", color: "transparent", minWidth: 36 }}>
            ···
          </span>
        ))}
      </div>
    );
  }

  if (!availability) return null;

  // Go JSON serialises PlatformResult fields as lowercase snake_case:
  // domain, instagram, x, tiktok, threads, youtube
  // and domain_unknown, instagram_unknown, x_unknown, tiktok_unknown, threads_unknown, youtube_unknown
  const probes = availability.probes as unknown as Record<string, boolean>;
  const pct = Math.round(availability.score);

  return (
    <div className="mt-1.5">
      {/* ── Score bar ──────────────────────────────────────────── */}
      <div className="flex items-center gap-2 mb-2">
        <div className="flex-1 h-1.5 rounded-full overflow-hidden"
          style={{ background: "rgba(255,255,255,0.08)" }}>
          <div
            className="h-full rounded-full transition-all duration-500"
            style={{
              width: `${pct}%`,
              background: availability.passes
                ? "linear-gradient(90deg,#22d3ee,#4ade80)"
                : pct >= 60
                  ? "linear-gradient(90deg,#f59e0b,#f97316)"
                  : "linear-gradient(90deg,#f87171,#e11d48)",
            }}
          />
        </div>
        <span className="text-xs font-bold tabular-nums"
          style={{ color: availability.passes ? "#4ade80" : pct >= 60 ? "#f59e0b" : "#f87171" }}>
          {pct}%
        </span>
        {availability.passes
          ? <span className="text-xs font-black" style={{ color: "#4ade80" }}>PASS ✓</span>
          : <span className="text-xs font-black" style={{ color: "#f87171" }}>PARTIAL</span>}
      </div>

      {/* ── Platform badges — 2-column grid for even distribution ── */}
      <div className="grid grid-cols-2 gap-x-2 gap-y-1">
        {DOT_PLATFORMS.map(p => {
          const avail   = probes[p.key] === true;
          const unknown = probes[`${p.key}_unknown`] === true;

          let bg    = "rgba(248,113,113,0.18)";
          let color = "#f87171";
          let icon  = "✕";
          if (unknown) { bg = "rgba(255,255,255,0.06)"; color = "#94a3b8"; icon = "?"; }
          else if (avail) { bg = "rgba(74,222,128,0.15)"; color = "#4ade80"; icon = "✓"; }

          return (
            <span key={p.key}
              className="inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-bold"
              title={unknown ? "Check inconclusive" : avail ? "Available" : "Taken"}
              style={{ background: bg, color }}>
              <span className="shrink-0">{p.label}</span>
              <span className="shrink-0">{icon}</span>
            </span>
          );
        })}
      </div>

      {/* ── Competitor radar warning ────────────────────────────── */}
      {availability.radar && (
        <p className="mt-1.5 text-xs leading-relaxed"
          style={{ color: "#f59e0b" }}>
          ⚠ {availability.radar}
        </p>
      )}
    </div>
  );
}
