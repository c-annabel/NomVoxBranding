# NomVox — Born from the Void

> **AI-powered brand identity platform. From void to voice — names, handles, logos, and landing pages in under 60 seconds.**

Built for the **IBM AI Builders Challenge — July 2026 · Creative Industries Theme**  
Primary development tool: **IBM Bob**

[![Live App](https://img.shields.io/badge/Live%20App-nomvox.vercel.app-8B5CF6?style=flat-square)](https://nomvox.vercel.app)
[![API](https://img.shields.io/badge/API-nomvox--api.fly.dev-22d3ee?style=flat-square)](https://nomvox-api.fly.dev/api/ping)
[![Built with IBM Bob](https://img.shields.io/badge/Built%20with-IBM%20Bob-0f62fe?style=flat-square)](http://ibm.biz/university-bob)

---

## Problem Statement

Every new business faces the same blank-page problem: *what do I call this, and does anyone already own that name?* Today that means juggling brainstorm sessions, domain registrar tabs, separate searches on Instagram/X/TikTok, a freelancer quote for logo concepts, and days of back-and-forth before a founder has anything shareable.

Current tools (Namelix, Namechk, Looka) are **one-shot generators with no memory** — a user cannot say "I liked the energy of option 3 but want it shorter" and have the tool genuinely respond to that signal. Founders waste hours cycling through unusable suggestions with no creative guidance.

## Solution Description

NomVox is a **conversational AI brand-identity partner**. A founder describes their idea using a structured 8-field intake form; NomVox generates coined names and taglines enriched with origin stories, brand scores, and brand-voice samples — then automatically filters out any name that doesn't clear a **60% cross-platform availability threshold** (domain + social handles).

The user reacts, refines, and selects. The AI explains every rejection, asks a clarifying question, and regenerates with full session memory. After selection, NomVox generates a complete visual identity pack: mood board (4 panels), three logo concepts (profile / app icon / business card), brand persona, and an AI-generated hero landing-page mockup — all exportable as a single ZIP.

### Key Features

| Feature | Description |
|---|---|
| **AI Name Generation** | 4 coined names per pass — invented vocabulary, never real words |
| **Origin Stories** | Etymology of each invented name, naming strategy explained |
| **Brand Score** | Memorability, Spellability, Global Safety, Squatter Risk |
| **Brand Voice Samples** | Instagram caption, Email subject, 404 page message |
| **Availability Engine** | RDAP domain check + 5 social platform probes in parallel |
| **60% Gate** | Names must clear domain + handle availability threshold |
| **Zero-pass Fallback** | Top-2 partials shown with amber banner if no name fully passes |
| **Session Memory** | Liked/rejected/notes injected into every AI call |
| **Style DNA Sliders** | Playful↔Premium + Abstract↔Descriptive influence next batch |
| **Anti-Name Reasoning** | AI asks a clarifying question after each rejection |
| **Competitor Radar** | Semantic brand clash detection via second watsonx call |
| **Logo Concepts** | 3 SVG/CSS styles: Geometric Bauhaus · Gradient Glassmorphism · Horizontal Wordmark |
| **Mood Board** | 4-panel visual board tailored to brand colours + personality (2×2 grid) |
| **Brand Persona** | "Brand as a person" card — age, occupation, voice, never-says, core values |
| **Landing Page Mockup** | AI-generated hero section HTML/CSS with brand colours, logo in nav, CTA |
| **Export Pack** | ZIP: brand-brief.json, brand-brief.html, landing-page.html, logos/, mood-board/, README.txt |

---

## AI Approach & Architecture

```
IBM Granite / Llama 3.3 70B (watsonx.ai) ──► All text & reasoning
  • Name + tagline generation
  • Origin stories, brand scores, voice samples
  • Brand persona card ("brand as a person")
  • Anti-name reasoning — explains rejections, asks clarifying Q
  • Competitor name radar (semantic clash detection)
  • Hero landing-page HTML/CSS mockup
  • SVG logo concepts (geometric, glassmorphic, wordmark)
  • SVG mood board tiles (4 abstract brand panels)

CSS Art Fallback (primary visual path in production) ──► All visual output
  • All visual generation uses dynamic CSS/SVG driven by user's brand palette
  • Reason: Imagen 3 (Google) free-tier credits depleted; IBM has no
    equivalent native image generation model in watsonx.ai at this time
  • CSS fallbacks are not placeholder stubs — they are fully designed,
    palette-responsive, personality-driven visual components
  • See section "Image Generation Decision" below for full context

Go API server (chi router, port 8080) ──► Availability engine + session memory
  • RDAP domain check (Verisign RDAP API — no scraping, no API key needed)
  • Parallel HTTP HEAD probes: Instagram, X, TikTok, Threads, YouTube
  • Weighted 60% scoring gate (domain=2pts, social=1pt each)
  • Redis session store (Upstash) — liked/rejected/notes/slider state
  • Concurrent goroutines for visual generation tasks
  • ZIP export streamed to browser (archive/zip, no temp files)

Next.js 16 + React 19 (Vercel) ──► Frontend
  • 5-screen state machine: Intake → Loading → Results → Visuals → Export
  • Numbered intake form (Q1–Q8) with prompt richness meter
  • Name cards: Like / Pass (with reason) / Choose
  • Style DNA sliders + Regenerate loop with session memory
  • Visual Identity: Logos → Mood Board (2×2 grid + Persona) → Landing Page (iframe 75% scale)
  • Graceful fallback at every failure point (CSS art → SVG → PNG)
```

### Image Generation Decision — Imagen → SVG/CSS

The original architecture specified **Imagen 3** (Google AI Studio) for logo and mood board image generation. During development this worked locally, but hit a permanent wall before the production demo:

**Why Imagen was replaced:**
1. Google AI Studio free-tier prepaid credits for Gemini/Imagen image generation were **exhausted permanently** — further calls require billing enabled on a Google Cloud project
2. **IBM watsonx.ai has no native image generation model** comparable to Imagen, DALL-E, or Stable Diffusion accessible through the same API at this time
3. The fallback path — asking the Llama 3.3 70B text model to output SVG markup — produces geometric/abstract results that are structurally valid but not design-quality raster images
4. Adding a paid third-party image API (Stability AI, Replicate, etc.) would violate the spirit of the challenge's IBM-first requirement

**What was built instead:**
- A fully dynamic **CSS/SVG visual system** driven entirely by `extractPalette()` — which reads the user's `color_mood` field and derives `bg`, `accent`, `accent2`, `text` tokens
- Logo placeholders: three distinct CSS compositions (geometric Bauhaus hexagonal mark / glassmorphic gradient app icon / horizontal wordmark lockup) — all palette-adaptive
- Mood board: four distinct CSS panels (Colour World / Brand Identity / Pattern DNA / Typography + CTA) — all palette-adaptive, all clearly visible
- Landing page: a full CSS hero section with nav, headline, tagline, CTA, and right-column geometric accent

This is not a degraded experience — it is a **designed system** that reliably produces on-brand visuals from user input with zero API cost and zero rate limit risk.

**Future path:** When IBM releases a native image generation model on watsonx.ai, or when a paid Stability AI / Replicate integration is scoped and budgeted, the CSS fallbacks are designed to be replaced by real AI images at the same component insertion points.

---

### Session Memory Loop (Creative Partner Pattern)

```
User types brand idea
    ↓
NomVox generates 4 coined names (Session created in Redis)
    ↓
Availability probed in parallel (domain + 5 socials)
    ↓
User: ♡ Like / ✕ Pass (with reason) / ✓ Choose
    ↓  ↓
    │  Pass + reason → AI asks clarifying question
    │                 → User answers → injected into session
    ↓
↻ Regenerate → Liked names + rejection notes + slider DNA
               injected into every LLM system prompt
    ↓
User chooses a name → Visual Identity Pack generated
    ↓
Logo selected → Targeted mood board regen with logo context
    ↓
Export ZIP pack
```

---

## Selected Challenge Theme

**Creative Industries** — spanning four solution areas simultaneously:
- ✅ **AI Creative Partner** — conversational name refinement loop with session memory
- ✅ **Creative Ideation & Brainstorming Platform** — multi-strategy coined name generation
- ✅ **AI-powered Design & Visual Concept Tools** — logos, mood boards, landing page
- ✅ **Personalized Creative Assistant** — Style DNA sliders, anti-name reasoning, per-session direction

---

## How IBM Bob Was Used

IBM Bob was the **primary development tool** throughout the full SDLC:

1. **Architecture design** — Technology selection, Go vs Node backend, Redis vs in-memory session, deployment topology (Fly.io + Vercel)
2. **SDLC planning** — Full 15-day sprint plan (ST-01 → ST-12), gap analysis, risk register, judging-criteria alignment
3. **Prompt engineering** — 5-strategy radical naming system prompt, anti-name reasoning loop, brand persona, competitor radar, mood board tile generation, SVG logo prompts, landing page HTML generator
4. **Code scaffolding** — All Go handlers, session store, availability engine (RDAP + social probers), ZIP export builder
5. **Bug diagnosis** — JSON truncation fix, IAM token refresh cycle, base64 SVG encoding, hydration mismatch, triple-fire guard, iframe 100vh scroll strip
6. **Deployment guidance** — Fly.io Dockerfile, `fly.toml` config, Vercel environment variables, CORS configuration
7. **Documentation** — README, SDLC plan HTML, DesignDoc, this file

---

## Tech Stack

| Layer | Technology | Notes |
|---|---|---|
| Frontend | Next.js 16.2, React 19, Tailwind CSS v4 | App Router, Turbopack dev |
| UI Font | Space Grotesk | via @fontsource/space-grotesk |
| Backend | Go 1.22+, chi v5 router | net/http standard library |
| AI — Text | Llama 3.3 70B via IBM watsonx.ai | `ca-tor.ml.cloud.ibm.com` · IAM token cache 50 min |
| AI — Visual | CSS/SVG (palette-driven) | Primary visual path; Imagen 3 removed (credits exhausted) |
| Session store | Redis via Upstash | 2-hour session TTL |
| Domain check | RDAP (Verisign) | `rdap.org/domain/{name}.com` — no scraping |
| Social probes | HTTP HEAD to platform profile URLs | Instagram, X, TikTok, Threads, YouTube |
| Deployment — API | Fly.io (`nomvox-api`, region `iad`) | shared-cpu-1x, auto-stop |
| Deployment — Frontend | Vercel | Auto-deploy from GitHub `main` |
| Primary dev tool | IBM Bob | Required for challenge compliance |

---

## Repository Structure

```
NomVoxBranding/
│
├── cmd/
│   ├── server/main.go          — Go API entry point (chi router, all routes)
│   ├── testllm/main.go         — Standalone watsonx.ai connection test
│   └── diagnose-imagen/main.go — Imagen 3 diagnostic tool (archived)
│
├── internal/
│   ├── ai/
│   │   ├── granite.go          — watsonx.ai HTTP client, IAM token cache
│   │   ├── imagen.go           — Google Imagen 3 client (legacy, rate-limited)
│   │   ├── prompts.go          — All LLM system prompts (names, persona, radar)
│   │   ├── visual_prompts.go   — Visual generation prompts + buildColourContext()
│   │   └── ai_test.go          — Unit tests for AI helpers
│   │
│   ├── availability/
│   │   ├── checker.go          — Parallel RDAP + social HEAD probes
│   │   └── availability_test.go
│   │
│   ├── handlers/
│   │   ├── generate.go         — POST /api/generate (name gen + availability gate)
│   │   ├── visuals.go          — POST /api/visuals (logos + mood board + mockup)
│   │   ├── export.go           — POST /api/export (ZIP stream)
│   │   ├── session.go          — GET/PATCH /api/session/*
│   │   ├── availability.go     — POST /api/availability
│   │   ├── debug_llm.go        — Debug endpoint (dev only)
│   │   ├── debug_visuals.go    — Debug endpoint (dev only)
│   │   └── diagnose.go         — Diagnose endpoint
│   │
│   ├── models/
│   │   └── types.go            — All shared Go structs
│   │
│   └── session/
│       ├── store.go            — Redis session CRUD (Upstash)
│       └── session_test.go
│
├── frontend/
│   ├── app/
│   │   ├── layout.tsx          — Root layout, metadata, favicon config
│   │   ├── page.tsx            — Server Component wrapper
│   │   ├── HomeClient.tsx      — Main state machine (5 screens)
│   │   ├── globals.css         — Brand tokens, animations, Space Grotesk
│   │   ├── icon.png            — App Router favicon (tab icon)
│   │   └── favicon.ico         — Legacy favicon fallback
│   │
│   ├── components/
│   │   ├── IntakeForm.tsx           — 8-field brand intake + richness meter
│   │   ├── NameCardComponent.tsx    — Name card: score, voice, like/pass/choose
│   │   ├── NameCardSkeleton.tsx     — Loading skeleton
│   │   ├── AvailabilityBadges.tsx   — Platform availability badge row
│   │   ├── BrandScoreCard.tsx       — Memorability/Spellability/Safety scores
│   │   ├── RejectionDialog.tsx      — Anti-name reasoning dialog
│   │   ├── StyleDNASlider.tsx       — Playful↔Premium / Abstract↔Descriptive
│   │   ├── PromptRichnessMeter.tsx  — Intake form richness indicator
│   │   ├── ErrorBoundary.tsx        — React error boundary
│   │   └── VisualIdentityPanel.tsx  — Steps 1–3: logos, mood board, landing page
│   │
│   ├── lib/
│   │   ├── types.ts            — TypeScript interfaces (mirrors Go structs)
│   │   ├── schemas.ts          — Zod validation schemas
│   │   └── api.ts              — API fetch wrappers
│   │
│   ├── public/
│   │   ├── favicon.ico         — Browser favicon
│   │   ├── nomvox-icon.png     — 512×512 brand icon
│   │   ├── nomvox-logo.png     — Full wordmark
│   │   └── nomvox-bg.png       — Landing background
│   │
│   └── vercel.json             — Cache headers for favicon assets
│
├── 01-Plans/
│   ├── v1–v4 HTML SDLC plans   — Full project planning history
│   └── v5-nomvox-full-sdlc-plan.html  — Latest comprehensive plan
│
├── Dockerfile                  — Multi-stage Go build for Fly.io
├── fly.toml                    — Fly.io app config (nomvox-api, region iad)
├── go.mod                      — Go module (github.com/c-annabel/NomVoxBranding)
├── go.sum
├── .env.example                — All required environment variables documented
├── DesignDoc.md                — Full technical design reference
├── nomvox-plan.md              — SDLC plan with sub-task status
├── README.md                   — This file
└── developer-bob-feedback.md   — Developer feedback on IBM Bob
```

---

## Local Setup

### Prerequisites

- Go 1.22+
- Node.js 20+
- Redis (or [Upstash](https://upstash.com) free account)
- IBM watsonx.ai account + Project ID (with WML service associated)
- IBM Cloud IAM API key

### Environment Variables

```bash
cp .env.example .env
```

Edit `.env`:

```env
# Redis (Upstash recommended)
REDIS_URL=rediss://:password@host.upstash.io:6379

# IBM watsonx.ai
WATSONX_API_KEY=your-ibm-cloud-iam-api-key
WATSONX_PROJECT_ID=your-watsonx-project-id
WATSONX_URL=https://ca-tor.ml.cloud.ibm.com

# CORS (development default: localhost:3000)
ALLOWED_ORIGIN=http://localhost:3000

# Google AI (optional — Imagen 3 legacy; currently rate-limited/depleted)
# GOOGLE_AI_API_KEY=your-google-ai-studio-key
```

> **Note on IBM watsonx.ai setup:** Your IBM Cloud IAM API key must have Watson Machine Learning access. The watsonx project must have the WML service **associated** in the project settings (Projects → Settings → Associated services). Without this, you will receive a 403 with no descriptive error.

### Run

```bash
# Terminal 1 — Go API (port 8080)
go mod download
go run ./cmd/server

# Terminal 2 — Next.js frontend (port 3000)
cd frontend
npm install
npm run dev
```

Open `http://localhost:3000`.

### Test watsonx.ai connection

```bash
go run ./cmd/testllm
```

---

## Production Deployment

### Step 1 — Fly.io (Go API)

```bash
# Install flyctl: https://fly.io/docs/hands-on/install-flyctl/
fly auth login
fly launch --name nomvox-api --region iad   # creates fly.toml
fly secrets set \
  WATSONX_API_KEY="..." \
  WATSONX_PROJECT_ID="..." \
  WATSONX_URL="https://ca-tor.ml.cloud.ibm.com" \
  REDIS_URL="rediss://..." \
  ALLOWED_ORIGIN="https://nomvox.vercel.app"
fly deploy
```

Verify: `https://nomvox-api.fly.dev/api/ping` → `{"status":"ok"}`

> **Cold start note:** `auto_stop_machines = "stop"` is enabled — the first request after inactivity may take 2–4 seconds to wake the machine. Subsequent requests are fast.

### Step 2 — Vercel (Next.js Frontend)

```bash
cd frontend
npx vercel --prod
# Set environment variable in Vercel dashboard or CLI:
vercel env add NEXT_PUBLIC_API_URL
# → enter: https://nomvox-api.fly.dev
vercel --prod   # redeploy with env var
```

Verify: `https://nomvox.vercel.app` loads the intake form.

---

## Live Demo

- **App:** https://nomvox.vercel.app
- **API ping:** https://nomvox-api.fly.dev/api/ping

## Demo Video

[3-minute walkthrough](#) ← *to be recorded before July 31 2026*

---

## What's Implemented (as of latest commit)

| Feature | Status |
|---|---|
| Intake form (8 fields + richness meter) | ✅ Complete |
| Name generation (4 coined names, watsonx) | ✅ Complete |
| Origin stories + brand scores + voice samples | ✅ Complete |
| Availability engine (RDAP + 5 social platforms) | ✅ Complete |
| 60% gate + zero-pass fallback (top-2 partial) | ✅ Complete |
| Session memory (Redis, liked/rejected/notes) | ✅ Complete |
| Anti-name reasoning (AI clarifying question) | ✅ Complete |
| Style DNA sliders (2-axis tonal direction) | ✅ Complete |
| Competitor radar (semantic clash detection) | ✅ Complete |
| Logo concepts — 3 CSS/SVG styles | ✅ Complete |
| Mood board — 4-panel CSS visual board | ✅ Complete |
| Brand persona card | ✅ Complete |
| Landing page mockup (AI HTML + CSS fallback) | ✅ Complete |
| Export ZIP | ✅ Complete |
| Deployment (Fly.io + Vercel) | ✅ Live |
| AI image generation (Imagen 3) | ⚠️ Credits depleted — CSS/SVG primary path |
| Handle screenshot (Puppeteer) | 🚫 Cut from scope — Fly.io container constraints |

---

## Future Roadmap

### Stage 1 — Immediate (next sprint)
- Confirm and document favicon behavior on deployed Vercel URL
- Record and publish 3-minute demo video
- IBM SkillsBuild certificate upload + submission page completion
- Replace `nomvox-plan-v3.md` reference in README with `nomvox-plan.md`

### Stage 2 — Near-term enhancements
- Integrate a paid image generation API (Stability AI / Replicate / Together.ai) when IBM watsonx image models become available — drop-in replacement at the same CSS fallback component insertion points
- Puppeteer-based handle screenshot service (hosted separately from Fly.io due to Chrome binary constraints)
- Color palette extractor from user-uploaded reference image (Gemini vision, when credits restored)
- Short + long brand description in-app editing with per-field regeneration

### Stage 3 — Product evolution
- Multi-name session persistence (save entire brand exploration to account)
- Collaborative sessions (share a session link with a co-founder)
- Brand voice style guide export (full PDF brief, not just ZIP)
- Domain registration handoff (deep-link to Namecheap / GoDaddy with prefilled name)
- Social handle reservation queue (notify when a handle becomes available)

### Stage 4 — Revenue potential
- **Freemium SaaS:** 3 free generations per month, paid tier for unlimited + premium AI images + PDF export
- **Agency white-label:** NomVox API + private-label frontend for brand consultancies
- **Shopify / Squarespace plugin:** Brand name + visual identity embedded in website builder onboarding
- **IBM Partner Program:** Position NomVox as a reference implementation of watsonx.ai for creative industries — potential IBM co-marketing opportunity
- **Startup accelerator tooling:** Partner with YC, Techstars, or similar to offer NomVox as part of pre-cohort onboarding

---

## A Note on This Experience

Building NomVox was a genuinely rewarding experience. In roughly 15 days, one developer went from zero to a deployed, full-stack AI web application — spanning Go backend, Next.js frontend, Redis session layer, real availability checking across 6 platforms, a conversational AI loop, and a complete visual identity system.

The process taught real lessons that no tutorial covers: the actual mechanics of IBM watsonx IAM authentication, the importance of training data quality in LLM output, why concurrent goroutines need careful timeout management, how Next.js App Router handles hydration differently from the Pages Router, and why CSS fallbacks are not a consolation prize — they are a design decision.

IBM Bob made the journey possible. It was a genuine creative and technical partner — not perfect, not always the most efficient, but consistently present, consistently knowledgeable, and genuinely educational. The gaps and frustrations documented in `developer-bob-feedback.md` are shared in the same spirit — honest feedback makes better tools.

The name itself — NomVox — came from exactly the kind of AI-assisted creative session this tool is designed to enable. *Nomen* (name) + *vox* (voice). The cosmos names things. Now so do you.

---

## Team

| Member | Role | IBM SkillsBuild Certificate |
|---|---|---|
| c-annabel | Full-stack development, AI integration, UI/UX design | *[upload link — required before July 31]* |

---

*© 2026 c-annabel — Developed with IBM Bob — All rights reserved.*  
*AI-generated brand assets are creative inspiration. Verify availability before registration.*
