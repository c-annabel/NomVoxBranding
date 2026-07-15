"use client";

import { useState, useEffect } from "react";
import { z } from "zod";
import type { IntakePayload } from "@/lib/types";
import { EMPTY_INTAKE } from "@/lib/types";
import PromptRichnessMeter from "./PromptRichnessMeter";

const schema = z.object({
  core_idea: z.string().min(2, "Please describe your core idea (at least 2 characters)."),
});

const EXAMPLES: Partial<IntakePayload>[] = [
  {
    core_idea: "eco-friendly coffee brand",
    target_audience: "young professionals 25–35",
    personality: "minimal, warm, honest",
    style: "earthy wordmark",
    industry: "food & beverage",
    color_mood: "green, cream, terracotta",
    name_length: "1 word, under 8 chars",
    avoid: "leaf, eco, green in the name",
  },
  {
    core_idea: "indie game studio specialising in puzzle games",
    target_audience: "casual gamers 18–40",
    personality: "playful, clever, slightly nerdy",
    style: "pixel art inspired, modern wordmark",
    industry: "gaming",
    color_mood: "electric blue, neon yellow, dark bg",
    name_length: "1–2 words",
    avoid: "pixel, game, play",
  },
  {
    core_idea: "online tutoring platform for kids aged 7–14",
    target_audience: "parents of school-age children",
    personality: "friendly, trustworthy, encouraging",
    style: "rounded, approachable",
    industry: "edtech",
    color_mood: "bright orange, sky blue, white",
    name_length: "no preference",
    avoid: "school, tutor, learn",
  },
  {
    core_idea: "B2B SaaS tool for invoice automation",
    target_audience: "small business owners, freelancers",
    personality: "professional, efficient, no-nonsense",
    style: "clean tech wordmark",
    industry: "fintech",
    color_mood: "deep navy, white, accent teal",
    name_length: "under 8 chars, invented word ok",
    avoid: "invoice, pay, bill",
  },
  {
    core_idea: "sustainable streetwear brand",
    target_audience: "Gen Z fashion-conscious shoppers",
    personality: "bold, rebellious, conscious",
    style: "abstract emblem, uppercase wordmark",
    industry: "fashion & apparel",
    color_mood: "black, off-white, acid green pop",
    name_length: "short punchy word or initialism",
    avoid: "eco, green, sustainable",
  },
];

// Field definitions — core_idea is index 0 (required), rest are 1-indexed optional
const FIELDS: {
  key: keyof IntakePayload;
  label: string;
  required?: boolean;
  hint: string;
  placeholder: string;
  span?: "full";
}[] = [
  {
    key: "core_idea",
    label: "Core Idea",
    required: true,
    hint: "What does this brand do or sell?",
    placeholder: "e.g. eco-friendly coffee · indie game studio · online tutoring for kids",
    span: "full",
  },
  {
    key: "target_audience",
    label: "Target Audience",
    hint: "Who is the primary customer?",
    placeholder: "e.g. young professionals 25–35 · parents of toddlers · B2B SaaS teams",
  },
  {
    key: "personality",
    label: "Brand Personality",
    hint: "What emotional tone should the name carry?",
    placeholder: "e.g. playful, bold · minimal, premium · warm, community-driven",
  },
  {
    key: "style",
    label: "Style / Aesthetic",
    hint: "Visual and naming style preference",
    placeholder: "e.g. modern wordmark · nature-inspired · techy acronym · made-up word",
  },
  {
    key: "industry",
    label: "Industry / Category",
    hint: "Helps avoid clashes with existing brands",
    placeholder: "e.g. food & beverage · fintech · health & wellness",
  },
  {
    key: "color_mood",
    label: "Colour Mood",
    hint: "Seeds mood board and logo palette",
    placeholder: "e.g. earth tones, green, cream · electric blue, white · no preference",
  },
  {
    key: "name_length",
    label: "Name Length Preference",
    hint: "Short names are easier handles; longer carry more meaning",
    placeholder: "e.g. 1–6 chars · 1 word · 2 words ok · no preference",
  },
  {
    key: "avoid",
    label: "Words / Sounds to Avoid",
    hint: "Prevents suggestions you'll immediately reject",
    placeholder: "e.g. avoid: green, eco, leaf · no made-up words · nothing religious",
    span: "full",
  },
];

interface IntakeFormProps {
  onSubmit: (payload: IntakePayload) => void;
  loading?: boolean;
  /** When provided, the form seeds from these values (used when navigating back from results). */
  initialValues?: IntakePayload;
}

export default function IntakeForm({ onSubmit, loading = false, initialValues }: IntakeFormProps) {
  const [values, setValues] = useState<IntakePayload>(initialValues ?? EMPTY_INTAKE);
  const [errors, setErrors] = useState<Partial<Record<keyof IntakePayload, string>>>({});
  const [exampleIdx, setExampleIdx] = useState(0);

  // Re-seed the form if the parent passes updated initialValues (e.g. navigating back
  // from results after the user had already submitted once). Only fires when
  // initialValues reference changes — not on every render.
  useEffect(() => {
    if (initialValues) {
      setValues(initialValues);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialValues]);

  const filledOptionalCount = FIELDS.filter(
    (f) => !f.required && values[f.key].trim() !== ""
  ).length;

  function handleChange(key: keyof IntakePayload, value: string) {
    setValues((prev) => ({ ...prev, [key]: value }));
    if (errors[key]) setErrors((prev) => ({ ...prev, [key]: undefined }));
  }

  function handleInspire() {
    setValues({ ...EMPTY_INTAKE, ...EXAMPLES[exampleIdx % EXAMPLES.length] });
    setExampleIdx((i) => i + 1);
    setErrors({});
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const result = schema.safeParse(values);
    if (!result.success) {
      const fieldErrors: Partial<Record<keyof IntakePayload, string>> = {};
      result.error.issues.forEach((i) => {
        fieldErrors[i.path[0] as keyof IntakePayload] = i.message;
      });
      setErrors(fieldErrors);
      return;
    }
    onSubmit(values);
  }

  // Question numbers: core_idea = Q1, rest = Q2–Q8
  const questionNumber = (idx: number) => idx + 1;

  return (
    <form onSubmit={handleSubmit} className="w-full space-y-6">

      {/* ── Section header ─────────────────────────────────────── */}
      <div>
        <h2 className="text-2xl font-black tracking-tight"
          style={{ color: "var(--color-text-primary)" }}>
          Describe your brand idea
        </h2>
        <p className="mt-1 text-base" style={{ color: "var(--color-text-secondary)" }}>
          One required field — the rest sharpen the result.
        </p>
      </div>

      {/* ── Richness meter ─────────────────────────────────────── */}
      <PromptRichnessMeter filledCount={filledOptionalCount} totalOptional={7} />

      {/* ── Fields grid ────────────────────────────────────────── */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
        {FIELDS.map((field, idx) => {
          const isCoreIdea = field.key === "core_idea";
          return (
            <div key={field.key} className={field.span === "full" ? "sm:col-span-2" : ""}>

              {/* Question number + label row */}
              <div className="flex items-center gap-2 mb-1">
                {/* Number badge */}
                <span
                  className="shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-[11px] font-black"
                  style={{
                    background: isCoreIdea
                      ? "rgba(251,146,60,0.20)"
                      : "rgba(139,92,246,0.18)",
                    color: isCoreIdea ? "#fb923c" : "var(--color-pulse)",
                    border: isCoreIdea
                      ? "1px solid rgba(251,146,60,0.35)"
                      : "1px solid rgba(139,92,246,0.30)",
                  }}
                >
                  {questionNumber(idx)}
                </span>

                {/* Label */}
                <label
                  htmlFor={field.key}
                  className="text-sm font-bold uppercase tracking-widest"
                  style={{
                    // Core idea label uses orange gradient text
                    color: isCoreIdea ? "transparent" : "var(--color-pulse)",
                    background: isCoreIdea
                      ? "linear-gradient(90deg, #fb923c, #f97316)"
                      : "none",
                    WebkitBackgroundClip: isCoreIdea ? "text" : undefined,
                    backgroundClip: isCoreIdea ? "text" : undefined,
                  }}
                >
                  {field.label}
                </label>

                {field.required && (
                  <span className="text-xs font-bold" style={{ color: "var(--color-signal)" }}>
                    required
                  </span>
                )}
              </div>

              {/* Hint text */}
              <p className="text-sm mb-2 leading-snug ml-8" style={{ color: "var(--color-text-secondary)" }}>
                {field.hint}
              </p>

              {/* Input — core_idea gets orange-gradient border */}
              <input
                id={field.key}
                type="text"
                value={values[field.key]}
                onChange={(e) => handleChange(field.key, e.target.value)}
                placeholder={field.placeholder}
                disabled={loading}
                className={`nv-input${errors[field.key] ? " nv-input-error" : ""}${isCoreIdea ? " nv-input-core" : ""}`}
                style={isCoreIdea ? {
                  borderColor: "rgba(251,146,60,0.55)",
                  boxShadow: "0 0 0 1px rgba(251,146,60,0.12) inset",
                } : {}}
                onFocus={isCoreIdea ? (e) => {
                  e.currentTarget.style.borderColor = "#fb923c";
                  e.currentTarget.style.boxShadow = "0 0 0 3px rgba(251,146,60,0.18), 0 0 16px rgba(251,146,60,0.12)";
                } : undefined}
                onBlur={isCoreIdea ? (e) => {
                  e.currentTarget.style.borderColor = "rgba(251,146,60,0.55)";
                  e.currentTarget.style.boxShadow = "0 0 0 1px rgba(251,146,60,0.12) inset";
                } : undefined}
              />

              {errors[field.key] && (
                <p className="mt-1 text-sm font-medium text-red-400">{errors[field.key]}</p>
              )}
            </div>
          );
        })}
      </div>

      {/* ── Actions ────────────────────────────────────────────── */}
      <div className="flex flex-wrap items-center gap-3 pt-1">
        {/* Generate — dimmed purple-blue matching logo V+X */}
        <button
          type="submit"
          disabled={loading}
          className="px-8 py-3 rounded-lg text-base font-black tracking-wide
            text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          style={{
            background: loading
              ? "rgba(80,50,160,0.5)"
              : "linear-gradient(135deg, rgba(109,40,217,0.85) 0%, rgba(30,90,180,0.80) 100%)",
            border: "1px solid rgba(139,92,246,0.40)",
            boxShadow: loading ? "none" : "0 0 18px rgba(80,50,180,0.30)",
          }}
          onMouseEnter={e => {
            if (!loading) {
              e.currentTarget.style.background = "linear-gradient(135deg, rgba(139,92,246,0.90) 0%, rgba(34,150,220,0.85) 100%)";
              e.currentTarget.style.boxShadow = "0 0 28px rgba(100,60,220,0.45)";
            }
          }}
          onMouseLeave={e => {
            if (!loading) {
              e.currentTarget.style.background = "linear-gradient(135deg, rgba(109,40,217,0.85) 0%, rgba(30,90,180,0.80) 100%)";
              e.currentTarget.style.boxShadow = "0 0 18px rgba(80,50,180,0.30)";
            }
          }}
        >
          {loading ? "Generating…" : "Generate Names"}
        </button>

        {/* Inspire me — border matches core idea orange */}
        <button
          type="button"
          onClick={handleInspire}
          disabled={loading}
          className="px-5 py-3 rounded-lg text-base font-semibold transition-all disabled:opacity-50"
          style={{
            border: "1px solid rgba(251,146,60,0.45)",
            color: "#fb923c",
            background: "transparent",
          }}
          onMouseEnter={e => {
            e.currentTarget.style.borderColor = "#fb923c";
            e.currentTarget.style.background = "rgba(251,146,60,0.08)";
            e.currentTarget.style.boxShadow = "0 0 12px rgba(251,146,60,0.20)";
          }}
          onMouseLeave={e => {
            e.currentTarget.style.borderColor = "rgba(251,146,60,0.45)";
            e.currentTarget.style.background = "transparent";
            e.currentTarget.style.boxShadow = "none";
          }}
        >
          Inspire me ✦
        </button>

        {/* Clear — dark pink / magenta hover */}
        <button
          type="button"
          onClick={() => { setValues(EMPTY_INTAKE); setErrors({}); }}
          disabled={loading}
          className="px-5 py-3 rounded-lg text-base font-semibold transition-all disabled:opacity-50"
          style={{
            border: "1px solid #9d3060",
            color: "#c4607a",
            background: "transparent",
          }}
          onMouseEnter={e => {
            e.currentTarget.style.borderColor = "#e040a0";
            e.currentTarget.style.color = "#e040a0";
            e.currentTarget.style.background = "rgba(224,64,160,0.08)";
          }}
          onMouseLeave={e => {
            e.currentTarget.style.borderColor = "#9d3060";
            e.currentTarget.style.color = "#c4607a";
            e.currentTarget.style.background = "transparent";
          }}
        >
          Clear ✕
        </button>
      </div>
    </form>
  );
}
