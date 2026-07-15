package ai

import (
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
func BuildMockupUserPrompt(card models.NameCard, intake models.IntakePayload) string {
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
- Target audience: %s

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
		personality,
		card.Name,
		card.Name,
	)
}
