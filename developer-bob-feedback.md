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
Bob held the full stack in context simultaneously: Go backend, TypeScript frontend, Redis session layer, deployment config (Fly.io `fly.toml`, Vercel `vercel.json`), and AI API integrations. Switching between layers in a single conversation without losing context is a genuine capability.

### Structured Prompting and AI Reasoning
Bob's help crafting the watsonx.ai system prompts — especially the 5-strategy radical naming system, the Anti-Name Reasoning clarifying question loop, and the Brand Persona card — produced noticeably better LLM output than the developer's own first attempts. The prompt engineering assistance was a clear force multiplier.

---

## ✧ Difficulties and Friction Points

### 1. watsonx.ai Setup — Confusing and Underdocumented

Connecting to IBM watsonx.ai was the single most frustrating part of the project. The challenges:

- **IAM token flow is non-obvious.** The documentation for obtaining a bearer token from `iam.cloud.ibm.com/identity/token` using an API key, caching it for 50 minutes, and refreshing it before expiry is scattered across multiple IBM Cloud docs pages that don't clearly link to each other. Bob helped eventually, but even Bob's first pass used a slightly outdated endpoint format that required a correction round.
- **Project ID vs. Space ID confusion.** The watsonx.ai inference endpoint requires a `project_id` parameter, but the IBM Cloud UI shows both "Project ID" and "Space ID" in different places, and the error messages when the wrong one is used are generic and unhelpful (`400 Bad Request` with no field-level detail).
- **Region endpoint inconsistency.** The `ca-tor` (Toronto) regional endpoint behaves differently from `us-south` in subtle ways — model availability, rate limits, and error message formats differ. This wasn't documented clearly anywhere.
- **WML service association.** The Watson Machine Learning service must be associated with the watsonx project before inference calls work. This step is buried in the UI and the error you get when it's missing does not tell you what's wrong — it just returns a 403. Bob helped diagnose this eventually, but multiple dead-end cycles were spent before the root cause was identified. The error message on IBM's side should say "WML service not associated with this project."

### 2. Gemini / Imagen Charging Obstacle — IBM Has No Direct Equivalent

The original architecture called for Imagen 3 (Google AI Studio) for logo and mood board image generation. This worked in local testing but hit a hard wall in production:

- **Gemini 2.5 Flash Image generation credits depleted permanently.** The free tier had exhausted its prepaid credit allocation, and further image generation requires billing enabled on a Google Cloud project. For a hackathon prototype with a fixed budget, this was a permanent blocker.
- **IBM does not have a comparable image generation model available via watsonx.ai** at the time of development. IBM Granite is excellent for text and code. There is no IBM-native equivalent of Stable Diffusion, DALL-E, or Imagen accessible through the same watsonx interface. This creates an awkward gap for any project in the "Creative Industries" theme that needs actual image output.
- **The fallback path (watsonx SVG generation)** worked but is inherently limited — an LLM generating SVG markup is not image generation, and the results are geometric/abstract rather than design-quality visuals.
- **CSS art as a tertiary fallback** is functional and actually looks polished in the UI, but it is not "AI-generated imagery" in the sense that judges or users expect.

The result: all visual generation in the deployed product uses CSS/SVG fallbacks, not real AI imagery. This is an honest limitation that needed to be designed around rather than hidden.

### 3. HTTP 429 — RESOURCE_EXHAUSTED at Runtime

Even text generation via watsonx.ai hit rate limits frequently during development:

- The watsonx.ai free tier enforces a **2 requests/second** limit. NomVox's visual generation phase sends 7 concurrent calls (3 logos + 4 mood board tiles + 1 persona + 1 mockup). All 7 would 429 simultaneously.
- The error response is `429 RESOURCE_EXHAUSTED` — the message is clear, but the fix (sequential calling with delays, exponential backoff, or upgrading to a paid tier) creates real architectural constraints for a demo product.
- Bob's suggested fix (sequential generation with `time.Sleep` between calls) worked but added 15-25 seconds to the visual generation phase, which significantly hurts the demo experience.
- Bobcoins were spent debugging 429s that turned out to be architectural (too many concurrent calls) rather than code bugs — the same issue diagnosed multiple ways before the root cause was accepted.

### 4. Training Data Limitations — Some Problems Were Genuinely Unsolvable

There were several technical questions where Bob consistently produced answers that didn't work, and it became apparent the training data was either outdated or insufficient for the specific domain:

- **Puppeteer headless screenshot on Fly.io** — multiple attempts, multiple approaches (custom buildpack, Chrome binary path, `--no-sandbox` flags), none worked reliably in the Fly.io container environment. Bob tried many configurations confidently, but each failed in slightly different ways. This feature (taking live screenshots of taken social handles) was eventually cut from scope.
- **Fly.io machine autosuspend + cold-start latency** — understanding the exact behavior of `auto_stop_machines = "stop"` and how it interacts with Upstash Redis connection pool teardown required reading Fly.io community forums, not Bob's answers, which were generic.
- **Next.js App Router `app/icon.png` favicon** — nearly 40 bobcoins were spent over multiple sessions trying to get the favicon to display correctly in the browser tab. Bob tried: `public/favicon.ico`, `app/icon.png`, `metadata.icons` in layout.tsx, `vercel.json` headers, and combinations thereof. In the final session, Bob declared the fix correct — "it should work on Vercel after deploy." It did not visibly change in local development (which Bob later acknowledged is expected browser behavior for localhost). The uncertainty was never fully resolved — the issue was closed by reasoning that localhost behavior is not representative, not by confirming it worked on Vercel. Forty bobcoins for a favicon is a real cost.

### 5. Outdated Instructions

Several specific areas where Bob's answers reflected older versions of tools:

- **Next.js 13 vs. App Router patterns** — early sessions produced `getServerSideProps` patterns that no longer apply in Next.js 14+/App Router. Had to explicitly ask Bob to use App Router conventions.
- **Tailwind CSS v4 syntax** — Bob defaulted to Tailwind v3 `@apply` and config-file patterns; Tailwind v4 uses a different CSS-variable-based approach. Several rounds of correction were needed.
- **watsonx.ai API schema** — the `parameters` field naming (`max_new_tokens` vs. `max_tokens`) changed between versions and Bob's initial code used the older format, producing silent failures.
- **Fly.io CLI flags** — `fly secrets set` flag syntax in some Bob suggestions used deprecated flag formats.

### 6. Unintended "Improvements" Without Consultation

On multiple occasions, Bob changed functioning code while fixing something minor:

- A request to adjust the color of a footer text element resulted in a refactor of the footer component's layout, removing a working gradient border that had taken time to tune.
- A request to fix the iframe scrollbar (a one-line CSS fix) produced a diff that also rewrote the entire `hasMockup` validation logic — the new logic was correct but the change was broader than requested and introduced a brief regression on a different field.
- A request to update a label string triggered a re-evaluation of the entire component's state management, adding an unnecessary `useEffect`.

The pattern: Bob reads the surrounding context and "improves" what it sees, even when the instruction was surgical. For a solo developer on a deadline, unexpected refactors are high-risk.

### 7. Bobcoin Consumption Rate

- **Simple questions consumed surprising coin amounts** — asking "what does this error mean?" for a single-line error output, or "is this import correct?", would sometimes trigger long exploratory reasoning chains that consumed significant coin before reaching a simple answer.
- **Redundant reviews** — Bob would sometimes re-read and re-summarize files it had already seen in the same session, spending coin on recap rather than action. Prefacing requests with "don't re-read the file, just answer" sometimes helped but wasn't always respected.
- **Thinking time** — on complex requests (multi-file diffs, architecture decisions), the "thinking" phase before output was sometimes 60+ seconds with no visible progress indicator from the developer's perspective — just waiting. This made it difficult to know whether Bob was working or had stalled.
- **Long response time on bug hits** — when Bob encountered a genuinely novel bug (not in training data), response time increased significantly as it explored approaches. These were also the sessions most likely to produce incorrect first attempts, meaning multiple expensive rounds.
- **Token limits on large files** — `VisualIdentityPanel.tsx` at 1,100+ lines frequently required splitting into range reads (e.g., lines 1-200, 200-450, etc.) because a single full-file read would hit context or cost limits. The overhead of re-assembling context across multiple reads added up over many sessions.

### 8. The Favicon Saga — A Case Study

The favicon issue deserves its own entry as a cautionary tale about diminishing returns:

Over multiple sessions, Bob attempted: writing `public/favicon.ico`, adding `metadata.icons` to `layout.tsx`, creating `app/icon.png` (App Router convention), adding `vercel.json` cache headers, and explaining that localhost shows a letter "N" which is normal browser behavior. Each attempt was confident, each consumed bobcoins, and the final verdict was "it's correct, it will work on Vercel."

The browser tab on localhost still shows the Edge-generated "N" for localhost. On Vercel, the correct icon may or may not be showing (unconfirmed at time of writing). **~40 bobcoins** were spent. The correct answer, identified in hindsight: `app/icon.png` is the right approach for Next.js App Router, and localhost Edge behavior is irrelevant. That insight could have been delivered in one response at the start.

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

1. **Add a "surgical edit" mode** — when a user says "change only X", Bob should scope its diff strictly to X and not refactor adjacent code.
2. **Improve watsonx.ai setup documentation** — a single "getting started" guide covering IAM token flow, WML service association, project ID location, and regional endpoint differences would save every watsonx-based project significant time.
3. **Acknowledge IBM's image generation gap honestly** — rather than suggesting Gemini/Imagen alternatives that may have cost barriers, Bob should proactively offer the SVG/CSS fallback path as a first-class option with honest tradeoffs stated upfront.
4. **Bobcoin cost previews** — show an estimated cost before executing long reasoning chains. Let the developer decide if the question is worth it.
5. **Reduce recap overhead** — within a session, don't re-summarize files already loaded. Trust the context.
6. **Version-pin recommendations** — when suggesting API patterns, state which version of the library/framework they apply to. "This is the Next.js 14 App Router pattern" vs. "This is the Pages Router pattern" would have saved multiple correction cycles.

---

*Written by c-annabel — NomVox developer — July 2026*  
*Primary development tool: IBM Bob*  
*© 2026 c-annabel. All rights reserved.*
