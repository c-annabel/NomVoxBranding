import { z } from "zod";
import type { NameCard, BrandScore } from "@/lib/types";

// ── Zod schema — validates each card from the API ─────────────────
const BrandScoreSchema = z.object({
  memorability:        z.number().min(1).max(10).default(5),
  spellability:        z.number().min(1).max(10).default(5),
  global_safety:       z.number().min(1).max(10).default(5),
  squatter_risk:       z.enum(["Low","Medium","High"]).default("Medium"),
  mem_reasoning:       z.string().default(""),
  spell_reasoning:     z.string().default(""),
  global_reasoning:    z.string().default(""),
  squatter_reasoning:  z.string().default(""),
});

const VoiceSamplesSchema = z.object({
  instagram_caption: z.string().default(""),
  email_subject:     z.string().default(""),
  not_found_message: z.string().default(""),
});

export const NameCardSchema = z.object({
  name:           z.string().min(1),
  tagline:        z.string().default(""),
  tone_reasoning: z.string().default(""),
  style_tags:     z.array(z.string()).default([]),
  short_desc:     z.string().default(""),
  long_desc:      z.string().default(""),
  origin_story:   z.string().default(""),
  score:          BrandScoreSchema.default({
    memorability: 5, spellability: 5, global_safety: 5, squatter_risk: "Medium" as const,
    mem_reasoning: "", spell_reasoning: "", global_reasoning: "", squatter_reasoning: "",
  }),
  voice_samples:  VoiceSamplesSchema.default({
    instagram_caption: "", email_subject: "", not_found_message: "",
  }),
});

export const GenerateResponseSchema = z.object({
  session_id: z.string(),
  cards:      z.array(NameCardSchema),
});

// Parse and silently drop malformed cards — never crash the UI.
export function parseGenerateResponse(raw: unknown): { session_id: string; cards: NameCard[] } {
  const result = GenerateResponseSchema.safeParse(raw);
  if (!result.success) {
    console.error("parseGenerateResponse:", result.error.issues);
    return { session_id: "", cards: [] };
  }
  return result.data as { session_id: string; cards: NameCard[] };
}
