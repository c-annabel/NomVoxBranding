# NomVox SDLC Plan v3

## Overview

NomVox is an AI-powered brand identity platform that transforms a raw business concept into a complete, launch-ready brand universe — validated names, social/domain handles, logo assets, mood boards, and landing page mockups — through a conversational creative-partner loop that learns from every user reaction.

**Sprint:** 15 days  
**Backend:** Go (net/http + chi router)  
**Frontend:** Next.js 14 + Tailwind CSS + Space Grotesk font  
**AI Core (text / reasoning):** IBM Granite via watsonx.ai — name generation, taglines, origin stories, brand scores, persona, voice samples, competitor radar, landing page HTML (required for challenge compliance)
**AI Core (vision + image gen):** Google AI Studio free tier — Gemini 2.0 Flash (vision image analysis, natively multimodal, no extra key) + Imagen 3 (mood board + logo image generation) — zero cost for demo prototype
**Session store:** Redis / Upstash  
**Deployment:** Vercel (frontend) + Fly.io (Go server)  
**Primary dev tool:** IBM Bob

---

## Brand Identity (Locked from Moodboard)

### Etymology
- **NOM** — Latin/French *nomen*: name, law, system. The foundation — structure, identity, registration.
- **VOX** — Latin: voice. Also references Alpha Vulpeculae (Anser), the brightest star in Vulpecula constellation — the cosmic voice. The expression — giving a voice to an idea in the cosmos.

### Tagline
`NAME IT. CLAIM IT. LAUNCH IT.` is the functional baseline.  
Elevated creative taglines to use in the app, pitch, and README:
- **"Your idea has a name. The universe just hasn't heard it yet."**
- **"From void to voice — your brand, synthesized."**
- **"The cosmos names things. Now so do you."**
- **"Signal sent. Brand received."** ← recommended for app hero (short, punchy, cosmic)

### Color Palette
| Token | Hex | Usage |
|---|---|---|
| `--color-void` | `#0B0F19` | Primary background (deep space) |
| `--color-nebula` | `#141824` | Cards, sidebar, secondary bg |
| `--color-pulse` | `#8B5CF6` | Primary accent — NOM purple, CTAs, active states |
| `--color-signal` | `#06B6D4` | Secondary accent — VOX cyan, links, badges |
| `--color-lunar` | `#E6E8F0` | Primary text, icons, UI chrome |

### Typography
- **Display / Logo:** Space Grotesk — Wide, Modern, Technical (Google Fonts, free)
- **Body:** Space Grotesk Regular / Medium
- **Monospace labels:** JetBrains Mono (availability badges, score readouts)

### Logo System
| Asset | Description | Usage |
|---|---|---|
| **Primary wordmark** | `NOM` white + `V` large purple chevron + `OX` cyan gradient + Vulpecula constellation arc above | Hero, README, landing page |
| **Icon mark** | Geometric `NVK` ligature (N+V purple, K-arm cyan) | Profile/social avatar, favicon |
| **App icon** | `NVK` mark on rounded-square `#141824` bg with purple-cyan corner glow | iOS/Android icon, browser tab |
| **Business lockup** | Icon mark + wordmark horizontal | Email signature, business card mockup |

---

## State Machine — User Interaction Flow

```
[Step 1: Keyword Intake Form]
    ↓ Submit (core_idea required, 7 optional fields)
[Step 2: AI Generation — Granite]
    ↓ Produces 8–12 NameCard objects
[Step A: Availability Gate — Go]
    ↓ 80% threshold filter (weighted score)
    → Names below 80%: silently discarded
    → Zero pass: auto-retry Granite (up to 3 cycles) → closest-match fallback
[Step B: Show Filtered Results to User]
    User chooses:
    ├── Option 1 (SELECT): Accept a name → proceed to Visual Identity
    ├── Option 2 (REJECT): Reject with optional note
    │       ↓
    │   [Anti-Name Reasoning: AI explains what it heard, asks 1 clarifying Q]
    │       ↓ User answers
    │   [Session memory updated: rejected[], direction_notes[]]
    │       ↓
    │   Loop back to AI Generation with full session context
    └── Option 3 (REPRODUCE): Fresh batch using accumulated session memory
            ↓ Loop back to AI Generation
[Step 3: Visual Identity — after SELECT]
    ├── Optional: Upload vision image → extracts palette/mood → seeds all visuals
    ├── Mood Board (4-panel, image-gen)
    ├── Logo Concepts (Profile / App / Business — 3 types, each regenerable)
    ├── Hero Landing-Page Mockup (Granite HTML/CSS, sandboxed iframe)
    └── Brand Voice Samples (Instagram caption, email subject, 404 message)
[Step C: Export]
    → ZIP: names, taglines, brand brief PDF, logos (SVG+PNG), mood board,
           hero HTML, color palette JSON, brand persona card, voice samples
```

---

## Feature Specification

| # | Feature | Build day(s) | MVP / Stretch |
|---|---|---|---|
| 1 | Keyword Intake Form (8 fields + prompt richness meter) | 2 | MVP |
| 2 | AI Name + Tagline generation (Granite, streamed) | 3–4 | MVP |
| 3 | Name Origin Story (etymology per card) | 3–4 | MVP |
| 4 | Brand Score Card (Memorability, Spellability, Global Safety, Squatter Risk) | 3–4 | MVP |
| 5 | 80% Availability Gate (Go weighted scoring, parallel probes) | 6–7 | MVP |
| 6 | Competitor Name Radar (Granite second-pass semantic overlap check) | 6–7 | MVP |
| 7 | Session Memory (liked/rejected/notes → Redis, injected every call) | 5 | MVP |
| 8 | Anti-Name Reasoning (AI explains rejection, asks 1 clarifying Q) | 5 | MVP |
| 9 | Select / Reject / Reproduce state machine (Step B UI) | 5 | MVP |
| 10 | Style DNA Slider (Playful↔Premium · Abstract↔Descriptive live regen) | 8 | MVP |
| 11 | Brand Persona Card (brand as a person — age, voice, reads, never says) | 8 | MVP |
| 12 | Vision Image Upload (post-name-selection, optional, seeds visuals) | 9 | MVP |
| 13 | Mood Board (4-panel image-gen, AI learns visual feedback) | 9–10 | MVP |
| 14 | Logo Concepts — Profile / App / Business (3 types, per-type regen) | 9–10 | MVP |
| 15 | Hero Landing-Page Mockup (Granite HTML/CSS, sandboxed iframe) | 11 | MVP |
| 16 | Brand Voice Samples (Instagram caption, email subject line, 404 message) | 8 | MVP |
| 17 | Short + Long Brand Descriptions (editable, per-field regen) | 4 | MVP |
| 18 | Export ZIP (Go archive/zip — all assets) | 12 | MVP |
| 19 | Screenshot of taken handles (Puppeteer → signed Vercel Blob URL) | 14 | Stretch |
| 20 | Color palette extractor from vision image | 9 | MVP (Gemini 2.0 Flash — free tier, natively multimodal, no extra API key beyond GOOGLE_AI_API_KEY) |

---

## Sub-Tasks

---

### ST-01: Environment, Scaffold & Repo [ ] pending

**Intent:** Establish the full monorepo structure, Go module, Next.js app, CI, secrets management, and shared type definitions. This is the foundation every other sub-task builds on.

**Expected Outcomes:**
- GitHub repo with Go `.gitignore` committed
- Go module initialised at `github.com/your-org/nomvox`
- Next.js 14 app scaffolded inside `frontend/` with Tailwind, Space Grotesk, shadcn/ui
- GitHub Actions CI: `go build ./...` + `go test ./...` + `npm run build`
- Upstash Redis provisioned; connection string in GitHub secrets
- All shared Go structs defined in `internal/models/types.go`
- All shared TypeScript types defined in `frontend/lib/types.ts`
- `.env.example` committed with all required keys documented

**Todo List:**
1. Create GitHub repo, add Go `.gitignore` (binaries, test artifacts, `.env`, `vendor/`, `.next/`, `node_modules/`)
2. `go mod init github.com/your-org/nomvox`; add deps: `chi`, `go-redis`, `encoding/json`, `archive/zip`, `sync`
3. Create `cmd/server/main.go` with chi router skeleton and health-check `/api/ping`
4. Scaffold `frontend/` with `npx create-next-app@latest`, add Tailwind, install `@fontsource/space-grotesk`, shadcn/ui init
5. Define Go structs: `IntakePayload`, `NameCard`, `AvailabilityResult`, `PlatformResult`, `BrandSession`, `VisionContext`, `VisualPack`
6. Define TypeScript interfaces mirroring the Go structs
7. Configure GitHub Actions workflow: lint + build + test on push to `main` and `dev`
8. Provision Upstash Redis; add `REDIS_URL` to secrets; test connection from Go health-check
9. Write `.env.example` documenting: `REDIS_URL`, `WATSONX_API_KEY`, `WATSONX_PROJECT_ID`, `GOOGLE_AI_API_KEY` (Gemini 2.0 Flash for vision + Imagen 3 for image generation — both free tier via Google AI Studio, single key), `VERCEL_BLOB_TOKEN`
10. Set up brand token CSS variables in `frontend/styles/globals.css` (all 5 colour tokens + Space Grotesk import)

**Relevant Context:**
- Go file layout: `cmd/server/`, `internal/handlers/`, `internal/ai/`, `internal/availability/`, `internal/session/`, `internal/models/`
- Frontend layout: `frontend/app/`, `frontend/components/`, `frontend/lib/`
- All colour tokens defined in section "Brand Identity — Color Palette" above

---

### ST-02: Keyword Intake Form [ ] pending

**Intent:** Build the structured 8-field intake form that replaces a free-text prompt box. This is the critical first step — gathering enough context in one shot to avoid thin-prompt regeneration loops. The Go handler assembles a deterministic, rich LLM prompt from the structured payload.

**Expected Outcomes:**
- 8-field form renders correctly on mobile (375px) and desktop (1280px)
- Prompt richness meter updates in real time as fields are filled (1 field = red, 3 = orange, 5+ = green)
- "Inspire me" button pre-fills with a random example scenario
- On submit: frontend sends `IntakePayload` JSON to `POST /api/generate`
- Go handler validates `core_idea` non-empty (400 if missing) and assembles the LLM prompt string
- All 7 optional fields are incorporated into the prompt when present

**Todo List:**
1. Build `IntakeForm` React component with 8 fields: core idea (required), target audience, brand personality, style/aesthetic, industry/category, colour mood, name length preference, words/sounds to avoid
2. Add inline hint text to each field (examples shown in plan v2 §1)
3. Build `PromptRichnessMeter` component: count non-empty optional fields → map to percentage → colour (red/orange/green) + tooltip "More context = better names on the first try"
4. Build "Inspire me" button: cycle through 5 pre-written brand scenarios, pre-fill all fields
5. Form validation: `core_idea` required; show inline error if empty on submit
6. On submit: serialize to `IntakePayload`, call `POST /api/generate` with `action: "init"`
7. Go handler `internal/handlers/generate.go`: validate payload, assemble deterministic prompt string from all filled fields, call Granite client

**Relevant Context:**
- `internal/models/types.go` — `IntakePayload` struct
- `frontend/lib/types.ts` — `IntakePayload` TypeScript interface
- Prompt assembly template: `"Generate brand names for a [core_idea] targeting [audience], personality: [personality], style: [style], industry: [industry], colors: [color_mood], length preference: [name_length]. Avoid: [avoid]."`

---

### ST-03: LLM Core — Name Generation, Origin Story, Brand Score, Voice Samples [ ] pending

**Intent:** Implement the IBM Granite LLM integration that generates name cards enriched with origin story, brand score, and voice samples — all in a single structured call. This makes every name feel consultant-grade on first delivery, not just a list.

**Expected Outcomes:**
- `POST /api/generate` streams a JSON array of enriched `NameCard` objects
- Each NameCard includes: name, tagline, tone reasoning, style tags, short description, long description, origin story (etymology + sound symbolism), brand score (memorability, spellability, global safety, squatter risk), brand voice samples (Instagram caption, email subject, 404 message)
- Streaming response renders cards progressively in the UI (first card visible before all are ready)
- Zod validation on client: malformed cards silently skipped, never crash the UI
- Go retry on malformed JSON: retry once with stricter format instruction

**Todo List:**
1. Write `internal/ai/granite.go`: HTTP client for watsonx.ai Granite endpoint, handles streaming, parses NDJSON chunks
2. Define system prompt template with session memory injection slots: `{rejected[]}`, `{liked[]}`, `{direction_notes[]}`, `{allGenerated[]}` — plus explicit exclusion instruction
3. Define NameCard JSON schema in the prompt (all fields listed above)
4. Add origin story instruction: "For each name, provide 1–2 sentences on linguistic roots, sound symbolism, and cultural associations"
5. Add brand score instruction: "Score each name on Memorability (1–10), Spellability (1–10), Global Safety (1–10), Domain Squatter Risk (Low/Medium/High) with one-line reasoning each"
6. Add brand voice samples instruction: "Generate 3 copy snippets in this brand's voice: an Instagram caption (≤15 words), a welcome email subject line (≤8 words), and a 404 page message (≤12 words, on-brand tone)"
7. Implement Go response parser: strict `encoding/json` decode; on error, retry once with stricter prompt; on second failure, return partial results
8. Build frontend `NameCard` component: collapsible sections for origin story, brand score table, voice samples, tone reasoning
9. Implement streaming render: use `ReadableStream` on the client, render each card as its JSON chunk completes

**Relevant Context:**
- `internal/models/types.go` — `NameCard` struct
- `internal/ai/granite.go` — Granite client
- `frontend/components/NameCard.tsx` — card component
- Session memory format: `BrandSession.Liked[]`, `.Rejected[]`, `.DirectionNotes[]`, `.AllGenerated[]`

---

### ST-04: Session Memory & Anti-Name Reasoning [ ] pending

**Intent:** Implement the Redis session store that makes NomVox a genuine creative partner. Every liked name, rejected name, rejection explanation, visual note, slider position, and uploaded vision context is persisted and injected into every subsequent AI call. Anti-Name Reasoning adds a clarifying question loop when a name is rejected, making the AI's learning visible to the user.

**Expected Outcomes:**
- Redis session created on first `/api/generate` call; UUID v4 session ID returned to client
- `BrandSession` object persisted: `Liked[]`, `Rejected[]`, `DirectionNotes[]`, `AllGenerated[]`, `VisualNotes[]`, `SlidersPosition`, `VisionContext`
- Every Granite call injects full session state as system prompt prefix
- Names in `AllGenerated[]` are never re-suggested
- On rejection with a note: AI responds with "I heard X — are you looking for Y, Z, or W?" before regenerating
- `PATCH /api/session/react` updates session correctly for all action types: like, reject, note, visual-note, slider

**Todo List:**
1. Write `internal/session/store.go`: Redis get/set/update using `go-redis`; TTL 2 hours; all operations use session ID as key
2. Define `BrandSession` struct with all fields; JSON serialise/deserialise to Redis string
3. Implement `POST /api/session` to create a new session; return `{ sessionId }`
4. Implement `PATCH /api/session/react` to handle: `like`, `reject` (with optional note), `visual-note`, `slider`, `direction-note`
5. Modify `internal/handlers/generate.go` to load session from Redis before building the Granite prompt
6. Build `AntiNameReasoning` response flow: when action is `reject` with a non-empty note, trigger a Granite call that generates a single clarifying question in response to the rejection note, before the next name batch
7. Build frontend `RejectionDialog` component: shows the AI's clarifying question, user answers in text field, answer stored as `DirectionNote`, then triggers regeneration
8. Add localStorage backup of `sessionId` on client; handle Redis TTL expiry with graceful "Start new session" recovery UI

**Relevant Context:**
- `internal/session/store.go` — session CRUD
- `internal/handlers/generate.go` — Granite prompt assembly
- `frontend/components/RejectionDialog.tsx` — anti-name reasoning UI
- Redis key format: `session:{sessionId}` → JSON string of `BrandSession`

---

### ST-05: Availability Engine & 80% Gate [ ] pending

**Intent:** Implement the Go availability engine that checks domain and social handle availability in parallel, applies the weighted 80% scoring gate, filters out failing names before they reach the user, and handles zero-pass retry cycles and closest-match fallback.

**Expected Outcomes:**
- `POST /api/availability` accepts a list of names and returns only those scoring ≥ 80%
- All 6 platform probes run concurrently using `sync.WaitGroup`; total probe time ≈ 1–2 seconds
- Domain checked via RDAP (`https://rdap.org/domain/{name}.com`)
- Social handles checked via HTTP HEAD: Instagram, X, TikTok, Threads, YouTube
- 429 / 5xx → "unknown" badge (does not count as available or unavailable in score)
- Results cached in Redis for 10 minutes per name
- Zero-pass: auto-retry Granite up to 3 cycles with "more unusual/coined words" instruction
- After 3 failed cycles: show closest-match fallback card with breakdown and platform toggle
- User can toggle which platforms matter; domain remains double-weighted and non-toggleable
- Competitor Name Radar: second Granite pass after gate, flags semantic overlap with known brands

**Todo List:**
1. Write `internal/availability/domain.go`: RDAP HTTP GET, parse JSON response for registrar presence (200 + registrar = taken, 404 = available)
2. Write `internal/availability/social.go`: HTTP HEAD with 5s timeout and randomised user-agent; map status codes to Available / Taken / Unknown
3. Write `internal/availability/checker.go`: `CheckAll()` using `sync.WaitGroup` for concurrent probes of all 6 platforms
4. Write `internal/availability/scorer.go`: weighted formula (domain=2, IG=1, X=1, TikTok=1, Threads=0.5, YouTube=0.5; total=6; threshold=4.8/6=80%); returns score float and pass bool
5. Write unit tests: `scorer_test.go` with table-driven cases; `checker_test.go` with mock HTTP server
6. Implement Redis cache in checker: check cache before probing; write result to cache with 10-minute TTL
7. Implement zero-pass retry loop in `internal/handlers/availability.go`: on 0 passing names, call Granite with modified prompt, re-probe, repeat up to 3 times
8. Implement closest-match fallback: if still 0 passes after 3 cycles, return top-scoring failed name with per-platform breakdown
9. Implement `POST /api/availability` handler: orchestrates check, gate, retry, fallback; returns `AvailabilityResult[]`
10. Write `internal/ai/granite.go` competitor radar function: second Granite call with prompt "Flag semantic overlap between these names and well-known brands in [industry]: [names]"
11. Build frontend availability matrix component: rows = names, columns = platforms, colour badges (green ✓ / red ✗ / amber ?); score pill per name; platform toggle checkboxes; closest-match fallback card

**Relevant Context:**
- `internal/availability/` — full availability package
- `internal/handlers/availability.go` — route handler
- Scoring formula defined in plan v2 §2 and above
- Platform probe URLs: `instagram.com/{h}`, `x.com/{h}`, `tiktok.com/@{h}`, `threads.net/@{h}`, `youtube.com/@{h}`

---

### ST-06: Select / Reject / Reproduce State Machine UI [ ] pending

**Intent:** Implement the three-option interaction model (Step B in the state machine) as a clear, deliberate UI component. This is the core creative-partner mechanic — the user is never forced into a linear flow.

**Expected Outcomes:**
- Each name card has three clearly labelled actions: Select (✓), Reject (✗), and a global Reproduce button
- Reject opens the `RejectionDialog` (Anti-Name Reasoning from ST-04)
- Reproduce triggers a new Granite batch with full session memory; does not reset anything
- Select moves the user to the Visual Identity section (Step 3)
- Style DNA Slider is visible above the card grid; moving it appends a tonal note to session and highlights that the next Reproduce will reflect it
- Session sidebar (collapsible) shows liked list, rejected list, and direction notes as a visible "creative conversation log"

**Todo List:**
1. Build `NameCardGrid` component: renders array of `NameCard` with Select / Reject / Reproduce controls
2. Add Select handler: stores chosen name in session (`.SelectedName`), navigates to Visual Identity section
3. Add Reject handler: opens `RejectionDialog`; on submit, calls `PATCH /api/session/react` with action `reject`
4. Add Reproduce button (global, not per-card): calls `POST /api/generate` with `action: "reproduce"` and current session state
5. Build `StyleDNASlider` component: two axes (Playful↔Premium, Abstract↔Descriptive); debounced onChange calls `PATCH /api/session/react` with `action: "slider"`; slider value injected into next Granite prompt as tonal instruction
6. Build `SessionSidebar` component: collapsible panel showing liked names, rejected names with notes, direction notes as a readable conversation log
7. Implement smooth section transitions (Framer Motion fade-slide) between intake form → name grid → visual identity

**Relevant Context:**
- `frontend/components/NameCardGrid.tsx`
- `frontend/components/StyleDNASlider.tsx`
- `frontend/components/SessionSidebar.tsx`
- `PATCH /api/session/react` — all session mutation calls go here

---

### ST-07: Brand Persona Card [ ] pending

**Intent:** After a name is selected, generate a brand persona card that describes the brand as if it were a person. This goes beyond naming into brand strategy and is the single feature most likely to make judges stop and say "I've never seen a tool do that."

**Expected Outcomes:**
- Brand persona card generated automatically after name selection (one Granite call)
- Card displays: age, occupation archetype, tone of voice, what they read/watch, what they would never say, their 3 core values
- Persona is consistent with the session's style direction, liked names, and intake form fields
- User can regenerate the persona card independently without affecting other assets
- Persona text is included in the export ZIP as `brand-persona.txt`

**Todo List:**
1. Write Granite prompt for persona generation: inject selected name, tone direction, intake payload, session context; request structured JSON persona object
2. Add persona route to `internal/handlers/generate.go` or create `internal/handlers/persona.go`
3. Build `BrandPersonaCard` frontend component: styled as a "character card" with the brand's "voice" in a large pull-quote format
4. Add "Regenerate Persona" button; calls persona endpoint with same session context
5. Include persona output in export ZIP assembly (ST-11)

**Relevant Context:**
- `internal/ai/granite.go` — reuse Granite client
- `frontend/components/BrandPersonaCard.tsx`
- Example output: `"Verdara is 28, reads Monocle, uses em-dashes, would never say 'synergy'. Core values: Quiet luxury, Environmental honesty, Long-game thinking."`

---

### ST-08: Vision Image Upload & Visual Identity [ ] pending

**Intent:** Implement the optional vision image upload (post-name-selection) that lets users show rather than describe their brand's visual direction. The extracted palette and mood feed into all image-gen calls for mood board, logos, and landing page. Also covers the full visual identity pipeline: mood board, all three logo types, and colour palette display.

**Expected Outcomes:**
- Optional "Upload your vision" prompt appears after name selection
- Uploaded image is processed by **Gemini 2.0 Flash** (Google AI Studio, natively multimodal — single `GOOGLE_AI_API_KEY`, free tier) to extract: dominant colours (5 HEX values), mood adjectives, style keywords, texture notes
- Extracted `VisionContext` stored in Redis session and injected into all subsequent Imagen 3 image-gen calls
- If no image uploaded: Imagen 3 calls use intake form colour mood + style fields directly
- Mood board: 4 images in a 2×2 grid, each individually regenerable
- Logo concepts: 3 types (Profile/Social avatar, App icon rounded-square, Business horizontal lockup), each individually regenerable
- All visual assets stored as signed Vercel Blob URLs
- Visual feedback notes (from user clicking "redo this one" with a comment) stored in session `.VisualNotes[]` and injected into next visual-gen call

**Todo List:**
1. Build `VisionUpload` component: drag-and-drop + click-to-browse; accepts PNG/JPG/WEBP ≤10MB; calls `POST /api/vision-context`
2. Write `internal/handlers/vision.go`: accept uploaded image, base64-encode it, call Gemini 2.0 Flash via Google AI SDK (`POST https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent`) with the image inline and a structured-output prompt requesting `{ colours: []string, mood: []string, style: []string, textures: []string }`; parse into `VisionContext` struct; store in Redis session — uses `GOOGLE_AI_API_KEY`, free tier, same key as Imagen 3
3. Fallback if Google AI API unavailable: show a colour picker (5 swatches) + mood adjective chips instead; all downstream Imagen 3 calls still work normally
4. Write `internal/ai/imagegen.go`: client for **Imagen 3** via Google AI Studio API (`POST https://generativelanguage.googleapis.com/v1beta/models/imagen-3.0-generate-002:predict`); accepts structured prompt + negative prompt + aspect ratio; uses `GOOGLE_AI_API_KEY`; free tier — use `imagen-3.0-generate-002` for logos (quality) and `imagen-3.0-fast-generate-001` for mood board panels (speed/cost)
5. Write mood board prompt builder: uses `VisionContext` (or intake fields) + brand name + style to generate 4 distinct but cohesive prompts
6. Write logo prompt builder for each of 3 types:
   - Profile: "minimal icon, {style}, {palette}, no text, works at 48px, clean background"
   - App icon: "app icon, {style}, {palette}, rounded square, flat design, subtle glow"
   - Business lockup: "{name} wordmark + icon, {palette}, professional, horizontal layout, SVG-style"
7. Write `POST /api/visuals` handler: orchestrates all image-gen calls; returns `VisualPack` with moodboard and logo URLs
8. Build `MoodBoard` frontend component: 2×2 grid, each tile has "Redo this" button with optional comment field
9. Build `LogoGrid` frontend component: 3 rows (Profile / App / Business), each with variant carousel and "Redo this type" button
10. Build `ColorPalette` component: displays 5 extracted or derived HEX swatches with copy-to-clipboard
11. Store all visual feedback notes: `PATCH /api/session/react` with `action: "visual-note"`; inject into next image-gen call

**Relevant Context:**
- `internal/handlers/vision.go` — vision processing
- `internal/ai/imagegen.go` — image generation client
- `internal/handlers/visuals.go` — visual pipeline orchestrator
- `frontend/components/MoodBoard.tsx`, `LogoGrid.tsx`, `ColorPalette.tsx`
- Logo type specs match the moodboard: Icon Mark = geometric NVK-style ligature; App Icon = same on rounded-square dark bg

---

### ST-09: Hero Landing-Page Mockup [ ] pending

**Intent:** Generate a single hero-section HTML/CSS mockup for the selected name, incorporating the brand palette, tagline, and logo slot. Scoped strictly to one section — not a full site — so it remains buildable in one day and demoable in 30 seconds.

**Expected Outcomes:**
- `POST /api/mockup` returns a self-contained hero-section HTML string
- HTML includes: brand name as H1, tagline as subheading, CTA button, logo placeholder slot (img tag with alt), inline CSS using the extracted/derived palette
- Rendered in a sandboxed `<iframe>` in the frontend (no external scripts)
- "Copy code" and "Download HTML" buttons available
- Mockup included in export ZIP as `hero-section.html`
- User can regenerate the mockup with a note (e.g., "make it darker") — note stored in session

**Todo List:**
1. Write `internal/handlers/mockup.go`: accepts selected name, tagline, palette, style; calls Granite with a prompt requesting a hero HTML/CSS section
2. Granite prompt: "Generate a self-contained HTML hero section for a brand called {name}. Tagline: {tagline}. Use this colour palette: {palette}. Style: {style}. Include: H1 (brand name), p (tagline), a.cta-button, img#logo (placeholder). All styles inline or in a style tag. No external dependencies."
3. Sanitise returned HTML on the Go side before sending to client (strip script tags)
4. Build `MockupPreview` frontend component: renders HTML in a sandboxed iframe with `sandbox="allow-same-origin"` attribute; overlaid "Copy Code" and "Download" buttons
5. Add "Regenerate with note" input field; note appended to session and sent with next mockup call

**Relevant Context:**
- `internal/handlers/mockup.go`
- `frontend/components/MockupPreview.tsx`
- Palette sourced from `VisionContext.colours` or derived from intake form colour mood

---

### ST-10: Export Pack [ ] pending

**Intent:** Bundle all session assets into a single downloadable ZIP that founders can immediately hand to a developer, printer, or registrar. The Go `archive/zip` package handles this without any external dependency.

**Expected Outcomes:**
- `POST /api/export` streams a ZIP binary to the client
- ZIP contains: `names-taglines.txt`, `brand-brief.pdf`, `brand-persona.txt`, `voice-samples.txt`, `mood-board.png`, `logo-profile.svg`, `logo-profile.png`, `logo-app.svg`, `logo-app.png`, `logo-business.svg`, `logo-business.png`, `hero-section.html`, `color-palette.json`
- Download button triggers correct browser download behaviour (`Content-Disposition: attachment`)
- All files are human-readable and immediately usable

**Todo List:**
1. Write `internal/handlers/export.go`: load all asset URLs from session; download each from Vercel Blob; write into `archive/zip` writer; stream ZIP binary in response
2. Generate `names-taglines.txt`: formatted list of all liked names with their taglines, origin stories, and brand scores
3. Generate `brand-persona.txt`: brand persona card content as plain text
4. Generate `voice-samples.txt`: all three brand voice copy snippets
5. Generate `color-palette.json`: array of `{ name, hex, usage }` objects
6. Generate `brand-brief.pdf`: use `chromedp` or `wkhtmltopdf` to render a styled HTML brief to PDF
7. Set response headers: `Content-Type: application/zip`, `Content-Disposition: attachment; filename="nomvox-brand-pack.zip"`
8. Build frontend "Export Pack" button: calls `/api/export`, triggers browser download via blob URL

**Relevant Context:**
- `internal/handlers/export.go`
- Go standard library: `archive/zip`, `net/http`
- All asset URLs stored in `BrandSession.VisualPack`

---

### ST-11: QA, Accessibility & Error States [ ] pending

**Intent:** Validate the full end-to-end flow, ensure all async sections have loading states, confirm mobile responsiveness, and confirm graceful degradation for every external dependency that could fail.

**Expected Outcomes:**
- Full happy-path end-to-end test passes: intake → generate → gate → select → visuals → export
- All loading states covered with skeleton components (no blank white flashes on dark bg)
- Mobile-responsive at 375px, 768px, 1280px
- WCAG AA basics: keyboard navigation, aria-labels, 4.5:1 colour contrast (note: dark bg requires checking `#E6E8F0` text on `#0B0F19` bg — should pass comfortably)
- Error states defined and handled: Granite API rate limit (watsonx.ai), Imagen 3 image-gen failure, Gemini 2.0 Flash vision failure (fallback to colour picker + mood chips), zero-pass after 3 retries, Redis TTL expiry, Google AI API unavailable (vision fallback degrades gracefully), Puppeteer timeout
- Go unit tests pass: `scorer_test.go`, `checker_test.go` (mock HTTP), `session_test.go`

**Todo List:**
1. Write table-driven unit tests for `scorer.go`: test all combinations including unknown platforms, edge cases at exactly 80%, domain unavailable (max possible 67%)
2. Write `checker_test.go` using `httptest.NewServer` to mock platform responses
3. Write `session_test.go` using `miniredis` (in-memory Redis mock for Go tests)
4. Build `SkeletonCard` component matching NameCard dimensions; build `SkeletonGrid` for mood board and logo grid
5. Test all error states: trigger each deliberately in dev, confirm graceful UI message is shown
6. Mobile pass: test on 375px (iPhone SE), 768px (iPad), 1280px (desktop); fix any layout breaks
7. Run contrast check: all text/bg combinations against WCAG AA 4.5:1 ratio

**Relevant Context:**
- All component files built in ST-02 through ST-09
- Go test files in same package as source (`_test.go` suffix)
- `github.com/alicebob/miniredis` for Redis mock in tests

---

### ST-12: Demo Preparation, Deployment & Submission [ ] pending

**Intent:** Deploy to production, record the 3-minute demo video, finalise all documentation, and complete the submission. This sub-task is the finish line.

**Expected Outcomes:**
- Go server deployed to Fly.io; Next.js deployed to Vercel; both publicly accessible
- 3-minute demo video recorded, edited, uploaded to YouTube (unlisted or public)
- GitHub README complete with all required sections
- Submission platform page complete with all required fields
- IBM SkillsBuild certificates uploaded for all team members
- Repo and video publicly accessible and verified

**Todo List:**
1. Write `Dockerfile` for Go server; deploy to Fly.io (`fly launch`, `fly deploy`)
2. Deploy Next.js to Vercel; set all environment variables in Vercel dashboard
3. Seed demo session: "eco-friendly coffee brand for young professionals, minimal aesthetic, green and cream tones" — pre-run to confirm all 6 steps work on production
4. Record demo video following the script below (§ Video Script)
5. Edit video to exactly ≤ 3:00; upload to YouTube
6. Write final README sections: Problem Statement, Solution Description, AI Approach & Architecture, Challenge Theme, How IBM Bob Was Used, Tech Stack, Local Setup, Live Demo URL, Demo Video URL, Team
7. Tag `v1.0.0` release on GitHub
8. Complete submission platform page; paste GitHub URL + video URL
9. Verify both links open correctly in an incognito browser window
10. Upload IBM SkillsBuild completion certificates

---

## Video Script (3 minutes)

| Time | Segment | On screen | Voiceover |
|---|---|---|---|
| 0:00–0:20 | Hook | NomVox landing page hero, galaxy vortex animation | "Every great brand starts as a feeling you can't quite name. NomVox changes that." |
| 0:20–0:50 | Intake Form | Fill all 8 fields live; prompt richness meter turns green | "Eight fields. That's all the engine needs. Watch the richness meter fill — the more you give it, the better the first pass." |
| 0:50–1:20 | Results + Gate | 8 name cards stream in; 2 auto-filtered (show score badges 67%, 50%); zoom on 100% and 83% badges | "Three names were filtered automatically — they didn't clear the 80% availability threshold across domains and handles. These two did." |
| 1:20–1:40 | Reject + Anti-Reasoning | Reject 'Foliaro' with note 'too corporate'; AI responds with clarifying question; user answers | "When I reject a name, NomVox doesn't just spin again. It asks what I actually meant." |
| 1:40–2:10 | Select + Visuals | Select 'Verdara'; upload vision image; mood board appears; 3 logo types render; hero mockup in iframe | "I upload a reference image. The engine reads its palette and mood — and builds the whole visual identity from it." |
| 2:10–2:30 | Brand Persona + Voice Samples | Show brand persona card; show voice samples panel | "'Verdara is 28, reads Monocle, would never say synergy.' This is the brand's voice — and here's what it sounds like writing an Instagram caption and a 404 page." |
| 2:30–2:50 | Export | Click Export Pack; ZIP downloads; briefly show contents | "One ZIP. Everything a founder needs to hand off to a developer, a printer, or a registrar." |
| 2:50–3:00 | Close + Architecture | Architecture diagram (5 seconds) | "Built with IBM Granite, watsonx.ai, Go, and IBM Bob. NomVox — signal sent, brand received." |

### AI Presenter / Voice Tools (for pitch video production)
- **ElevenLabs** — generate a consistent AI voice narrator for the voiceover (custom voice, cosmic/confident tone)
- **HeyGen** — AI video presenter avatar for intro/outro segments if a human presenter is not on camera
- **Runway ML** — for the galaxy/vortex background animation in the hero section of the demo
- **Descript** — for editing, auto-transcript, and overdub corrections without re-recording
- **CapCut / DaVinci Resolve** — for final cut, lower-thirds, and the architecture diagram transition

---

## Go .gitignore

```gitignore
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
nomvox-server

# Test binary
*.test

# Go build output
/cmd/server/server
/bin/

# Go workspace
go.work
go.work.sum

# Module download cache
/vendor/

# Environment secrets
.env
.env.local
.env.*.local

# IDE / OS
.idea/
.vscode/
*.DS_Store
Thumbs.db

# Test and coverage
*.out
coverage.html

# Temp / generated
tmp/
dist/
.next/
node_modules/

# Vercel
.vercel

# Fly.io
fly.toml.bak
```

---

## 15-Day Sprint Map

| Day | Sub-task(s) | Gate |
|---|---|---|
| 1 | ST-01: Env, scaffold, repo, CI | Go `ping` endpoint live; Next.js builds |
| 2 | ST-02: Keyword intake form | Form submits valid `IntakePayload` to Go |
| 3–4 | ST-03: LLM core (name gen, origin story, brand score, voice samples) | Streamed NameCards render in UI |
| 5 | ST-04: Session memory + anti-name reasoning | Reject flow triggers clarifying Q; session persists in Redis |
| 6–7 | ST-05: Availability engine + 80% gate + competitor radar | Matrix renders; sub-80% names filtered; zero-pass retry works |
| 8 | ST-06: Select/Reject/Reproduce state machine + style DNA slider | All three options work; slider appends to session |
| 8 | ST-07: Brand persona card | Persona card renders after name selection |
| 9–10 | ST-08: Vision upload + full visual identity (mood board + logos) | All 3 logo types + mood board render; per-asset regen works |
| 11 | ST-09: Hero landing-page mockup | HTML renders in sandboxed iframe; download works |
| 12 | ST-10: Export ZIP | ZIP downloads with all assets |
| 13 | ST-11: QA, unit tests, error states, mobile pass | All tests pass; no blank states |
| 14 | ST-12 (part 1): Deploy + demo recording | Production URL live; video recorded |
| 15 | ST-12 (part 2): README + submission | Submission page complete; public links verified |
