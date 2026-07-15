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

// ── SVG Logo Generation (Granite / watsonx fallback) ─────────────────────────
//
// When Gemini image generation is unavailable, watsonx generates inline SVG logos.
// Each logoType receives a distinct prompt that produces valid, self-contained SVG.

// BuildSVGLogoSystemPrompt returns the system prompt for SVG logo generation.
func BuildSVGLogoSystemPrompt() string {
	return `You are NomVox, a brand design AI specialising in vector logo design.
Generate a SINGLE self-contained SVG logo based on the brand description.

STRICT RULES:
- Return ONLY the raw SVG markup starting with <svg — no markdown, no explanation, no code fences.
- The SVG must be fully self-contained: all styles inline, no external references.
- viewBox="0 0 200 200" for square logos (profile, app). viewBox="0 0 400 200" for landscape (business).
- No <image> tags, no foreignObject, no scripts.
- Use only SVG primitives: <rect>, <circle>, <ellipse>, <path>, <polygon>, <polyline>, <text>, <g>, <defs>, <linearGradient>, <radialGradient>, <filter>.
- Maximum 80 lines of SVG.
- The result must render correctly in a browser <img> tag with src="data:image/svg+xml;base64,..."`
}

// BuildSVGLogoUserPrompt returns the user turn for SVG logo generation.
// logoType: "profile" | "app" | "business"
func BuildSVGLogoUserPrompt(card models.NameCard, intake models.IntakePayload, logoType string) string {
	colourHint := strings.TrimSpace(intake.ColorMood)
	if colourHint == "" {
		colourHint = "deep indigo #3b0764 and electric cyan #22d3ee"
	}
	personality := strings.TrimSpace(intake.Personality)
	if personality == "" {
		personality = "modern and distinctive"
	}
	industry := strings.TrimSpace(intake.Industry)
	if industry == "" {
		industry = "technology"
	}
	initial := ""
	if len(card.Name) > 0 {
		initial = string([]rune(card.Name)[:1])
	}
	initials := initial
	if len([]rune(card.Name)) > 1 {
		initials = string([]rune(card.Name)[:2])
	}

	switch logoType {
	case "profile":
		return fmt.Sprintf(
			`Create a square SVG logo mark (viewBox="0 0 200 200") for brand "%s" in the %s industry.
Style: Bauhaus-inspired geometric abstract mark. Bold shapes, strong negative space.
Brand personality: %s. Colour palette: %s.
Design: Draw ONE central abstract geometric shape (hexagon, diamond, overlapping circles, or star) centred at 100,100.
Fill with a gradient using the brand colours. Add a subtle outer ring or frame.
Include the brand initial "%s" centred in the shape using a bold sans-serif font, white fill.
Dark background (#0b0f19 or brand dark tone).
NO brand name text below — only the mark and initial.`,
			card.Name, industry, personality, colourHint, initial,
		)
	case "app":
		return fmt.Sprintf(
			`Create a square SVG app icon (viewBox="0 0 200 200") for brand "%s" — "%s". Industry: %s.
Style: App store icon — rounded-rectangle frame with gradient background. Glassmorphic feel.
Brand personality: %s. Colour palette: %s.
Design: Draw a rounded rectangle (rx="36") filling most of the canvas with a gradient background.
Inside: draw a bold abstract symbol or monogram "%s" centred, white or light coloured.
Add a subtle radial gradient glow in the centre. Use deep rich colours for the background gradient.
The icon should look premium and recognisable at 48×48px.`,
			card.Name, card.Tagline, industry, personality, colourHint, initials,
		)
	case "business":
		return fmt.Sprintf(
			`Create a landscape SVG wordmark lockup (viewBox="0 0 400 200") for brand "%s" — "%s".
Industry: %s. Brand personality: %s. Colour palette: %s.
Style: Clean horizontal lockup on white/light background for print (business cards, letterhead).
Design:
1. Left side (x=40): Draw a small abstract geometric icon mark (hexagon or diamond, 48×48) filled with brand colour gradient.
2. Right of icon: Brand name "%s" in bold sans-serif, large (font-size="36"), dark colour (#1a0a2e or similar).
3. Below brand name: tagline "%s" in lighter font (font-size="14"), brand accent colour, tracking-wide.
4. Below everything: a thin horizontal rule line in brand accent colour.
Light background (#faf9f7 or #f5f0f8). Total composition centred vertically in the canvas.`,
			card.Name, card.Tagline, industry, personality, colourHint,
			card.Name, card.Tagline,
		)
	default:
		return fmt.Sprintf(
			`Create a square SVG logo (viewBox="0 0 200 200") for brand "%s". %s colour palette. Geometric style.`,
			card.Name, colourHint,
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
// Each tile represents a different facet of the brand's visual world.
func BuildSVGMoodBoardPrompts(card models.NameCard, intake models.IntakePayload) []string {
	colour1 := "#8B5CF6" // pulse purple
	colour2 := "#22d3ee" // signal cyan
	colour3 := "#0b0f19" // void
	if c := strings.TrimSpace(intake.ColorMood); c != "" {
		// Use the provided colour as a hint for accent
		colour1 = "#8B5CF6"
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

	return []string{
		// Tile 1: Colour field / texture
		fmt.Sprintf(
			`Create an abstract SVG mood tile (viewBox="0 0 300 300") for brand "%s" in %s industry.
Style: Colour field — no shapes, pure gradient atmosphere.
Use: deep background %s, radiating circles or concentric rings in %s and %s.
Multiple overlapping semi-transparent ellipses/circles creating depth and texture.
NO text, NO letters. Pure abstract colour study.`,
			name, industry, colour3, colour1, colour2,
		),
		// Tile 2: Geometric pattern
		fmt.Sprintf(
			`Create an abstract SVG geometric pattern tile (viewBox="0 0 300 300") for brand "%s".
Style: Repeating geometric forms — hexagons, triangles, or diamonds in a grid.
Colours: %s background, %s and %s shapes with varying opacity (0.3 to 0.9).
Brand personality: %s. Pure geometry, no text, no letters.
Create visual rhythm through size variation and overlapping.`,
			name, colour3, colour1, colour2, personality,
		),
		// Tile 3: Abstract marks / motion
		fmt.Sprintf(
			`Create an abstract SVG motion tile (viewBox="0 0 300 300") for brand "%s".
Style: Flowing curves or diagonal lines suggesting motion and energy.
Use curved <path> elements, arcs, and bezier curves.
Colours: %s background, %s and %s strokes of varying width (1px to 8px).
No text, no letters. Express brand personality: %s.`,
			name, colour3, colour1, colour2, personality,
		),
		// Tile 4: Brand mark echo
		fmt.Sprintf(
			`Create an abstract SVG brand echo tile (viewBox="0 0 300 300") for brand "%s" in %s.
Style: Central focal point — one large central shape (circle or hexagon) filled with gradient %s→%s.
Surrounded by diminishing rings or orbits in lighter tones.
Dark background %s. No text, no letters. Aspirational, premium feel.`,
			name, industry, colour1, colour2, colour3,
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
- The design MUST visually reflect the exact brand colours and style described — do not use generic dark/grey.
- Include: nav bar (brand name as text logo), large hero headline, tagline subheading, short description paragraph, one CTA button.
- The hero section should fill the viewport height with a strong branded background colour.
- Use the provided personality and style to drive ALL design decisions: layout, typography weight, spacing, border-radius, button style.
- If "playful" or "bold" — use large fonts, bright accents, rounded corners.
- If "minimal" or "premium" — use thin fonts, lots of whitespace, clean lines.
- If "earthy" or "organic" — use warm tones, soft backgrounds, natural textures via CSS gradients.
- Max 250 lines of HTML. Must render correctly in a sandboxed iframe.`
}

// BuildMockupUserPrompt returns the user turn for landing page generation.
// selectedLogoKey ("profile"|"app"|"business") and selectedLogoStyle are optional;
// when provided they add logo context to the design directives.
func BuildMockupUserPrompt(card models.NameCard, intake models.IntakePayload, selectedLogoKey, selectedLogoStyle string) string {
	colourHint := intake.ColorMood
	if colourHint == "" {
		colourHint = "deep navy background, white text, electric blue accent"
	}
	personality := strings.TrimSpace(intake.Personality)
	if personality == "" {
		personality = "modern, clean"
	}
	styleHint := strings.TrimSpace(intake.Style)
	if styleHint == "" {
		styleHint = "minimal wordmark"
	}

	// Derive logo context clause
	logoCtx := ""
	logoStyle := strings.TrimSpace(selectedLogoStyle)
	if logoStyle == "" {
		logoStyle = logoStyleDescription(selectedLogoKey)
	}
	if logoStyle != "" && logoStyle != "clean modern mark" {
		logoCtx = fmt.Sprintf("\n- Logo style chosen: %s — echo this aesthetic in nav typography, button shape, and icon use.", logoStyle)
	}

	return fmt.Sprintf(
		`Create a hero landing page for brand "%s".

Brand details:
- Tagline: "%s"
- Short description: %s
- Long description: %s
- Industry: %s
- Colour palette: %s
- Brand personality: %s
- Visual style: %s
- Target audience: %s%s

Design requirements:
- Background colour must directly use the specified colour palette (not generic dark grey).
- Typography weight and style must match the personality (%s).
- CTA button label: "%s — Get Started"
- Navigation: brand name "%s" as text logo on the left, two placeholder nav links on the right.
- Hero layout: centred content, large H1 in brand font, italic tagline, short description, CTA button.
- Include a subtle brand-coloured gradient or background pattern that reflects the industry and style.`,
		card.Name,
		card.Tagline,
		card.ShortDesc,
		card.LongDesc,
		strings.TrimSpace(intake.Industry),
		colourHint,
		personality,
		styleHint,
		strings.TrimSpace(intake.TargetAudience),
		logoCtx,
		personality,
		card.Name,
		card.Name,
	)
}
