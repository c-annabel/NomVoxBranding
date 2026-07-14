"use client";

// Counts filled optional fields and renders a colour-coded progress bar.
interface Props {
  filledCount: number;
  totalOptional: number;
}

export default function PromptRichnessMeter({ filledCount, totalOptional }: Props) {
  const pct = Math.round((filledCount / totalOptional) * 100);

  const { label, barColor, textColor } =
    pct >= 70
      ? { label: "Rich — great first pass", barColor: "bg-emerald-400", textColor: "text-emerald-300" }
      : pct >= 40
      ? { label: "Good — can still improve", barColor: "bg-amber-400", textColor: "text-amber-300" }
      : { label: "Thin — add more context", barColor: "bg-red-400", textColor: "text-red-300" };

  return (
    <div
      className="space-y-1.5"
      title="More fields filled = better names on the first try, fewer regeneration loops"
    >
      <div className="flex items-center justify-between">
        <span
          className="text-[11px] font-bold uppercase tracking-widest"
          style={{ color: "var(--color-text-hint)" }}
        >
          Prompt richness
        </span>
        <span className={`text-xs font-bold ${textColor}`}>{label}</span>
      </div>
      <div className="h-1.5 w-full rounded-full overflow-hidden" style={{ background: "rgba(255,255,255,0.08)" }}>
        <div
          className={`h-full rounded-full transition-all duration-300 ${barColor}`}
          style={{ width: `${Math.max(pct, filledCount > 0 ? 5 : 0)}%` }}
        />
      </div>
      <p className="text-xs" style={{ color: "var(--color-text-hint)" }}>
        {filledCount} of {totalOptional} optional fields filled — more context = better names on the first try
      </p>
    </div>
  );
}
