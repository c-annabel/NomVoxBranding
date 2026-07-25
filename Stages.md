
---

# Build Stages & Completion Checklist

| Stage | Title | Description | Status |
|:---:|---|---|:---:|
| **ST-01** | Environment, scaffold, repo | Go module scaffold, chi router, `/api/ping`, Redis session store, `.env.example`, Dockerfile, GitHub Actions CI, `README.md` | ✅ Complete |
| **ST-02** | Keyword intake form | 8-field intake form, richness meter, Inspire Me button, brand CSS tokens, star-field background | ✅ Complete |
| **ST-03** | LLM core — name gen, scores, voice | IBM watsonx.ai integration — IAM token exchange, `/ml/v1/text/chat`, `ExtractJSON` with truncation repair, file logging | ✅ Complete |
| **ST-04** | Session memory + anti-name reasoning | Redis session memory (`BrandSession`), `POST /api/session`, `PATCH /api/session/react` (6 action types), Anti-Name Reasoning, `RejectionDialog` | ✅ Complete |
| **ST-05** | Availability engine + 60% gate | Availability engine — RDAP domain, 6-platform parallel goroutines, weighted 60% gate (lowered from 80%), zero-pass fallback (top-2 partials + ⚠ banner), Competitor Radar | ✅ Complete |
| **ST-06** | Select/reject/reproduce state machine | Full state machine UI — Like / Reject / Select, `StyleDNASlider`, liked names summary banner | ✅ Complete |
| **ST-07** | Style DNA sliders | `NameCardComponent` — 2-row 7-col table layout, `rejectedReason` banner, column head colours | ✅ Complete |
| **ST-08** | Visual identity pipeline | `/api/visuals` — 3-tier fallback chain (Imagen 3 → watsonx SVG → CSS art), `throttleGranite()` mutex, `graniteCallMu`, `Base64ToDataURI` RawURLEncoding fix | ✅ Complete |
| **ST-09** | Landing page mockup | Landing-page mockup via Granite — `BuildMockupSystemPrompt/UserPrompt`, iframe 75% scale, 100vh strip, `hasMockup` validator | ✅ Complete |
| **ST-10** | Export ZIP | `/api/export` — ZIP with `brand-brief.json`, `landing-page.html`, `mood-board/`, `logos/`, `README.txt`. `dataURIToBytes` RawURLEncoding fix. Selected logo copy. | ✅ Complete |
| **ST-11** | Token issue | CSS/SVG visual system — `extractPalette()` bigram parser, palette-responsive logo/moodboard/landing components, mood board always 4 tiles (pad with CSS if AI returns fewer) | ✅ Complete |
| **ST-12** | QA | Radical name-invention prompt — 5 strategies (syllable forge, phonaesthetics, neologism splice, void-word, ancient root mutation) | ✅ Complete |
| **ST-13** | Unit tests, error states | Unit tests — `go test ./...` 23/23 pass, `tsc --noEmit` 0 errors, `next build` clean | ✅ Complete |
| **ST-14** | Deployment | Deploy to Fly.io (Go API, `nomvox-api`, region `iad`) + Vercel (Next.js frontend, auto-deploy from GitHub `main`) | ✅ Complete |
| **ST-15** | Submission | Demo video, IBM SkillsBuild certificates, submission page | ✅ Complete |