package ai

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/c-annabel/NomVoxBranding/internal/models"
)

// ── Imagen 3 Prompt Builders ──────────────────────────────────────────────────

// logoStyleDescription returns a human-readable description of the chosen logo type
// for use in prompts, so the mood board visually references the logo's aesthetic.
func logoStyleDescription(logoKey string) string {
	switch logoKey {
	case "profile":
		return "flat geometric Bauhaus mark — bold shapes, strong negative space, minimal palette"
	case "app":
		return "vibrant gradient icon — deep colour depth, glassmorphism, neon glow accent"
	case "business":
		return "horizontal wordmark lockup — clean white background, corporate typography, single accent colour"
	default:
		return "clean modern mark"
	}
}

// MoodBoardPrompt keeps the original signature for backwards compatibility.
// It returns the "texture" panel prompt.
func MoodBoardPrompt(card models.NameCard, intake models.IntakePayload, vision *models.VisionContext, selectedLogoKey, selectedLogoStyle string) string {
	return MoodBoardPanelPrompt(card, intake, vision, selectedLogoKey, selectedLogoStyle, "texture")
}

// MoodBoardPanelPrompt builds a prompt for ONE single photographic image.
// panel is "texture" or "atmosphere".
//
// The previous version asked for a 2x2 grid of four panels, which made Gemini
// return a single collage image complete with white gutters. Each call now
// requests one unbroken frame instead.
func MoodBoardPanelPrompt(card models.NameCard, intake models.IntakePayload, vision *models.VisionContext, selectedLogoKey, selectedLogoStyle, panel string) string {
	// ── Colour palette ────────────────────────────────────────────────
	colourHint := strings.TrimSpace(intake.ColorMood)
	if vision != nil && len(vision.Colours) > 0 {
		colourHint = strings.Join(vision.Colours, ", ")
	}
	if colourHint == "" {
		colourHint = "deep midnight blue, electric cyan, warm cream"
	}

	// ── Visual mood ───────────────────────────────────────────────────
	moodHint := strings.TrimSpace(intake.Personality)
	if vision != nil && len(vision.Mood) > 0 {
		moodHint = strings.Join(vision.Mood, ", ")
	}
	if moodHint == "" {
		moodHint = "distinctive, aspirational"
	}

	// ── Visual style ──────────────────────────────────────────────────
	styleHint := strings.TrimSpace(intake.Style)
	if vision != nil && len(vision.Style) > 0 {
		styleHint = strings.Join(vision.Style, ", ")
	}
	if styleHint == "" {
		styleHint = "editorial minimalism"
	}

	// ── Industry ──────────────────────────────────────────────────────
	industry := strings.TrimSpace(intake.Industry)
	if industry == "" {
		industry = "creative"
	}

	// ── Audience ──────────────────────────────────────────────────────
	audience := strings.TrimSpace(intake.TargetAudience)
	if audience == "" {
		audience = "modern professionals"
	}

	// ── Logo style reference ──────────────────────────────────────────
	logoStyle := strings.TrimSpace(selectedLogoStyle)
	if logoStyle == "" {
		logoStyle = logoStyleDescription(selectedLogoKey)
	}

	// ── Name etymology / origin ───────────────────────────────────────
	originContext := strings.TrimSpace(card.OriginStory)
	if originContext == "" {
		originContext = card.ShortDesc
	}

	// ── Avoid hint ────────────────────────────────────────────────────
	avoidClause := ""
	if avoidHint := strings.TrimSpace(intake.Avoid); avoidHint != "" {
		avoidClause = fmt.Sprintf("Do NOT include: %s. ", avoidHint)
	}

	// ── Panel subject ─────────────────────────────────────────────────
	subject := fmt.Sprintf(
		`A single macro material study. Surface texture rendered in %s, `+
			`with a tactile quality that suggests %s. The material should feel `+
			`like something from the world of %s. Shallow depth of field, `+
			`dramatic side lighting, fills the entire frame.`,
		colourHint, moodHint, industry)

	if panel == "atmosphere" {
		subject = fmt.Sprintf(
			`A single atmospheric environment shot. An interior or architectural `+
				`space that a %s brand serving %s would inhabit. Light, scale and `+
				`mood expressed in %s, matching a %s aesthetic. Wide angle, `+
				`cinematic lighting, no people, fills the entire frame.`,
			industry, audience, colourHint, styleHint)
	}

	return fmt.Sprintf(
		`ONE single photographic image. This is NOT a collage, NOT a grid, `+
			`NOT a multi-panel composition. One unbroken frame with no borders, `+
			`no white margins, and no dividing lines.

Brand context — the image must feel like it belongs to THIS brand:
- Brand: "%s", a %s brand
- Name meaning: %s
- Tagline: "%s"
- Personality: %s
- Aesthetic: %s
- Chosen logo style: %s

Subject: %s

Strict rules: magazine-quality commercial photography. No text, no letters, `+
			`no numbers, no logos, no watermarks, no human faces, no borders. `+
			`The image must fill the frame edge to edge. %s`,
		card.Name, industry,
		originContext,
		card.Tagline,
		moodHint,
		styleHint,
		logoStyle,
		subject,
		avoidClause,
	)
}

// logoPalette resolves the user's free-text colour mood into concrete hex codes.
//
// Passing raw text like "electric blue, neon yellow, dark bg" straight into an
// image prompt makes the model guess, and "dark bg" gets read as a mark colour.
// Resolving to hex first gives predictable, on-brand output.
//
// Returns: dark background hex, primary mark hex, secondary mark hex.
func logoPalette(intake models.IntakePayload) (darkBG, primary, secondary string) {
	lower := strings.ToLower(strings.TrimSpace(intake.ColorMood))

	// Ordered so multi-word keys match before their single-word prefixes.
	rules := []struct{ key, hex string }{
		{"neon yellow", "#facc15"}, {"electric blue", "#3b82f6"},
		{"sky blue", "#38bdf8"}, {"navy", "#1e3a8a"},
		{"orange", "#f97316"}, {"coral", "#fb7185"},
		{"yellow", "#facc15"}, {"gold", "#eab308"}, {"amber", "#f59e0b"},
		{"teal", "#22d3ee"}, {"cyan", "#22d3ee"}, {"aqua", "#22d3ee"},
		{"turquoise", "#2dd4bf"},
		{"emerald", "#10b981"}, {"lime", "#84cc16"}, {"green", "#10b981"},
		{"violet", "#7c3aed"}, {"purple", "#8b5cf6"}, {"indigo", "#6366f1"},
		{"magenta", "#e879f9"}, {"pink", "#ec4899"}, {"rose", "#f43f5e"},
		{"crimson", "#dc2626"}, {"scarlet", "#f43f5e"}, {"red", "#ef4444"},
		{"blue", "#3b82f6"},
	}

	// Collect matches in the order the user wrote them.
	type hit struct {
		pos int
		hex string
	}
	var hits []hit
	seen := map[string]bool{}
	for _, r := range rules {
		if idx := strings.Index(lower, r.key); idx >= 0 && !seen[r.hex] {
			seen[r.hex] = true
			hits = append(hits, hit{idx, r.hex})
		}
	}
	for i := 1; i < len(hits); i++ {
		for j := i; j > 0 && hits[j].pos < hits[j-1].pos; j-- {
			hits[j], hits[j-1] = hits[j-1], hits[j]
		}
	}

	primary, secondary = "#3b82f6", "#facc15"
	if len(hits) > 0 {
		primary = hits[0].hex
	}
	if len(hits) > 1 {
		secondary = hits[1].hex
	} else if len(hits) == 1 {
		secondary = "#ffffff"
	}

	darkBG = "#0a1628"
	switch {
	case strings.Contains(lower, "midnight"), strings.Contains(lower, "void"),
		strings.Contains(lower, "black"):
		darkBG = "#050810"
	case strings.Contains(lower, "charcoal"), strings.Contains(lower, "slate"):
		darkBG = "#1e293b"
	}
	return darkBG, primary, secondary
}

// LogoConceptPrompt builds a Gemini image prompt for one logo format.
// logoType: "profile" | "app" | "business"
//
// Prompt design notes:
//   - Hard constraints lead. Image models weight early tokens most, so the
//     "one shape, no letters" rule must not be buried at the end.
//   - Colours are given as explicit hex codes, never as free text.
//   - The core idea is included so the symbol relates to what the brand does.
//   - Each format has a distinct role, so the three read as one system rather
//     than three unrelated images.

func LogoConceptPrompt(card models.NameCard, intake models.IntakePayload, logoType string) string {
	darkBG, primary, secondary := logoPalette(intake)

	coreIdea := strings.TrimSpace(intake.CoreIdea)
	if coreIdea == "" {
		coreIdea = strings.TrimSpace(card.ShortDesc)
	}
	if coreIdea == "" {
		coreIdea = "a modern brand"
	}

	personality := strings.TrimSpace(intake.Personality)
	if personality == "" {
		personality = "modern and distinctive"
	}
	industry := strings.TrimSpace(intake.Industry)
	if industry == "" {
		industry = "technology"
	}

	// Shared hard rules, repeated at the top and bottom of every prompt.
	const noText = `ABSOLUTE RULE: the image must contain NO letters, NO words, ` +
		`NO numbers, NO monogram, NO initials and NO typography of any kind. ` +
		`Do not render the brand name. Any letterform makes the output unusable.`

	switch logoType {
	case "profile":
		return fmt.Sprintf(
			`%s
 
Design ONE flat vector logo symbol. A single geometric mark, nothing else.
 
What the brand does: %s (industry: %s).
The symbol should abstractly suggest that idea through pure geometry —
simple, confident shapes, not an illustration or a scene.
 
Exact colours, use these and nothing else:
- Symbol primary: %s
- Symbol secondary accent: %s
- Background: pure white #ffffff, completely plain
 
Rules: flat vector only. Solid fills. Two colours maximum on the symbol.
Thick strong strokes, generous negative space, perfectly centred,
occupying about 60%% of the frame. No gradients, no shadows, no glow,
no 3D, no texture, no outlines around the canvas, no mockup or device frame.
Tone: %s. Must stay legible at 32x32 pixels.
Reference quality: the mark should feel as clean and inevitable as the
Nike swoosh or the Airbnb Belo.
 
%s`,
			noText, coreIdea, industry, primary, secondary, personality, noText)

	case "app":
		return fmt.Sprintf(
			`%s
 
Design ONE app icon. A rounded-square tile with a single symbol centred on it.
 
What the brand does: %s (industry: %s).
The symbol abstractly suggests that idea. ONE shape only — not a composition,
not several objects combined, not a scene.
 
Exact colours, use these and nothing else:
- Tile background: solid %s, completely flat, edge to edge
- Symbol: %s
- Optional small accent detail: %s
 
Rules: flat vector only. Solid fills. The symbol occupies about 55%% of the
tile with generous even padding on all sides. Thick simple forms.
No glow, no neon, no bloom, no gradients, no glassmorphism, no drop shadows,
no 3D, no bevel, no fine detail, no background pattern.
Tone: %s. Must be instantly recognisable at 48x48 pixels.
Reference quality: Spotify, Airbnb or Duolingo — one bold shape a child
could redraw from memory.
 
%s`,
			noText, coreIdea, industry, darkBG, primary, secondary, personality, noText)

	case "business":
		// The only format where type is intentional, so noText does not apply.
		return fmt.Sprintf(
			`Design a horizontal brand lockup for a business card.
 
Layout, left to right:
1. A small flat vector symbol, ONE simple geometric shape only.
2. A thin vertical divider rule.
3. The brand name "%s" set in a bold geometric sans-serif,
   with the tagline "%s" directly beneath it in a smaller regular weight.
 
What the brand does: %s (industry: %s).
The symbol abstractly suggests that idea.
 
Exact colours, use these and nothing else:
- Background: pure white #ffffff, completely plain
- Symbol: %s with a %s accent
- Brand name text: near-black #111111
- Tagline text: %s
 
Rules: flat vector only. Solid fills. Landscape composition with generous
margins, the lockup centred and occupying about 70%% of the width.
No gradients, no shadows, no glow, no 3D, no photographic card mockup,
no perspective, no background texture, no extra words beyond the brand
name and tagline given above.
Tone: %s. Print quality, the typographic confidence of a Fortune 500 rebrand.`,
			card.Name, card.Tagline, coreIdea, industry,
			primary, secondary, primary, personality)

	default:
		return fmt.Sprintf(
			`%s
 
ONE flat vector logo symbol for a %s brand. Symbol in %s with a %s accent
on a pure white background. Simple geometry, solid fills, centred.
 
%s`,
			noText, industry, primary, secondary, noText)
	}
}

// ── SVG Logo Generation (Granite / watsonx) ──────────────────────────────────
//
// These prompts must faithfully reflect the user's intake inputs (colour_mood,
// personality, industry, tagline). The AI must NOT use generic brand colours.

// BuildSVGLogoSystemPrompt returns the system prompt for SVG logo generation.
func BuildSVGLogoSystemPrompt() string {
	return `You are NomVox, a precision brand design AI. Generate a self-contained SVG logo.

ABSOLUTE RULES — follow exactly or the output is invalid:
1. Return ONLY raw SVG starting with <svg — zero markdown, zero explanation, zero code fences.
2. The SVG must be fully self-contained. No <image>, no foreignObject, no scripts, no external hrefs.
3. viewBox="0 0 200 200" for square formats. viewBox="0 0 400 200" for landscape wordmark.
4. COLOUR RULE: You MUST use the EXACT colours described in the user prompt. Do NOT substitute generic purple/indigo. If the user says "deep navy, white, teal" — use deep navy for background, white for text, teal for accents.
5. Use only: <rect>, <circle>, <ellipse>, <path>, <polygon>, <text>, <g>, <defs>, <linearGradient>, <radialGradient>.
6. Maximum 80 lines of SVG output.
7. The SVG must render correctly in an <img> src="data:image/svg+xml;base64,..." tag.`
}

// buildColourContext extracts colour guidance from user intake.
// Handles phrases like "bright orange, sky blue, white" correctly.
// Returns (bg=always dark, primary=#ffffff, accent=first named colour).
func buildColourContext(intake models.IntakePayload) (bg, primary, accent string) {
	raw := strings.TrimSpace(intake.ColorMood)
	if raw == "" {
		return "#0a1628", "#ffffff", "#f97316"
	}
	lower := strings.ToLower(raw)
	primary = "#ffffff"

	// Background — always keep dark; "white" means text colour, not bg
	bg = "#0a1628"
	if strings.Contains(lower, "midnight") || strings.Contains(lower, "void") || strings.Contains(lower, "black") {
		bg = "#050810"
	} else if strings.Contains(lower, "dark") {
		bg = "#0d0d1a"
	}

	// Colour → hex mapping (first match wins for accent)
	type rule struct{ key, hex string }
	rules := []rule{
		{"bright orange", "#f97316"}, {"orange", "#f97316"}, {"coral", "#f97316"},
		{"neon yellow", "#facc15"}, {"yellow", "#facc15"}, {"gold", "#f59e0b"}, {"amber", "#f59e0b"},
		{"sky blue", "#38bdf8"}, {"sky", "#38bdf8"}, {"azure", "#38bdf8"},
		{"electric blue", "#3b82f6"}, {"electric", "#3b82f6"},
		{"teal", "#22d3ee"}, {"cyan", "#22d3ee"}, {"aqua", "#22d3ee"},
		{"green", "#10b981"}, {"emerald", "#10b981"}, {"lime", "#84cc16"},
		{"purple", "#8B5CF6"}, {"violet", "#7c3aed"}, {"indigo", "#6366f1"},
		{"pink", "#ec4899"}, {"rose", "#f43f5e"}, {"magenta", "#e879f9"},
		{"red", "#ef4444"}, {"crimson", "#dc2626"}, {"scarlet", "#f43f5e"},
		{"blue", "#3b82f6"},
	}
	accent = "#f97316" // safe default (orange)
	for _, r := range rules {
		if strings.Contains(lower, r.key) {
			accent = r.hex
			break
		}
	}
	return bg, primary, accent
}

// BuildSVGLogoUserPrompt returns the user turn for SVG logo generation.
// logoType: "profile" | "app" | "business"
func BuildSVGLogoUserPrompt(card models.NameCard, intake models.IntakePayload, logoType string) string {
	colourMood := strings.TrimSpace(intake.ColorMood)
	if colourMood == "" {
		colourMood = "deep navy background, white text, teal/cyan accent"
	}
	bg, primary, accent := buildColourContext(intake)

	personality := strings.TrimSpace(intake.Personality)
	if personality == "" {
		personality = "modern and distinctive"
	}
	industry := strings.TrimSpace(intake.Industry)
	if industry == "" {
		industry = "technology"
	}
	tagline := strings.TrimSpace(card.Tagline)
	initial := ""
	if len([]rune(card.Name)) > 0 {
		initial = string([]rune(card.Name)[:1])
	}
	initials := initial
	if len([]rune(card.Name)) > 1 {
		initials = string([]rune(card.Name)[:2])
	}

	// Colour instruction — always explicit, always references user input
	colourInstr := fmt.Sprintf(
		`USER'S COLOUR PALETTE (you MUST use these — do not substitute):
"%s"
Translated: background = %s, text/marks = %s, accent/highlight = %s.
Every colour decision MUST come from this palette.`,
		colourMood, bg, primary, accent,
	)

	switch logoType {
	case "profile":
		return fmt.Sprintf(
			`Brand: "%s" · Tagline: "%s" · Industry: %s · Personality: %s

%s

Create a square SVG profile/social logo (viewBox="0 0 200 200"):
- Background: fill the entire canvas with %s (the brand background colour).
- Central mark: ONE bold geometric shape (hexagon, diamond, or interlocking circles) centred at 100,100.
  Fill the shape with a gradient from %s to %s.
- Brand initial "%s" centred in white (#ffffff), font-size="64", font-weight="bold", font-family="Arial, sans-serif".
- Optional: one thin ring or orbit circle around the mark in %s at 25%% opacity.
- NO brand name text anywhere — mark + initial only.
- The design must feel like it was made specifically for this brand, not a generic monogram.`,
			card.Name, tagline, industry, personality,
			colourInstr, bg, accent, bg, initial, accent,
		)
	case "app":
		return fmt.Sprintf(
			`Brand: "%s" · Tagline: "%s" · Industry: %s · Personality: %s

%s

Create a square SVG app store icon (viewBox="0 0 200 200"):
- Outer shape: rounded rectangle x="10" y="10" width="180" height="180" rx="40".
  Fill: linear gradient from %s (top) to a slightly lighter version of %s (bottom).
- Inside the rounded rect: monogram "%s" centred at 100,100 in %s (#ffffff or accent), font-size="72", font-weight="900".
- Add a subtle radial gradient glow at cx="100" cy="100" r="60" from %s at 30%% opacity to transparent.
- Border stroke on the rounded rect: %s at 40%% opacity, stroke-width="2".
- The icon must be recognisable at 32×32px — keep it simple and bold.`,
			card.Name, tagline, industry, personality,
			colourInstr, bg, bg, initials, primary, accent, accent,
		)
	case "business":
		return fmt.Sprintf(
			`Brand: "%s" · Tagline: "%s" · Industry: %s · Personality: %s

%s

Create a landscape SVG wordmark lockup (viewBox="0 0 400 200"):
- Background: fill entire canvas with %s (light background — if user palette has "white" or "light", use #f8fafc or #ffffff; if dark palette, use %s).
- Left side (x=50, y=100 centred): small geometric mark (hexagon polygon or diamond, ~40×40px) filled with %s.
- Brand name "%s": text element at x="110" y="95", font-size="38", font-weight="900", fill="%s" (dark text for light bg, white for dark bg), font-family="Arial, sans-serif".
- Tagline "%s": text at x="112" y="120", font-size="13", font-weight="400", fill="%s", letter-spacing="0.08em", font-family="Arial, sans-serif".
- Thin horizontal rule: line from x1="112" to x2="340" at y="132" stroke="%s" stroke-width="1".
- IMPORTANT: if user colour has "navy" — use white (#ffffff) for brand name text and %s light background. If "white/light" — use dark (#0d1b2a) for brand name text.`,
			card.Name, tagline, industry, personality,
			colourInstr,
			primary, bg,
			accent,
			card.Name, "#0d1b2a",
			tagline, accent,
			accent, "#f0f4f8",
		)
	default:
		return fmt.Sprintf(
			`Create a square SVG logo (viewBox="0 0 200 200") for brand "%s".
%s
Simple geometric mark, brand initial "%s", tailored to the palette above.`,
			card.Name, colourInstr, initial,
		)
	}
}

// SVGToDataURI converts raw SVG string to a data URI for use in <img src=...>.
// Uses base64 encoding for maximum browser compatibility.
func SVGToDataURI(svg string) string {
	if svg == "" {
		return ""
	}
	svg = strings.TrimSpace(svg)

	// Ensure it starts with <svg
	if !strings.HasPrefix(svg, "<svg") {
		if idx := strings.Index(svg, "<svg"); idx >= 0 {
			svg = svg[idx:]
		} else {
			return ""
		}
	}

	// Reject truncated output instead of forcing a close tag.
	// Malformed XML renders as a broken image; empty lets the
	// frontend LogoMark fallback take over cleanly.
	if !strings.Contains(svg, "</svg>") {
		return ""
	}

	// Drop anything after the final </svg> (stray model commentary).
	if idx := strings.LastIndex(svg, "</svg>"); idx >= 0 {
		svg = svg[:idx+len("</svg>")]
	}

	// REQUIRED for data-URI SVG in <img>. Browsers parse it as a
	// standalone XML document and refuse to render without this.
	if !strings.Contains(svg, "xmlns=") {
		svg = strings.Replace(svg, "<svg", `<svg xmlns="http://www.w3.org/2000/svg"`, 1)
	}

	// Escape bare ampersands, which break XML parsing.
	svg = escapeBareAmps(svg)

	encoded := base64.StdEncoding.EncodeToString([]byte(svg))
	return "data:image/svg+xml;base64," + encoded
}

// escapeBareAmps escapes & that is not already part of an XML entity.
// Go's regexp is RE2 and has no negative lookahead, so we over-escape
// then restore the legitimate entities.
func escapeBareAmps(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}
	s = strings.ReplaceAll(s, "&", "&amp;")
	for _, e := range []string{"amp", "lt", "gt", "quot", "apos"} {
		s = strings.ReplaceAll(s, "&amp;"+e+";", "&"+e+";")
	}
	return numericEntityRe.ReplaceAllString(s, "&$1;")
}

// numericEntityRe restores numeric entities like &#39; and &#x27;.
var numericEntityRe = regexp.MustCompile(`&amp;(#\d+|#x[0-9a-fA-F]+);`)

// ── SVG Mood Board Tiles ──────────────────────────────────────────────────────

// BuildSVGTileSystemPrompt returns the system prompt for a single SVG mood board tile.
func BuildSVGTileSystemPrompt() string {
	return `You are NomVox, a brand design AI specialising in abstract SVG art.
Generate a SINGLE self-contained SVG brand mood tile.

STRICT RULES:
- Return ONLY the raw SVG markup starting with <svg — no markdown, no explanation, no code fences.
- viewBox="0 0 300 300" always.
- The SVG must be purely abstract — no text, no brand names, no letters.
- Use only geometric primitives and gradients to create visual mood.
- Maximum 50 lines of SVG.
- Use brand colours provided in the prompt.`
}

// BuildSVGMoodBoardPrompts returns 4 distinct SVG tile prompts for the mood board.
// CRITICAL: Must use the user's actual colour palette — not hardcoded brand colours.
func BuildSVGMoodBoardPrompts(card models.NameCard, intake models.IntakePayload) []string {
	// Extract actual colours from user input
	bg, primary, accent := buildColourContext(intake)
	colourMood := strings.TrimSpace(intake.ColorMood)
	if colourMood == "" {
		colourMood = "deep navy, white, teal"
	}

	personality := strings.TrimSpace(intake.Personality)
	if personality == "" {
		personality = "modern, aspirational"
	}
	industry := strings.TrimSpace(intake.Industry)
	if industry == "" {
		industry = "creative"
	}
	name := card.Name
	tagline := strings.TrimSpace(card.Tagline)

	// Colour context block used in every tile prompt
	colourCtx := fmt.Sprintf(
		`REQUIRED PALETTE from user input "%s": background=%s, marks=%s, accent=%s. Use ONLY these colours.`,
		colourMood, bg, primary, accent,
	)

	return []string{
		// Tile 1: Colour atmosphere / gradient field
		fmt.Sprintf(
			`Abstract SVG mood tile (viewBox="0 0 300 300") for brand "%s" — %s industry.
%s
Style: Pure colour atmosphere. Fill canvas with %s. Add 4–6 overlapping ellipses/circles
in %s and %s at 15–35%% opacity, varying sizes, suggesting depth and warmth.
No text, no letters. Evoke the feeling: "%s".`,
			name, industry, colourCtx, bg, accent, primary, personality,
		),
		// Tile 2: Geometric pattern grid
		fmt.Sprintf(
			`Abstract SVG geometric pattern tile (viewBox="0 0 300 300") for brand "%s".
%s
Style: Repeating geometric grid — use hexagons or diamonds tiled across the canvas.
Background: %s. Shape fill: %s at 40–70%% opacity. Accent stroke: %s.
Create a structured, purposeful pattern. Brand personality: %s.
No text, no letters.`,
			name, colourCtx, bg, accent, primary, personality,
		),
		// Tile 3: Motion lines / flow
		fmt.Sprintf(
			`Abstract SVG motion tile (viewBox="0 0 300 300") for brand "%s".
%s
Style: Flowing diagonal lines or curves suggesting movement and energy.
Background: %s. Use <path> bezier curves and arcs.
Line colours: %s (2px) and %s (1px) alternating, varying opacity 0.4–0.9.
The motion should feel: %s. No text, no letters.`,
			name, colourCtx, bg, accent, primary, personality,
		),
		// Tile 4: Focal brand mark echo — uses tagline as context
		fmt.Sprintf(
			`Abstract SVG focal tile (viewBox="0 0 300 300") for brand "%s" — "%s".
%s
Style: Single dominant focal shape expressing the brand essence.
Background: %s. Central circle or hexagon radius 80, filled with linear gradient %s→%s.
Surrounded by 3 concentric rings at radius 95, 110, 125 in %s at decreasing opacity (30%%, 20%%, 10%%).
The composition should feel %s and aspirational. No text, no letters.`,
			name, tagline, colourCtx, bg, accent, bg, accent, personality,
		),
	}
}

// ── Brand Persona Prompts ─────────────────────────────────────────────────────

// BuildPersonaSystemPrompt returns the system prompt for generating a BrandPersona.
func BuildPersonaSystemPrompt() string {
	return `You are NomVox, a brand strategy AI. Given a brand name and description, return ONLY a valid JSON object — no markdown, no explanation.

Schema (copy field names exactly):
{"age":28,"occupation":"Freelance photographer","voice":"Warm, direct, no jargon","reads":["Kinfolk","It's Nice That"],"never_says":["synergy","leverage"],"core_values":["authenticity","craft","community"]}

Rules:
- age: integer, the brand's personified age (18–60)
- occupation: one job title that embodies the brand
- voice: max 8 words describing communication style
- reads: 2–3 publications/media the brand personality would read
- never_says: 2–3 words/phrases the brand would never use
- core_values: 3 single-word values
Return ONLY the JSON object.`
}

// BuildPersonaUserPrompt returns the user turn for persona generation.
func BuildPersonaUserPrompt(card models.NameCard, intake models.IntakePayload) string {
	return fmt.Sprintf(
		`Brand: "%s". Tagline: "%s". Description: %s. Industry: %s. Personality: %s. Target audience: %s. Generate the brand persona.`,
		card.Name, card.Tagline,
		card.ShortDesc,
		strings.TrimSpace(intake.Industry),
		strings.TrimSpace(intake.Personality),
		strings.TrimSpace(intake.TargetAudience),
	)
}

// trailingCommaRe matches a comma followed by a closing } or ] — the most
// common LLM JSON malformation (e.g. ["a","b",]) that breaks encoding/json
// with "invalid character '}' after array element".
var trailingCommaRe = regexp.MustCompile(`,\s*([}\]])`)

// ParsePersona extracts a BrandPersona from raw LLM JSON output.
func ParsePersona(raw string) (*models.BrandPersona, error) {
	jsonStr := ExtractJSONObject(raw)
	if jsonStr == "" {
		return nil, fmt.Errorf("parsePersona: no JSON object found in: %s", TruncateStr(raw, 200))
	}
	// Repair trailing commas before closing braces/brackets so slightly
	// malformed LLM output still parses.
	jsonStr = trailingCommaRe.ReplaceAllString(jsonStr, "$1")

	var p models.BrandPersona
	if err := unmarshalPersona(jsonStr, &p); err != nil {
		return nil, fmt.Errorf("parsePersona: unmarshal: %w", err)
	}
	return &p, nil
}

func unmarshalPersona(jsonStr string, p *models.BrandPersona) error {
	// Use a raw map for lenient parsing
	var m map[string]interface{}
	if err := parseJSON([]byte(jsonStr), &m); err != nil {
		return err
	}

	if v, ok := m["age"].(float64); ok {
		p.Age = int(v)
	}
	p.Occupation = strField(m, "occupation")
	p.Voice = strField(m, "voice")
	p.Reads = strSliceField(m, "reads")
	p.NeverSays = strSliceField(m, "never_says")
	p.CoreValues = strSliceField(m, "core_values")
	return nil
}

// ── Landing page mockup prompt ────────────────────────────────────────────────

// BuildMockupSystemPrompt returns the system prompt for the landing page HTML generator.
func BuildMockupSystemPrompt() string {
	return `You are NomVox, a brand design AI. Generate a self-contained HTML hero section for a brand landing page.

Rules:
- Return ONLY valid HTML starting with <!DOCTYPE html> — no markdown fences, no explanation.
- Inline ALL CSS in a <style> block; no external stylesheets, no CDN links, no font imports.
- No JavaScript.
- Use system fonts: font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif.
- Use the EXACT hex colours provided — never substitute generic black or grey.
- NO decorative shapes. Do NOT add circles, hexagons, blobs, or abstract SVG paths.
  Do NOT place anything behind the headline text. The layout must stay clean and structural.
- The only images allowed are the logo placeholders given in the user prompt.
- Max 200 lines of HTML. Must render correctly in a sandboxed iframe at 75% scale.`
}

// BuildMockupUserPrompt returns the user turn for landing page generation.
// selectedLogoKey, selectedLogoStyle, and selectedLogoDataURI are optional;
// when a data URI is provided it is embedded directly into the nav <img> tag.
func BuildMockupUserPrompt(card models.NameCard, intake models.IntakePayload, selectedLogoKey, selectedLogoStyle string) string {
	return BuildMockupUserPromptWithLogo(card, intake, selectedLogoKey, selectedLogoStyle, "")
}

// BuildMockupUserPromptWithLogo is the full version that accepts a logo data URI.
func BuildMockupUserPromptWithLogo(card models.NameCard, intake models.IntakePayload, selectedLogoKey, selectedLogoStyle, logoDataURI string) string {
	colourHint := strings.TrimSpace(intake.ColorMood)
	if colourHint == "" {
		colourHint = "deep indigo #3b0764 as background, electric cyan #22d3ee as accent, white text"
	}
	personality := strings.TrimSpace(intake.Personality)
	if personality == "" {
		personality = "modern, clean"
	}
	styleHint := strings.TrimSpace(intake.Style)
	if styleHint == "" {
		styleHint = "minimal wordmark"
	}

	// Derive logo style description
	logoStyle := strings.TrimSpace(selectedLogoStyle)
	if logoStyle == "" {
		logoStyle = logoStyleDescription(selectedLogoKey)
	}

	// Logo clause — embed data URI if available, otherwise use text logo
	// Logo clauses — placeholders only. The real data URI is substituted
	// into the returned HTML by the handler.
	heroLogo := fmt.Sprintf(
		`<img src="%s" alt="%s logo" style="width:220px;height:220px;object-fit:contain;border-radius:20px">`,
		"__NOMVOX_LOGO_HERO__", card.Name)

	return fmt.Sprintf(
		`Create a hero landing page for brand "%s" using EXACTLY this structure.

Brand details:
- Tagline: "%s"
- Short description: %s
- Industry: %s
- Colour palette: %s
- Brand personality: %s
- Visual style: %s
- Target audience: %s

COLOURS — use these exactly, no generic black or grey:
- Page background: a dark tone derived from "%s"
- Headings: white
- Accent (buttons, eyebrow text, underline bar): the most vibrant colour in the palette

REQUIRED STRUCTURE — follow this layout precisely:

1. NAV BAR (full width, subtle bottom border):
   LEFT: the brand name "%s" in bold white text. No logo image in the nav.
   RIGHT: text links "About" and "Features" in the accent colour, then a
   solid accent-coloured "Get Started" button with white text.

2. HERO — a two-column flex row, generous padding:

   LEFT COLUMN (60%% width):
   - Small uppercase eyebrow text reading "%s" in the accent colour,
     letter-spacing 0.15em, font-size 0.75rem, bold.
   - H1 with the brand name "%s" — very large (clamp(2.5rem, 6vw, 4rem)),
     heavy weight, white, tight line-height.
   - A short horizontal bar underneath: 48px wide, 4px tall, accent colour.
   - The tagline in italics, wrapped in curly quotes, accent colour, ~1.15rem.
   - The short description in a muted light grey, ~1rem, max-width 32rem.
   - Two buttons side by side: a filled accent button labelled "Start with %s →"
     with white text, and an outlined button labelled "Learn More" with a
     1px light border and white text.

   RIGHT COLUMN (40%% width, centred):
   - %s
   - Beneath it, the brand personality words in uppercase, letter-spacing
	 0.1em, font-size 0.8rem, font-weight 700, in white at 75%% opacity
	 (rgba(255,255,255,0.75)), with 16px of space above. It must be clearly
	 legible against the dark background — never use a dark or saturated colour here.

3. BOTTOM STRIP (full width, thin top border, three equal cells divided by
   1px vertical borders, each centred, uppercase, 0.7rem, letter-spacing
   0.1em, accent colour):
   Cell 1: "%s"   Cell 2: "%s"   Cell 3: "%s"

Do not add any other sections, decorative shapes, or background graphics.`,
		card.Name,
		card.Tagline,
		card.ShortDesc,
		strings.TrimSpace(intake.Industry),
		colourHint,
		personality,
		logoStyle,
		strings.TrimSpace(intake.TargetAudience),
		colourHint,
		card.Name,
		strings.ToUpper(strings.TrimSpace(intake.Industry)),
		card.Name,
		card.Name,
		heroLogo,
		strings.ToUpper(strings.TrimSpace(intake.Industry)),
		strings.ToUpper(personality),
		strings.ToUpper(strings.TrimSpace(intake.TargetAudience)),
	)
}
