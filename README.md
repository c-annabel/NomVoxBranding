# NomVox

**Signal Sent. Brand Received.**

NomVox is an AI-powered brand identity platform that transforms a raw business concept into a complete, launch-ready brand universe — validated names, social/domain handles, logo assets, mood boards, and landing-page mockups — through a conversational creative-partner loop that learns from every user reaction.

> Built for the **IBM AI Builders Challenge — July 2026 · Creative Industries Theme**
> Primary development tool: **IBM Bob**

---

## Problem Statement

Every new business faces the same blank-page problem: *what do I call this, and does anyone already own that name?* Today that means juggling name-brainstorm sessions, domain registrar tabs, separate searches on Instagram/X/TikTok, a freelancer quote for logo concepts, and days of back-and-forth before you have anything shareable. Current tools (Namelix, Namechk, Looka) are one-shot generators with no conversational memory — a user cannot say "I liked the energy of option 3 but want it shorter" and have the tool genuinely respond to that signal.

## Solution Description

NomVox is a conversational AI brand-identity partner. A founder describes their idea using a structured 8-field intake form; NomVox generates names and taglines enriched with origin stories, brand scores, and voice samples — then automatically filters out any name that doesn't clear an **80% cross-platform availability threshold** (domain + social handles). The user reacts, refines, and selects. The AI explains every rejection, asks a clarifying question, and regenerates with full session memory. After selection, NomVox generates a visual identity pack: mood board, three logo types (profile/app/business), and a hero landing-page mockup.

## AI Approach & Architecture

```
IBM Granite (watsonx.ai) ──► All text & reasoning
  • Name + tagline generation (streamed)
  • Origin stories, brand scores, voice samples
  • Brand persona card ("brand as a person")
  • Anti-name reasoning (explains rejections, asks clarifying Q)
  • Competitor name radar
  • Hero landing-page HTML/CSS

Gemini 2.0 Flash (Google AI Studio, free tier) ──► Vision analysis
  • User uploads reference image → extracts palette, mood, style

Imagen 3 (Google AI Studio, free tier) ──► Image generation
  • Mood board (4 panels)
  • Logo concepts (Profile / App / Business)

Go API server (chi router) ──► Availability engine + session memory
  • RDAP domain check + parallel HTTP HEAD probes (6 platforms)
  • Weighted 80% scoring gate (domain double-weighted)
  • Redis session store (liked/rejected/notes injected into every AI call)
  • ZIP export (archive/zip)

Next.js 14 (Vercel) ──► Frontend
  • Streaming name card render
  • Select / Reject / Reproduce state machine
  • Style DNA slider, session sidebar, visual identity UI
```

## Selected Challenge Theme

**Creative Industries** — spanning three solution areas:
- AI Creative Partner
- Creative Ideation & Brainstorming Platform
- AI-powered Design & Visual Concept Tools

## How IBM Bob Was Used

IBM Bob was the **primary development tool** throughout the full SDLC:
- Architecture planning and technology selection
- Full SDLC plan generation (v1 → v3) including gap analysis, sprint breakdown, and risk register
- Prompt engineering for IBM Granite (name generation, persona, anti-name reasoning, competitor radar)
- Code scaffolding for Go server, session store, availability engine, and all handlers
- Go struct design, TypeScript type definitions, and API route specification
- QA planning, unit test design, and error-state definitions

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Next.js 14, React, Tailwind CSS, Space Grotesk, shadcn/ui |
| Backend | Go 1.22+, chi router, net/http |
| AI — Text | IBM Granite via watsonx.ai |
| AI — Vision | Gemini 2.0 Flash (Google AI Studio, free tier) |
| AI — Images | Imagen 3 (Google AI Studio, free tier) |
| Session store | Redis (Upstash free tier) |
| Asset storage | Vercel Blob |
| Deployment | Vercel (frontend) + Fly.io (Go server) |

## Local Setup

```bash
# 1. Clone
git clone https://github.com/your-org/nomvox.git
cd nomvox

# 2. Copy and fill environment variables
cp .env.example .env
# Edit .env with your actual keys

# 3. Start the Go API server
go mod download
go run ./cmd/server

# 4. Start the Next.js frontend (separate terminal)
cd frontend
npm install
npm run dev
```

The Go server runs on `http://localhost:8080`.
The Next.js app runs on `http://localhost:3000`.

## Live Demo

<!-- Added after Day 14 deployment -->
- **App:** https://nomvox.vercel.app
- **API:** https://nomvox-api.fly.dev/api/ping

## Demo Video

<!-- Added after Day 14 recording -->
[3-minute walkthrough on YouTube](#)

## Team

| Member | IBM SkillsBuild Certificate |
|---|---|
| <!-- name --> | <!-- link --> |
