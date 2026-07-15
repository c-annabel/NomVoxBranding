# NomVox — Born from the Void

> **AI-powered brand identity platform. From void to voice — names, handles, logos, and landing pages in under 60 seconds.**

Built for the **IBM AI Builders Challenge — July 2026 · Creative Industries Theme**  
Primary development tool: **IBM Bob**

---

## Problem Statement

Every new business faces the same blank-page problem: *what do I call this, and does anyone already own that name?* Today that means juggling name-brainstorm sessions, domain registrar tabs, separate searches on Instagram/X/TikTok, a freelancer quote for logo concepts, and days of back-and-forth before you have anything shareable.

Current tools (Namelix, Namechk, Looka) are **one-shot generators with no memory** — a user cannot say "I liked the energy of option 3 but want it shorter" and have the tool genuinely respond to that signal. Founders waste hours cycling through unusable suggestions with no creative guidance.

## Solution Description

NomVox is a **conversational AI brand-identity partner**. A founder describes their idea using a structured 8-field intake form; NomVox generates coined names and taglines enriched with origin stories, brand scores, and brand-voice samples — then automatically filters out any name that doesn't clear a **60% cross-platform availability threshold** (domain + social handles).

The user reacts, refines, and selects. The AI explains every rejection, asks a clarifying question, and regenerates with full session memory. After selection, NomVox generates a complete visual identity pack: mood board (4 panels), three logo concepts (profile/app/business card), brand persona, and an AI-generated hero landing-page mockup — all exportable as a ZIP.

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
| **Competitor Radar** | Semantic brand clash detection via second Granite call |
| **Logo Concepts** | 3 distinct SVG styles: Geometric Bauhaus / Gradient Glassmorphism / Horizontal Wordmark |
| **Mood Board** | 4-panel SVG abstract tile board tailored to brand colours + personality (2×2 grid) |
| **Brand Persona** | "Brand as a person" card — age, occupation, voice, never-says, core values |
| **Landing Page Mockup** | AI-generated hero section HTML/CSS with brand colours, logo in nav, CTA button |
| **Export Pack** | ZIP: brand-brief.json, brand-brief.html, landing-page.html, logos/*.svg, mood-board/*.svg, README.txt |

## AI Approach & Architecture

```
IBM Granite (watsonx.ai) ──► All text & reasoning
  • Name + tagline generation  (meta-llama/llama-3-3-70b-instruct)
  • Origin stories, brand scores, voice samples
  • Brand persona card ("brand as a person")
  • Anti-name reasoning — explains rejections, asks clarifying Q
  • Competitor name radar (semantic clash detection)
  • Hero landing-page HTML/CSS mockup

Imagen 3 (Google AI Studio) ──► Image generation [fallback: watsonx SVG]
  • Mood board — 4 abstract brand tiles (SVG via watsonx when Gemini credits absent)
  • Logo concepts — Profile / App Icon / Business Card (3 parallel calls)
  • Fallback chain: Gemini PNG → watsonx SVG → CSS art placeholders

Gemini 2.0 Flash (Google AI Studio, free tier) ──► Vision [optional]
  • User uploads reference image → extracts palette, mood, style
  • Injected into mood board + logo prompts

Go API server (chi router) ──► Availability engine + session memory
  • RDAP domain check (verisign RDAP API, no scraping)
  • Parallel HTTP HEAD probes for: Instagram, X, TikTok, Threads, YouTube
  • Weighted 60% scoring gate (domain=2pts, social=1pt each, minor=0.5pt)
  • Redis session store (Upstash) — liked/rejected/notes/slider state
  • Concurrent goroutines for all 5 visual generation tasks
  • ZIP export streamed directly to browser (archive/zip)

Next.js 16 + React 19 (Vercel) ──► Frontend
  • 5-screen state machine: Intake → Loading → Results → Visuals → Export
  • Numbered intake form (Q1–Q8) with prompt richness meter
  • Name cards: Like / Pass (with reason) / Choose
  • Style DNA sliders + Regenerate loop with session memory
  • Visual Identity: Step 1 Logos → Step 2 Mood Board (2×2 SVG grid + Persona) → Step 3 Landing Page (iframe 75% scale)
  • Graceful fallback at every failure point (CSS art → SVG → PNG)
```

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

## Selected Challenge Theme

**Creative Industries** — spanning four solution areas:
- ✅ AI Creative Partner (conversational name refinement loop)
- ✅ Creative Ideation & Brainstorming Platform (multi-strategy name generation)
- ✅ AI-powered Design & Visual Concept Tools (logos, mood boards, landing page)
- ✅ Personalized Creative Assistant (session memory, Style DNA sliders)

## How IBM Bob Was Used

IBM Bob was the **primary development tool** throughout the full SDLC:

1. **Architecture design** — Technology selection, tradeoffs between Go vs Node backend, Redis vs in-memory session, Imagen 3 vs DALL-E
2. **SDLC planning** — Full sprint-by-sprint plan (ST-01 → ST-15), gap analysis, risk register, judging-criteria alignment
3. **Prompt engineering** — 5-strategy radical naming system prompt, anti-name reasoning, persona, competitor radar, mood board, logo concepts, landing page
4. **Code scaffolding** — All Go handlers, session store, availability engine, export ZIP builder
5. **Bug diagnosis** — JSON truncation fix, Base64 encoding fix, hydration mismatch fix, triple-fire guard
6. **QA planning** — Unit test design, mobile responsive review, error-state coverage

## Tech Stack

| Layer | Technology | Notes |
|---|---|---|
| Frontend | Next.js 16.2, React 19, Tailwind CSS v4 | App Router, Turbopack |
| UI Font | Space Grotesk | via @fontsource |
| Backend | Go 1.22+, chi v5 router | net/http standard library |
| AI — Text | IBM Granite / Llama 3.3 70B via IBM watsonx.ai | IAM token cache 50min |
| AI — Images | Imagen 3 via Google AI Studio | Free tier |
| AI — Vision | Gemini 2.0 Flash via Google AI Studio | Optional feature |
| Session store | Redis via Upstash | 2-hour session TTL |
| Domain check | RDAP (Verisign) | No scraping, no API key |
| Social probes | HTTP HEAD to platform profile URLs | 6 platforms, 8s timeout each |
| Deployment | Vercel (frontend) + Fly.io (Go API) | |
| Primary dev tool | IBM Bob | Required for challenge |

## Local Setup

### Prerequisites

- Go 1.22+
- Node.js 20+
- Redis (or [Upstash](https://upstash.com) free account)
- IBM watsonx.ai account + Project ID
- Google AI Studio API key (for Imagen 3 + optional Gemini vision)

### Environment Variables

```bash
cp .env.example .env
```

Edit `.env`:

```
# Redis
REDIS_URL=rediss://:password@host.upstash.io:6379

# IBM watsonx.ai
WATSONX_API_KEY=your-ibm-cloud-api-key
WATSONX_PROJECT_ID=your-project-id
WATSONX_URL=https://ca-tor.ml.cloud.ibm.com

# Google AI (Imagen 3 + Gemini)
GOOGLE_AI_API_KEY=your-google-ai-studio-key

# CORS (development default: localhost:3000)
ALLOWED_ORIGIN=http://localhost:3000
```

### Run

```bash
# Terminal 1 — Go API server (port 8080)
go mod download
go run ./cmd/server

# Terminal 2 — Next.js frontend (port 3000)
cd frontend
npm install
npm run dev
```

Open `http://localhost:3000`.

### Test the AI connection

```bash
go run ./cmd/testllm
```

## Production Deployment

### Fly.io (Go API)

```bash
fly launch                        # creates fly.toml
fly secrets set WATSONX_API_KEY=... WATSONX_PROJECT_ID=... WATSONX_URL=... \
  GOOGLE_AI_API_KEY=... REDIS_URL=... \
  ALLOWED_ORIGINS=https://nomvox.vercel.app
fly deploy
```

### Vercel (Next.js)

```bash
cd frontend
vercel                            # follow prompts
vercel env add NEXT_PUBLIC_API_URL   # set to https://nomvox-api.fly.dev
```

## Live Demo

- **App:** https://nomvox.vercel.app  
- **API ping:** https://nomvox-api.fly.dev/api/ping  

## Demo Video

[3-minute walkthrough](#) ← *to be added by July 31 2026*

## Repository Structure

```
NomVoxBranding/
├── cmd/
│   ├── server/         Go API entry point
│   └── testllm/        Standalone LLM connection test
├── internal/
│   ├── ai/             Granite/Imagen/Gemini clients + prompt builders
│   ├── availability/   Domain RDAP + social platform probers
│   ├── handlers/       HTTP handlers (generate, visuals, export, session…)
│   ├── models/         Shared Go structs
│   └── session/        Redis session store
├── frontend/
│   ├── app/            Next.js App Router (layout, page, HomeClient)
│   ├── components/     React UI components
│   └── lib/            TypeScript types, Zod schemas, API wrappers
├── DesignDoc.md        Full technical reference
├── nomvox-plan-v3.md   SDLC plan ST-01 → ST-15
└── .env.example        Environment variable template
```

## Team

| Member | Role | IBM SkillsBuild Certificate |
|---|---|---|
| c-annabel | Full-stack + AI + Design | *[upload link]* |

---

© 2026 c-annabel — Developed with IBM Bob — All rights reserved.  
*These assets are AI-generated and intended as creative inspiration. Always verify brand name availability before registration.*
