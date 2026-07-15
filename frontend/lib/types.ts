// ── Intake form ───────────────────────────────────────────────────
export interface IntakePayload {
  core_idea: string;          // required
  target_audience: string;
  personality: string;
  style: string;
  industry: string;
  color_mood: string;
  name_length: string;
  avoid: string;
}

export const EMPTY_INTAKE: IntakePayload = {
  core_idea: "",
  target_audience: "",
  personality: "",
  style: "",
  industry: "",
  color_mood: "",
  name_length: "",
  avoid: "",
};

// ── Name cards ────────────────────────────────────────────────────
export interface BrandScore {
  memorability: number;
  spellability: number;
  global_safety: number;
  squatter_risk: "Low" | "Medium" | "High";
  mem_reasoning: string;
  spell_reasoning: string;
  global_reasoning: string;
  squatter_reasoning: string;
}

export interface VoiceSamples {
  instagram_caption: string;
  email_subject: string;
  not_found_message: string;
}

export interface NameCard {
  name: string;
  tagline: string;
  tone_reasoning: string;
  style_tags: string[];
  short_desc: string;
  long_desc: string;
  origin_story: string;
  score: BrandScore;
  voice_samples: VoiceSamples;
}

// ── Availability ──────────────────────────────────────────────────
export interface PlatformProbes {
  domain: boolean;
  instagram: boolean;
  x: boolean;
  tiktok: boolean;
  threads: boolean;
  youtube: boolean;
  domain_unknown: boolean;
  instagram_unknown: boolean;
  x_unknown: boolean;
  tiktok_unknown: boolean;
  threads_unknown: boolean;
  youtube_unknown: boolean;
}

export interface AvailabilityResult {
  name: string;
  probes: PlatformProbes;
  score: number;   // 0–100
  passes: boolean; // score >= 80
  radar: string;   // competitor name radar warning
}

// ── Session ───────────────────────────────────────────────────────
export interface SessionReactPayload {
  session_id: string;
  name?: string;
  action: "like" | "reject" | "note" | "visual-note" | "slider" | "select";
  note?: string;
  slider_playful?: number;
  slider_abstract?: number;
}

// ── API responses ─────────────────────────────────────────────────
export interface GenerateResponse {
  session_id: string;
  cards: NameCard[];
}

export interface SessionResponse {
  session_id: string;
}

export interface ReactResponse {
  ok: boolean;
  clarifying_question?: string; // only set on reject+note
}

// ── UI state machine ──────────────────────────────────────────────
export type AppStep =
  | "intake"       // user filling the form
  | "generating"   // AI generating names
  | "results"      // name cards shown, user selecting/rejecting
  | "clarifying"   // AI asked a question after rejection
  | "selected"     // user selected a name → visual identity phase
  | "visuals"      // visuals being generated / shown
  | "export";      // export phase

export interface NameCardState {
  card: NameCard;
  availability?: AvailabilityResult;
  reaction?: "liked" | "rejected" | null;
}

// ── Style DNA Sliders ─────────────────────────────────────────────
export interface StyleDNA {
  playful: number;   // 0 = premium, 1 = playful
  abstract: number;  // 0 = descriptive, 1 = abstract
}

// ── Visual Identity ───────────────────────────────────────────────
export interface BrandPersona {
  age: number;
  occupation: string;
  voice: string;
  reads: string[];
  never_says: string[];
  core_values: string[];
}

export interface VisualsResponse {
  mood_board: string[];      // data URIs
  logo_profile: string;
  logo_app: string;
  logo_business: string;
  mockup_html: string;
  persona?: BrandPersona;
}

export interface ExportRequest {
  session_id: string;
  brand_name: string;
  card: NameCard;
  intake: IntakePayload;
  mood_board: string[];
  logo_profile: string;
  logo_app: string;
  logo_business: string;
  selected_logo_key: string; // "profile" | "app" | "business"
  mockup_html: string;
  persona?: BrandPersona;
}
