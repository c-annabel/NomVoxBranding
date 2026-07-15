"use client";

// Skeleton shown while name cards are loading.
export default function NameCardSkeleton() {
  return (
    <div className="rounded-xl p-5 space-y-3 animate-pulse"
      style={{
        background: "rgba(14,18,32,0.90)",
        border: "1px solid rgba(139,92,246,0.12)",
      }}>
      {/* Name */}
      <div className="h-7 w-36 rounded-md" style={{ background: "rgba(255,255,255,0.07)" }} />
      {/* Tagline */}
      <div className="h-4 w-56 rounded-md" style={{ background: "rgba(255,255,255,0.05)" }} />
      {/* Short desc */}
      <div className="space-y-1.5">
        <div className="h-3 w-full rounded-md" style={{ background: "rgba(255,255,255,0.05)" }} />
        <div className="h-3 w-4/5 rounded-md"  style={{ background: "rgba(255,255,255,0.04)" }} />
      </div>
      {/* Origin block */}
      <div className="h-12 w-full rounded-lg" style={{ background: "rgba(6,182,212,0.05)", borderLeft: "3px solid rgba(6,182,212,0.15)" }} />
    </div>
  );
}
