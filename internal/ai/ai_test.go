package ai_test

import (
	"strings"
	"testing"

	"github.com/c-annabel/NomVoxBranding/internal/ai"
	"github.com/c-annabel/NomVoxBranding/internal/models"
)

// ── ExtractJSON ───────────────────────────────────────────────────────────────

func TestExtractJSON_CleanArray(t *testing.T) {
	input := `[{"name":"Veyrak","tagline":"Signal sent."}]`
	got := ai.ExtractJSON(input)
	if got != input {
		t.Errorf("expected clean array passthrough, got %q", got)
	}
}

func TestExtractJSON_WithMarkdownFences(t *testing.T) {
	input := "```json\n[{\"name\":\"Sorveth\"}]\n```"
	got := ai.ExtractJSON(input)
	if !strings.HasPrefix(got, "[") {
		t.Errorf("expected JSON array, got %q", got)
	}
}

func TestExtractJSON_TruncatedArray(t *testing.T) {
	// Simulates a token-truncated LLM response — missing closing ] and }
	input := `[{"name":"Qalune","tagline":"Silence forged."},{"name":"Myrrix"`
	got := ai.ExtractJSON(input)
	// After repair, should end with ]
	if !strings.HasSuffix(strings.TrimSpace(got), "]") {
		t.Errorf("expected repaired array ending with ], got %q", got)
	}
}

func TestExtractJSON_NoArray(t *testing.T) {
	// If there is no [ at all, returns the original text (caller will fail gracefully)
	input := `{"name":"not an array"}`
	got := ai.ExtractJSON(input)
	// Should return the original (no array found — start == -1 returns text)
	if got == "" {
		t.Error("expected non-empty return for no-array input")
	}
}

// ── ExtractJSONObject ─────────────────────────────────────────────────────────

func TestExtractJSONObject_CleanObject(t *testing.T) {
	input := `{"age":28,"occupation":"Designer"}`
	got := ai.ExtractJSONObject(input)
	if got != input {
		t.Errorf("expected clean passthrough, got %q", got)
	}
}

func TestExtractJSONObject_WithMarkdownFences(t *testing.T) {
	input := "```json\n{\"age\":32}\n```"
	got := ai.ExtractJSONObject(input)
	if !strings.HasPrefix(got, "{") || !strings.HasSuffix(got, "}") {
		t.Errorf("expected JSON object, got %q", got)
	}
}

func TestExtractJSONObject_WithPreamble(t *testing.T) {
	input := `Here is the persona: {"age":25,"voice":"Calm, direct"} Thank you.`
	got := ai.ExtractJSONObject(input)
	if !strings.HasPrefix(got, "{") {
		t.Errorf("expected extracted object, got %q", got)
	}
}

// ── ParseNameCards ────────────────────────────────────────────────────────────

func TestParseNameCards_ValidJSON(t *testing.T) {
	input := `[
		{
			"name": "Veyrak",
			"tagline": "Signal sent.",
			"tone_reasoning": "Bold plosive sound",
			"style_tags": ["tech", "minimal"],
			"short_desc": "A brand that moves.",
			"long_desc": "For those who act first.",
			"origin_story": "V from Sanskrit vega + rak from Old Norse rek.",
			"score": {
				"memorability": 8,
				"spellability": 9,
				"global_safety": 8,
				"squatter_risk": "Low",
				"mem_reasoning": "Short, punchy.",
				"spell_reasoning": "Phonetic.",
				"global_reasoning": "No clashes.",
				"squatter_reasoning": "Invented word."
			},
			"voice_samples": {
				"instagram_caption": "Move like Veyrak.",
				"email_subject": "Veyrak is live",
				"not_found_message": "This page moved fast."
			}
		}
	]`

	cards, err := ai.ParseNameCards(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Name != "Veyrak" {
		t.Errorf("expected name Veyrak, got %q", cards[0].Name)
	}
	if cards[0].Score.Memorability != 8 {
		t.Errorf("expected memorability 8, got %d", cards[0].Score.Memorability)
	}
	if cards[0].Score.SquatterRisk != "Low" {
		t.Errorf("expected squatter risk Low, got %q", cards[0].Score.SquatterRisk)
	}
	if cards[0].VoiceSamples.EmailSubject != "Veyrak is live" {
		t.Errorf("expected email subject, got %q", cards[0].VoiceSamples.EmailSubject)
	}
}

func TestParseNameCards_EmptyName_Filtered(t *testing.T) {
	input := `[{"name": "", "tagline": "test"}, {"name": "Sorveth", "tagline": "ok"}]`
	cards, err := ai.ParseNameCards(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card after filtering empty name, got %d", len(cards))
	}
	if cards[0].Name != "Sorveth" {
		t.Errorf("expected Sorveth, got %q", cards[0].Name)
	}
}

func TestParseNameCards_InvalidJSON(t *testing.T) {
	_, err := ai.ParseNameCards("not json at all")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// ── BuildSystemPrompt ────────────────────────────────────────────────────────

func TestBuildSystemPrompt_NoSession(t *testing.T) {
	got := ai.BuildSystemPrompt(nil)
	if !strings.Contains(got, "NomVox") {
		t.Error("expected NomVox in system prompt")
	}
	if !strings.Contains(got, "JSON array") {
		t.Error("expected JSON array instruction in system prompt")
	}
}

func TestBuildSystemPrompt_WithSession(t *testing.T) {
	sess := &models.BrandSession{
		SessionID: "test-123",
		Liked:     []string{"Veyrak", "Sorveth"},
		Rejected:  []string{"EcoBrand"},
		DirectionNotes: []string{"prefer abstract names"},
		SliderPlayful:  0.8,
		SliderAbstract: 0.6,
	}
	got := ai.BuildSystemPrompt(sess)

	// Should include rejected names
	if !strings.Contains(got, "EcoBrand") {
		t.Error("expected rejected name EcoBrand in system prompt")
	}
	// Should include liked names
	if !strings.Contains(got, "Veyrak") {
		t.Error("expected liked name Veyrak in system prompt")
	}
	// Should include direction note
	if !strings.Contains(got, "abstract names") {
		t.Error("expected direction note in system prompt")
	}
	// Should include slider direction
	if !strings.Contains(got, "playful") {
		t.Error("expected playful tone direction in system prompt")
	}
}

// ── BuildUserPrompt ───────────────────────────────────────────────────────────

func TestBuildUserPrompt_MinimalInput(t *testing.T) {
	p := models.IntakePayload{CoreIdea: "eco-friendly coffee brand"}
	got := ai.BuildUserPrompt(p, 4)
	if !strings.Contains(got, "4 brand names") {
		t.Errorf("expected count in prompt, got %q", got)
	}
	if !strings.Contains(got, "eco-friendly coffee brand") {
		t.Errorf("expected core idea in prompt, got %q", got)
	}
}

func TestBuildUserPrompt_AllFields(t *testing.T) {
	p := models.IntakePayload{
		CoreIdea:       "sustainable streetwear",
		TargetAudience: "Gen Z",
		Personality:    "bold",
		Style:          "abstract",
		Industry:       "fashion",
		ColorMood:      "black and acid green",
		NameLength:     "under 8 chars",
		Avoid:          "eco, green",
	}
	got := ai.BuildUserPrompt(p, 4)
	if !strings.Contains(got, "Gen Z") {
		t.Error("expected target audience in prompt")
	}
	if !strings.Contains(got, "Avoid") {
		t.Error("expected avoid clause in prompt")
	}
	if !strings.Contains(got, "JSON array") {
		t.Error("expected JSON instruction in prompt")
	}
}

// ── TruncateStr ───────────────────────────────────────────────────────────────

func TestTruncateStr(t *testing.T) {
	s := "hello world this is a long string"
	got := ai.TruncateStr(s, 10)
	if len(got) > 13 { // 10 + "…"
		t.Errorf("expected truncated string, got %q (len=%d)", got, len(got))
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("expected ellipsis suffix, got %q", got)
	}
}

func TestTruncateStr_ShortString(t *testing.T) {
	s := "hi"
	got := ai.TruncateStr(s, 100)
	if got != s {
		t.Errorf("expected passthrough for short string, got %q", got)
	}
}
