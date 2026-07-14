package models

// IntakePayload is the structured input from the keyword intake form.
// CoreIdea is required; all other fields are optional but improve generation quality.
type IntakePayload struct {
	CoreIdea        string `json:"core_idea"`         // required
	TargetAudience  string `json:"target_audience"`   // optional
	Personality     string `json:"personality"`        // optional
	Style           string `json:"style"`              // optional
	Industry        string `json:"industry"`           // optional
	ColorMood       string `json:"color_mood"`         // optional
	NameLength      string `json:"name_length"`        // optional
	Avoid           string `json:"avoid"`              // optional
}

// BrandScore holds the multi-dimension quality score for a generated name.
type BrandScore struct {
	Memorability  int    `json:"memorability"`   // 1–10
	Spellability  int    `json:"spellability"`   // 1–10
	GlobalSafety  int    `json:"global_safety"`  // 1–10
	SquatterRisk  string `json:"squatter_risk"`  // "Low" | "Medium" | "High"
	MemReasoning  string `json:"mem_reasoning"`
	SpellReasoning string `json:"spell_reasoning"`
	GlobalReasoning string `json:"global_reasoning"`
	SquatterReasoning string `json:"squatter_reasoning"`
}

// VoiceSamples holds brand copy snippets generated in the brand's voice.
type VoiceSamples struct {
	InstagramCaption  string `json:"instagram_caption"`
	EmailSubject      string `json:"email_subject"`
	NotFoundMessage   string `json:"not_found_message"` // 404 page copy
}

// NameCard is a single generated brand name with all enrichment data.
type NameCard struct {
	Name          string       `json:"name"`
	Tagline       string       `json:"tagline"`
	ToneReasoning string       `json:"tone_reasoning"`
	StyleTags     []string     `json:"style_tags"`
	ShortDesc     string       `json:"short_desc"`
	LongDesc      string       `json:"long_desc"`
	OriginStory   string       `json:"origin_story"`
	Score         BrandScore   `json:"score"`
	VoiceSamples  VoiceSamples `json:"voice_samples"`
}

// PlatformResult holds the raw availability boolean for each platform.
// Available = true means the handle/domain is free to register.
// Unknown = true means the probe returned an inconclusive status (e.g. 429).
type PlatformResult struct {
	Domain    bool `json:"domain"`
	Instagram bool `json:"instagram"`
	X         bool `json:"x"`
	TikTok    bool `json:"tiktok"`
	Threads   bool `json:"threads"`
	YouTube   bool `json:"youtube"`

	// Unknown flags — set when probe returned 429/5xx
	DomainUnknown    bool `json:"domain_unknown"`
	InstagramUnknown bool `json:"instagram_unknown"`
	XUnknown         bool `json:"x_unknown"`
	TikTokUnknown    bool `json:"tiktok_unknown"`
	ThreadsUnknown   bool `json:"threads_unknown"`
	YouTubeUnknown   bool `json:"youtube_unknown"`
}

// AvailabilityResult combines the name, its platform probe results, and the
// computed availability score. Only names with Score >= 80.0 are returned
// to the frontend.
type AvailabilityResult struct {
	Name     string         `json:"name"`
	Probes   PlatformResult `json:"probes"`
	Score    float64        `json:"score"`    // 0–100
	Passes   bool           `json:"passes"`   // score >= 80.0
	Radar    string         `json:"radar"`    // competitor name radar warning, if any
}

// VisionContext holds the structured output from Gemini 2.0 Flash vision
// analysis of the user's uploaded reference image.
type VisionContext struct {
	Colours  []string `json:"colours"`  // HEX values, e.g. ["#3d6b4f","#f5efe0"]
	Mood     []string `json:"mood"`     // e.g. ["warm","organic","grounded"]
	Style    []string `json:"style"`    // e.g. ["earthy minimalism","flat lay"]
	Textures []string `json:"textures"` // e.g. ["linen","matte ceramic"]
}

// VisualPack holds all generated visual asset URLs for a selected brand name.
type VisualPack struct {
	MoodBoard []string `json:"mood_board"` // 4 Vercel Blob URLs
	LogoProfile  string `json:"logo_profile"`
	LogoApp      string `json:"logo_app"`
	LogoBusiness string `json:"logo_business"`
}

// BrandPersona is the "brand as a person" character card.
type BrandPersona struct {
	Age        int      `json:"age"`
	Occupation string   `json:"occupation"`
	Voice      string   `json:"voice"`
	Reads      []string `json:"reads"`
	NeverSays  []string `json:"never_says"`
	CoreValues []string `json:"core_values"`
}

// BrandSession is the full Redis-persisted session object.
// It accumulates all user feedback across the session so that every
// subsequent Granite call has complete context.
type BrandSession struct {
	SessionID      string        `json:"session_id"`
	Intake         IntakePayload `json:"intake"`
	Liked          []string      `json:"liked"`
	Rejected       []string      `json:"rejected"`
	DirectionNotes []string      `json:"direction_notes"`
	AllGenerated   []string      `json:"all_generated"`
	VisualNotes    []string      `json:"visual_notes"`
	SliderPlayful  float64       `json:"slider_playful"`  // 0=premium, 1=playful
	SliderAbstract float64       `json:"slider_abstract"` // 0=descriptive, 1=abstract
	SelectedName   string        `json:"selected_name"`
	VisionCtx      *VisionContext `json:"vision_context,omitempty"`
	Visuals        *VisualPack   `json:"visuals,omitempty"`
	Persona        *BrandPersona `json:"persona,omitempty"`
}
