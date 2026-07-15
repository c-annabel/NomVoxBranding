package handlers

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/c-annabel/NomVoxBranding/internal/models"
)

// exportRequest is the payload sent from the frontend.
type exportRequest struct {
	SessionID       string               `json:"session_id"`
	BrandName       string               `json:"brand_name"`
	Card            models.NameCard      `json:"card"`
	Intake          models.IntakePayload `json:"intake"`
	MoodBoard       []string             `json:"mood_board"`        // data URIs
	LogoProfile     string               `json:"logo_profile"`      // data URI
	LogoApp         string               `json:"logo_app"`          // data URI
	LogoBusiness    string               `json:"logo_business"`     // data URI
	SelectedLogoKey string               `json:"selected_logo_key"` // "profile", "app", or "business"
	MockupHTML      string               `json:"mockup_html"`
	Persona         *models.BrandPersona `json:"persona,omitempty"`
}

// ExportHandler handles POST /api/export.
// It assembles all brand assets into a ZIP file and streams it to the browser.
func ExportHandler(w http.ResponseWriter, r *http.Request) {
	var req exportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	slug := slugify(req.BrandName)
	if slug == "" {
		slug = "brand"
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-nomvox-pack.zip"`, slug))
	w.Header().Set("Cache-Control", "no-store")

	zw := zip.NewWriter(w)
	defer zw.Close()

	now := time.Now().UTC()

	// ── brand-brief.json ─────────────────────────────────────────────
	if brief, err := json.MarshalIndent(brandBrief{
		GeneratedAt: now.Format(time.RFC3339),
		BrandName:   req.BrandName,
		Card:        req.Card,
		Intake:      req.Intake,
		Persona:     req.Persona,
	}, "", "  "); err == nil {
		writeZipEntry(zw, slug+"/brand-brief.json", brief)
	}

	// ── brand-brief.html — human-readable version ─────────────────────
	writeZipEntry(zw, slug+"/brand-brief.html", []byte(buildBrandBriefHTML(req, now)))

	// ── landing-page.html ────────────────────────────────────────────
	if strings.TrimSpace(req.MockupHTML) != "" {
		writeZipEntry(zw, slug+"/landing-page.html", []byte(req.MockupHTML))
	}

	// ── mood-board — supports SVG or PNG data URIs ────────────────────
	for i, uri := range req.MoodBoard {
		ext := dataURIExt(uri)
		if ext == "" {
			continue
		}
		if data := dataURIToBytes(uri); len(data) > 0 {
			writeZipEntry(zw, fmt.Sprintf("%s/mood-board/panel-%d.%s", slug, i+1, ext), data)
		}
	}

	// ── logos — supports SVG or PNG ──────────────────────────────────
	logoMap := map[string]string{
		"profile":  req.LogoProfile,
		"app":      req.LogoApp,
		"business": req.LogoBusiness,
	}
	baseNames := map[string]string{
		"profile":  "logo-profile",
		"app":      "logo-app-icon",
		"business": "logo-business-card",
	}
	for key, uri := range logoMap {
		if uri == "" {
			continue
		}
		ext := dataURIExt(uri)
		if ext == "" {
			continue
		}
		if data := dataURIToBytes(uri); len(data) > 0 {
			writeZipEntry(zw, slug+"/logos/"+baseNames[key]+"."+ext, data)
		}
	}
	// Write the selected logo as a dedicated easy-to-find file
	if req.SelectedLogoKey != "" {
		if uri, ok := logoMap[req.SelectedLogoKey]; ok && uri != "" {
			ext := dataURIExt(uri)
			if ext != "" {
				if data := dataURIToBytes(uri); len(data) > 0 {
					writeZipEntry(zw, slug+"/logos/selected-logo."+ext, data)
				}
			}
		}
	}

	// ── README.txt ───────────────────────────────────────────────────
	readme := buildReadme(req)
	writeZipEntry(zw, slug+"/README.txt", []byte(readme))
}

// ── helpers ───────────────────────────────────────────────────────────────────

type brandBrief struct {
	GeneratedAt string               `json:"generated_at"`
	BrandName   string               `json:"brand_name"`
	Card        models.NameCard      `json:"name_card"`
	Intake      models.IntakePayload `json:"intake"`
	Persona     *models.BrandPersona `json:"persona,omitempty"`
}

func writeZipEntry(zw *zip.Writer, name string, data []byte) {
	f, err := zw.Create(name)
	if err != nil {
		return
	}
	f.Write(data) //nolint:errcheck
}

// dataURIExt returns the file extension for a data URI based on its MIME type.
// Returns "svg", "png", "jpg", or "webp". Returns "" if the URI is empty or unrecognised.
func dataURIExt(uri string) string {
	if uri == "" {
		return ""
	}
	lower := strings.ToLower(uri)
	switch {
	case strings.HasPrefix(lower, "data:image/svg"):
		return "svg"
	case strings.HasPrefix(lower, "data:image/png"):
		return "png"
	case strings.HasPrefix(lower, "data:image/jpeg"), strings.HasPrefix(lower, "data:image/jpg"):
		return "jpg"
	case strings.HasPrefix(lower, "data:image/webp"):
		return "webp"
	}
	return ""
}

// dataURIToBytes decodes a data URI (data:image/…;base64,… or data:image/svg+xml;base64,…)
// to raw bytes. Supports PNG (binary base64) and SVG (base64 encoded UTF-8 text).
func dataURIToBytes(uri string) []byte {
	if uri == "" {
		return nil
	}
	comma := strings.Index(uri, ",")
	if comma == -1 {
		return nil
	}
	payload := uri[comma+1:]
	header := strings.ToLower(uri[:comma])

	// SVG can also be stored as plain text after the comma (not base64).
	// Check if the header says base64 — if not, treat payload as raw text.
	if strings.Contains(header, "image/svg") && !strings.Contains(header, "base64") {
		return []byte(payload)
	}

	// Standard base64 decode — try all variants.
	if data, err := base64.StdEncoding.DecodeString(payload); err == nil {
		return data
	}
	if data, err := base64.RawStdEncoding.DecodeString(payload); err == nil {
		return data
	}
	if data, err := base64.RawURLEncoding.DecodeString(payload); err == nil {
		return data
	}
	// For SVG data URIs that failed base64 — return as raw text (the SVG markup itself).
	if strings.Contains(header, "image/svg") {
		return []byte(payload)
	}
	return nil
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	return b.String()
}

func buildReadme(req exportRequest) string {
	year := time.Now().UTC().Format("2006")
	var b strings.Builder
	b.WriteString("════════════════════════════════════════════════\n")
	b.WriteString("  NomVox Brand Identity Pack\n")
	b.WriteString("════════════════════════════════════════════════\n")
	b.WriteString("Generated by NomVox — https://nomvox.app\n")
	b.WriteString("Built with IBM Bob · AI-powered brand identity\n")
	b.WriteString("Date: " + time.Now().UTC().Format("2006-01-02") + "\n")
	b.WriteString("© " + year + " c-annabel. Developed with IBM Bob.\n\n")

	b.WriteString("── Brand Name ──────────────────────────────────\n")
	b.WriteString("Name:        " + req.BrandName + "\n")
	if req.Card.Tagline != "" {
		b.WriteString("Tagline:     " + req.Card.Tagline + "\n")
	}
	if req.Card.OriginStory != "" {
		b.WriteString("Origin:      " + req.Card.OriginStory + "\n")
	}
	if req.Card.ShortDesc != "" {
		b.WriteString("Short desc:  " + req.Card.ShortDesc + "\n")
	}
	if req.Card.LongDesc != "" {
		b.WriteString("Long desc:   " + req.Card.LongDesc + "\n")
	}

	b.WriteString("\n── Brand Score ─────────────────────────────────\n")
	b.WriteString(fmt.Sprintf("Memorability:  %d/10  — %s\n", req.Card.Score.Memorability, req.Card.Score.MemReasoning))
	b.WriteString(fmt.Sprintf("Spellability:  %d/10  — %s\n", req.Card.Score.Spellability, req.Card.Score.SpellReasoning))
	b.WriteString(fmt.Sprintf("Global Safety: %d/10  — %s\n", req.Card.Score.GlobalSafety, req.Card.Score.GlobalReasoning))
	b.WriteString(fmt.Sprintf("Squatter Risk: %s     — %s\n", req.Card.Score.SquatterRisk, req.Card.Score.SquatterReasoning))

	if len(req.Card.StyleTags) > 0 {
		b.WriteString("\n── Style Tags ──────────────────────────────────\n")
		b.WriteString(strings.Join(req.Card.StyleTags, " · ") + "\n")
	}

	b.WriteString("\n── Brand Voice Samples ─────────────────────────\n")
	if req.Card.VoiceSamples.InstagramCaption != "" {
		b.WriteString("Instagram:  " + req.Card.VoiceSamples.InstagramCaption + "\n")
	}
	if req.Card.VoiceSamples.EmailSubject != "" {
		b.WriteString("Email subj: " + req.Card.VoiceSamples.EmailSubject + "\n")
	}
	if req.Card.VoiceSamples.NotFoundMessage != "" {
		b.WriteString("404 page:   " + req.Card.VoiceSamples.NotFoundMessage + "\n")
	}

	b.WriteString("\n── Brand Inputs ────────────────────────────────\n")
	if req.Intake.CoreIdea != "" {
		b.WriteString("Core idea:  " + req.Intake.CoreIdea + "\n")
	}
	if req.Intake.TargetAudience != "" {
		b.WriteString("Audience:   " + req.Intake.TargetAudience + "\n")
	}
	if req.Intake.Industry != "" {
		b.WriteString("Industry:   " + req.Intake.Industry + "\n")
	}
	if req.Intake.Personality != "" {
		b.WriteString("Personality:" + req.Intake.Personality + "\n")
	}
	if req.Intake.Style != "" {
		b.WriteString("Style:      " + req.Intake.Style + "\n")
	}
	if req.Intake.ColorMood != "" {
		b.WriteString("Colours:    " + req.Intake.ColorMood + "\n")
	}

	b.WriteString("\n── Files Included ──────────────────────────────\n")
	b.WriteString("brand-brief.json    — Machine-readable full brand brief\n")
	b.WriteString("brand-brief.html    — Human-readable brand brief (open in browser)\n")
	b.WriteString("landing-page.html   — AI hero section mockup (open in browser)\n")
	b.WriteString("mood-board/         — Brand mood board panels (SVG vector files)\n")
	b.WriteString("logos/              — 3 logo concepts + selected-logo (SVG vector)\n")
	b.WriteString("                      Open .svg files in any browser or Figma\n")

	b.WriteString("\n── Legal ───────────────────────────────────────\n")
	b.WriteString("These assets were generated by AI and are intended as creative inspiration.\n")
	b.WriteString("Always verify brand name availability before registration.\n")
	b.WriteString("© " + year + " c-annabel. Developed with IBM Bob. All rights reserved.\n")
	return b.String()
}

// buildBrandBriefHTML generates a standalone human-readable HTML file with all brand data.
func buildBrandBriefHTML(req exportRequest, now time.Time) string {
	year := now.Format("2006")
	date := now.Format("2006-01-02")

	// Availability score from the card — we don't have the full probe results in export,
	// so we surface the Brand Score and voice samples instead.
	scoreColor := func(v int) string {
		if v >= 7 { return "#4ade80" }
		if v >= 4 { return "#f59e0b" }
		return "#f87171"
	}
	riskColor := func(r string) string {
		switch r {
		case "Low":  return "#4ade80"
		case "High": return "#f87171"
		default:     return "#f59e0b"
		}
	}
	tagsHTML := ""
	for _, t := range req.Card.StyleTags {
		tagsHTML += fmt.Sprintf(`<span class="tag">%s</span>`, escHTML(t))
	}
	personaHTML := ""
	if req.Persona != nil {
		cvHTML := ""
		for _, v := range req.Persona.CoreValues {
			cvHTML += fmt.Sprintf(`<span class="tag">%s</span>`, escHTML(v))
		}
		nsHTML := ""
		for _, v := range req.Persona.NeverSays {
			nsHTML += fmt.Sprintf(`<span class="tag red">%s</span>`, escHTML(v))
		}
		personaHTML = fmt.Sprintf(`
		<section>
		  <h2>Brand Persona</h2>
		  <table><tr><th>Age</th><td>%d</td></tr>
		  <tr><th>Occupation</th><td>%s</td></tr>
		  <tr><th>Voice</th><td>%s</td></tr>
		  <tr><th>Reads</th><td>%s</td></tr>
		  <tr><th>Core Values</th><td>%s</td></tr>
		  <tr><th>Never Says</th><td>%s</td></tr>
		  </table>
		</section>`,
			req.Persona.Age,
			escHTML(req.Persona.Occupation),
			escHTML(req.Persona.Voice),
			escHTML(strings.Join(req.Persona.Reads, ", ")),
			cvHTML, nsHTML,
		)
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s — NomVox Brand Brief</title>
<style>
  :root{--void:#0B0F19;--pulse:#8B5CF6;--signal:#22d3ee;--lunar:#E6E8F0;}
  *{box-sizing:border-box;margin:0;padding:0;}
  body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
    background:var(--void);color:var(--lunar);min-height:100vh;padding:40px 24px;}
  .wrap{max-width:820px;margin:0 auto;}
  header{border-bottom:2px solid var(--pulse);padding-bottom:24px;margin-bottom:32px;}
  .brand-name{font-size:3rem;font-weight:900;letter-spacing:-1px;color:#fff;}
  .tagline{font-size:1.25rem;font-style:italic;color:var(--signal);margin-top:6px;}
  .meta{font-size:.8rem;color:#64748b;margin-top:8px;}
  section{margin-bottom:32px;background:rgba(255,255,255,0.03);
    border:1px solid rgba(255,255,255,0.08);border-radius:12px;padding:24px;}
  h2{font-size:.7rem;font-weight:900;text-transform:uppercase;letter-spacing:.12em;
    color:var(--pulse);margin-bottom:16px;}
  p{line-height:1.7;color:#cbd5e1;font-size:.95rem;margin-bottom:8px;}
  table{width:100%%;border-collapse:collapse;}
  th{text-align:left;width:140px;color:#64748b;font-size:.82rem;
    font-weight:700;padding:6px 0;vertical-align:top;}
  td{color:var(--lunar);font-size:.9rem;padding:6px 0;padding-left:12px;}
  .score-row{display:flex;align-items:center;gap:12px;margin-bottom:8px;}
  .score-label{width:110px;font-size:.82rem;color:#94a3b8;}
  .score-bar{flex:1;height:6px;background:rgba(255,255,255,0.08);
    border-radius:9999px;overflow:hidden;max-width:120px;}
  .score-fill{height:100%%;border-radius:9999px;}
  .score-val{font-size:.9rem;font-weight:700;width:24px;text-align:right;}
  .score-reason{font-size:.78rem;color:#64748b;flex:1;}
  .tag{display:inline-block;background:rgba(139,92,246,0.2);color:#c4b5fd;
    border:1px solid rgba(139,92,246,0.4);border-radius:9999px;
    padding:3px 10px;font-size:.75rem;font-weight:700;margin:2px;}
  .tag.red{background:rgba(248,113,113,0.15);color:#fca5a5;border-color:rgba(248,113,113,0.35);}
  .voice-item{margin-bottom:10px;}
  .voice-label{font-size:.75rem;font-weight:900;color:var(--signal);
    text-transform:uppercase;letter-spacing:.08em;}
  .voice-text{font-size:.9rem;font-style:italic;color:#cbd5e1;margin-top:2px;}
  footer{text-align:center;margin-top:48px;padding-top:24px;
    border-top:1px solid rgba(255,255,255,0.06);
    font-size:.75rem;color:#334155;}
</style>
</head>
<body>
<div class="wrap">
  <header>
    <div class="brand-name">%s</div>
    <div class="tagline">"%s"</div>
    <div class="meta">Generated %s · NomVox · Built with IBM Bob</div>
    <div style="margin-top:10px;">%s</div>
  </header>

  <section>
    <h2>Brand Story</h2>
    <p><strong>Origin:</strong> %s</p>
    <p>%s</p>
    <p style="color:#94a3b8;">%s</p>
  </section>

  <section>
    <h2>Brand Score</h2>
    <div class="score-row">
      <span class="score-label">Memorability</span>
      <div class="score-bar"><div class="score-fill" style="width:%d%%;background:%s;"></div></div>
      <span class="score-val" style="color:%s;">%d</span>
      <span class="score-reason">%s</span>
    </div>
    <div class="score-row">
      <span class="score-label">Spellability</span>
      <div class="score-bar"><div class="score-fill" style="width:%d%%;background:%s;"></div></div>
      <span class="score-val" style="color:%s;">%d</span>
      <span class="score-reason">%s</span>
    </div>
    <div class="score-row">
      <span class="score-label">Global Safety</span>
      <div class="score-bar"><div class="score-fill" style="width:%d%%;background:%s;"></div></div>
      <span class="score-val" style="color:%s;">%d</span>
      <span class="score-reason">%s</span>
    </div>
    <div class="score-row">
      <span class="score-label">Squatter Risk</span>
      <span style="color:%s;font-weight:700;">%s</span>
      <span class="score-reason">%s</span>
    </div>
  </section>

  <section>
    <h2>Brand Voice</h2>
    <div class="voice-item">
      <div class="voice-label">Instagram Caption</div>
      <div class="voice-text">"%s"</div>
    </div>
    <div class="voice-item">
      <div class="voice-label">Email Subject Line</div>
      <div class="voice-text">"%s"</div>
    </div>
    <div class="voice-item">
      <div class="voice-label">404 Page Message</div>
      <div class="voice-text">"%s"</div>
    </div>
  </section>

  <section>
    <h2>Brand Inputs</h2>
    <table>
      <tr><th>Core Idea</th><td>%s</td></tr>
      <tr><th>Audience</th><td>%s</td></tr>
      <tr><th>Industry</th><td>%s</td></tr>
      <tr><th>Personality</th><td>%s</td></tr>
      <tr><th>Style</th><td>%s</td></tr>
      <tr><th>Colours</th><td>%s</td></tr>
      <tr><th>Name Length</th><td>%s</td></tr>
      <tr><th>Avoid</th><td>%s</td></tr>
    </table>
  </section>
  %s
  <footer>
    &copy; %s c-annabel &mdash; Developed with IBM Bob &mdash; Generated by NomVox &mdash; https://nomvox.app<br>
    These assets are AI-generated and intended as creative inspiration.
    Always verify brand name availability before registration.
  </footer>
</div>
</body>
</html>`,
		escHTML(req.BrandName),
		// header
		escHTML(req.BrandName), escHTML(req.Card.Tagline), date, tagsHTML,
		// brand story
		escHTML(req.Card.OriginStory), escHTML(req.Card.ShortDesc), escHTML(req.Card.LongDesc),
		// memorability
		req.Card.Score.Memorability*10, scoreColor(req.Card.Score.Memorability),
		scoreColor(req.Card.Score.Memorability), req.Card.Score.Memorability,
		escHTML(req.Card.Score.MemReasoning),
		// spellability
		req.Card.Score.Spellability*10, scoreColor(req.Card.Score.Spellability),
		scoreColor(req.Card.Score.Spellability), req.Card.Score.Spellability,
		escHTML(req.Card.Score.SpellReasoning),
		// global safety
		req.Card.Score.GlobalSafety*10, scoreColor(req.Card.Score.GlobalSafety),
		scoreColor(req.Card.Score.GlobalSafety), req.Card.Score.GlobalSafety,
		escHTML(req.Card.Score.GlobalReasoning),
		// squatter risk
		riskColor(req.Card.Score.SquatterRisk), escHTML(req.Card.Score.SquatterRisk),
		escHTML(req.Card.Score.SquatterReasoning),
		// voice
		escHTML(req.Card.VoiceSamples.InstagramCaption),
		escHTML(req.Card.VoiceSamples.EmailSubject),
		escHTML(req.Card.VoiceSamples.NotFoundMessage),
		// inputs
		escHTML(req.Intake.CoreIdea), escHTML(req.Intake.TargetAudience),
		escHTML(req.Intake.Industry), escHTML(req.Intake.Personality),
		escHTML(req.Intake.Style), escHTML(req.Intake.ColorMood),
		escHTML(req.Intake.NameLength), escHTML(req.Intake.Avoid),
		// persona
		personaHTML,
		// footer
		year,
	)
}

// escHTML escapes HTML special characters to prevent XSS in the generated file.
func escHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&#34;")
	return s
}
