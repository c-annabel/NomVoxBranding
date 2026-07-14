"use client";

import { useState } from "react";
import { z } from "zod";
import type { IntakePayload } from "@/lib/types";
import { EMPTY_INTAKE } from "@/lib/types";
import PromptRichnessMeter from "./PromptRichnessMeter";

// ── Validation ────────────────────────────────────────────────────
const schema = z.object({
  core_idea: z.string().min(2, "Please describe your core idea (at least 2 characters)."),
});

// ── Inspire-me examples ───────────────────────────────────────────
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

// ── Field definitions ─────────────────────────────────────────────
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
    label: "Core idea",
    required: true,
    hint: "What does this brand do or sell?",
    placeholder: "e.g. eco-friendly coffee · indie game studio · online tutoring for kids",
    span: "full",
  },
  {
    key: "target_audience",
    label: "Target audience",
    hint: "Who is the primary customer?",
    placeholder: "e.g. young professionals 25–35 · parents of toddlers · B2B SaaS teams",
  },
  {
    key: "personality",
    label: "Brand personality",
    hint: "What emotional tone should the name carry?",
    placeholder: "e.g. playful, bold · minimal, premium · warm, community-driven",
  },
  {
    key: "style",
    label: "Style / aesthetic",
    hint: "Visual and naming style preference",
    placeholder: "e.g. modern wordmark · nature-inspired · techy acronym · made-up word",
  },
  {
    key: "industry",
    label: "Industry / category",
    hint: "Helps avoid clashes with existing brands",
    placeholder: "e.g. food & beverage · fintech · health & wellness",
  },
  {
    key: "color_mood",
    label: "Colour mood",
    hint: "Seeds mood board and logo palette",
    placeholder: "e.g. earth tones, green, cream · electric blue, white · no preference",
  },
  {
    key: "name_length",
    label: "Name length preference",
    hint: "Short names are easier handles; longer carry more meaning",
    placeholder: "e.g. 1–6 chars · 1 word · 2 words ok · no preference",
  },
  {
    key: "avoid",
    label: "Words / sounds to avoid",
    hint: "Prevents suggestions you'll immediately reject",
    placeholder: "e.g. avoid: green, eco, leaf · no made-up words · nothing religious",
    span: "full",
  },
];

interface IntakeFormProps {
  onSubmit: (payload: IntakePayload) => void;
  loading?: boolean;
}

export default function IntakeForm({ onSubmit, loading = false }: IntakeFormProps) {
  const [values, setValues] = useState<IntakePayload>(EMPTY_INTAKE);
  const [errors, setErrors] = useState<Partial<Record<keyof IntakePayload, string>>>({});
  const [exampleIdx, setExampleIdx] = useState(0);

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

  return (
    <form onSubmit={handleSubmit} className="w-full space-y-5">

      {/* ── Section header ─────────────────────────────────────── */}
      <div>
        <h2
          className="text-lg font-bold tracking-tight"
          style={{ color: "var(--color-text-primary)" }}
        >
          Describe your brand idea
        </h2>
        <p className="mt-0.5 text-sm" style={{ color: "var(--color-text-secondary)" }}>
          One required field — the rest sharpen the result.
        </p>
      </div>

      {/* ── Richness meter ─────────────────────────────────────── */}
      <PromptRichnessMeter filledCount={filledOptionalCount} totalOptional={7} />

      {/* ── Fields grid ────────────────────────────────────────── */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {FIELDS.map((field) => (
          <div key={field.key} className={field.span === "full" ? "sm:col-span-2" : ""}>
            {/* Label + required tag */}
            <div className="flex items-baseline gap-1.5 mb-0.5">
              <label
                htmlFor={field.key}
                className="text-[11px] font-bold uppercase tracking-widest"
                style={{ color: "var(--color-pulse)" }}
              >
                {field.label}
              </label>
              {field.required && (
                <span className="text-[11px] font-bold" style={{ color: "var(--color-signal)" }}>
                  required
                </span>
              )}
            </div>

            {/* Hint text */}
            <p className="text-[12px] mb-1.5 leading-snug" style={{ color: "var(--color-text-secondary)" }}>
              {field.hint}
            </p>

            {/* Input */}
            <input
              id={field.key}
              type="text"
              value={values[field.key]}
              onChange={(e) => handleChange(field.key, e.target.value)}
              placeholder={field.placeholder}
              disabled={loading}
              className={`nv-input${errors[field.key] ? " nv-input-error" : ""}`}
            />

            {errors[field.key] && (
              <p className="mt-1 text-xs font-medium text-red-400">{errors[field.key]}</p>
            )}
          </div>
        ))}
      </div>

      {/* ── Actions ────────────────────────────────────────────── */}
      <div className="flex flex-wrap items-center gap-3">
        <button
          type="submit"
          disabled={loading}
          className="px-8 py-3 rounded-md text-sm font-bold tracking-wide
            text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          style={{
            background: loading
              ? "var(--color-pulse-lo)"
              : "linear-gradient(135deg, var(--color-pulse), var(--color-signal))",
            boxShadow: loading ? "none" : "0 0 20px rgba(139,92,246,0.35)",
          }}
        >
          {loading ? "Generating…" : "Generate Names"}
        </button>

        <button
          type="button"
          onClick={handleInspire}
          disabled={loading}
          className="px-5 py-3 rounded-md text-sm font-semibold transition-all
            disabled:opacity-50"
          style={{
            border: "1px solid rgba(139,92,246,0.35)",
            color: "var(--color-text-secondary)",
            background: "transparent",
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLButtonElement).style.borderColor = "var(--color-signal)";
            (e.currentTarget as HTMLButtonElement).style.color = "var(--color-signal)";
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLButtonElement).style.borderColor = "rgba(139,92,246,0.35)";
            (e.currentTarget as HTMLButtonElement).style.color = "var(--color-text-secondary)";
          }}
        >
          Inspire me ✦
        </button>

        <button
          type="button"
          onClick={() => { setValues(EMPTY_INTAKE); setErrors({}); }}
          disabled={loading}
          className="px-5 py-3 rounded-md text-sm font-semibold transition-all
            disabled:opacity-50"
          style={{
            border: "1px solid #9d3060",
            color: "#c4607a",
            background: "transparent",
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLButtonElement).style.borderColor = "#e040a0";
            (e.currentTarget as HTMLButtonElement).style.color = "#e040a0";
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLButtonElement).style.borderColor = "#9d3060";
            (e.currentTarget as HTMLButtonElement).style.color = "#c4607a";
          }}
        >
          Clear ✕
        </button>
      </div>
    </form>
  );
}
