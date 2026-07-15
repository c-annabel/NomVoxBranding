"use client";

import Image from "next/image";
import { useState, useCallback, useEffect } from "react";
import IntakeForm from "@/components/IntakeForm";
import NameCardComponent from "@/components/NameCardComponent";
import RejectionDialog from "@/components/RejectionDialog";
import StyleDNASlider from "@/components/StyleDNASlider";
import VisualIdentityPanel from "@/components/VisualIdentityPanel";
import { parseGenerateResponse } from "@/lib/schemas";
import type {
  IntakePayload,
  NameCard,
  AvailabilityResult,
  ReactResponse,
  StyleDNA,
  VisualsResponse,
} from "@/lib/types";

type Step = "intake" | "loading" | "results" | "clarifying" | "selected" | "visuals-loading" | "visuals";

interface CardState {
  card: NameCard;
  availability?: AvailabilityResult;
  availabilityLoading: boolean;
  reaction?: "liked" | "rejected" | null;
  rejectedReason?: string;  // captured when user passes with a note
}

const BACKEND = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function apiFetch<T>(path: string, opts?: RequestInit): Promise<T> {
  return fetch(BACKEND + path, opts).then(async r => {
    const body = await r.json().catch(() => ({}));
    if (!r.ok) throw new Error((body as { error?: string }).error ?? `HTTP ${r.status}`);
    return body as T;
  });
}

export default function HomeClient() {
  const [step, setStep]                 = useState<Step>("intake");
  const [cardStates, setCardStates]     = useState<CardState[]>([]);
  const [sessionId, setSessionId]       = useState<string>("");
  const [apiError, setApiError]         = useState<string>("");
  const [availabilityMsg, setAvailabilityMsg] = useState<string>(""); // timeout / partial fallback message
  // intakeValues is the single source of truth for the form — persists when
  // navigating back from results so the user can refine without re-typing
  const [intakeValues, setIntakeValues] = useState<IntakePayload | null>(null);
  const [lastIntake, setLastIntake]     = useState<IntakePayload | null>(null);
  const [rejecting, setRejecting]       = useState<{ name: string; q: string } | null>(null);
  const [styleDNA, setStyleDNA]         = useState<StyleDNA>({ playful: 0.5, abstract: 0.5 });
  const [selectedName, setSelectedName] = useState<string>("");
  const [selectedCard, setSelectedCard] = useState<NameCard | null>(null);
  const [likedNames, setLikedNames]     = useState<string[]>([]);
  const [visuals, setVisuals]           = useState<VisualsResponse | null>(null);
  const [visualsError, setVisualsError] = useState<string>("");
  const [generatingLogos, setGeneratingLogos] = useState(false);
  // generating guards all name-gen calls so rapid double-clicks fire only once
  const [generating, setGenerating]     = useState(false);

  useEffect(() => {
    const saved = localStorage.getItem("nomvox_session_id");
    if (saved) setSessionId(saved);
  }, []);

  useEffect(() => {
    if (sessionId) localStorage.setItem("nomvox_session_id", sessionId);
  }, [sessionId]);

  const AVAILABILITY_TIMEOUT_MS = 60_000; // 60 s wall-clock limit

  const runAvailability = useCallback(async (cards: NameCard[], sid: string) => {
    const names = cards.map(c => c.name);
    setAvailabilityMsg("");

    // Race the real fetch against a 60-second timeout
    let timedOut = false;
    const timeoutId = setTimeout(() => { timedOut = true; }, AVAILABILITY_TIMEOUT_MS);

    try {
      const fetchPromise = apiFetch<{ passing: AvailabilityResult[]; partials: AvailabilityResult[] }>(
        "/api/availability",
        {
          method: "POST",
          headers: { "Content-Type": "application/json", "X-Session-ID": sid },
          body: JSON.stringify({ session_id: sid, names }),
          signal: AbortSignal.timeout(AVAILABILITY_TIMEOUT_MS),
        }
      );

      const resp = await fetchPromise;
      clearTimeout(timeoutId);

      // Only keep passing names (≥60%). If zero pass, keep top-3 partials as fallback.
      const passing  = resp.passing  ?? [];
      const partials = resp.partials ?? [];

      if (passing.length === 0 && partials.length === 0) {
        // Truly nothing — clear loading, show message
        setCardStates(prev => prev.map(cs => ({ ...cs, availabilityLoading: false })));
        setAvailabilityMsg("Unable to find available and creative names matching the threshold. Please regenerate to try again.");
        return;
      }

      const toShow  = passing.length > 0 ? passing : partials.slice(0, 2);
      const passMap = Object.fromEntries(toShow.map(r => [r.name, r]));

      if (passing.length === 0) {
        // Fallback: show up to 2 best partials with a notice
        setAvailabilityMsg(
          `Unable to find available and creative names in 1 minute — showing ${toShow.length} closest match${toShow.length === 1 ? "" : "es"}. Please regenerate to try again.`
        );
      }

      setCardStates(prev => {
        const updated = prev.map(cs => ({
          ...cs,
          availability: passMap[cs.card.name] ?? undefined,
          availabilityLoading: false,
        }));
        // Filter: only show names that are in toShow
        return updated.filter(cs => passMap[cs.card.name] !== undefined);
      });

    } catch (err) {
      clearTimeout(timeoutId);
      const isAbort = err instanceof Error && (err.name === "AbortError" || err.name === "TimeoutError" || timedOut);
      setCardStates(prev => prev.map(cs => ({ ...cs, availabilityLoading: false })));
      if (isAbort) {
        setAvailabilityMsg("Unable to find available and creative names in 1 minute — please regenerate to try again.");
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Soft back — preserves all card states, session, and intake values.
  // User can tweak their form fields and re-generate without losing context.
  function handleBackToIntake() {
    setStep("intake");
    setApiError("");
  }

  async function handleGenerate(payload: IntakePayload) {
    if (generating) return;           // block double-submit
    setGenerating(true);
    setIntakeValues(payload);         // persist so form re-seeds when user navigates back
    setStep("loading");
    setCardStates([]);
    setApiError("");
    setAvailabilityMsg("");
    setRejecting(null);
    setLastIntake(payload);
    setVisuals(null);
    setVisualsError("");

    try {
      // If this session has already generated many names (>12 = 3+ rounds) the
      // AllGenerated exclusion list gets so long the LLM starts repeating the same
      // pool of "safe" names. Drop the session ID to start a fresh context while
      // keeping liked/direction notes alive via the new session on next call.
      const effectiveSessionId = sessionId || "";

      const raw = await apiFetch<unknown>("/api/generate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          session_id: effectiveSessionId,
          intake: payload,
          action: effectiveSessionId ? "reproduce" : "init",
        }),
      });

      const { session_id, cards } = parseGenerateResponse(raw);
      const newSid = session_id || sessionId;
      if (session_id) setSessionId(session_id);

      if (cards.length === 0) {
        setApiError("No names were generated — please try again or rephrase your idea.");
        setStep("intake");
      } else {
        setCardStates(cards.map(c => ({ card: c, availabilityLoading: true })));
        setStep("results");
        runAvailability(cards, newSid);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setApiError(msg.replace(/^API.*?:\s*/, ""));
      setStep("intake");
    } finally {
      setGenerating(false);   // always unlock so next call can proceed
    }
  }

  async function handleRegenerate() {
    if (generating) return;   // block double-click on Regenerate button
    if (!lastIntake) return;
    if (sessionId && (styleDNA.playful !== 0.5 || styleDNA.abstract !== 0.5)) {
      try {
        await apiFetch("/api/session/react", {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            session_id: sessionId,
            action: "slider",
            slider_playful: styleDNA.playful,
            slider_abstract: styleDNA.abstract,
          }),
        });
      } catch { /* non-fatal */ }
    }
    handleGenerate(lastIntake);
  }

  async function handleLike(name: string) {
    const nowLiked = cardStates.find(cs => cs.card.name === name)?.reaction !== "liked";
    setCardStates(prev => prev.map(cs => ({
      ...cs,
      reaction: cs.card.name === name
        ? (cs.reaction === "liked" ? null : "liked")
        : cs.reaction,
    })));
    setLikedNames(prev =>
      nowLiked ? [...prev.filter(n => n !== name), name] : prev.filter(n => n !== name)
    );
    if (!sessionId) return;
    try {
      await apiFetch("/api/session/react", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ session_id: sessionId, action: "like", name }),
      });
    } catch { /* non-fatal */ }
  }

  async function handleReject(name: string, reason?: string) {
    setCardStates(prev => prev.map(cs => ({
      ...cs,
      reaction: cs.card.name === name ? "rejected" : cs.reaction,
      rejectedReason: cs.card.name === name ? (reason ?? cs.rejectedReason) : cs.rejectedReason,
    })));
    setLikedNames(prev => prev.filter(n => n !== name));
    if (!sessionId) return;
    try {
      const resp = await apiFetch<ReactResponse>("/api/session/react", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          session_id: sessionId,
          action: "reject",
          name,
          ...(reason ? { note: reason } : {}),
        }),
      });
      if (resp.clarifying_question) {
        setRejecting({ name, q: resp.clarifying_question });
        setStep("clarifying");
      }
    } catch { /* non-fatal */ }
  }

  async function handleSelect(name: string) {
    const chosenCard = cardStates.find(cs => cs.card.name === name)?.card ?? null;
    setSelectedName(name);
    setSelectedCard(chosenCard);

    if (sessionId) {
      try {
        await apiFetch("/api/session/react", {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ session_id: sessionId, action: "select", name }),
        });
      } catch { /* non-fatal */ }
    }

    // Immediately kick off visual generation
    if (chosenCard && lastIntake) {
      setStep("visuals-loading");
      setVisualsError("");
      try {
        const v = await apiFetch<VisualsResponse>("/api/visuals", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            session_id: sessionId,
            card: chosenCard,
            intake: lastIntake,
          }),
        });
        setVisuals(v);
        setStep("visuals");
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        setVisualsError(msg);
        // Fall back to "selected" step so user still sees their name
        setStep("selected");
      }
    } else {
      setStep("selected");
    }
  }

  async function handleClarifyAnswer(answer: string) {
    if (sessionId && answer) {
      try {
        await apiFetch("/api/session/react", {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ session_id: sessionId, action: "note", note: answer }),
        });
      } catch { /* non-fatal */ }
    }
    setRejecting(null);
    setStep("results");
    handleRegenerate();
  }

  function handleStartOver() {
    setStep("intake");
    setCardStates([]);
    setSessionId("");
    setApiError("");
    setAvailabilityMsg("");
    setRejecting(null);
    setSelectedName("");
    setSelectedCard(null);
    setLikedNames([]);
    setVisuals(null);
    setVisualsError("");
    setGeneratingLogos(false);
    setIntakeValues(null);   // wipe form — only cleared on explicit Start Over
    localStorage.removeItem("nomvox_session_id");
  }

  async function handleRegenerateLogos() {
    if (!selectedCard || !lastIntake) return;
    setGeneratingLogos(true);
    try {
      const v = await apiFetch<VisualsResponse>("/api/visuals", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          session_id: sessionId,
          card: selectedCard,
          intake: lastIntake,
        }),
      });
      setVisuals(v);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      setVisualsError(msg);
    } finally {
      setGeneratingLogos(false);
    }
  }

  // generatingMoodboard tracks the targeted mood board regen (after logo selection)
  const [generatingMoodboard, setGeneratingMoodboard] = useState(false);

  async function handleRegenerateMoodboard(logoKey: string, logoStyle: string) {
    if (!selectedCard || !lastIntake) return;
    setGeneratingMoodboard(true);
    try {
      const v = await apiFetch<VisualsResponse>("/api/visuals", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          session_id: sessionId,
          card: selectedCard,
          intake: lastIntake,
          selected_logo_key: logoKey,
          selected_logo_style: logoStyle,
        }),
      });
      // Merge — only update mood_board; keep existing logos
      setVisuals(prev => prev ? { ...prev, mood_board: v.mood_board } : v);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      setVisualsError(msg);
    } finally {
      setGeneratingMoodboard(false);
    }
  }

  const isResultsScreen = step === "results" || step === "clarifying" || step === "selected" || step === "loading";

  return (
    <>
      {/* ── Star-field background ──────────────────────────────── */}
      <div className="fixed inset-0 -z-10 overflow-hidden">
        <Image src="/nomvox-bg.png" alt="" fill priority quality={90}
          className="object-cover object-center" style={{ opacity: 0.85 }} />
        <div className="absolute inset-0" style={{
          background:
            "radial-gradient(ellipse at 50% 0%, rgba(139,92,246,0.07) 0%, transparent 65%), " +
            "radial-gradient(ellipse at 50% 100%, rgba(5,7,15,0.7) 0%, transparent 70%)",
        }} />
      </div>

      <main className="flex min-h-screen flex-col items-center px-3 sm:px-6 pt-8 pb-16">

        {/* ── Logo — always visible ──────────────────────────── */}
        <div className="mb-6 flex flex-col items-center text-center">
          <Image src="/nomvox-logo2.png" alt="NomVox — Born from the Void"
            width={600} height={200} priority
            className="w-[280px] h-auto drop-shadow-[0_0_32px_rgba(139,92,246,0.55)]" />
          <p className="mt-2 text-sm" style={{ color: "var(--color-text-secondary)" }}>
            Name your brand in 60 seconds — handles, logos &amp; landing pages included.
          </p>
        </div>

        {/* ══════════════════════════════════════════════════════
            SCREEN 1 — INTAKE
        ══════════════════════════════════════════════════════ */}
        {!isResultsScreen && step !== "visuals-loading" && step !== "visuals" && (
          <div className="w-full max-w-2xl lg:max-w-3xl xl:max-w-4xl">

            {/* ── Return-to-results banner — shown when user navigated back from results */}
            {cardStates.length > 0 && (
              <div className="mb-4 flex items-center justify-between rounded-xl border px-5 py-3"
                style={{ background: "rgba(34,211,238,0.05)", borderColor: "rgba(34,211,238,0.22)" }}>
                <p className="text-sm font-semibold" style={{ color: "var(--color-signal)" }}>
                  ✦ Your previous results are still saved — edit inputs below or go back to names.
                </p>
                <button type="button"
                  onClick={() => setStep("results")}
                  className="ml-4 shrink-0 px-4 py-1.5 rounded-lg text-sm font-black transition-all"
                  style={{ background: "rgba(34,211,238,0.12)", border: "1px solid rgba(34,211,238,0.35)", color: "var(--color-signal)" }}
                  onMouseEnter={e => { e.currentTarget.style.background = "rgba(34,211,238,0.22)"; }}
                  onMouseLeave={e => { e.currentTarget.style.background = "rgba(34,211,238,0.12)"; }}>
                  ← Back to Names
                </button>
              </div>
            )}

            {/* Error from a previous attempt */}
            {apiError && (
              <div className="mb-4 rounded-xl border px-5 py-4 text-sm"
                style={{ background: "rgba(248,113,113,0.07)", borderColor: "rgba(248,113,113,0.35)", color: "#fca5a5" }}>
                <span className="font-bold">Error: </span>{apiError}
                <button type="button" onClick={() => setApiError("")}
                  className="ml-3 text-xs opacity-60 hover:opacity-100">✕ dismiss</button>
              </div>
            )}

            <div className="rounded-2xl border p-6 sm:p-8 backdrop-blur-sm"
              style={{
                background: "rgba(11,15,28,0.90)",
                borderColor: "rgba(139,92,246,0.28)",
                boxShadow: "0 0 0 1px rgba(139,92,246,0.10), 0 20px 50px rgba(5,7,15,0.65), 0 0 60px rgba(139,92,246,0.07)",
              }}>
              <IntakeForm
                onSubmit={handleGenerate}
                loading={false}
                initialValues={intakeValues ?? undefined}
              />
            </div>
          </div>
        )}

        {/* ══════════════════════════════════════════════════════
            SCREEN 2 — LOADING (animated dots, not skeletons)
        ══════════════════════════════════════════════════════ */}
        {step === "loading" && (
          <div className="w-full max-w-xl">
            <div className="rounded-2xl border p-10 text-center"
              style={{ background: "rgba(16,20,36,0.95)", borderColor: "rgba(139,92,246,0.25)" }}>
              <p className="text-2xl font-black tracking-tight mb-3"
                style={{ color: "var(--color-text-primary)" }}>
                NomVox is thinking
                <span className="nv-dot1" style={{ color: "var(--color-pulse)" }}>.</span>
                <span className="nv-dot2" style={{ color: "var(--color-pulse)" }}>.</span>
                <span className="nv-dot3" style={{ color: "var(--color-pulse)" }}>.</span>
              </p>
              <p className="text-sm font-medium mb-6" style={{ color: "var(--color-text-secondary)" }}>
                Generating names, scores &amp; availability checks — takes about 20 seconds.
              </p>
              <div className="flex flex-col gap-3">
                {[
                  { label: "✦  Crafting brand names",            color: "var(--color-pulse)" },
                  { label: "✦  Scoring memorability & safety",   color: "var(--color-signal)" },
                  { label: "✦  Probing domain + handles",        color: "#f59e0b" },
                  { label: "✦  Checking competitor radar",       color: "#4ade80" },
                ].map((item, i) => (
                  <div key={i} className="flex items-center gap-3 rounded-lg px-4 py-3"
                    style={{ background: "rgba(139,92,246,0.07)", border: "1px solid rgba(139,92,246,0.15)" }}>
                    <span className="text-sm font-semibold" style={{ color: item.color }}>{item.label}</span>
                    <span className="ml-auto text-xs font-bold" style={{ color: item.color, opacity: 0.6 }}>
                      <span className="nv-dot1">·</span><span className="nv-dot2">·</span><span className="nv-dot3">·</span>
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* ══════════════════════════════════════════════════════
            SCREEN 3 — RESULTS / CLARIFYING / SELECTED
        ══════════════════════════════════════════════════════ */}
        {(step === "results" || step === "clarifying" || step === "selected") && (cardStates.length > 0 || availabilityMsg) && (
          <div className="w-full max-w-full xl:max-w-[1400px]">

            {/* ── Results panel wrapper — brighter bg to distinguish ── */}
            <div className="rounded-2xl border overflow-hidden"
              style={{
                background: "rgba(18,22,42,0.97)",
                borderColor: "rgba(99,210,255,0.22)",
                boxShadow: "0 0 0 1px rgba(34,211,238,0.08), 0 24px 64px rgba(5,7,15,0.7)",
              }}>

              {/* ── Results header bar ─────────────────────────── */}
              <div className="flex items-center justify-between px-6 py-4 border-b"
                style={{ borderColor: "rgba(255,255,255,0.07)", background: "rgba(14,18,32,0.60)" }}>

                <div className="flex items-center gap-3">
                  {/* Soft back — returns to intake with form pre-filled, session preserved */}
                  <button type="button" onClick={handleBackToIntake}
                    className="flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-bold transition-all"
                    style={{ border: "1px solid rgba(99,210,255,0.30)", color: "var(--color-signal)", background: "transparent" }}
                    onMouseEnter={e => { e.currentTarget.style.background = "rgba(34,211,238,0.10)"; e.currentTarget.style.borderColor = "var(--color-signal)"; }}
                    onMouseLeave={e => { e.currentTarget.style.background = "transparent"; e.currentTarget.style.borderColor = "rgba(99,210,255,0.30)"; }}>
                    ← Edit Inputs
                  </button>

                  <span className="text-base font-bold" style={{ color: "var(--color-text-primary)" }}>
                    {cardStates.length} name{cardStates.length !== 1 ? "s" : ""} generated
                  </span>
                </div>

                {/* Regenerate — disabled while a generation is already running */}
                <button type="button" onClick={handleRegenerate}
                  disabled={generating}
                  className="flex items-center gap-2 px-5 py-2 rounded-lg text-sm font-black transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                  style={{
                    background: "linear-gradient(135deg, rgba(109,40,217,0.85) 0%, rgba(30,90,180,0.80) 100%)",
                    border: "1px solid rgba(139,92,246,0.40)",
                    color: "#fff",
                    boxShadow: "0 0 14px rgba(80,50,180,0.30)",
                  }}
                  onMouseEnter={e => { if (!generating) { e.currentTarget.style.background = "linear-gradient(135deg, rgba(139,92,246,0.90) 0%, rgba(34,150,220,0.85) 100%)"; e.currentTarget.style.boxShadow = "0 0 24px rgba(100,60,220,0.50)"; }}}
                  onMouseLeave={e => { e.currentTarget.style.background = "linear-gradient(135deg, rgba(109,40,217,0.85) 0%, rgba(30,90,180,0.80) 100%)"; e.currentTarget.style.boxShadow = "0 0 14px rgba(80,50,180,0.30)"; }}>
                  {generating ? <>Generating<span className="nv-dot1">.</span><span className="nv-dot2">.</span><span className="nv-dot3">.</span></> : "↻ Regenerate"}
                </button>
              </div>

              {/* ── Liked summary banner (when ≥1 liked) ──────── */}
              {likedNames.length > 0 && (
                <div className="px-6 py-3 border-b"
                  style={{ borderColor: "rgba(74,222,128,0.15)", background: "rgba(74,222,128,0.06)" }}>
                  <p className="text-sm font-semibold" style={{ color: "#4ade80" }}>
                    ♥ Liked so far:{" "}
                    {likedNames.map((n, i) => (
                      <span key={n}>
                        <span className="font-black">{n}</span>
                        {i < likedNames.length - 1 ? ", " : ""}
                      </span>
                    ))}
                    {" "}— hit Regenerate to get more in this direction.
                  </p>
                </div>
              )}

              {/* ── Availability timeout / partial fallback message ── */}
              {availabilityMsg && (
                <div className="px-6 py-3 border-b flex items-start gap-3"
                  style={{ borderColor: "rgba(245,158,11,0.20)", background: "rgba(245,158,11,0.06)" }}>
                  <span className="mt-0.5 text-base leading-none" style={{ color: "#f59e0b" }}>⚠</span>
                  <div className="flex-1">
                    <p className="text-sm font-semibold" style={{ color: "#fcd34d" }}>{availabilityMsg}</p>
                  </div>
                  <button type="button" onClick={() => setAvailabilityMsg("")}
                    className="text-xs font-bold shrink-0"
                    style={{ color: "rgba(245,158,11,0.50)" }}>✕</button>
                </div>
              )}

              {/* ── Selected name banner ───────────────────────── */}
              {step === "selected" && (
                <div className="px-6 py-5 border-b"
                  style={{ borderColor: "rgba(139,92,246,0.25)", background: "rgba(139,92,246,0.08)" }}>
                  <p className="text-xs font-bold uppercase tracking-widest mb-1"
                    style={{ color: "var(--color-pulse)" }}>✓ Name Selected</p>
                  <p className="text-3xl font-black tracking-tight"
                    style={{ color: "var(--color-text-primary)" }}>{selectedName}</p>
                  {visualsError && (
                    <p className="mt-1 text-sm" style={{ color: "#fca5a5" }}>
                      Visual generation failed: {visualsError} — retry below or export manually.
                    </p>
                  )}
                  <p className="mt-1 text-sm" style={{ color: "var(--color-text-secondary)" }}>
                    Visual identity, logos &amp; landing page coming next.
                  </p>
                  <button type="button"
                    onClick={() => selectedCard && lastIntake && handleSelect(selectedName)}
                    className="mt-2 px-4 py-2 rounded-lg text-sm font-bold transition-all"
                    style={{
                      background: "linear-gradient(135deg, rgba(109,40,217,0.85) 0%, rgba(30,90,180,0.80) 100%)",
                      border: "1px solid rgba(139,92,246,0.40)",
                      color: "#fff",
                    }}>
                    ✦ Generate Visual Identity
                  </button>
                  <button type="button" onClick={() => setStep("results")}
                    className="ml-3 mt-2 text-sm font-bold transition-colors"
                    style={{ color: "var(--color-text-hint)" }}
                    onMouseEnter={e => { e.currentTarget.style.color = "var(--color-signal)"; }}
                    onMouseLeave={e => { e.currentTarget.style.color = "var(--color-text-hint)"; }}>
                    ← Back to all names
                  </button>
                </div>
              )}

              {/* ── Clarifying dialog ──────────────────────────── */}
              {step === "clarifying" && rejecting && (
                <div className="px-6 py-5 border-b" style={{ borderColor: "rgba(139,92,246,0.20)" }}>
                  <RejectionDialog
                    rejectedName={rejecting.name}
                    clarifyingQuestion={rejecting.q}
                    onAnswer={handleClarifyAnswer}
                    onSkip={() => { setRejecting(null); setStep("results"); handleRegenerate(); }}
                  />
                </div>
              )}

              {/* ── Style DNA Sliders ──────────────────────────── */}
              <div className="px-6 py-4 border-b"
                style={{ borderColor: "rgba(255,255,255,0.06)", background: "rgba(14,18,32,0.40)" }}>
                <StyleDNASlider dna={styleDNA} onChange={setStyleDNA} />
              </div>

              {/* ── Name cards — full-width table rows ────────── */}
              <div className="p-4">
                <div className="flex flex-col gap-4">
                  {cardStates.map((cs, i) => (
                    <NameCardComponent
                      key={cs.card.name + i}
                      card={cs.card}
                      index={i}
                      availability={cs.availability}
                      availabilityLoading={cs.availabilityLoading}
                      reaction={cs.reaction}
                      rejectedReason={cs.rejectedReason}
                      onLike={handleLike}
                      onReject={handleReject}
                      onSelect={handleSelect}
                    />
                  ))}
                </div>
              </div>

            </div>{/* end results panel */}
          </div>
        )}

        {/* ══════════════════════════════════════════════════════
            SCREEN 4 — VISUALS LOADING (animated dots)
        ══════════════════════════════════════════════════════ */}
        {step === "visuals-loading" && (
          <div className="w-full max-w-xl">
            <div className="rounded-2xl border p-10 text-center"
              style={{ background: "rgba(18,22,42,0.97)", borderColor: "rgba(99,210,255,0.22)" }}>
              <p className="text-xl font-black mb-1"
                style={{ color: "var(--color-text-primary)" }}>{selectedName}</p>
              <p className="text-2xl font-black mb-4"
                style={{ color: "var(--color-pulse)" }}>
                Building your visual identity
                <span className="nv-dot1">.</span><span className="nv-dot2">.</span><span className="nv-dot3">.</span>
              </p>
              <div className="flex flex-col gap-3 text-left mb-4">
                {[
                  { label: "🎨  Generating mood board (×4)",     color: "#f59e0b" },
                  { label: "🔷  Creating logo concepts (×3)",     color: "var(--color-signal)" },
                  { label: "📄  Building landing page mockup",    color: "var(--color-pulse)" },
                  { label: "👤  Crafting brand persona",          color: "#4ade80" },
                ].map(item => (
                  <div key={item.label} className="flex items-center gap-3 rounded-lg px-4 py-3"
                    style={{ background: "rgba(139,92,246,0.07)", border: "1px solid rgba(139,92,246,0.15)" }}>
                    <span className="text-sm font-semibold" style={{ color: item.color }}>{item.label}</span>
                    <span className="ml-auto text-xs" style={{ color: item.color, opacity: 0.7 }}>
                      <span className="nv-dot1">·</span><span className="nv-dot2">·</span><span className="nv-dot3">·</span>
                    </span>
                  </div>
                ))}
              </div>
              <p className="text-xs" style={{ color: "var(--color-text-hint)" }}>
                This takes about 30–60 seconds — Imagen 3 is running.
              </p>
            </div>
          </div>
        )}

        {/* ══════════════════════════════════════════════════════
            SCREEN 5 — VISUAL IDENTITY (3-step sub-flow)
        ══════════════════════════════════════════════════════ */}
        {step === "visuals" && visuals && selectedCard && lastIntake && (
          <VisualIdentityPanel
            onRegenerateMoodboard={handleRegenerateMoodboard}
            generatingMoodboard={generatingMoodboard}
            brandName={selectedName}
            card={selectedCard}
            intake={lastIntake}
            sessionId={sessionId}
            visuals={visuals}
            onBack={() => setStep("results")}
            onStartOver={handleStartOver}
            onRegenerateLogos={handleRegenerateLogos}
            generatingLogos={generatingLogos}
          />
        )}

      </main>
    </>
  );
}
