# NomVox SDLC Plan — v5 (Final, Deployed)

> Last updated: July 2026
> Status: Prototype complete · Deployed to production 

---

## Project Overview

NomVox is an AI-powered brand identity platform that transforms a raw business concept into a complete, launch-ready brand universe — validated names, social/domain handles, logo assets, mood boards, and landing page mockups — through a conversational creative-partner loop that learns from every user reaction.

**Sprint:** 15 days (completed)  
**Module:** `github.com/c-annabel/NomVoxBranding`  
**Backend:** Go 1.22 (net/http + chi v5 router) — deployed on Fly.io (`nomvox-api`, region `iad`)  
**Frontend:** Next.js 16.2 + React 19 + Tailwind CSS v4 + Space Grotesk — deployed on Vercel  
**AI Core (text/reasoning):** IBM Granite / Llama 3.3 70B via watsonx.ai (`ca-tor.ml.cloud.ibm.com`) — name generation, taglines, origin stories, brand scores, persona, voice samples, competitor radar, SVG logos, SVG mood board tiles, landing page HTML  
**AI Visual (production):** CSS/SVG art system (palette-driven, personality-responsive) — primary visual path; replaces Imagen 3 which hit free-tier credit exhaustion. See "Image Generation Decision" section.  
**Session store:** Redis via Upstash (2-hour TTL)  
**Primary dev tool:** IBM Bob

---

## Brand Identity

### Etymology
- **NOM** — Latin/French *nomen*: name, law, system. Structure, identity, registration.
- **VOX** — Latin: voice. Also references Alpha Vulpeculae (Anser), the brightest star in the Vulpecula constellation — the cosmic voice. Expression — giving a voice to an idea in the cosmos.

### Taglines
- **"From void to voice — your brand, synthesized."** ← app hero
- **"Signal sent. Brand received."** ← short punchy variant
- **"NAME IT. CLAIM IT. LAUNCH IT."** ← functional baseline
- **"Your idea has a name. The universe just hasn't heard it yet."** ← video hook

### Colour Palette
| Token | Hex | Usage |
|---|---|---|
| `--color-void` / `#0B0F19` | Deep space | Primary background |
| `--color-nebula` / `#141824` | Dark card | Secondary bg, panels |
| `--color-pulse` / `#8B5CF6` | NOM purple | CTAs, active states, accent |
| `--color-signal` / `#22d3ee` | VOX cyan | Links, badges, secondary accent |
| `--color-lunar` / `#E6E8F0` | White | Primary text, icons |

### Typography
- **Display + logo:** Space Grotesk — Wide, Modern, Technical
- **Body:** Space Grotesk Regular/Medium
- **Monospace:** JetBrains Mono (score readouts, availability badges)

---

## Image Generation Decision: Imagen 3 → CSS/SVG System

### What was planned
Imagen 3 via Google AI Studio for logo concepts (3 types) and mood board tiles (4 panels). Gemini 2.0 Flash for optional vision image analysis (palette extraction from uploaded reference image).

### What happened
During development, local tests with Imagen 3 worked. Before the production demo, Google AI Studio free-tier prepaid credits for image generation were **exhausted permanently**. Further calls return HTTP 429 `RESOURCE_EXHAUSTED`. Re-enabling requires billing on Google Cloud.

Concurrently, watsonx.ai also hit 429 `RESOURCE_EXHAUSTED` on the SVG generation fallback — the free tier enforces a 2 requests/second limit, and NomVox's 7-concurrent visual calls all fail simultaneously.

IBM watsonx.ai does not currently offer a native image generation model comparable to Imagen, DALL-E, or Stable Diffusion.

### What was built instead
A fully dynamic **CSS/SVG visual system** — `extractPalette()` parses the user's `color_mood` field (using bigram matching: "bright orange", "sky blue", etc.) and derives `bg`, `accent`, `accent2`, `text` tokens. Every visual component renders exclusively from these tokens:

- **Logo placeholders (3 CSS compositions):**
  - *Profile/Social:* Geometric hexagonal mark with orbital ring and corner accent dots
  - *App Icon:* Glassmorphic rounded-square card with ambient glow blobs
  - *Business/Print:* Horizontal wordmark lockup with icon mark and tagline

- **Mood board (4 CSS panels — always 4 visible regardless of AI response count):**
  - *Colour World:* Full-saturation accent gradient background with palette swatches
  - *Brand Identity:* Semi-dark bg with brand name, hexagonal mark, tagline
  - *Pattern DNA:* Hexagonal shape grid with accent-tinted background
  - *Typography + CTA:* Type scale display with solid CTA button

- **Landing page (full CSS hero section):**
  - Nav bar with logo mark + brand name + nav links
  - Two-column hero: headline + tagline + CTA (left) / geometric accent (right)
  - Personality-driven button radius (pill for playful, sharp for edgy, rounded for standard)
  - Feature strip with industry, personality, target audience

### The Quota Cascade — What Actually Happened in Production

On a session with multiple name regenerations, the watsonx.ai free-tier daily token quota can be exhausted entirely before the user reaches the visual phase. When this happens:

```
Imagen 3 → HTTP 429 RESOURCE_EXHAUSTED
  "Your prepayment credits are depleted."
  (Google AI Studio free-tier credits exhausted permanently)

watsonx SVG fallback → HTTP 403 token_quota_reached
  "Request of 1 token(s) from quota was rejected"
  (watsonx free-tier daily budget consumed by name gen calls)

Result: all 9 visual tasks fail simultaneously
  visuals: persona → 403
  visuals: svg logo profile → 403
  visuals: svg logo app → 403
  visuals: svg logo business → 403
  visuals: svg mood tile 0 → 403
  visuals: svg mood tile 1 → 403
  visuals: svg mood tile 2 → 403
  visuals: svg mood tile 3 → 403
  visuals: mockup html → 403
```

**The 403 `token_quota_reached` and 429 `RESOURCE_EXHAUSTED` are two different failures:**
- `429` = per-second rate limit → fixed by `throttleGranite()` serialization with 600ms gaps
- `403` = daily token budget exhausted → only resets at midnight UTC or on plan upgrade

The `throttleGranite()` mutex prevents the rate-limit 429s. Nothing can prevent the quota 403s except spending fewer tokens overall or upgrading the account.

### Future path
When IBM watsonx.ai releases a native image generation model, or when a paid Stability AI / Replicate / Together.ai integration is scoped and budgeted, the CSS components are designed for drop-in replacement at the same JSX insertion points (`VisualIdentityPanel.tsx`).

---

## Architecture: Actual Production Stack

```
User Browser (nomvox.vercel.app)
    │
    │ HTTPS
    ▼
Vercel Edge (Next.js 16.2, React 19, Tailwind v4)
    │  frontend/app/HomeClient.tsx — 5-screen state machine
    │  frontend/components/VisualIdentityPanel.tsx — visual steps 1-3
    │
    │ HTTPS REST (NEXT_PUBLIC_API_URL=https://nomvox-api.fly.dev)
    ▼
Fly.io shared-cpu-1x (Go 1.22, chi v5)
    │  cmd/server/main.go — routes all handlers
    │  internal/handlers/ — generate, visuals, export, session, availability
    │
    ├──► Upstash Redis (session store, availability cache, IAM token cache)
    │
    ├──► IBM watsonx.ai ca-tor (Llama 3.3 70B)
    │       IAM token from iam.cloud.ibm.com/identity/token
    │       POST /ml/v1/text/generation
    │       Rate limit: 2 req/s — all calls now sequential with delay
    │
    ├──► Verisign RDAP API (domain availability — no key)
    │       GET https://rdap.org/domain/{name}.com
    │
    └──► Social platform HEAD probes (5 platforms, parallel goroutines)
            instagram.com/{h}, x.com/{h}, tiktok.com/@{h},
            threads.net/@{h}, youtube.com/@{h}
```

---

## Complete File Structure

```
NomVoxBranding/
│
├── cmd/
│   ├── server/main.go              ← Go entry point, all routes registered
│   ├── testllm/main.go             ← watsonx.ai connection test
│   └── diagnose-imagen/main.go     ← Imagen 3 diagnostic (archived)
│
├── internal/
│   ├── ai/
│   │   ├── granite.go              ← watsonx.ai client, IAM token cache (50 min)
│   │   ├── imagen.go               ← Imagen 3 client (legacy, 429 in prod)
│   │   ├── prompts.go              ← Name gen, persona, competitor radar prompts
│   │   ├── visual_prompts.go       ← SVG logo/moodboard prompts, buildColourContext()
│   │   └── ai_test.go
│   │
│   ├── availability/
│   │   ├── checker.go              ← RDAP + social HEAD probes, 60% gate
│   │   └── availability_test.go
│   │
│   ├── handlers/
│   │   ├── generate.go             ← POST /api/generate (4 names, availability gate)
│   │   ├── visuals.go              ← POST /api/visuals (logos+moodboard+mockup, 2-phase)
│   │   ├── export.go               ← POST /api/export (ZIP stream, SVG-aware)
│   │   ├── export_fallback.go      ← POST /api/export_fallback (ZIP stream, SVG-aware)
│   │   ├── session.go              ← GET/PATCH /api/session/*
│   │   ├── availability.go         ← POST /api/availability
│   │   ├── debug_llm.go            ← Dev debug endpoint
│   │   ├── debug_visuals.go        ← Dev debug endpoint
│   │   └── diagnose.go             ← System diagnose endpoint
│   │
│   ├── models/
│   │   └── types.go                ← All shared Go structs
│   │
│   └── session/
│       ├── store.go                ← Redis CRUD (Upstash)
│       └── session_test.go
│
├── frontend/
│   ├── app/
│   │   ├── layout.tsx              ← Root layout + metadata + favicon
│   │   ├── page.tsx                ← Landing main page
│   │   ├── landing.css             ← Landing page style
│   │   ├── HomeClient.tsx          ← 5-screen state machine + all API calls
│   │   ├── globals.css             ← Brand tokens, animations, nv-dot keyframes
│   │   ├── icon.png                ← App Router favicon (Next.js auto-generates <link>)
│   │   └── build/                   
│   │       └── page.tsx            ← Server component wrapper
│   │
│   ├── components/
│   │   ├── IntakeForm.tsx           ← 8-field form (Q1–Q8) + richness meter
│   │   ├── NameCardComponent.tsx    ← Card: like/pass/choose, score, voice, origin
│   │   ├── NameCardSkeleton.tsx     ← Loading skeleton (pulse animation)
│   │   ├── AvailabilityBadges.tsx   ← Platform badge row (✓/✗/?)
│   │   ├── BrandScoreCard.tsx       ← Memorability/Spellability/Safety/Squatter
│   │   ├── RejectionDialog.tsx      ← Anti-name reasoning dialog + AI question
│   │   ├── StyleDNASlider.tsx       ← Playful↔Premium / Abstract↔Descriptive
│   │   ├── PromptRichnessMeter.tsx  ← Intake richness progress bar
│   │   ├── ErrorBoundary.tsx        ← React error boundary wrapper
│   │   └── VisualIdentityPanel.tsx  ← Step 1 logos / Step 2 moodboard / Step 3 landing
│   │
│   ├── lib/
│   │   ├── types.ts                ← TypeScript interfaces (mirrors Go structs)
│   │   ├── schemas.ts              ← Zod validation schemas
│   │   └── api.ts                  ← Fetch wrappers
│   │
│   ├── public/
│   │   ├── favicon.ico
│   │   ├── nomvox-icon.png         ← 512×512 brand icon
│   │   ├── nomvox-logo.png
│   │   ├── nomvox-logo2.png
│   │   └── nomvox-bg.png
│   │
│   └── vercel.json                 ← Cache headers for favicon assets
│
├── 01-Plans/                       ← Full SDLC plan history (v1–v5 HTML)
├── 00-SkillBuildLab/               ← IBM SkillsBuild certificates
├── Dockerfile                      ← Multi-stage Go build for Fly.io
├── fly.toml                        ← Fly.io config (nomvox-api, iad, shared-cpu-1x)
├── go.mod                          ← module github.com/c-annabel/NomVoxBranding
├── go.sum
├── .env.example
├── DesignDoc.md
├── nomvox-plan.md                  ← This file
├── README.md
└── developer-bob-feedback.md       ← Candid IBM Bob dev feedback
```

**Favicon resolution:** Browser tab icon on deployed Vercel URL served from `app/icon.png` (Next.js App Router convention). `public/favicon.ico` serves as legacy fallback. localhost Edge behavior (shows "N") is browser behavior, not a code bug.

---

## Development Process Summary

### What was genuinely hard
- **watsonx IAM setup** — IAM token flow, WML service association, Project vs Space ID, regional endpoint differences: all underdocumented. Required multiple dead-end debug cycles before first successful call.
- **The quota cascade** — Imagen 3 depleted (permanent 429 `RESOURCE_EXHAUSTED`), watsonx SVG hits 403 `token_quota_reached` when daily budget runs out from name generation calls. Two different errors, two different causes, both required architectural decisions rather than code fixes.
- **Concurrent goroutine rate limiting** — 7 concurrent watsonx calls at 2 req/s → all 429. Added `throttleGranite()` mutex. Adds latency but prevents cascade failure.
- **JSON truncation repair** — LLM occasionally truncates JSON array at token limit. `ExtractJSON()` with repair loop handles partial arrays; still occasionally fails with `unexpected end of JSON input` (logged, auto-retry triggered).
- **Puppeteer on Fly.io** — Chrome binary constraints on shared-cpu containers. Multiple approaches, none reliable. Feature cut.
- **Favicon on localhost** — App Router `app/icon.png` is correct. Browser shows "N" on localhost — browser behavior, not a code bug.

### What worked well from day one
- **RDAP domain availability** — reliable, no API key, no scraping, Verisign-hosted. Always returns clean 200/404.
- **Session memory in Redis** — clean `BrandSession` struct, `PATCH /api/session/react` pattern, full inject on every generate call. The creative-partner loop works exactly as designed.
- **ZIP export** — `archive/zip` streamed with no temp files, `dataURIToBytes()` with RawURLEncoding fallback. Reliable across all asset types.
- **CSS/SVG visual system** — `extractPalette()` bigram parser + palette-responsive component tree. Works at all times, zero cost, zero rate limits.
- **Anti-name reasoning loop** — rejection dialog + AI clarifying question + session inject works cleanly end-to-end.

---

## Sub-Task Status (Final)

| ST | Title | Status | Notes |
|---|---|---|---|
| ST-01 | Environment, scaffold, repo | ✅ Complete | Module: `github.com/c-annabel/NomVoxBranding` |
| ST-02 | Keyword intake form | ✅ Complete | 8 fields, richness meter, inspire-me button |
| ST-03 | LLM core — name gen, scores, voice | ✅ Complete | 4 names/pass, full NameCard JSON |
| ST-04 | Session memory + anti-name reasoning | ✅ Complete | Redis/Upstash, clarifying Q loop |
| ST-05 | Availability engine + 60% gate | ✅ Complete | RDAP + 5 social HEAD probes, zero-pass fallback |
| ST-06 | Select/reject/reproduce state machine | ✅ Complete | Like / Pass / Choose + regenerate |
| ST-07 | Style DNA sliders | ✅ Complete | 2-axis tonal direction |
| ST-08 | Visual identity pipeline | ✅ Complete | CSS/SVG logos + mood board + brand persona |
| ST-09 | Landing page mockup | ✅ Complete | AI HTML + CSS fallback, iframe 75% scale |
| ST-10 | Export ZIP | ✅ Complete | SVG-aware archive stream |
| ST-11 | QA, unit tests, error states | ✅ Complete | `tsc --noEmit` + `go test` pass; 23/23 |
| ST-12 | Deployment + submission | ✅ Complete | Fly.io + Vercel live; video + cert uploaded |

---

## Testing Results

```
go test ./... — 23/23 pass
npx tsc --noEmit — 0 errors
npx next build — Compiled successfully (Turbopack, 1831ms)
```

---

## What Didn't Make It (and Why)

| Feature | Decision | Reason |
|---|---|---|
| Imagen 3 real image generation | ⚠️ Replaced with CSS/SVG | Free-tier credits exhausted; no IBM equivalent |
| Puppeteer handle screenshots | 🚫 Cut | Chrome binary constraints on Fly.io shared-cpu container |
| Vision image upload (palette extract) | 🚫 Deferred | Gemini API credits also depleted |
| PDF brand brief export | 🚫 Deferred | `wkhtmltopdf`/`chromedp` adds significant container size |
| 80% availability gate | ⚠️ Adjusted to 60% | 80% threshold left too few passing names for good UX |

---

## Future Implementation Roadmap

### Stage 1 — Submission sprint (immediate)
- [ ] Record and publish 3-minute demo video (upload to YouTube)
- [ ] Complete IBM SkillsBuild certificate upload
- [ ] Complete submission platform page with GitHub + video links
- [ ] Tag `v1.0.0` on GitHub

### Stage 2 — Post-submission enhancements
- [ ] Real AI image generation: integrate Stability AI or Replicate API (drop-in at CSS fallback insertion points in `VisualIdentityPanel.tsx`)
- [ ] Puppeteer screenshot service: separate lightweight container (not Fly.io shared-cpu) — show screenshots of taken handles
- [ ] Vision image upload: restore Gemini palette extraction when credits restored
- [ ] Editable brand descriptions: short/long in-app edit + per-field regeneration
- [ ] "Download HTML" button for landing page mockup

### Stage 3 — Product evolution
- [ ] User accounts: save and revisit brand exploration sessions
- [ ] Collaborative sessions: share session link with co-founders
- [ ] Full brand brief PDF export (styled HTML → PDF via headless Chrome)
- [ ] Domain registration handoff (deep-link to Namecheap/GoDaddy with prefilled name)
- [ ] Handle availability notification (email when a handle opens up)
- [ ] Multi-language brand naming (Mandarin, Spanish, Arabic romanisation safety checks)

### Stage 4 — Revenue model
- [ ] **Freemium SaaS:** 3 free generations/month; paid plan ($19/mo) for unlimited + real AI images + PDF export
- [ ] **Agency white-label:** Private-label NomVox API + frontend for brand consultancies ($299/mo)
- [ ] **Shopify / Squarespace plugin:** Brand naming embedded in website builder onboarding
- [ ] **IBM co-marketing:** Reference implementation of watsonx.ai for Creative Industries
- [ ] **Accelerator partnerships:** Pre-cohort brand naming tool for YC, Techstars, On Deck cohorts

---

## A Note on the Experience

This project was built in roughly 15 days by one developer, going from a blank canvas to a deployed, full-stack AI application. The process was genuinely educational:

**Learned for the first time:**
- IBM watsonx.ai IAM authentication flow (token endpoint, 50-min cache, refresh cycle)
- Why training data quality is the actual ceiling of LLM output quality — not prompt length
- Go's `sync.WaitGroup` for concurrent goroutines with proper cancellation/timeout
- Next.js App Router hydration model and client/server component boundary rules
- How `archive/zip` streaming works without writing to disk
- Why CSS fallbacks are a design decision, not a consolation prize
- The importance of scoping visual features to what's actually achievable on a free-tier stack

**Had genuine fun:** Watching AI generate coined names like "Verdara" and "Nuvelo" with full origin stories, then seeing the mood board render those brand colours in real time, was exactly the kind of demo-friendly magic the judges are looking for — and the kind of tool a real founder would actually use.

The name NomVox itself came from an AI-assisted creative session. *Nomen* (name) + *vox* (voice). The cosmos names things. Now so do you.

---

*© 2026 c-annabel — Developed with IBM Bob — IBM AI Builders Challenge — Creative Industries*
