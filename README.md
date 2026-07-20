# NomVox — Born from the Void

> **AI-powered brand identity platform.**  
> From a single idea to a complete brand universe — coined names, availability checks, logo concepts, mood boards, and a landing page mockup — synthesized in one session.

Built for the **IBM AI Builders Challenge · July 2026 · Creative Industries**  
**Primary development tool: IBM Bob**

[![Live](https://img.shields.io/badge/Live_App-nomvox.vercel.app-8B5CF6?style=flat-square)](https://nomvox.vercel.app)
[![API](https://img.shields.io/badge/API-nomvox--api.fly.dev-22d3ee?style=flat-square)](https://nomvox-api.fly.dev/api/ping)
[![IBM Bob](https://img.shields.io/badge/Built_with-IBM_Bob-0f62fe?style=flat-square)](http://ibm.biz/university-bob)

---

## The Problem

Every founder, creator, and side-hustle builder hits the same blank-page wall: *what do I call this, and does anyone already own it?*

Today that means opening ten tabs — a name brainstorm doc, a domain registrar, Instagram search, X search, TikTok search, a freelancer quote for a logo — and spending days in back-and-forth before having anything shareable. And when a naming tool spits out a static list, there's no way to say "I liked the energy of #3 but want it shorter" and have it actually respond.

**NomVox closes that gap.** It acts as a creative partner, not a generator — it remembers what you liked, what you rejected, and why, and it gets smarter with every reaction.

---

## What NomVox Does

A founder types their idea. NomVox synthesizes up to five **coined brand names** — invented words built from phonaesthetics, syllable forging, ancient root mutation, and neologism splicing. Each name arrives with:

- A **tagline**, **origin story**, and **tone reasoning** — explaining *why* the word works
- A **brand score** across Memorability, Spellability, Global Safety, and Squatter Risk
- **Three brand voice samples** — an Instagram caption, email subject line, and 404 message
- **Live availability checks** across `.com` domain and five social platforms simultaneously
- A **competitor radar** flag if the name semantically clashes with a known brand in the category

The user **likes**, **passes** (with a reason), or **chooses** a name. Passing triggers the AI's **anti-name reasoning** — it asks one clarifying question to understand what was wrong. That answer, plus all liked names and rejection notes, is injected into every subsequent generation call. This is the **creative-partner loop**: the AI learns within the session.

Once a name is chosen, NomVox generates a complete **visual identity pack**: three logo concept directions, a four-panel mood board, a brand persona card, and an AI-generated hero landing page mockup — all calibrated to the user's brand palette and personality. Everything exports as a ZIP.

---

## Why This Wins on Every Judging Criterion

### ✦ Technical Execution
- **IBM watsonx.ai** (Llama 3.3 70B via `ca-tor.ml.cloud.ibm.com`) powers all AI reasoning: name generation with five distinct invention strategies, origin stories, brand scores, voice samples, competitor radar, persona cards, SVG logo concepts, mood board tiles, and HTML landing page mockups
- **Go API server** (chi router) orchestrates parallel goroutines for simultaneous availability probing across six platforms and concurrent visual generation
- **Redis session store** (Upstash) persists the full creative conversation — liked names, rejections, direction notes, slider positions — across every API call
- **Three-tier visual fallback chain**: Imagen 3 PNG → watsonx SVG → CSS art system — the app never shows a broken state

### ✦ Innovation
- **Creative-partner loop**: the AI asks clarifying questions when names are rejected and injects the user's full reaction history into every subsequent prompt — this is not a list generator
- **Anti-name reasoning**: structured rejection flow that makes refusal productive rather than frustrating
- **Style DNA sliders**: two-axis tonal control (Playful↔Premium, Abstract↔Descriptive) that shifts the AI's generation character in real time
- **Competitor radar**: a second AI pass that semantically compares generated names against known brands in the user's industry
- **CSS/SVG visual system**: when AI image generation is unavailable, a fully dynamic palette-responsive design system renders every visual component from the user's stated colour mood — no static placeholders

### ✦ Challenge Fit — Creative Industries
NomVox spans **four of the six example solution areas** simultaneously:
- **AI Creative Partner** — the session-memory-driven name refinement loop
- **Creative Ideation & Brainstorming Platform** — multi-strategy coined name generation with scoring
- **AI-powered Design & Visual Concept Tools** — logos, mood boards, landing page
- **Personalized Creative Assistant** — every session is unique; the AI adapts to individual direction

### ✦ Feasibility
- Deployed and publicly accessible today at [nomvox.vercel.app](https://nomvox.vercel.app)
- Zero-dependency visual fallbacks mean the demo works regardless of API quota state
- Full ZIP export means output is immediately usable by a developer, printer, or registrar

### ✦ Real-World Impact
- Universal audience: anyone launching anything — a startup, a side project, a band, a nonprofit — needs a name and handle before anything else
- Instant demo value: a judge types "eco-friendly coffee brand for young professionals" and watches AI generate names, check availability, and build a visual identity in real time
- Addresses a real pain point with a clear workflow, not a toy

---

## Core Features

| Capability | What it does |
|---|---|
| **Coined name generation** | 5 invention strategies: syllable forge · phonaesthetics · neologism splice · void-word · ancient root mutation |
| **Origin stories** | Etymology and sound-symbolism reasoning for each invented name |
| **Brand scoring** | Memorability · Spellability · Global Safety · Squatter Risk with reasoning |
| **Brand voice samples** | Instagram caption · Email subject · 404 page message — written in the brand's voice |
| **Live availability engine** | Parallel RDAP domain check + HEAD probes: Instagram · X · TikTok · Threads · YouTube |
| **60% availability gate** | Names must clear a weighted threshold; zero-pass fallback shows top-2 partials |
| **Session memory** | Liked/rejected/direction notes persist in Redis and inject into every AI call |
| **Anti-name reasoning** | AI asks one clarifying question when a name is rejected with a note |
| **Style DNA sliders** | Playful↔Premium and Abstract↔Descriptive axes shift generation tone in real time |
| **Competitor radar** | Second AI pass flags semantic overlap with known brands in the user's industry |
| **Logo concepts** | Three CSS/SVG styles: Geometric Bauhaus mark · Glassmorphic app icon · Horizontal wordmark |
| **Mood board** | Four-panel visual board — Colour World · Brand Identity · Pattern DNA · Typography — palette-driven |
| **Brand persona** | "Brand as a person" — age, occupation, voice, what it reads, what it would never say |
| **Landing page mockup** | AI-generated hero section HTML/CSS + CSS art fallback, rendered in a sandboxed iframe |
| **Export ZIP** | brand-brief.json · brand-brief.html · logos/ · mood-board/ · landing-page.html · README.txt |

---

## IBM Technologies at the Core

| IBM Technology | Role |
|---|---|
| **IBM Bob** | Primary development tool across the full SDLC — architecture, code generation, prompt engineering, debugging, documentation |
| **IBM watsonx.ai** | All AI reasoning: name generation, scoring, persona, competitor radar, SVG visuals, landing page HTML |
| **Meta Llama 3.3 70B** via watsonx | Primary model — `ca-tor.ml.cloud.ibm.com` · chat API (`/ml/v1/text/chat`) |
| **IBM Granite** via watsonx | Fallback model — `us-south.ml.cloud.ibm.com` — same API, automatic endpoint selection |
| **IBM Cloud IAM** | API key → Bearer token exchange with 50-minute in-process cache |

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go · chi router · archive/zip · sync.WaitGroup |
| Frontend | Next.js · React · TypeScript · Tailwind CSS · Zod |
| AI — Text & SVG | IBM watsonx.ai (Llama 3.3 70B / IBM Granite) |
| AI — Images | Imagen 3 via Google AI Studio (fallback: watsonx SVG → CSS art) |
| Session store | Redis via Upstash |
| Availability | RDAP (Verisign) + parallel HTTP HEAD probes |
| Deployment | Fly.io (Go API) · Vercel (Next.js) |
| Font | Space Grotesk |
| Analytics | Vercel Analytics |

---

## Quick Start

```bash
# 1. Copy env template
cp .env.example .env
# Fill in: WATSONX_API_KEY, WATSONX_PROJECT_ID, WATSONX_URL, REDIS_URL

# 2. Start Go API (port 8080)
go build -o nomvox-server.exe ./cmd/server
.\nomvox-server.exe

# 3. Start frontend (port 3000)
cd frontend && npm install && npm run dev
```

Open `http://localhost:3000`. Full setup details in `DesignDoc.md`.

---

## Repository

```
NomVoxBranding/
├── cmd/server/          Go API entry point
├── cmd/testllm/         watsonx.ai standalone connection test
├── internal/ai/         watsonx client, Imagen client, all prompt builders
├── internal/availability/ RDAP + social HEAD probes, 60% gate
├── internal/handlers/   HTTP handlers: generate, visuals, export, session
├── internal/models/     Shared Go structs
├── internal/session/    Redis session CRUD
├── frontend/app/        Next.js App Router — layout, page, HomeClient
├── frontend/components/ IntakeForm, NameCard, VisualIdentityPanel, and more
├── frontend/lib/        TypeScript types, Zod schemas, API wrappers
├── 01-Plans/            Full SDLC plan history (HTML)
├── DesignDoc.md         Complete technical reference
├── nomvox-plan.md       SDLC plan with sub-task status
└── developer-bob-feedback.md  Candid feedback on IBM Bob
```

---

## Demo

- **Live app:** https://nomvox.vercel.app
- **API health:** https://nomvox-api.fly.dev/api/ping
- **Demo video:** https://www.youtube.com/@c-annabel

---

## Team

| Member | Role | IBM SkillsBuild |
|---|---|---|
| c-annabel | Full-stack · AI integration · Design | *Uploaded* |

---

*© 2026 c-annabel — Developed with IBM Bob — IBM AI Builders Challenge — All rights reserved.*  
*AI-generated brand assets are creative inspiration. Verify name availability before registration.*
