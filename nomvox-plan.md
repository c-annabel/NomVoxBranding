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
**AI Core (reasoning):** IBM watsonx.ai chat completion API. The model is selected by endpoint: Llama 3.3 70B Instruct on `ca-tor.ml.cloud.ibm.com`, IBM Granite 3 8B Instruct on `us-south.ml.cloud.ibm.com`. Covers name generation, taglines, origin stories, brand scores, persona, voice samples, competitor radar, SVG logo markup, SVG mood board tiles, and landing page HTML.  
**AI Visual:** Google Gemini 2.5 Flash Image for raster logo concepts and mood board textures, with a watsonx SVG tier and a deterministic CSS brand system behind it. See "Image Generation Decision".  
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

## Image Generation Decision: Imagen 3 to Gemini 2.5 Flash Image

### What was planned
Imagen 3 via Google AI Studio for logo concepts (3 formats) and mood board tiles. Gemini 2.0 Flash for optional vision analysis, extracting a palette from an uploaded reference image.

### What happened
Imagen 3 worked in local testing. Before the production demo, Google AI Studio prepaid credits were exhausted and every call returned HTTP 429 `RESOURCE_EXHAUSTED`. Restoring it required enabling billing on Google Cloud.

Concurrently watsonx.ai returned 429 on the SVG fallback, because the free tier enforces 2 requests per second and the seven concurrent visual calls all fired at once.

### What was built instead
Two changes, in sequence.

First, a fully dynamic **CSS/SVG brand system**. `extractPalette()` parses the user's `color_mood` field (using bigram matching: "bright orange", "sky blue", etc.) and derives `bg`, `accent`, `accent2`, `text` tokens. Every visual component renders exclusively from these tokens:

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
Gemini image generation → HTTP 429 RESOURCE_EXHAUSTED
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

## The Fallback Trap

The most instructive bug of the build was invisible until the day it mattered.

The SVG fallback path had three defects that had never surfaced, because while Gemini had credits the code took the first branch every time and the fallback never executed. When credits ran out, all three appeared at once:

1. The generated SVG had no `xmlns` attribute. An SVG loaded through an `img` tag is parsed as a standalone XML document and will not render without it. Inline SVG in HTML does not need it, which is what makes this easy to miss.
2. The markdown fence stripper searched for the first occurrence of a fence in a loop. After removing an opening fence it matched the closing one and discarded everything before it, returning an empty string.
3. Truncated output had a closing tag appended blindly, producing malformed XML that the browser rejected without reporting an error.

Redis had the same shape of problem. Session writes had been failing from the first deploy, but the handler logged and continued, so every request still returned 200. Nothing surfaced until Upstash sent an inactivity notice for a database that had never recorded a single command.

**The lesson is not about SVG or Redis.** An untested error path is not a fallback, it is a second bug waiting for the first one. Two changes came out of this: generated assets are validated before use rather than assumed well formed, and degraded states are logged at startup rather than only at the point of failure.

---

## What Prompt Debugging Taught

Four rules came out of getting usable output from the image and HTML models.

**Resolve colours to hex before prompting.** Passing free text such as "electric blue, neon yellow, dark bg" makes the model guess, and a background instruction gets read as a mark colour. `logoPalette()` now parses the mood string into explicit hex values, preserving the order the user wrote them.

**Put hard constraints first and repeat them.** Image models weight early tokens most heavily. A "no text" rule placed at the end of a long style description was ignored, and produced app icons containing the brand's initials. The same rule stated first and restated last holds.

**Never ask an image model for a grid.** A prompt requesting a four-panel mood board returns one collage image with visible gutters, not four images.

**Check the prompt before blaming the model.** The landing page kept rendering a circle behind the headline. The prompt was asking for a decorative SVG accent shape, and the model was doing exactly as told.

---

### Then image generation was restored

Image generation now runs on **Gemini 2.5 Flash Image** rather than Imagen 3. It has a free tier, uses the standard `generateContent` endpoint with `responseModalities` set to include `IMAGE`, and does not require prepaid credits. The client type and filename kept the `imagen` name to avoid breaking the debug endpoint contract.

### Why watsonx does not do this job

IBM watsonx.ai does not currently offer a native raster image generation model comparable to Gemini, DALL-E, or Stable Diffusion. It can be prompted to emit SVG markup, which is code rather than pixels, and that is exactly how the second tier of the fallback chain works. For photographic texture, an image model is required.

### Where the mood board landed

The mood board is a hybrid, and the split is deliberate.

Panels that must be exact are rendered deterministically from the parsed palette: the colour swatches, the typography specimen, and the three logo lockups. An image model cannot render a specific hex value, set type in Space Grotesk, or place the real logo. Only the two texture panels come from Gemini, because texture and atmosphere are what an image model is genuinely good at.

The original implementation asked one Gemini call for a four-panel grid. That returns a single collage image complete with white gutters, not four images. It now makes two separate calls, each requesting one unbroken frame.

### Future path

If IBM watsonx.ai releases a native image generation model, the Gemini tier can be swapped for it without touching the frontend, because every visual component already reads from the same palette tokens and data URI contract.

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
    ├──► IBM watsonx.ai (Llama 3.3 70B on ca-tor, Granite 3 8B on us-south)
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
│   └── diagnose-imagen/main.go     ← image generation diagnostic (archived)
│
├── internal/
│   ├── ai/
│   │   ├── granite.go              ← watsonx.ai client, model selected by endpoint,
│   │   │                              IAM token cache (50 min)
│   │   ├── imagen.go               ← Gemini image client (legacy filename)
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
- **The quota cascade.** Imagen 3 depleted with a 429 `RESOURCE_EXHAUSTED`, while watsonx SVG hit 403 `token_quota_reached` once the daily budget was spent on name generation. Two different errors, two different causes, both needing architectural decisions rather than code fixes.
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

### Hardening pass

| Area | Before | After |
|---|---|---|
| Visuals response size | ~12.7 MB | ~5.9 MB |
| Visuals generation time | ~40 s | ~29 s |
| Session writes | Failing silently since first deploy | Verified at startup, succeeding |
| SVG fallback | Rendered as broken images | Validated, renders or cleanly defers |
| Landing page prompt | Overflowed context window when a logo was attached | Placeholder substitution, well within budget |
| Mood board images | One collage per call, white gutters | Two calls, one clean frame each |

---

## What Didn't Make It (and Why)

| Feature | Decision | Reason |
|---|---|---|
| Imagen 3 | ⚠️ Replaced | Prepaid credits exhausted; migrated to Gemini 2.5 Flash Image, which has a free tier |
| Puppeteer handle screenshots | 🚫 Cut | Chrome binary constraints on Fly.io shared-cpu container |
| Vision image upload (palette extract) | 🚫 Deferred | Implemented in the client but not wired into the UI flow |
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
- [ ] Vision image upload: wire the existing Gemini palette extraction into the intake flow
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
