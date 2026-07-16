package ai

import (
	"encoding/base64"
	"fmt"
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

// MoodBoardPrompt returns a contextually rich Imagen 3 prompt for a 4-panel mood board.
// It fuses: brand name + its invented etymology, tagline meaning, chosen logo style,
// all intake fields (industry, audience, personality, colour mood, style), and any
// vision context from an uploaded reference image.
// The resulting mood board visually expresses the brand world — not generic stock photography.
func MoodBoardPrompt(card models.NameCard, intake models.IntakePayload, vision *models.VisionContext, selectedLogoKey, selectedLogoStyle string) string {
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
	avoidHint := strings.TrimSpace(intake.Avoid)

	// Build the prompt
	avoidClause := ""
	if avoidHint != "" {
		avoidClause = fmt.Sprintf("Do NOT include: %s. ", avoidHint)
	}

	return fmt.Sprintf(
		`Brand identity mood board for "%s" — a %s brand. `+
			`Name meaning & origin: %s. `+
			`Brand tagline: "%s". `+
			`Target audience: %s. `+
			`Brand personality: %s. `+
			`Visual aesthetic: %s. `+
			`Colour palette: %s. `+
			`Chosen logo style: %s. `+
			`\n`+
			`Create a 2×2 grid of 4 cohesive but distinct photographic panels that DIRECTLY express this brand's world: `+
			`(1) A material or texture close-up that embodies the brand colour palette — `+
			`dominant hues from "%s", surface quality matching "%s" personality — `+
			`(2) An editorial lifestyle scene showing the brand's ideal customer (%s) `+
			`in their natural environment, product or concept implied — `+
			`(3) An abstract macro or colour study — ink in water, smoke, light refraction — `+
			`using ONLY the brand's colour palette (%s) — `+
			`(4) An environment or architectural detail that a "%s" brand would inhabit — `+
			`space, scale, and light matching "%s" aesthetic. `+
			`\n`+
			`Strict rules: magazine-quality commercial photography, no text, no logos, no human faces, `+
			`no generic stock imagery. Every panel must feel like it belongs to THIS specific brand. `+
			`%s`+
			`Crisp focus, natural or dramatic lighting, high resolution, `+
			`colour palette must be visually consistent across all 4 panels.`,
		card.Name, industry,
		originContext,
		card.Tagline,
		audience,
		moodHint,
		styleHint,
		colourHint,
		logoStyle,
		// panel 1
		colourHint, moodHint,
		// panel 2
		audience,
		// panel 3
		colourHint,
		// panel 4
		card.Name, styleHint,
		// avoid
		avoidClause,
	)
}

// LogoConceptPrompt returns an Imagen 3 prompt for one of three logo concept types.
// Each type uses a deliberately different visual style so the three logos look distinct.
// logoType: "profile" | "app" | "business"
func LogoConceptPrompt(card models.NameCard, intake models.IntakePayload, logoType string) string {
	colourHint := strings.TrimSpace(intake.ColorMood)
	if colourHint == "" {
		colourHint = "deep indigo and electric cyan"
	}
	personality := strings.TrimSpace(intake.Personality)
	if personality == "" {
		personality = "modern and distinctive"
	}
	industry := strings.TrimSpace(intake.Industry)
	if industry == "" {
		industry = "technology"
	}

	switch logoType {
	case "profile":
		// Style A: bold geometric monogram — a single strong abstract mark
		return fmt.Sprintf(
			`Minimalist geometric logo mark for brand "%s" in the %s industry. `+
				`Single abstract symbol — NOT text, NOT a wordmark. `+
				`Inspired by: %s. Color palette: %s. `+
				`Style: flat vector, bold geometric shapes, strong negative space, Bauhaus-inspired. `+
				`White background, centered mark, no gradients, no shadows. `+
				`The shape should suggest motion, growth, or the brand's core concept through pure geometry. `+
				`Professional, iconic, scalable to 32×32 pixels. No text, no letters, no words.`,
			card.Name, industry, personality, colourHint,
		)
	case "app":
		// Style B: vibrant gradient icon — rich depth, app-store aesthetic
		return fmt.Sprintf(
			`App store icon for brand "%s" — %s. Industry: %s. `+
				`Style: rounded-square canvas with deep gradient background (%s tones). `+
				`A single bold abstract icon or symbol floats centred with soft drop shadow. `+
				`Rich colour depth, glassmorphism or neon glow accent. `+
				`Personality: %s. High detail, 3D-feeling depth without being photorealistic. `+
				`No text, no letters. Looks premium in a 256×256 app store grid.`,
			card.Name, card.Tagline, industry, colourHint, personality,
		)
	case "business":
		// Style C: horizontal wordmark lockup — text + icon side by side on clean background
		return fmt.Sprintf(
			`Horizontal brand wordmark lockup for "%s" — "%s". `+
				`Industry: %s. Personality: %s. `+
				`Style: clean white or very light background, sans-serif bold wordmark on the right, `+
				`small abstract icon mark on the left, separated by thin rule or negative space. `+
				`Colour palette: %s accents on white. `+
				`Corporate print quality, business card format (landscape 3.5"×2"). `+
				`Professional typographic hierarchy, no gradients on background, `+
				`single accent colour on the icon. Looks like a Fortune 500 rebrand.`,
			card.Name, card.Tagline, industry, personality, colourHint,
		)
	default:
		return fmt.Sprintf(`Clean modern logo for brand "%s". %s colour palette. Vector style, white background.`, card.Name, colourHint)
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
	// Trim any leading/trailing whitespace and ensure it starts with <svg
	svg = strings.TrimSpace(svg)
	if !strings.HasPrefix(svg, "<svg") {
		// Try to extract just the SVG block
		if idx := strings.Index(svg, "<svg"); idx >= 0 {
			svg = svg[idx:]
		} else {
			return ""
		}
	}
	// Close tag safety check
	if !strings.Contains(svg, "</svg>") {
		svg += "</svg>"
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(svg))
	return "data:image/svg+xml;base64," + encoded
}

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

// ParsePersona extracts a BrandPersona from raw LLM JSON output.
func ParsePersona(raw string) (*models.BrandPersona, error) {
	jsonStr := ExtractJSONObject(raw)
	if jsonStr == "" {
		return nil, fmt.Errorf("parsePersona: no JSON object found in: %s", TruncateStr(raw, 200))
	}

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
	return `You are NomVox, a brand design AI. Generate a beautiful, self-contained HTML hero section for a brand landing page.

Rules:
- Return ONLY valid HTML starting with <!DOCTYPE html> — no markdown fences, no explanation.
- Inline ALL CSS in a <style> block; no external stylesheets, no CDN links, no Google Fonts imports.
- No JavaScript.
- Use system fonts: font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif.
- The design MUST use the EXACT hex colours provided — never substitute generic black/dark-grey.
- If a logo data URI is provided, embed it with <img src="LOGO_DATA_URI" alt="logo"> in the nav bar.
- Include: nav bar with logo, large hero headline (H1), italic tagline, short description, one CTA button.
- The hero section fills viewport height with the specified background gradient using the brand colours.
- Use the provided personality to drive ALL decisions: font weight, spacing, border-radius, CTA shape.
- If "playful" or "bold" — large fonts, bright accent, rounded corners (border-radius: 32px on CTA).
- If "minimal" or "premium" — thin fonts, generous whitespace, sharp lines (border-radius: 4px on CTA).
- If "earthy" or "organic" — warm gradients, soft radii, natural tones.
- If "edgy" or "dark" — high contrast, angular shapes, electric accent on dark bg.
- Include an inline SVG accent shape in the hero (a decorative circle, hexagon, or abstract path)
  that reuses the brand's accent colour — this makes the page feel visually designed not plain text.
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
	logoClause := fmt.Sprintf(`In the nav bar, display the brand name "%s" as a bold text logo on the left.`, card.Name)
	if logoDataURI != "" {
		logoClause = fmt.Sprintf(
			`In the nav bar, display this logo image on the left: <img src="%s" alt="%s logo" style="height:40px;width:auto;object-fit:contain;vertical-align:middle"> followed by the brand name "%s" in bold text. The logo style is: %s.`,
			logoDataURI, card.Name, card.Name, logoStyle,
		)
	}

	// Colour extraction for explicit hex usage
	colourInstructions := fmt.Sprintf(
		`CRITICAL: Use these exact colours — do NOT default to black or generic grey.
Primary background: derive a rich gradient from "%s".
All headings: white or the lightest colour from the palette.
CTA button: the most vibrant accent colour from the palette as background, white text.
Decorative SVG element: the accent colour at 30%% opacity.`,
		colourHint,
	)

	return fmt.Sprintf(
		`Create a hero landing page for brand "%s".

%s

Brand details:
- Tagline: "%s"
- Short description: %s
- Industry: %s
- Colour palette: %s
- Brand personality: %s
- Visual style: %s (logo aesthetic) — echo this in typography weight, button shape, and spacing.
- Target audience: %s

%s

Design requirements:
1. BACKGROUND: rich CSS gradient using the brand colour palette — NO pure black (#000), NO generic grey.
2. NAV BAR: %s — two nav links ("About", "Get Started") on the right in the accent colour.
3. HERO: large H1 brand name, italic tagline below, short description (max 20 words), CTA button.
4. SVG DECORATION: include one inline SVG shape (60px circle or hexagon) in accent colour at 25%% opacity, floating to the side of the hero text.
5. CTA BUTTON: label "%s — Get Started", background = accent colour, white text, border-radius matches personality.
6. Ensure the page looks DESIGNED not plain — use gradients, letter-spacing on tagline, subtle text-shadow on H1.`,
		card.Name,
		logoClause,
		card.Tagline,
		card.ShortDesc,
		strings.TrimSpace(intake.Industry),
		colourHint,
		personality,
		logoStyle,
		strings.TrimSpace(intake.TargetAudience),
		colourInstructions,
		logoClause,
		card.Name,
	)
}
