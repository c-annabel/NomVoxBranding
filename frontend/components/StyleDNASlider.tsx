"use client";

import type { StyleDNA } from "@/lib/types";

interface Props {
  dna: StyleDNA;
  onChange: (dna: StyleDNA) => void;
  fullWidth?: boolean;
}

export default function StyleDNASlider({ dna, onChange, fullWidth }: Props) {
  return (
    <div className={`w-full ${fullWidth ? "" : "max-w-xl"} rounded-xl px-5 py-4`}
      style={{
        background: "rgba(139,92,246,0.06)",
        border: "1px solid rgba(139,92,246,0.22)",
      }}>
      <p className="text-sm font-bold uppercase tracking-widest mb-4"
        style={{ color: "var(--color-pulse)" }}>
        Style DNA — Tune the next batch
      </p>

      <div className="flex flex-col gap-5">

        {/* ── Playful ↔ Premium ─────────────────────────────── */}
        <div>
          <div className="flex justify-between items-center mb-2">
            <span className="text-sm font-semibold" style={{ color: "var(--color-text-secondary)" }}>
              Premium
            </span>
            <span className="text-sm font-black px-3 py-0.5 rounded-full"
              style={{
                background: "rgba(139,92,246,0.22)",
                color: "#c4b5fd",
                border: "1px solid rgba(139,92,246,0.40)",
              }}>
              {dna.playful < 0.3 ? "Premium" : dna.playful > 0.7 ? "Playful" : "Balanced"}
            </span>
            <span className="text-sm font-semibold" style={{ color: "var(--color-text-secondary)" }}>
              Playful
            </span>
          </div>
          <input
            type="range" min={0} max={1} step={0.05} value={dna.playful}
            onChange={e => onChange({ ...dna, playful: parseFloat(e.target.value) })}
            className="nv-range w-full cursor-pointer"
            style={{ "--thumb-color": "#8B5CF6", "--track-fill": "#8B5CF6" } as React.CSSProperties}
          />
        </div>

        {/* ── Descriptive ↔ Abstract ─────────────────────────── */}
        <div>
          <div className="flex justify-between items-center mb-2">
            <span className="text-sm font-semibold" style={{ color: "var(--color-text-secondary)" }}>
              Descriptive
            </span>
            <span className="text-sm font-black px-3 py-0.5 rounded-full"
              style={{
                background: "rgba(34,211,238,0.14)",
                color: "#67e8f9",
                border: "1px solid rgba(34,211,238,0.35)",
              }}>
              {dna.abstract < 0.3 ? "Descriptive" : dna.abstract > 0.7 ? "Abstract" : "Hybrid"}
            </span>
            <span className="text-sm font-semibold" style={{ color: "var(--color-text-secondary)" }}>
              Abstract
            </span>
          </div>
          <input
            type="range" min={0} max={1} step={0.05} value={dna.abstract}
            onChange={e => onChange({ ...dna, abstract: parseFloat(e.target.value) })}
            className="nv-range nv-range-cyan w-full cursor-pointer"
          />
        </div>
      </div>

      <p className="mt-4 text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
        Move sliders then hit ↻ Regenerate to apply your direction.
      </p>
    </div>
  );
}
