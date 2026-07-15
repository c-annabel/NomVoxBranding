# NomVox — Design & Developer Reference

> **"From void to voice — your brand, synthesized."**  
> AI-powered brand identity platform for the IBM AI Builders Challenge (Creative Industries, July 2026)

---

## Table of Contents

1. [Project Summary](#1-project-summary)
2. [Tech Stack & Requirements](#2-tech-stack--requirements)
3. [Local Dev — Running Procedure](#3-local-dev--running-procedure)
4. [Environment Variables](#4-environment-variables)
5. [File Structure](#5-file-structure)
6. [Data Models & ERD](#6-data-models--erd)
7. [API Reference](#7-api-reference)
8. [Request / Response Flow](#8-request--response-flow)
9. [Build Stages & Completion Checklist](#9-build-stages--completion-checklist)
10. [Troubleshooting FAQ](#10-troubleshooting-faq)

---

## 1. Project Summary

NomVox transforms a raw brand idea into a complete, launch-ready identity in one session:

- **Coined names** — invented words using phonaesthetics, ancient root mutation, syllable forging, and neologism splicing. Never real dictionary words.
- **Availability gate** — parallel probes of `.com` domain + Instagram / X / TikTok / Threads / YouTube. 60% weighted threshold. Unknown probes (rate-limited) treated as taken.
- **Creative-partner loop** — session memory (Redis) accumulates liked/rejected/direction notes. Every Regenerate call injects full history into the LLM prompt.
- **Visual identity** — Imagen 3 logos (Profile / App / Business) + mood board (4 panels), Granite HTML landing-page mockup, brand persona card.
- **Export** — ZIP containing `brand-brief.json`, `landing-page.html`, `logos/*.png`, `mood-board/*.png`, `README.txt`.

---

## 2. Tech Stack & Requirements

### Runtime Requirements

| Tool | Version | Install |
|---|---|---|
| **Go** | 1.22+ | https://go.dev/dl/ |
| **Node.js** | 20 LTS+ | https://nodejs.org/ |
| **npm** | 10+ | bundled with Node |
| **Git** | any | https://git-scm.com/ |

### Key Dependencies

| Layer | Package | Purpose |
|---|---|---|
| Go | `github.com/go-chi/chi/v5 v5.1.0` | HTTP router |
| Go | `github.com/redis/go-redis/v9 v9.6.1` | Redis session store |
| Go | `github.com/google/uuid v1.6.0` | Session ID generation |
| Go | `github.com/joho/godotenv v1.5.1` | `.env` file loading in dev |
| JS | `next 16.2.10` | React framework (SSR + API routes) |
| JS | `react 19.2.4` | UI library |
| JS | `tailwindcss ^4` | Utility-first CSS |
| JS | `zod ^4.4.3` | Runtime schema validation |
| JS | `@fontsource/space-grotesk ^5.2.10` | Brand typography |

### External Services (all required in `.env`)

| Service | Purpose | Key name |
|---|---|---|
| IBM watsonx.ai | LLM — name generation, persona, mockup, competitor radar | `WATSONX_API_KEY`, `WATSONX_PROJECT_ID`, `WATSONX_URL` |
| Google AI Studio | Imagen 3 image gen + Gemini 2.0 Flash vision | `GOOGLE_AI_API_KEY` |
| Upstash Redis | Session persistence (2hr TTL) | `REDIS_URL` |

---

## 3. Full Technology Reference

> Complete reference for interview, README, and submission documentation.

### Primary Development Tool
| Tool | Role | URL |
|---|---|---|
| **IBM Bob** | Primary AI development assistant — used for code generation, architecture design, debugging, SDLC planning, prompt engineering, and all documentation across the full project lifecycle | http://ibm.biz/university-bob |

### AI / Machine Learning
| Technology | Provider | Role in NomVox | API/Endpoint |
|---|---|---|---|
| **Meta Llama 3.3 70B Instruct** | IBM watsonx.ai (ca-tor region) | Brand name generation, taglines, brand scores, origin stories, voice samples, competitor radar, brand persona, landing page HTML | `POST /ml/v1/text/chat` |
| **IBM Granite 3 8B Instruct** | IBM watsonx.ai (us-south region) | Same as Llama — fallback model when us-south endpoint is configured | `POST /ml/v1/text/chat` |
| **Imagen 3** (`imagen-3.0-generate-002`) | Google AI Studio | AI image generation — 3 logo concepts (Profile / App / Business) + 4-panel mood board | `POST /v1beta/models/imagen-3.0-generate-002:predict` |
| **Gemini 2.0 Flash** | Google AI Studio | Vision analysis — extracts palette/mood/style from user-uploaded reference images | `POST /v1beta/models/gemini-2.0-flash:generateContent` |
| **IBM Cloud IAM** | IBM Cloud | API key → Bearer token exchange for all watsonx.ai calls | `POST https://iam.cloud.ibm.com/identity/token` |

### Backend
| Technology | Version | Role |
|---|---|---|
| **Go** | 1.22 | API server — all business logic, AI orchestration, availability checks, export |
| **chi router** | v5.1.0 | HTTP routing, middleware chain (RequestID, RealIP, Logger, Recoverer, CORS) |
| **go-redis** | v9.6.1 | Redis client — session CRUD with 2-hour TTL |
| **godotenv** | v1.5.1 | `.env` file loading in local development |
| **google/uuid** | v1.6.0 | Session ID generation |
| **encoding/json** | stdlib | JSON marshal/unmarshal for all API payloads |
| **archive/zip** | stdlib | ZIP assembly for brand identity export |
| **sync** | stdlib | WaitGroup for concurrent Imagen 3 calls; Mutex for IAM token cache |
| **net/http** | stdlib | HTTP client for watsonx.ai, Imagen 3, RDAP, and social platform probes |

### Frontend
| Technology | Version | Role |
|---|---|---|
| **Next.js** | 16.2.10 | React framework — SSR, file-based routing, `<Image>` optimisation, metadata/favicon |
| **React** | 19.2.4 | UI component model |
| **TypeScript** | 5.x | Static typing across all components and API boundaries |
| **Tailwind CSS** | v4 | Utility-first CSS — all layout, spacing, responsive breakpoints |
| **Zod** | v4.4.3 | Runtime schema validation for API responses (`parseGenerateResponse`) |
| **Space Grotesk** | v5.2.10 | Brand display font (Google Fonts via `@fontsource`) |
| **Turbopack** | bundled with Next.js 16 | Dev server bundler — sub-second HMR |

### Data / Storage
| Technology | Provider | Role |
|---|---|---|
| **Redis** | Upstash (managed serverless) | Session store — persists `BrandSession` (liked/rejected/direction notes/all generated names) with 2hr TTL. Session memory feeds every LLM call for the creative-partner loop. |
| **IAM Token Cache** | In-process (Go `sync.Mutex`) | Caches the IBM Cloud IAM bearer token for 50 minutes — eliminates 1-2s token exchange overhead on every LLM call |

### Availability Checking (Platform Probes)
| Platform | Method | How |
|---|---|---|
| **Domain (.com)** | RDAP lookup | `GET https://rdap.verisign.com/com/v1/domain/{name}.com` — 404 = available, 200 = registered |
| **Instagram** | HTTP probe | `GET https://www.instagram.com/{handle}/` — 404 = free, 200 = taken |
| **X (Twitter)** | HTTP probe | `GET https://x.com/{handle}` |
| **TikTok** | HTTP probe | `GET https://www.tiktok.com/@{handle}` |
| **Threads** | HTTP probe | `GET https://www.threads.net/@{handle}` |
| **YouTube** | HTTP probe | `GET https://www.youtube.com/@{handle}` |

**Scoring:** Domain=2.0pts, IG/X/TikTok=1.0pt, Threads/YouTube=0.5pt. Total=6.0pts. Gate=60% (3.6pts). All 6 probes run in parallel goroutines with 8s timeout each.

### Deployment & Infrastructure
| Platform | Service | Role |
|---|---|---|
| **Fly.io** | Container hosting (Docker) | Go API server — `fly.toml` configured, multi-stage Dockerfile (golang:1.22-alpine → alpine:3.20) |
| **Vercel** | Serverless frontend | Next.js app — zero-config deployment, automatic HTTPS, Edge CDN |
| **GitHub** | Source control + CI | Repository + GitHub Actions CI (`go build ./...` + `go test ./...` + `npm run build` on every push) |
| **Upstash** | Serverless Redis | Session store — free tier, TLS (`rediss://`), global low-latency |

### Challenge Compliance
| Requirement | How NomVox meets it |
|---|---|
| **IBM Bob as primary dev tool** | Used throughout full SDLC: architecture, all code generation, debugging, prompt engineering, documentation |
| **AI as core functional component** | Llama 3.3 70B via watsonx.ai generates all names, scores, taglines, personas, and HTML mockups. Imagen 3 generates all visual assets. |
| **IBM Granite** | `ibm/granite-3-8b-instruct` on us-south endpoint; Llama on ca-tor. Both via IBM watsonx.ai. |
| **watsonx** | All LLM calls routed through `https://ca-tor.ml.cloud.ibm.com/ml/v1/text/chat` |
| **LangChain / LangFlow** | Not used — custom Go client gives tighter control over token budget and JSON repair |
| **Python / Node.js / React / Next.js** | Next.js 16 + React 19 frontend |

---

## 4. Local Dev — Running Procedure

### First-time setup (do once)

```powershell
# 1. Clone the repo
git clone https://github.com/c-annabel/NomVoxBranding
cd NomVoxBranding\NomVoxBranding

# 2. Copy the env template and fill in your keys
Copy-Item .env.example .env
# Edit .env in your editor — set WATSONX_API_KEY, WATSONX_PROJECT_ID,
# WATSONX_URL, GOOGLE_AI_API_KEY, REDIS_URL

# 3. Download Go dependencies
go mod download

# 4. Install frontend dependencies
cd frontend
npm install
cd ..
```

### Every dev session — standard startup order

Open **two separate PowerShell terminals**.

**Terminal 1 — Go API server**
```powershell
# From the repo root (where go.mod is)
cd C:\Users\c-ann\Desktop\20260410\PROJECTS\012-NomVoxBranding\NomVoxBranding\NomVoxBranding

# Build the binary (only needed after Go source changes)
go build -o nomvox-server.exe ./cmd/server

# Start the server
.\nomvox-server.exe
```

You should see:
```
NomVox API server listening on :8080
```

**Terminal 2 — Next.js frontend**
```powershell
cd frontend
npm run dev
```

You should see:
```
▲ Next.js 16.2.10
- Local: http://localhost:3000
```

Open **http://localhost:3000** in your browser.

### Quick rebuild after Go changes (no full restart)

```powershell
# In Terminal 1 — stop server with Ctrl+C, then:
go build -o nomvox-server.exe ./cmd/server ; .\nomvox-server.exe
```

### Verify everything is working

```powershell
# Quick health check
Invoke-RestMethod http://localhost:8080/api/ping

# Expected: { "status": "ok", "service": "nomvox" }

# Full AI + Redis diagnosis
Invoke-RestMethod http://localhost:8080/api/diagnose

# Test Imagen 3 + watsonx connectivity
Invoke-RestMethod http://localhost:8080/api/debug-visuals
```

---

### Do I need to kill the old process every time?

**No — not usually.** If you stop the server cleanly with `Ctrl+C`, the port is released and you can start again immediately.

You only need to force-kill if:
- The server **crashed** without releasing the port (you see `bind: address already in use`)
- You closed the terminal window **without pressing Ctrl+C**
- The process is stuck and not responding

**Force-kill command** (only when port 8080 is stuck):
```powershell
# Kill by process name
Get-Process -Name "nomvox-server" -ErrorAction SilentlyContinue | Stop-Process -Force

# Or kill by port — find PID then kill
netstat -ano | findstr ":8080"
# Note the PID from the last column, then:
Stop-Process -Id <PID> -Force
```

---

### Do I need to delete the log?

**No.** `nomvox.log` is an append-only file. It never causes the server to fail. You can safely delete it at any time to start fresh, but it is not required.

```powershell
# Optional — clear the log (not required for normal operation)
Clear-Content nomvox.log

# Or delete it entirely (server recreates it automatically)
Remove-Item nomvox.log -ErrorAction SilentlyContinue
```

The log accumulates timestamped LLM call records like:
```
[14:22:07] generate: calling LLM (nameCount=5, sessionID=abc-123)
[14:22:19] generate: LLM responded in 12.3s, raw len=2847
[14:22:19] generate: parsed 5 cards OK
```

---

### TypeScript-only changes (no Go rebuild needed)

Changes to `frontend/` files are picked up **live** by `npm run dev` via Hot Module Replacement. You do **not** need to restart the frontend. You do **not** need to restart the Go server.

---

### Full clean rebuild (after dependency changes)

```powershell
# Go — after editing go.mod or go.sum
go mod tidy
go build -o nomvox-server.exe ./cmd/server

# Frontend — after editing package.json
cd frontend
npm install
npm run dev
```

---

## 4. Environment Variables

All variables live in `.env` at the repo root. The server loads this file automatically in dev. In production (Fly.io / Vercel) they are injected by the platform.

| Variable | Required | Description |
|---|---|---|
| `WATSONX_API_KEY` | ✅ | IBM Cloud IAM API key (from cloud.ibm.com/iam/apikeys) |
| `WATSONX_PROJECT_ID` | ✅ | watsonx.ai project ID (from watsonx console) |
| `WATSONX_URL` | ✅ | Endpoint base URL, e.g. `https://ca-tor.ml.cloud.ibm.com` |
| `WATSONX_CPD_URL` | ⬜ | Only set for Cloud Pak for Data deployments — leave empty for public IBM Cloud |
| `GOOGLE_AI_API_KEY` | ✅ | Google AI Studio key — used for both Imagen 3 (image gen) and Gemini 2.0 Flash (vision) |
| `REDIS_URL` | ✅ | Upstash Redis connection string: `rediss://:password@host.upstash.io:6379` |
| `ALLOWED_ORIGIN` | ⬜ | CORS origin for the frontend. Defaults to `http://localhost:3000` |
| `PORT` | ⬜ | HTTP port. Defaults to `8080` |
| `VERCEL_BLOB_TOKEN` | ⬜ | Vercel Blob storage (for future screenshot feature — not used in current build) |

**Model selection is automatic** — `capsForEndpoint()` in `internal/ai/granite.go` picks the model based on `WATSONX_URL`:
- `ca-tor.*` → `meta-llama/llama-3-3-70b-instruct` + chat API
- `us-south.*` → `ibm/granite-3-8b-instruct` + chat API

---

## 5. File Structure

```
NomVoxBranding/                        ← repo root (go.mod lives here)
│
├── cmd/
│   ├── server/
│   │   └── main.go                    ← HTTP server entry point, chi router, CORS
│   └── testllm/
│       └── main.go                    ← Standalone LLM smoke test (no server needed)
│
├── internal/
│   ├── ai/
│   │   ├── granite.go                 ← watsonx.ai client: IAM token exchange, chat/text API,
│   │   │                                 ExtractJSON / ExtractJSONObject, truncation repair
│   │   ├── imagen.go                  ← Imagen 3 image gen + Gemini 2.0 Flash vision analysis,
│   │   │                                 Base64ToDataURI (RawURLEncoding fallback chain)
│   │   ├── prompts.go                 ← systemPromptBase (5 naming strategies), BuildSystemPrompt,
│   │   │                                 BuildUserPrompt, ParseNameCards
│   │   └── visual_prompts.go          ← MoodBoardPrompt, LogoConceptPrompt,
│   │                                     BuildPersonaSystemPrompt/UserPrompt,
│   │                                     BuildMockupSystemPrompt/UserPrompt
│   │
│   ├── availability/
│   │   └── checker.go                 ← 6-platform parallel probes (RDAP + HTTP),
│   │                                     weighted 60% gate, partials fallback, sanitiseHandle
│   │
│   ├── handlers/
│   │   ├── generate.go                ← POST /api/generate — session load, LLM call,
│   │   │                                 ParseNameCards, session save, file logging
│   │   ├── availability.go            ← POST /api/availability — CheckNames, Competitor Radar
│   │   ├── session.go                 ← POST /api/session · PATCH /api/session/react
│   │   │                                 (6 action types: like/reject/note/visual-note/slider/select)
│   │   ├── visuals.go                 ← POST /api/visuals — 5 concurrent goroutines:
│   │   │                                 mood board, 3 logos, mockup HTML, brand persona
│   │   ├── export.go                  ← POST /api/export — ZIP assembly, dataURIToBytes
│   │   │                                 (RawURLEncoding fallback), selected-logo.png
│   │   ├── diagnose.go                ← GET /api/diagnose — env + Redis check (DEV)
│   │   ├── debug_llm.go               ← GET /api/debug-llm — direct LLM smoke test (DEV)
│   │   └── debug_visuals.go           ← GET /api/debug-visuals — Imagen 3 + Granite test (DEV)
│   │
│   ├── models/
│   │   └── types.go                   ← All shared Go structs (IntakePayload, NameCard,
│   │                                     BrandScore, VoiceSamples, PlatformResult,
│   │                                     AvailabilityResult, VisionContext, VisualPack,
│   │                                     BrandPersona, BrandSession)
│   │
│   └── session/
│       └── store.go                   ← Redis CRUD wrapper (Get/Set/Ping), 2hr TTL
│
├── frontend/
│   ├── app/
│   │   ├── layout.tsx                 ← Root layout — font, global CSS, metadata
│   │   ├── page.tsx                   ← Server component → renders HomeClient
│   │   ├── HomeClient.tsx             ← Main client component: 5-screen state machine,
│   │   │                                 runAvailability (60s timeout), all handlers
│   │   └── globals.css                ← Brand CSS tokens, .nv-range slider styles,
│   │                                     .nv-dot1/2/3 loading animations
│   │
│   ├── components/
│   │   ├── IntakeForm.tsx             ← 8-field intake, richness meter, Inspire Me button
│   │   ├── NameCardComponent.tsx      ← 2-row 7-col table card: name, tagline, origin,
│   │   │                                 availability, brand score, brand voice, actions
│   │   ├── AvailabilityBadges.tsx     ← Platform dots, score bar, competitor radar warning
│   │   ├── VisualIdentityPanel.tsx    ← 3-step visual flow: logos → mood board → mockup
│   │   ├── StyleDNASlider.tsx         ← Playful↔Premium / Abstract↔Descriptive sliders
│   │   ├── RejectionDialog.tsx        ← AI clarifying question after reject+note
│   │   ├── BrandScoreCard.tsx         ← (standalone score card — unused in main table)
│   │   └── NameCardSkeleton.tsx       ← (loading skeleton — unused, replaced by dot animation)
│   │
│   ├── lib/
│   │   ├── types.ts                   ← All TypeScript interfaces (mirrors models/types.go)
│   │   ├── schemas.ts                 ← Zod validation, parseGenerateResponse()
│   │   └── api.ts                     ← Typed API fetch wrappers
│   │
│   └── public/
│       ├── nomvox-logo2.png           ← Active logo (header)
│       └── nomvox-bg.png              ← Star-field background texture
│
├── .env                               ← Local secrets (git-ignored)
├── .env.example                       ← Template — commit this, never .env
├── .gitignore
├── Dockerfile                         ← Multi-stage Go build → alpine runtime
├── fly.toml                           ← Fly.io deployment config
├── go.mod                             ← Go module: github.com/c-annabel/NomVoxBranding
├── go.sum
├── nomvox.log                         ← Append-only LLM call log (git-ignored)
├── nomvox-server.exe                  ← Built binary (git-ignored)
├── nomvox-plan-v3.md                  ← Full SDLC plan
├── DesignDoc.md                       ← This file
├── README.md
├── test-diagnose.ps1                  ← Dev test script
└── test-generate.ps1                  ← Dev test script
```

---

## 6. Data Models & ERD

### Core Data Flow

```
Browser (Next.js)
    │  POST /api/generate  { session_id, intake, action }
    ▼
GenerateHandler (Go)
    │  reads ──────────────────────────────────► Redis
    │  BrandSession { liked[], rejected[],        (2hr TTL)
    │    direction_notes[], all_generated[], ... }
    │
    │  builds prompts via BuildSystemPrompt(sess) + BuildUserPrompt(intake, n)
    │
    │  POST to watsonx.ai ──────────────────────► IBM Granite / Llama
    │  { model_id, project_id, messages[system,user], parameters }
    │
    │  ExtractJSON + ParseNameCards → []NameCard
    │
    │  writes ─────────────────────────────────► Redis
    │  (appends names to sess.AllGenerated)
    │
    ▼
{ session_id, cards: []NameCard }
    │
    ▼  (frontend immediately fires)
POST /api/availability  { session_id, names[] }
    │
    │  6 goroutines in parallel:
    │    RDAP(.com)  Instagram  X  TikTok  Threads  YouTube
    │    each returns (available bool, unknown bool)
    │
    │  weighted score  (60% gate = 3.6/6.0 pts)
    │  passing → show  |  zero passing → top-2 partials + warning banner
    │
    ▼
{ passing: []AvailabilityResult, partials: []AvailabilityResult }
    │
    ▼  (user selects a name)
POST /api/visuals  { session_id, card, intake }
    │
    │  5 goroutines in parallel:
    │    Imagen 3 ──► 4 mood board images  (base64 → data URI)
    │    Imagen 3 ──► logo profile          (1:1)
    │    Imagen 3 ──► logo app icon         (1:1)
    │    Imagen 3 ──► logo business card    (16:9)
    │    Granite  ──► landing page HTML     (sandboxed iframe)
    │    Granite  ──► brand persona JSON
    │
    ▼
{ mood_board[], logo_profile, logo_app, logo_business, mockup_html, persona }
    │
    ▼  (user clicks Export)
POST /api/export  { brand_name, card, intake, mood_board[], logos, selected_logo_key, ... }
    │
    │  ZIP assembly:
    │    brand-brief.json
    │    landing-page.html
    │    logos/logo-profile.png
    │    logos/logo-app-icon.png
    │    logos/logo-business-card.png
    │    logos/selected-logo.png        ← copy of user's chosen logo
    │    mood-board/panel-1..4.png
    │    README.txt
    │
    ▼
application/zip  (streamed to browser)
```

### Entity Relationship Diagram

```
┌──────────────────────────────────────────────────────┐
│                    IntakePayload                      │
│  core_idea (required)     target_audience             │
│  personality              style                       │
│  industry                 color_mood                  │
│  name_length              avoid                       │
└──────────────┬───────────────────────────────────────┘
               │ 1
               │ produces
               │ N
┌──────────────▼───────────────────────────────────────┐
│                      NameCard                         │
│  name          tagline        tone_reasoning          │
│  style_tags[]  short_desc     long_desc               │
│  origin_story                                         │
│                                                       │
│  ┌───────────────────┐  ┌──────────────────────────┐ │
│  │    BrandScore     │  │      VoiceSamples        │ │
│  │  memorability     │  │  instagram_caption       │ │
│  │  spellability     │  │  email_subject           │ │
│  │  global_safety    │  │  not_found_message       │ │
│  │  squatter_risk    │  └──────────────────────────┘ │
│  │  *_reasoning ×4   │                               │
│  └───────────────────┘                               │
└──────────────┬───────────────────────────────────────┘
               │ 1
               │ checked by
               │ 1
┌──────────────▼───────────────────────────────────────┐
│                  AvailabilityResult                   │
│  name      score (0–100)    passes (≥60%)             │
│  radar (competitor warning)                           │
│                                                       │
│  ┌──────────────────────────────────────────────────┐ │
│  │               PlatformResult                     │ │
│  │  domain   instagram   x   tiktok   threads       │ │
│  │  youtube  + *_unknown flags for each             │ │
│  └──────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────┐
│                    BrandSession                       │  ← stored in Redis
│  session_id    intake (IntakePayload)                 │     key: "session:{id}"
│  liked[]       rejected[]     all_generated[]         │     TTL: 2 hours
│  direction_notes[]            visual_notes[]          │
│  slider_playful  slider_abstract                      │
│  selected_name                                        │
│                                                       │
│  vision_ctx? (VisionContext)                          │
│  visuals?    (VisualPack)                             │
│  persona?    (BrandPersona)                           │
└──────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────┐
│                    VisualPack                         │
│  mood_board[]  (4 data URIs)                          │
│  logo_profile  logo_app  logo_business  (data URIs)  │
└──────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────┐
│                    BrandPersona                       │
│  age   occupation   voice                             │
│  reads[]   never_says[]   core_values[]               │
└──────────────────────────────────────────────────────┘
```

### Availability Scoring Weights

```
Platform     Weight    Notes
──────────── ─────── ──────────────────────────────────────
domain        2.0    RDAP lookup — verisign.com/com/v1/
instagram     1.0    GET instagram.com/{handle}/
x             1.0    GET x.com/{handle}
tiktok        1.0    GET tiktok.com/@{handle}
threads       0.5    GET threads.net/@{handle}
youtube       0.5    GET youtube.com/@{handle}
──────────── ─────── ──────────────────────────────────────
TOTAL         6.0
THRESHOLD     3.6    60% gate
──────────── ─────── ──────────────────────────────────────

Response code logic (all platforms except domain):
  404  → available = true   (handle is free)
  200  → available = false  (handle is taken)
  429 / 5xx / error → unknown = true → treated as TAKEN (conservative)
  3xx / 403 / other → available = false (conservative)
```

---

## 7. API Reference

All routes are prefixed `/api`. The server runs on `http://localhost:8080` in dev.

| Method | Path | Timeout | Description |
|---|---|---|---|
| `GET` | `/api/ping` | — | Health check. Returns `{"status":"ok"}` |
| `GET` | `/api/diagnose` | 30s | DEV: checks env vars + Redis connectivity |
| `GET` | `/api/debug-llm` | 60s | DEV: sends simple prompt to watsonx, returns raw response |
| `GET` | `/api/debug-visuals` | 90s | DEV: fires test prompt at Imagen 3 + Granite, returns `{imagen_ok, granite_ok, ...}` |
| `POST` | `/api/generate` | 120s | Main generation — LLM call, session load/save, returns `{session_id, cards[]}` |
| `POST` | `/api/session` | 30s | Create or retrieve a session |
| `PATCH` | `/api/session/react` | 30s | Record user reaction — actions: `like/reject/note/visual-note/slider/select` |
| `POST` | `/api/availability` | 60s | Probe all 6 platforms, return `{passing[], partials[]}` |
| `POST` | `/api/visuals` | 120s | Generate mood board + 3 logos + mockup HTML + persona (5 concurrent) |
| `POST` | `/api/export` | 30s | Assemble and stream a ZIP of all brand assets |

### POST /api/generate

```json
// Request
{
  "session_id": "",          // empty on first call, populated on regenerate
  "intake": {
    "core_idea": "eco coffee for young professionals",
    "target_audience": "25-35 urban millennials",
    "personality": "warm, approachable",
    "style": "minimalist",
    "industry": "food & beverage",
    "color_mood": "earthy greens",
    "name_length": "short, 1-2 syllables",
    "avoid": "generic words like 'bean' or 'brew'"
  },
  "action": "init"           // "init" | "reproduce"
}

// Response
{
  "session_id": "uuid-here",
  "cards": [
    {
      "name": "Veyrak",
      "tagline": "Warmth brewed from the void.",
      "tone_reasoning": "Plosive V anchors boldness.",
      "style_tags": ["earthy", "bold"],
      "short_desc": "Coffee crafted for minds that move.",
      "long_desc": "Veyrak is a ritual, not a routine. Built for the deliberate.",
      "origin_story": "Old Norse 'ver' (spring) + Basque 'rak' (root fragment), void-word method.",
      "score": {
        "memorability": 9, "spellability": 8, "global_safety": 9,
        "squatter_risk": "Low",
        "mem_reasoning": "Two clean syllables, punchy.",
        "spell_reasoning": "Phonetic, no silent letters.",
        "global_reasoning": "No cultural clash detected.",
        "squatter_reasoning": "Invented — unlikely registered."
      },
      "voice_samples": {
        "instagram_caption": "Every sip is a signal. ✦ #Veyrak",
        "email_subject": "Your ritual just got rarer.",
        "not_found_message": "This page dissolved into the void."
      }
    }
  ]
}
```

### PATCH /api/session/react

```json
// Like
{ "session_id": "uuid", "action": "like", "name": "Veyrak" }

// Reject with reason (triggers clarifying question in response)
{ "session_id": "uuid", "action": "reject", "name": "Veyrak", "note": "too harsh sounding" }

// Direction note
{ "session_id": "uuid", "action": "note", "note": "I want something softer and more mysterious" }

// Slider update
{ "session_id": "uuid", "action": "slider", "slider_playful": 0.7, "slider_abstract": 0.8 }

// Select (locks the name, begins visual phase)
{ "session_id": "uuid", "action": "select", "name": "Veyrak" }
```

---

## 8. Request / Response Flow

### Screen state machine (frontend)

```
"intake"
  │  user submits form
  ▼
"loading"
  │  POST /api/generate  (≈12-18s)
  │  POST /api/availability runs in parallel (≈5-15s)
  │  60-second wall-clock timeout on availability
  ▼
"results"
  │  cardStates filtered to ≥60% passing names
  │  (or top-2 partials + ⚠ banner if none pass)
  │
  ├─ user clicks Like     → PATCH /api/session/react {action:"like"}
  │                          → session memory updated
  │
  ├─ user clicks Pass     → PATCH /api/session/react {action:"reject"}
  │                          → if note provided → clarifying_question returned
  │                          → step → "clarifying"
  │
  ├─ user clicks Regen    → handleRegenerate → POST /api/generate (with session_id)
  │
  └─ user clicks Select   → PATCH /api/session/react {action:"select"}
                          → POST /api/visuals  (≈30-90s, 5 concurrent goroutines)
                          → step → "visuals-loading" → "visuals"
                          │
                          └─ user clicks Export → POST /api/export
                                                → ZIP downloaded
```

### Session memory accumulation

Every `POST /api/generate` call:
1. Loads existing `BrandSession` from Redis by `session_id`
2. Appends new names to `AllGenerated[]`
3. `BuildSystemPrompt(sess)` injects into system message:
   - `Do NOT suggest: [all_generated]`
   - `Do NOT repeat rejected: [rejected]`
   - `User liked: [liked] — stay consistent`
   - `Direction notes: [direction_notes]`
   - `Tone: lean [playful/premium] and [abstract/descriptive]`
4. Writes updated session back to Redis (resets 2hr TTL)

---

## 9. Build Stages & Completion Checklist

| Stage | Description | Status |
|---|---|---|
| **ST-01** | Go module scaffold, chi router, `/api/ping`, Redis session store, `.env.example`, Dockerfile, GitHub Actions CI, `README.md` | ✅ Complete |
| **ST-02** | 8-field intake form, richness meter, Inspire Me button, brand CSS tokens, star-field background | ✅ Complete |
| **ST-03** | IBM watsonx.ai integration — IAM token exchange, `/ml/v1/text/chat`, `ExtractJSON` with truncation repair, file logging | ✅ Complete |
| **ST-04** | Redis session memory (`BrandSession`), `POST /api/session`, `PATCH /api/session/react` (6 action types), Anti-Name Reasoning, `RejectionDialog` | ✅ Complete |
| **ST-05** | Availability engine — RDAP domain, 6-platform parallel goroutines, weighted 60% gate (lowered from 80%), zero-pass fallback (top-2 partials + ⚠ banner), Competitor Radar | ✅ Complete |
| **ST-06** | Full state machine UI — Like / Reject / Select, `StyleDNASlider`, liked names summary banner, two-screen flow | ✅ Complete |
| **ST-07** | `NameCardComponent` — 2-row 7-col table layout, `rejectedReason` banner, column head colours | ✅ Complete |
| **ST-08** | `/api/visuals` — 5 concurrent goroutines (mood board ×4, 3 logos, mockup HTML, persona). `Base64ToDataURI` fixed to `RawURLEncoding` | ✅ Complete |
| **ST-09** | Landing-page mockup via Granite — `BuildMockupSystemPrompt/UserPrompt`, 640px iframe in step 3/3 | ✅ Complete |
| **ST-10** | `/api/export` — ZIP with `brand-brief.json`, `landing-page.html`, `mood-board/`, `logos/`, `README.txt`. `dataURIToBytes` fixed to `RawURLEncoding`. `selected_logo_key` → `logos/selected-logo.png` | ✅ Complete |
| **ST-11** | `GET /api/debug-visuals`, Brand Score `text-sm` + `max-w-[80px]` bars, Brand Voice `text-sm`, availability 60s timeout | ✅ Complete |
| **ST-12** | Radical name-invention prompt — 5 strategies (syllable forge, phonaesthetics, neologism splice, void-word, ancient root mutation) | ✅ Complete |
| **ST-13** | Unit tests, mobile responsive check (375px / 768px / 1280px) | ⬜ Pending |
| **ST-14** | Deploy to Fly.io (Go server) + Vercel (Next.js frontend) | ⬜ Pending |
| **ST-15** | Demo video (3 min max), GitHub README final polish, submission page on challenge platform | ⬜ Pending |

---

## 10. Troubleshooting FAQ

---

### Server won't start — "bind: address already in use"

Port 8080 is already occupied. Kill whatever is holding it:

```powershell
# Option A — kill by name
Get-Process -Name "nomvox-server" -ErrorAction SilentlyContinue | Stop-Process -Force

# Option B — kill by port
netstat -ano | findstr ":8080"
# Read the PID from the last column, e.g. 12345
Stop-Process -Id 12345 -Force
```

---

### Server starts but `/api/ping` returns nothing / connection refused

The server binary may not have been rebuilt after source changes:
```powershell
go build -o nomvox-server.exe ./cmd/server
.\nomvox-server.exe
```

---

### "AI service unavailable — check WATSONX_API_KEY"

1. Check `.env` has `WATSONX_API_KEY` and `WATSONX_PROJECT_ID` set (no extra spaces)
2. Run `Invoke-RestMethod http://localhost:8080/api/diagnose` — it lists which env vars are missing
3. Confirm your IBM Cloud API key has access to the watsonx.ai project

---

### Names generate but all show 0% / nothing passes the gate

This is expected behaviour. Social platforms (Instagram, X, TikTok) actively block server-side HTTP probes with 403 or 429. The app treats these as "taken" (conservative). You will often see:

- Domain probe works (RDAP is reliable — 2.0 pts)
- Social probes return unknown (treated as taken — 0 pts)
- Maximum reliable score without social confirmation ≈ 2.5/6 = 41%

The 60% gate is already set to the minimum viable threshold. When nothing passes, the app falls back to showing the **top 2 partials** with an amber banner:

> _"Unable to find available and creative names in 1 minute — showing 2 closest matches. Please regenerate to try again."_

Hit **Regenerate** — the AI creates a fresh batch of coined names using session memory, and the availability probe runs again. Invented names (void-words, syllable fusions) are significantly less likely to be registered than real words.

---

### Logos / mood board show "Not available" after visuals generate

This was a known base64 bug — now fixed. Verify your `.env` has `GOOGLE_AI_API_KEY` set, then run:

```powershell
Invoke-RestMethod http://localhost:8080/api/debug-visuals
```

Expected good response:
```json
{ "imagen_ok": true, "data_uri_len": 340000, "granite_ok": true }
```

If `imagen_ok: false`, check `imagen_error` in the response — it will show the actual error from Google AI Studio (e.g. quota exceeded, invalid key, region not supported).

---

### "LLM returned 0 parseable name cards"

The model returned something that couldn't be parsed as a JSON array. Causes:

1. **Token limit hit** — the response was truncated. The server auto-repairs truncated JSON, but very long outputs can still fail. Check `nomvox.log` for `raw len=` — if it's >4000 the response was cut.
2. **Model returned prose** — the LLM ignored the JSON-only instruction. Run `Invoke-RestMethod http://localhost:8080/api/debug-llm` to see a raw response.
3. **Retry automatically occurs** — `generate.go` retries once with a simpler prompt. If both fail, a 502 is returned.

---

### Session keeps repeating the same names (even after Regenerate)

Old session ID is cached in `localStorage`. Open browser DevTools → Application → Local Storage → `localhost:3000` → delete `nomvox_session_id`. Then click **Start Over** in the UI. This forces a fresh session with no accumulated `AllGenerated[]` history.

---

### Export ZIP downloads but images are empty / 0 bytes

The base64 decoder fix (`RawURLEncoding`) is applied in `export.go`. If you still see empty images:

1. Check that visuals actually generated (logos should be visible in Step 1 of the visual panel)
2. The export reads directly from the data URIs passed by the frontend — if the image failed to generate server-side, there's nothing to export
3. Run `/api/debug-visuals` to confirm Imagen 3 is working

---

### `npm run dev` shows TypeScript errors

Run the type check explicitly:
```powershell
cd frontend
npx tsc --noEmit
```

Common causes:
- Added a field to `ExportRequest` in `types.ts` but didn't add it to the component
- Mismatched interface name between `types.ts` and a component import

---

### Redis errors on startup ("REDIS_URL not set" or "dial tcp: connection refused")

The server degrades gracefully — it runs without Redis but session memory is disabled (every generate call starts fresh). To enable full session memory:

1. Create a free Upstash account at https://upstash.com
2. Create a Redis database → copy the `rediss://` connection string
3. Add `REDIS_URL=rediss://:password@host.upstash.io:6379` to `.env`
4. Restart the server

---

### How to rebuild just the Go binary (fastest cycle)

```powershell
go build -o nomvox-server.exe ./cmd/server ; .\nomvox-server.exe
```

The semicolon (`;`) runs both commands sequentially in PowerShell even if the first fails. Use `if ($?) { .\nomvox-server.exe }` if you want the second to only run on success.

---

### How to run the LLM standalone (no server, no browser)

```powershell
# From repo root — sends a test prompt directly to watsonx.ai
go run ./cmd/testllm/
```

This prints the raw LLM response and parsed cards to stdout. Useful for prompt engineering without starting the full server.

---

*Last updated: 2026-07 — NomVox v1 build sprint*
