# Developer's Feedback on IBM Bob

> Honest, constructive observations from a solo developer who built a full-stack AI web application  
> (NomVox — IBM AI Builders Challenge, July 2026) using IBM Bob as the **primary development tool**  
> from blank canvas to deployed production system.

---

## Overview

IBM Bob was used end-to-end across the full SDLC: architecture planning, prompt engineering, Go backend scaffolding, Next.js frontend development, deployment configuration, debugging, and documentation. This feedback is written with genuine care — the goal is to help IBM and the Bob team make a better product, not to complain. The good was real. The friction was real too.

---

## ✦ What Bob Did Well

### Documentation Friendliness
Bob is exceptional at producing structured, well-formatted documentation. README files, SDLC plans, architecture diagrams as HTML, API reference tables, `.env.example` templates — all came out clean, professional, and immediately usable. This alone saved multiple days of writing and formatting work. The quality was high enough to publish directly.

### Debugging Autonomy and Systematic Diagnosis
When presented with a bug and the relevant file contents, Bob's diagnostic process was methodical and impressively thorough. It would trace execution paths, identify root causes across multiple files simultaneously, and explain the reasoning clearly. For issues like JSON truncation in LLM responses, Redis TTL edge cases, and base64 encoding of SVG assets, Bob not only fixed the immediate issue but explained *why* it occurred — which helped build real understanding of the codebase.

### Teaching Through Doing
Building NomVox with Bob was a genuine learning experience. The developer came in knowing the rough shape of the architecture and left with a working understanding of:
- How IBM watsonx.ai IAM token authentication works, including token caching and refresh cycles
- Why Go's `sync.WaitGroup` with concurrent goroutines is the right pattern for parallel availability probing
- How Next.js App Router handles metadata, hydration, and server/client component boundaries
- The difference between `sandbox="allow-same-origin"` and full sandbox for iframes
- How `archive/zip` streaming works in Go without writing to disk
- The actual mechanics of RDAP domain checking vs. fragile DNS scraping

Bob made these concepts accessible and practical — not abstract.

### Multi-Stack Fluency
Bob held the full stack in context simultaneously: Go backend, TypeScript frontend, Redis session layer, deployment config, and AI API integrations. Switching between layers in a single conversation without losing context is a genuine capability.

### Structured Prompting and AI Reasoning
Bob's help crafting the watsonx.ai system prompts — especially the 5-strategy radical naming system, the Anti-Name Reasoning clarifying question loop, and the Brand Persona card — produced noticeably better LLM output than the developer's own first attempts. The prompt engineering assistance was a clear force multiplier.

---

## ✧ Difficulties and Friction Points

### 1. watsonx.ai Setup — Confusing and Underdocumented

Connecting to IBM watsonx.ai was the single most frustrating part of the project. The challenges:

- **IAM token flow is non-obvious.** The documentation for obtaining a bearer token from `iam.cloud.ibm.com/identity/token` using an API key, caching it for 50 minutes, and refreshing it before expiry is scattered across multiple IBM Cloud docs pages that don't clearly link to each other. Bob helped eventually, but even Bob's first pass used a slightly outdated endpoint format that required a correction round.
- **Project ID vs. Space ID confusion.** The watsonx.ai inference endpoint requires a `project_id` parameter, but the IBM Cloud UI shows both "Project ID" and "Space ID" in different places, and the error messages when the wrong one is used are generic and unhelpful (`400 Bad Request` with no field-level detail).
- **Region endpoint inconsistency.** The `ca-tor` (Toronto) regional endpoint behaves differently from `us-south` in subtle ways — model availability, rate limits, and error message formats differ. This wasn't documented clearly anywhere.
- **WML service association.** The Watson Machine Learning service must be associated with the watsonx project before inference calls work. This step is buried in the UI and the error you get when it's missing does not tell you what's wrong — it just returns a 403. Bob helped diagnose this eventually, but multiple dead-end cycles were spent before the root cause was identified.

### 2. Gemini / Imagen Charging Obstacle — IBM Has No Direct Equivalent

The original architecture called for Imagen 3 (Google AI Studio) for logo and mood board image generation. This worked in local testing but hit a permanent wall in production:

- **Google AI Studio prepaid credits exhausted permanently.** The free tier had depleted its prepaid credit allocation, and further image generation requires billing enabled on Google Cloud. For a hackathon prototype with a fixed budget, this was a permanent blocker.
- **IBM does not have a comparable image generation model available via watsonx.ai** at the time of development. IBM Granite is excellent for text and code. There is no IBM-native equivalent of Stable Diffusion, DALL-E, or Imagen accessible through the same watsonx interface.

The actual error received from Gemini/Imagen in production:
```
gemini-image: status 429: {
  "error": {
    "code": 429,
    "message": "Your prepayment credits are depleted. Please go to AI Studio
    at https://ai.studio/projects to manage your project and billing.",
    "status": "RESOURCE_EXHAUSTED"
  }
}
```
This appeared simultaneously for every visual call: logo profile, logo app, logo business, and all four mood board tiles. **The entire visual generation pipeline failed at once.**

### 3. The SVG Fallback Path and watsonx Token Quota

When Imagen/Gemini failed, the code attempted to fall back to watsonx SVG generation via Granite. This also failed — with a different, more serious error:

```
granite: chat status 403: {
  "errors": [{
    "code": "token_quota_reached",
    "message": "Request of 1 token(s) from quota was rejected",
    "more_info": "https://cloud.ibm.com/apidocs/watsonx-ai#text-chat"
  }],
  "status_code": 403
}
```

This `token_quota_reached` (403) error appeared across **all nine visual generation tasks simultaneously**:
- `visuals: svg logo profile` → 403
- `visuals: svg logo app` → 403
- `visuals: svg logo business` → 403
- `visuals: svg mood tile 0` → 403
- `visuals: svg mood tile 1` → 403
- `visuals: svg mood tile 2` → 403
- `visuals: svg mood tile 3` → 403
- `visuals: mockup html` → 403
- `visuals: persona` → 403

The watsonx free tier enforces a **2 requests/second** rate limit, and NomVox's visual phase was making 7–9 concurrent calls simultaneously. Even with the `throttleGranite()` serialization lock added to enforce 600ms gaps between calls, when the *daily token quota* for the free account was reached, every call failed with 403 regardless of timing.

**Root cause:** The free-tier token quota covers both text generation (name generation, session memory calls) and visual generation. A long name-generation session — multiple regenerate cycles, several users — could exhaust the daily quota before any visual generation was attempted.

**How the code handles it:** The Go `visuals.go` handler tolerates individual failures gracefully. Each goroutine logs the error and returns an empty result. The `visualsResponse` struct fields default to empty strings and nil arrays. The frontend `VisualIdentityPanel.tsx` then renders the **CSS art system** as the guaranteed primary display path — a fully palette-responsive visual design built from `extractPalette()` on the user's `color_mood` field. The user sees a branded, professional result regardless of API state.

**The CSS art system is not a degraded experience.** It is a deliberately designed fallback that works at all times, with zero API calls, with zero cost, and with zero rate limits. It was a design decision driven by necessity that produced a more reliable product.

### 4. HTTP 429 — RESOURCE_EXHAUSTED at the Rate Level

Even within-quota, the watsonx.ai free tier enforced a 2 requests/second limit. NomVox sends 7 concurrent calls (3 logos + 4 mood board tiles) when all generate simultaneously. The fix required:

1. A `graniteCallMu` mutex and `graniteLastCall` timestamp to serialize *when* each Granite call starts
2. A `throttleGranite()` function that enforces a minimum 600ms gap between call starts
3. Accepting that the visual generation phase now takes 5–8 seconds sequentially rather than 1–2 seconds concurrently

This added significant latency to the demo but prevented the cascade of 429 errors that made the visual phase appear broken.

### 5. Training Data Limitations — Genuinely Unsolvable Problems

Several areas where Bob's answers were consistently off or outdated:

- **Puppeteer headless screenshots on Fly.io** — multiple approaches, multiple Chrome binary path configurations, none worked reliably in the container environment. Each attempt was confident; each failed differently. This feature was cut.
- **Fly.io cold-start + Upstash Redis** — understanding the exact interaction between `auto_stop_machines = "stop"` and the Redis connection pool required reading Fly.io community forums, not Bob's answers.
- **Next.js App Router favicon** — approximately 40 bobcoins spent across multiple sessions. Bob tried `public/favicon.ico`, `app/icon.png`, `metadata.icons` in layout.tsx, and `vercel.json` headers. Final verdict: "it's correct, it will work on Vercel." The localhost "N" is browser behavior. This was technically accurate but took far too long to arrive at.

### 6. Outdated Instructions

- **Next.js 13 vs App Router patterns** — early sessions produced `getServerSideProps` patterns that no longer apply. Had to explicitly ask for App Router conventions.
- **Tailwind CSS v4 syntax** — Bob defaulted to v3 `@apply` and config-file patterns. Multiple correction rounds needed.
- **watsonx.ai API schema** — `max_new_tokens` vs `max_tokens` field naming changed between versions; Bob's initial code used the older format, producing silent failures.
- **Fly.io CLI flags** — some suggested flag formats were deprecated.

### 7. Unintended "Improvements" Without Consultation

On multiple occasions, Bob changed functioning code while fixing something minor:

- A request to adjust footer text color resulted in a refactor of the footer component's layout, removing a working gradient border.
- A request to fix the iframe scrollbar produced a diff that also rewrote the `hasMockup` validation logic — the new logic was correct but the scope was broader than requested.
- A request to update a label string triggered a re-evaluation of the component's state management, adding an unnecessary `useEffect`.

The pattern: Bob reads surrounding context and "improves" what it sees, even when the instruction was surgical. For a solo developer on a deadline, unexpected refactors are high-risk.

### 8. Bobcoin Consumption Rate

- **Simple questions consumed disproportionate coin.** Asking "what does this error mean?" for a single-line error output sometimes triggered long exploratory reasoning chains before a simple answer.
- **Redundant session recaps.** Bob would sometimes re-read and re-summarize files it had already seen in the same session, spending coin on recap rather than action.
- **Long response time on novel bugs.** When Bob encountered a genuinely novel bug (not in training data), response time increased significantly and the first attempt was often incorrect.
- **Token limits on large files.** `VisualIdentityPanel.tsx` at 1,100+ lines frequently required splitting into range reads (lines 1-200, 200-450, etc.) because a single full-file read would hit context limits. The overhead of re-assembling context across multiple reads added up.

---

## Summary Table

| Area | Rating | Notes |
|---|---|---|
| Documentation generation | ⭐⭐⭐⭐⭐ | Exceptional — publish-ready output |
| Debugging diagnosis | ⭐⭐⭐⭐ | Systematic and educational when it works |
| Multi-stack fluency | ⭐⭐⭐⭐ | Holds full stack in context well |
| Prompt engineering help | ⭐⭐⭐⭐⭐ | Major force multiplier |
| watsonx setup guidance | ⭐⭐ | Gaps in IAM, WML association, region config |
| Handling novel bugs | ⭐⭐ | Expensive and slow; training data limits visible |
| Following surgical instructions | ⭐⭐⭐ | Tendency to over-improve beyond scope |
| Bobcoin efficiency | ⭐⭐ | Simple Q&A is disproportionately expensive |
| Up-to-date knowledge | ⭐⭐⭐ | Some outdated patterns (Next.js, Tailwind, Fly) |
| Image generation guidance | ⭐⭐ | IBM has no equivalent tool; gap left unaddressed |

---

## Suggestions for IBM Bob Team

1. **Add a "surgical edit" mode** — when a user says "change only X", scope the diff strictly to X.
2. **Improve watsonx.ai setup documentation** — a single getting-started guide covering IAM token flow, WML service association, project ID location, and regional endpoint differences would save every project significant time.
3. **Acknowledge IBM's image generation gap honestly** — rather than suggesting Gemini/Imagen alternatives that may have cost barriers, Bob should proactively offer the SVG/CSS fallback path as a first-class option with honest tradeoffs stated upfront.
4. **Bobcoin cost previews** — show an estimated cost before executing long reasoning chains.
5. **Reduce recap overhead** — within a session, don't re-summarize files already loaded.
6. **Version-pin recommendations** — state which framework version a pattern applies to.
7. **Token quota documentation** — explicitly document the difference between the 2 req/s rate limit (HTTP 429) and the daily token quota (HTTP 403 `token_quota_reached`). These are two different limits that produce different errors and require different fixes. Many developers will confuse them.

---

*Written by c-annabel — NomVox developer — July 2026*  
*Primary development tool: IBM Bob*  
*© 2026 c-annabel. All rights reserved.*
