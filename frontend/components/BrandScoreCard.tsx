"use client";

import { useState } from "react";
import type { BrandScore } from "@/lib/types";

interface ScoreDotProps { value: number; max?: number; }

function ScoreDots({ value, max = 10 }: ScoreDotProps) {
  return (
    <span className="inline-flex gap-0.5">
      {Array.from({ length: max }).map((_, i) => (
        <span
          key={i}
          className="inline-block w-1.5 h-1.5 rounded-full"
          style={{
            background: i < value
              ? "var(--color-pulse)"
              : "rgba(255,255,255,0.12)",
          }}
        />
      ))}
    </span>
  );
}

function riskColor(risk: string) {
  if (risk === "Low")    return "text-emerald-400";
  if (risk === "High")   return "text-red-400";
  return "text-amber-400";
}

interface BrandScoreCardProps { score: BrandScore; }

export default function BrandScoreCard({ score }: BrandScoreCardProps) {
  const [open, setOpen] = useState(false);

  return (
    <div className="mt-3 rounded-lg overflow-hidden"
      style={{ border: "1px solid rgba(255,255,255,0.07)", background: "rgba(255,255,255,0.03)" }}>
      {/* Toggle header */}
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        className="w-full flex items-center justify-between px-3 py-2 text-left"
      >
        <span className="text-[11px] font-bold uppercase tracking-widest"
          style={{ color: "var(--color-text-hint)" }}>
          Brand Score
        </span>
        <span className="text-xs" style={{ color: "var(--color-text-hint)" }}>
          {open ? "▲" : "▼"}
        </span>
      </button>

      {open && (
        <div className="px-3 pb-3 space-y-2">
          {[
            { label: "Memorability",   value: score.memorability,  reason: score.mem_reasoning },
            { label: "Spellability",   value: score.spellability,  reason: score.spell_reasoning },
            { label: "Global Safety",  value: score.global_safety, reason: score.global_reasoning },
          ].map(row => (
            <div key={row.label}>
              <div className="flex items-center justify-between mb-0.5">
                <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
                  {row.label}
                </span>
                <div className="flex items-center gap-2">
                  <ScoreDots value={row.value} />
                  <span className="text-xs font-bold tabular-nums"
                    style={{ color: "var(--color-text-primary)" }}>
                    {row.value}/10
                  </span>
                </div>
              </div>
              {row.reason && (
                <p className="text-[11px] leading-snug" style={{ color: "var(--color-text-hint)" }}>
                  {row.reason}
                </p>
              )}
            </div>
          ))}
          <div className="flex items-center gap-2 pt-1">
            <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
              Domain squatter risk:
            </span>
            <span className={`text-xs font-bold ${riskColor(score.squatter_risk)}`}>
              {score.squatter_risk}
            </span>
          </div>
          {score.squatter_reasoning && (
            <p className="text-[11px]" style={{ color: "var(--color-text-hint)" }}>
              {score.squatter_reasoning}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
