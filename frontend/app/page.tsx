"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import "./landing.css";

// Update this once with your published YouTube video ID
// (the part after youtu.be/ or ?v= in the URL).
const DEMO_VIDEO_ID = "Kt73AI9WI8o";
const DEMO_VIDEO_URL = `https://youtu.be/${DEMO_VIDEO_ID}`;

export default function LandingPage() {
  const starsRef = useRef<SVGSVGElement>(null);
  const [playing, setPlaying] = useState(false);

  // Generate the starfield once on mount — same approach as the standalone
  // HTML build, just wrapped in a React effect instead of an inline <script>.
  useEffect(() => {
    const svg = starsRef.current;
    if (!svg) return;
    const w = window.innerWidth;
    const h = Math.max(document.body.scrollHeight, window.innerHeight);
    svg.setAttribute("viewBox", `0 0 ${w} ${h}`);
    svg.setAttribute("preserveAspectRatio", "none");
    svg.style.width = "100%";
    svg.style.height = `${h}px`;

    const count = Math.floor((w * h) / 10000);
    const ns = "http://www.w3.org/2000/svg";
    const frag = document.createDocumentFragment();
    for (let i = 0; i < count; i++) {
      const c = document.createElementNS(ns, "circle");
      const r = Math.random() < 0.85 ? Math.random() * 0.9 + 0.3 : Math.random() * 1.4 + 1;
      c.setAttribute("cx", String(Math.random() * w));
      c.setAttribute("cy", String(Math.random() * h));
      c.setAttribute("r", String(r));
      c.setAttribute("fill", Math.random() < 0.12 ? "#a78bfa" : "#e6e8f0");
      c.setAttribute("opacity", (Math.random() * 0.6 + 0.25).toFixed(2));
      frag.appendChild(c);
    }
    svg.appendChild(frag);
  }, []);

  return (
    <div className="landing-root">
      <div className="nebula nebula-a" />
      <div className="nebula nebula-b" />
      <svg ref={starsRef} id="stars" aria-hidden="true" />

      <main>
        <nav>
          <div className="wrap">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="nav-logo" src="/nomvox-logo.png" alt="NomVox" />
            <Link className="btn-nav" href="/build">Start</Link>
          </div>
        </nav>

        <section className="hero">
          <div className="wrap">
            <div className="eyebrow">Prototype &middot; IBM SkillsBuild AI Builders Challenge &middot; Creative Industries</div>

            <div className="hero-logo-wrap">
              <div className="hero-glow" />
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img className="hero-logo" src="/nomvox-logo2.png" alt="NomVox — Born from the void" />
            </div>

            <h1 className="headline">Give it an idea. <span className="accent">Get back a brand.</span></h1>
            <p className="subhead">NomVox turns a one-line business idea into a coined name, checked
              domains and handles, logo directions, a mood board, and a landing page &mdash;
              through one conversation that learns what you like as you go.</p>

            <div className="cta-row">
              <Link className="btn-primary" href="/build">Start naming your brand &rarr;</Link>
              <a className="btn-secondary" href="https://github.com/c-annabel/NomVoxBranding" target="_blank" rel="noopener">
                <svg viewBox="0 0 16 16" fill="currentColor"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0016 8c0-4.42-3.58-8-8-8z"/></svg>
                View source
              </a>
              <a className="btn-watch btn-watch-sm" href="#demo">
                <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5v14l11-7z" /></svg>
                Watch the demo
              </a>
            </div>
          </div>
        </section>

        <section className="origin">
          <div className="wrap">
            <div className="section-label">Why NomVox exists</div>
            <div className="origin-row">
              <svg className="constellation" width="86" height="86" viewBox="0 0 86 86" aria-hidden="true">
                <g stroke="rgba(139,92,246,0.55)" strokeWidth="1">
                  <line x1="10" y1="60" x2="30" y2="20"/><line x1="30" y1="20" x2="52" y2="34"/>
                  <line x1="52" y1="34" x2="72" y2="12"/><line x1="30" y1="20" x2="18" y2="8"/>
                  <line x1="52" y1="34" x2="60" y2="58"/>
                </g>
                <g fill="#E6E8F0">
                  <circle cx="10" cy="60" r="2.2"/><circle cx="30" cy="20" r="2.6"/>
                  <circle cx="18" cy="8" r="1.6"/><circle cx="52" cy="34" r="3"/>
                  <circle cx="72" cy="12" r="2"/><circle cx="60" cy="58" r="1.8"/>
                </g>
              </svg>
              <p><strong>Naming a brand usually means twelve open tabs</strong> &mdash; a
                thesaurus, a domain checker, five social apps, a logo tool, and a blank
                design file, all disconnected. NomVox collapses that into one guided
                conversation: describe the idea once, and the name, the handles, and the
                visuals all come from that same thread &mdash; getting sharper every time
                you say yes or no to a direction.</p>
            </div>
          </div>
        </section>

        <section className="problems">
          <div className="wrap">
            <div className="section-label">What it solves</div>
            <h2 className="section-title">Four places brand naming usually stalls.</h2>

            <div className="grid-2">
              <div className="ps-card">
                <span className="ps-tag problem">Problem</span>
                <p className="problem-text">Naming paralysis.</p>
                <p className="ps-solution">Coins genuinely new words &mdash; never real
                  dictionary terms &mdash; so you're not fighting over a name someone
                  already owns.</p>
              </div>
              <div className="ps-card">
                <span className="ps-tag problem">Problem</span>
                <p className="problem-text">A name you love, a domain you can&apos;t get.</p>
                <p className="ps-solution">Every suggestion is checked live against domain
                  registries and five social platforms before it&apos;s shown to you.</p>
              </div>
              <div className="ps-card">
                <span className="ps-tag problem">Problem</span>
                <p className="problem-text">Branding needs a designer you don&apos;t have.</p>
                <p className="ps-solution">Logo directions, a mood board, and a landing page
                  mockup generate straight from your brand&apos;s palette and personality.</p>
              </div>
              <div className="ps-card">
                <span className="ps-tag problem">Problem</span>
                <p className="problem-text">AI tools forget what you just told them.</p>
                <p className="ps-solution">Every like, pass, and note is remembered for the
                  rest of the session, so the next batch actually reflects your taste.</p>
              </div>
            </div>
          </div>
        </section>

        <section className="how">
          <div className="wrap">
            <div className="section-label">How it works</div>
            <h2 className="section-title">Four steps, one conversation.</h2>

            <div className="steps">
              <div className="step">
                <div className="step-num">01</div>
                <div className="step-title">Describe your idea</div>
                <div className="step-desc">What it does, who it&apos;s for, how it should feel.</div>
              </div>
              <div className="step">
                <div className="step-num">02</div>
                <div className="step-title">Choose a name</div>
                <div className="step-desc">Coined names arrive with brand scores and live
                  availability.</div>
              </div>
              <div className="step">
                <div className="step-num">03</div>
                <div className="step-title">Build the identity</div>
                <div className="step-desc">Pick a logo direction, get a matching mood board
                  and landing page.</div>
              </div>
              <div className="step">
                <div className="step-num">04</div>
                <div className="step-title">Export the pack</div>
                <div className="step-desc">Everything downloads as one ready-to-use brand
                  kit.</div>
              </div>
            </div>
          </div>
        </section>

        <section className="holds-up">
          <div className="wrap">
            <div className="section-label">What&apos;s underneath</div>
            <h2 className="section-title">A brand platform built to hold up, not just look good.</h2>

            <div className="grid-2" style={{ marginTop: 22 }}>
              <div className="ps-card">
                <p className="problem-text" style={{ marginBottom: 6 }}>Actually original</p>
                <p className="ps-solution">Every name is coined, not pulled from a
                  dictionary or a stock list &mdash; the naming engine invents vocabulary
                  and explains the etymology behind it.</p>
              </div>
              <div className="ps-card">
                <p className="problem-text" style={{ marginBottom: 6 }}>Real technical depth</p>
                <p className="ps-solution">Live domain and social-handle checks run in
                  parallel against real registries, and session memory carries every
                  reaction into the next AI call &mdash; not a single scripted demo path.</p>
              </div>
              <div className="ps-card">
                <p className="problem-text" style={{ marginBottom: 6 }}>Designed, not just generated</p>
                <p className="ps-solution">Logo directions, mood boards, and landing pages
                  all derive from one consistent palette and personality, so the output
                  reads as a brand system, not disconnected assets.</p>
              </div>
              <div className="ps-card">
                <p className="problem-text" style={{ marginBottom: 6 }}>Solves a real problem</p>
                <p className="ps-solution">Naming and early branding is a genuine bottleneck
                  for anyone starting something new &mdash; this shortens it from weeks
                  of scattered tools to one sitting.</p>
              </div>
            </div>
          </div>
        </section>

        <section className="demo-video" id="demo">
          <div className="wrap">
            <div className="section-label">See it in action</div>
            <h2 className="section-title">Watch one idea become a full brand.</h2>
            <p className="subhead">AI-coined names, live domain checks, generated logos,
              and a landing page mockup. All of it, in under three minutes.</p>

            {playing ? (
              <div className="video-frame is-playing">
                <iframe
                  src={`https://www.youtube-nocookie.com/embed/${DEMO_VIDEO_ID}?autoplay=1&rel=0`}
                  title="NomVox demo video"
                  allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                  allowFullScreen
                />
              </div>
            ) : (
              <button type="button" className="video-frame"
                onClick={() => setPlaying(true)}
                aria-label="Play the NomVox demo video">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src="/youtube-thumbnail.png" alt="" />
                <span className="duration">2:59</span>
                <span className="play-btn">
                  <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5v14l11-7z" /></svg>
                </span>
              </button>
            )}

            <a className="btn-watch" href={DEMO_VIDEO_URL} target="_blank" rel="noopener">
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5v14l11-7z" /></svg>
              Watch on YouTube
            </a>
          </div>
        </section>

        <section className="powered">
          <div className="wrap">
            <div className="section-label">Built on IBM</div>
            <h2 className="section-title">Every layer of NomVox runs on IBM.</h2>

            <div className="ibm-panel">
              <div className="ibm-badges">
                <span className="ibm-badge"><span className="dot" />IBM watsonx.ai</span>
                <span className="ibm-badge"><span className="dot" />IBM Granite</span>
                <span className="ibm-badge"><span className="dot" />IBM Bob</span>
              </div>
              <p className="lede">The naming engine, the SVG logo and mood board generator, and
                the brand persona writer all run on <strong>IBM watsonx.ai</strong> with{" "}
                <strong>IBM Granite</strong>. The build itself was carried end-to-end by{" "}
                <strong>IBM Bob</strong> &mdash; from architecture to deployment, as the primary
                development tool for this project.</p>
              <div className="ibm-chips" style={{ marginTop: 16 }}>
                <span className="ibm-chip"><b>Full-stack context</b> &mdash; held Go, TypeScript, and infra config together without losing track</span>
                <span className="ibm-chip"><b>Autonomous debugging</b> &mdash; traced issues across multiple files unaided</span>
                <span className="ibm-chip"><b>Docs as it built</b> &mdash; README and SDLC plan came out publish-ready</span>
              </div>
            </div>
          </div>
        </section>

        <section className="closing" id="start">
          <div className="wrap">
            <h2 className="section-title" style={{ marginLeft: "auto", marginRight: "auto" }}>Ready to see what your brand could be called?</h2>
            <p className="subhead">Takes about a minute to describe your idea.</p>
            <Link className="btn-primary" href="/build">Start naming your brand &rarr;</Link>
          </div>
        </section>

        <footer>
          <div className="wrap">
            <p>We hope you enjoy the experience.</p>
            <p className="small">&copy; 2026{" "}
              <a className="author-link" href="https://github.com/c-annabel" target="_blank" rel="noopener">c-annabel</a>
              {" "}&middot; Developed with IBM Bob &middot; NomVox Brand Identity Platform</p>
          </div>
        </footer>
      </main>
    </div>
  );
}
