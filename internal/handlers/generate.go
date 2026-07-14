package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/c-annabel/NomVoxBranding/internal/models"
)

// GenerateHandler handles POST /api/generate.
// In ST-03 this will call the Granite LLM client and stream name cards.
// For now it validates the payload and returns the assembled prompt string
// so the intake form can be tested end-to-end before the LLM is wired up.
func GenerateHandler(w http.ResponseWriter, r *http.Request) {
	var payload models.IntakePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Validate required field.
	if strings.TrimSpace(payload.CoreIdea) == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "core_idea is required",
			"field": "core_idea",
		})
		return
	}

	// Assemble deterministic LLM prompt from all filled fields.
	// ST-03 will pass this to the Granite client.
	prompt := assembleLLMPrompt(payload)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"prompt": prompt,
	})
}

// assembleLLMPrompt builds a rich, deterministic prompt string from the
// structured intake payload. Only appends clauses for non-empty fields.
func assembleLLMPrompt(p models.IntakePayload) string {
	var b strings.Builder
	b.WriteString("Generate brand names for a ")
	b.WriteString(strings.TrimSpace(p.CoreIdea))

	if v := strings.TrimSpace(p.TargetAudience); v != "" {
		b.WriteString(", targeting ")
		b.WriteString(v)
	}
	if v := strings.TrimSpace(p.Personality); v != "" {
		b.WriteString(", personality: ")
		b.WriteString(v)
	}
	if v := strings.TrimSpace(p.Style); v != "" {
		b.WriteString(", style: ")
		b.WriteString(v)
	}
	if v := strings.TrimSpace(p.Industry); v != "" {
		b.WriteString(", industry: ")
		b.WriteString(v)
	}
	if v := strings.TrimSpace(p.ColorMood); v != "" {
		b.WriteString(", colour mood: ")
		b.WriteString(v)
	}
	if v := strings.TrimSpace(p.NameLength); v != "" {
		b.WriteString(", name length preference: ")
		b.WriteString(v)
	}
	if v := strings.TrimSpace(p.Avoid); v != "" {
		b.WriteString(". Avoid: ")
		b.WriteString(v)
	}
	b.WriteString(".")

	return b.String()
}
