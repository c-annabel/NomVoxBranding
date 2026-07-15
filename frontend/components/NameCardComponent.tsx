"use client";

import { useState } from "react";
import type { NameCard, AvailabilityResult } from "@/lib/types";
import AvailabilityBadges from "./AvailabilityBadges";

interface NameCardProps {
  card: NameCard;
  index: number;
  availability?: AvailabilityResult;
  availabilityLoading?: boolean;
  reaction?: "liked" | "rejected" | null;
  rejectedReason?: string;
  onLike: (name: string) => void;
  onReject: (name: string, reason?: string) => void;
  onSelect: (name: string) => void;
}

function riskBadge(risk?: string) {
  if (!risk) return null;
  const color = risk === "Low" ? "#4ade80" : risk === "High" ? "#f87171" : "#f59e0b";
  const bg    = risk === "Low" ? "rgba(74,222,128,0.14)" : risk === "High" ? "rgba(248,113,113,0.14)" : "rgba(245,158,11,0.14)";
  return (
    <span className="text-xs font-bold px-2 py-0.5 rounded-full"
      style={{ background: bg, color, border: `1px solid ${color}40` }}>
      {risk}
    </span>
  );
}

function ColHead({ children, color }: { children: React.ReactNode; color?: string }) {
  return (
    <p className="text-[11px] font-black uppercase tracking-widest mb-1.5"
      style={{ color: color ?? "var(--color-text-hint)" }}>
      {children}
    </p>
  );
}

export default function NameCardComponent({
  card, index, availability, availabilityLoading, reaction, rejectedReason,
  onLike, onReject, onSelect,
}: NameCardProps) {
  const [showRejectBox, setShowRejectBox] = useState(false);
  const [rejectReason,   setRejectReason] = useState("");

  const isLiked    = reaction === "liked";
  const isRejected = reaction === "rejected";
  const availPasses = availability?.passes ?? true;

  function submitReject() {
    onReject(card.name, rejectReason.trim() || undefined);
    setShowRejectBox(false);
    setRejectReason("");
  }

  // Subtle per-card accent colours — rotates through 6 hues, stays dark so it's a separator not a distraction
  const CARD_ACCENT = [
    "rgba(139,92,246,0.35)",   // purple
    "rgba(34,211,238,0.30)",   // cyan
    "rgba(251,191,36,0.30)",   // amber
    "rgba(74,222,128,0.28)",   // green
    "rgba(251,113,133,0.30)",  // rose
    "rgba(167,139,250,0.28)",  // violet-light
  ];
  const accentColor = CARD_ACCENT[index % CARD_ACCENT.length];

  return (
    <div
      className="rounded-xl overflow-hidden transition-all"
      style={{
        background: isRejected ? "rgba(14,18,32,0.50)" : "rgba(16,20,38,0.97)",
        border: isLiked
          ? "1.5px solid rgba(74,222,128,0.50)"
          : isRejected
            ? "1px solid rgba(248,113,113,0.25)"
            : `1.5px solid ${accentColor}`,
        boxShadow: isLiked
          ? "0 4px 24px rgba(74,222,128,0.12)"
          : `0 4px 20px rgba(5,7,15,0.45), 0 0 0 1px ${accentColor}`,
        opacity: isRejected ? 0.60 : 1,
        animationDelay: `${index * 80}ms`,
      }}
    >
      {/* ══════════════════════════════════════════════════════════════
          ROW 1 — Name+Tags | Tagline | Origin | Description
          Mobile: single column stack; Desktop (lg): 4-column grid
      ══════════════════════════════════════════════════════════════ */}
      <div className="flex flex-col lg:grid"
        style={{ gridTemplateColumns: "190px 1fr 1fr 1.5fr" }}>

        {/* Name + tags */}
        <div className="px-4 pt-4 pb-3 lg:border-r border-b"
          style={{ borderColor: "rgba(255,255,255,0.08)" }}>
          <h3 className="text-2xl font-black tracking-tight leading-none mb-2"
            style={{ color: "var(--color-text-primary)" }}>
            {card.name}
          </h3>
          <div className="flex flex-wrap gap-1">
            {card.style_tags?.map(tag => (
              <span key={tag}
                className="text-[11px] font-bold px-2 py-0.5 rounded-full uppercase tracking-wide"
                style={{ background: "rgba(139,92,246,0.22)", color: "#c4b5fd", border: "1px solid rgba(139,92,246,0.40)" }}>
                {tag}
              </span>
            ))}
          </div>
        </div>

        {/* Tagline */}
        <div className="px-4 pt-3 pb-3 lg:border-r border-b"
          style={{ borderColor: "rgba(255,255,255,0.08)" }}>
          <ColHead color="var(--color-signal)">Tagline</ColHead>
          <p className="text-base italic font-semibold leading-snug"
            style={{ color: "var(--color-signal)" }}>
            "{card.tagline}"
          </p>
        </div>

        {/* Origin */}
        <div className="px-4 pt-3 pb-3 lg:border-r border-b"
          style={{ borderColor: "rgba(255,255,255,0.08)" }}>
          <ColHead color="#a78bfa">Origin</ColHead>
          <p className="text-sm leading-relaxed" style={{ color: "var(--color-text-secondary)" }}>
            {card.origin_story || "—"}
          </p>
        </div>

        {/* Description */}
        <div className="px-4 pt-3 pb-3 border-b"
          style={{ borderColor: "rgba(255,255,255,0.08)" }}>
          <ColHead color="#94a3b8">Description</ColHead>
          <p className="text-sm leading-relaxed" style={{ color: "var(--color-text-secondary)" }}>
            {card.short_desc || "—"}
          </p>
          {card.long_desc && (
            <p className="mt-1 text-sm leading-relaxed" style={{ color: "var(--color-text-hint)" }}>
              {card.long_desc}
            </p>
          )}
        </div>
      </div>

      {/* ══════════════════════════════════════════════════════════════
          ROW 2 — Availability | Brand Score | Brand Voice | Actions
          Mobile: single column stack; Desktop (lg): 4-column grid
      ══════════════════════════════════════════════════════════════ */}
      <div className="flex flex-col lg:grid"
        style={{ gridTemplateColumns: "1fr 1fr 1fr 120px" }}>

        {/* Availability */}
        <div className="px-4 pt-3 pb-4 lg:border-r border-b lg:border-b-0"
          style={{ borderColor: "rgba(255,255,255,0.08)" }}>
          <ColHead color={availPasses ? "#4ade80" : "#f87171"}>Availability</ColHead>
          <AvailabilityBadges availability={availability} loading={availabilityLoading} />
        </div>

        {/* Brand Score */}
        <div className="px-4 pt-3 pb-4 lg:border-r border-b lg:border-b-0"
          style={{ borderColor: "rgba(255,255,255,0.08)" }}>
          <ColHead color="#f59e0b">Brand Score</ColHead>
          {card.score ? (
            <div className="space-y-2">
              {[
                { label: "Memory", value: card.score.memorability },
                { label: "Spell",  value: card.score.spellability },
                { label: "Safety", value: card.score.global_safety },
              ].map(row => (
                <div key={row.label} className="flex items-center gap-1.5">
                  <span className="text-sm w-14 shrink-0 font-medium"
                    style={{ color: "var(--color-text-hint)" }}>
                    {row.label}
                  </span>
                  <div className="h-1.5 rounded-full overflow-hidden max-w-[80px] flex-1"
                    style={{ background: "rgba(255,255,255,0.08)" }}>
                    <div className="h-full rounded-full"
                      style={{
                        width: `${(row.value / 10) * 100}%`,
                        background: row.value >= 7 ? "#4ade80" : row.value >= 4 ? "#f59e0b" : "#f87171",
                      }} />
                  </div>
                  <span className="text-sm font-bold tabular-nums w-6 text-right"
                    style={{ color: "var(--color-text-primary)" }}>{row.value}</span>
                </div>
              ))}
              <div className="flex items-center gap-1.5 mt-0.5">
                <span className="text-sm w-14 font-medium" style={{ color: "var(--color-text-hint)" }}>Risk</span>
                {riskBadge(card.score.squatter_risk)}
              </div>
            </div>
          ) : (
            <p className="text-sm" style={{ color: "var(--color-text-hint)" }}>—</p>
          )}
        </div>

        {/* Brand Voice */}
        <div className="px-4 pt-3 pb-4 lg:border-r border-b lg:border-b-0"
          style={{ borderColor: "rgba(255,255,255,0.08)" }}>
          <ColHead color="#67e8f9">Brand Voice</ColHead>
          <div className="space-y-2">
            {[
              { label: "IG",    value: card.voice_samples?.instagram_caption },
              { label: "Email", value: card.voice_samples?.email_subject },
              { label: "404",   value: card.voice_samples?.not_found_message },
            ].filter(s => s.value).map(s => (
              <div key={s.label}>
                <span className="text-sm font-bold" style={{ color: "#67e8f9" }}>{s.label}: </span>
                <span className="text-sm italic" style={{ color: "var(--color-text-secondary)" }}>"{s.value}"</span>
              </div>
            ))}
          </div>
        </div>

        {/* Actions */}
        <div className="flex flex-row lg:flex-col justify-center flex-wrap px-4 py-3 gap-2">

          <button type="button" onClick={() => onLike(card.name)}
            className="w-full py-2 rounded-lg text-sm font-bold transition-all text-center"
            style={{
              background: isLiked ? "rgba(74,222,128,0.16)" : "rgba(255,255,255,0.06)",
              border: isLiked ? "1px solid rgba(74,222,128,0.50)" : "1px solid rgba(255,255,255,0.12)",
              color: isLiked ? "#4ade80" : "var(--color-text-secondary)",
            }}
            onMouseEnter={e => { if (!isLiked) { e.currentTarget.style.background = "rgba(74,222,128,0.10)"; e.currentTarget.style.color = "#4ade80"; }}}
            onMouseLeave={e => { if (!isLiked) { e.currentTarget.style.background = "rgba(255,255,255,0.06)"; e.currentTarget.style.color = "var(--color-text-secondary)"; }}}>
            {isLiked ? "♥ Liked" : "♡ Like"}
          </button>

          {!isRejected && !showRejectBox && (
            <button type="button" onClick={() => setShowRejectBox(true)}
              className="w-full py-2 rounded-lg text-sm font-bold transition-all"
              style={{ background: "rgba(255,255,255,0.06)", border: "1px solid rgba(255,255,255,0.12)", color: "var(--color-text-secondary)" }}
              onMouseEnter={e => { e.currentTarget.style.background = "rgba(248,113,113,0.12)"; e.currentTarget.style.color = "#f87171"; e.currentTarget.style.borderColor = "rgba(248,113,113,0.40)"; }}
              onMouseLeave={e => { e.currentTarget.style.background = "rgba(255,255,255,0.06)"; e.currentTarget.style.color = "var(--color-text-secondary)"; e.currentTarget.style.borderColor = "rgba(255,255,255,0.12)"; }}>
              ✕ Pass
            </button>
          )}

          {!isRejected && (
            <button type="button" onClick={() => onSelect(card.name)}
              className="w-full py-2 rounded-lg text-sm font-black transition-all"
              style={{
                background: "linear-gradient(135deg, rgba(109,40,217,0.88) 0%, rgba(30,90,180,0.85) 100%)",
                border: "1px solid rgba(139,92,246,0.45)",
                color: "#fff",
                boxShadow: "0 2px 12px rgba(100,50,200,0.25)",
              }}
              onMouseEnter={e => { e.currentTarget.style.background = "linear-gradient(135deg,rgba(139,92,246,0.95) 0%,rgba(34,150,220,0.90) 100%)"; e.currentTarget.style.boxShadow = "0 4px 20px rgba(100,50,200,0.45)"; }}
              onMouseLeave={e => { e.currentTarget.style.background = "linear-gradient(135deg, rgba(109,40,217,0.88) 0%, rgba(30,90,180,0.85) 100%)"; e.currentTarget.style.boxShadow = "0 2px 12px rgba(100,50,200,0.25)"; }}>
              ✓ Choose
            </button>
          )}

          {isRejected && (
            <div className="flex flex-col gap-1.5 items-center">
              <span className="text-sm font-bold" style={{ color: "#f87171" }}>✕ Passed</span>
              <button type="button" onClick={() => onLike(card.name)}
                className="text-sm font-bold transition-colors"
                style={{ color: "var(--color-text-hint)" }}
                onMouseEnter={e => { e.currentTarget.style.color = "var(--color-signal)"; }}
                onMouseLeave={e => { e.currentTarget.style.color = "var(--color-text-hint)"; }}>
                ↩ Undo
              </button>
            </div>
          )}
        </div>
      </div>

      {/* ── Reject reason textarea ──────────────────────────────────── */}
      {showRejectBox && !isRejected && (
        <div className="px-4 py-3 border-t"
          style={{ borderColor: "rgba(248,113,113,0.20)", background: "rgba(248,113,113,0.04)" }}>
          <p className="text-sm font-semibold mb-2" style={{ color: "#fca5a5" }}>
            Tell NomVox why — helps future suggestions (optional):
          </p>
          <textarea
            value={rejectReason}
            onChange={e => setRejectReason(e.target.value)}
            placeholder="e.g. too corporate, hard to pronounce, sounds like a competitor…"
            rows={2}
            className="w-full rounded-lg px-3 py-2 text-sm resize-none outline-none"
            style={{ background: "rgba(255,255,255,0.04)", border: "1px solid rgba(248,113,113,0.35)", color: "var(--color-text-primary)" }}
          />
          <div className="flex gap-2 mt-2">
            <button type="button" onClick={submitReject}
              className="flex-1 py-2 rounded-lg text-sm font-black"
              style={{ background: "rgba(248,113,113,0.18)", border: "1px solid rgba(248,113,113,0.45)", color: "#f87171" }}>
              ✕ Confirm Pass
            </button>
            <button type="button" onClick={() => { setShowRejectBox(false); setRejectReason(""); }}
              className="px-5 py-2 rounded-lg text-sm font-bold"
              style={{ border: "1px solid rgba(255,255,255,0.12)", color: "var(--color-text-hint)" }}>
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* ── Rejection reason banner ─────────────────────────────────── */}
      {isRejected && rejectedReason && (
        <div className="px-4 py-2.5 border-t flex items-start gap-2"
          style={{ borderColor: "rgba(248,113,113,0.18)", background: "rgba(248,113,113,0.05)" }}>
          <span className="text-[11px] font-black uppercase tracking-widest shrink-0 mt-0.5"
            style={{ color: "#f87171" }}>Your note:</span>
          <span className="text-sm italic" style={{ color: "#fca5a5" }}>
            "{rejectedReason}"
          </span>
        </div>
      )}
    </div>
  );
}
