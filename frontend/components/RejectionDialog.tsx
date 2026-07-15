"use client";

import { useState } from "react";

interface Props {
  rejectedName: string;
  clarifyingQuestion: string;
  onAnswer: (answer: string) => void;
  onSkip: () => void;
}

export default function RejectionDialog({ rejectedName, clarifyingQuestion, onAnswer, onSkip }: Props) {
  const [answer, setAnswer] = useState("");

  return (
    <div
      className="rounded-2xl border p-5 w-full"
      style={{
        background: "rgba(14,18,32,0.96)",
        borderColor: "rgba(139,92,246,0.35)",
        boxShadow: "0 0 30px rgba(139,92,246,0.12)",
      }}
    >
      {/* ── AI icon + header ──────────────────────────────────── */}
      <div className="flex items-start gap-3 mb-4">
        <div
          className="shrink-0 w-8 h-8 rounded-full flex items-center justify-center text-xs font-black"
          style={{ background: "rgba(139,92,246,0.25)", color: "var(--color-pulse)" }}
        >
          NV
        </div>
        <div>
          <p className="text-[11px] font-bold uppercase tracking-widest mb-1"
            style={{ color: "var(--color-text-hint)" }}>
            NomVox heard you reject "{rejectedName}"
          </p>
          <p className="text-sm leading-relaxed font-medium"
            style={{ color: "var(--color-text-primary)" }}>
            {clarifyingQuestion}
          </p>
        </div>
      </div>

      {/* ── Answer input ──────────────────────────────────────── */}
      <textarea
        value={answer}
        onChange={e => setAnswer(e.target.value)}
        placeholder="Tell me more… (or skip to regenerate directly)"
        rows={2}
        className="w-full rounded-lg px-3 py-2 text-sm resize-none outline-none transition-all"
        style={{
          background: "rgba(255,255,255,0.04)",
          border: "1px solid rgba(99,210,255,0.35)",
          color: "var(--color-text-primary)",
        }}
        onFocus={e => {
          e.currentTarget.style.borderColor = "rgba(99,210,255,0.75)";
          e.currentTarget.style.boxShadow = "0 0 0 3px rgba(99,210,255,0.12)";
        }}
        onBlur={e => {
          e.currentTarget.style.borderColor = "rgba(99,210,255,0.35)";
          e.currentTarget.style.boxShadow = "none";
        }}
      />

      {/* ── Action buttons ────────────────────────────────────── */}
      <div className="flex gap-2 mt-3">
        <button
          type="button"
          onClick={() => onAnswer(answer.trim())}
          disabled={!answer.trim()}
          className="flex-1 rounded-lg py-2 text-sm font-bold transition-all disabled:opacity-40"
          style={{
            background: "linear-gradient(135deg, #8B5CF6 0%, #6d28d9 100%)",
            color: "#fff",
          }}
        >
          Refine &amp; Regenerate →
        </button>
        <button
          type="button"
          onClick={onSkip}
          className="rounded-lg px-4 py-2 text-sm font-bold transition-colors"
          style={{
            border: "1px solid rgba(255,255,255,0.1)",
            color: "var(--color-text-hint)",
          }}
          onMouseEnter={e => { e.currentTarget.style.color = "var(--color-text-secondary)"; }}
          onMouseLeave={e => { e.currentTarget.style.color = "var(--color-text-hint)"; }}
        >
          Skip
        </button>
      </div>
    </div>
  );
}
