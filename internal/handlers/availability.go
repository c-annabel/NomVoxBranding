package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/c-annabel/NomVoxBranding/internal/ai"
	"github.com/c-annabel/NomVoxBranding/internal/availability"
	"github.com/c-annabel/NomVoxBranding/internal/models"
)

// ── POST /api/availability ─────────────────────────────────────────────────────
// Accepts a list of names, probes all platforms in parallel, applies the 80% gate,
// and returns passing results (+ closest-match partials if zero pass).
// Also runs Competitor Name Radar on all names via a single Granite call.

type availabilityRequest struct {
	SessionID string   `json:"session_id,omitempty"`
	Names     []string `json:"names"`
}

type availabilityResponse struct {
	Passing  []models.AvailabilityResult `json:"passing"`
	Partials []models.AvailabilityResult `json:"partials"` // top-3 closest matches when zero pass
}

func CheckAvailability(w http.ResponseWriter, r *http.Request) {
	var req availabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if len(req.Names) == 0 {
		jsonError(w, "names array is required", http.StatusBadRequest)
		return
	}
	// Cap to prevent abuse
	if len(req.Names) > 20 {
		req.Names = req.Names[:20]
	}

	// ── Probe all platforms ────────────────────────────────────────
	passing, partials := availability.CheckNames(req.Names)

	// ── Competitor radar (Granite second-pass) ─────────────────────
	allNames := append(passing, partials...)
	if len(allNames) > 0 {
		radarMap := runCompetitorRadar(r.Context(), namesFromResults(allNames))
		for i, res := range passing {
			passing[i].Radar = radarMap[res.Name]
		}
		for i, res := range partials {
			partials[i].Radar = radarMap[res.Name]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availabilityResponse{
		Passing:  passing,
		Partials: partials,
	})
}

// ── Competitor Name Radar ──────────────────────────────────────────────────────
// Sends all names to Granite in one call and asks for semantic overlap warnings.
// Returns a map[name]warningString. Empty string = no concern.

func runCompetitorRadar(ctx context.Context, names []string) map[string]string {
	result := make(map[string]string)
	if len(names) == 0 {
		return result
	}

	client, err := ai.NewGraniteClient()
	if err != nil {
		log.Printf("radar: granite client: %v", err)
		return result
	}

	systemPrompt := `You are a brand name trademark and competitor radar expert.
For each brand name provided, assess whether it sounds too similar to any well-known existing brand (global or regional).
Return ONLY a JSON object where keys are the brand names and values are either:
- An empty string "" if the name is safe
- A brief warning string (≤15 words) if it sounds too similar to an existing brand
Example: {"Nestly": "Sounds similar to Nestlé — global food brand, potential confusion", "Voxify": ""}`

	userPrompt := "Analyse these brand names for competitor overlap: " + strings.Join(names, ", ") + ". Return JSON only."

	raw, err := client.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		log.Printf("radar: generate: %v", err)
		return result
	}

	// Extract JSON object
	jsonStr := extractJSONObject(raw)
	if jsonStr == "" {
		return result
	}

	var parsed map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		log.Printf("radar: parse: %v (raw=%q)", err, raw)
		return result
	}
	return parsed
}

func extractJSONObject(text string) string {
	text = strings.TrimSpace(text)
	// Strip markdown fences
	if idx := strings.Index(text, "```json"); idx != -1 {
		text = text[idx+7:]
	} else if idx := strings.Index(text, "```"); idx != -1 {
		text = text[idx+3:]
	}
	if idx := strings.LastIndex(text, "```"); idx != -1 {
		text = text[:idx]
	}
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start != -1 && end != -1 && end > start {
		return strings.TrimSpace(text[start : end+1])
	}
	return ""
}

func namesFromResults(results []models.AvailabilityResult) []string {
	names := make([]string, len(results))
	for i, r := range results {
		names[i] = r.Name
	}
	return names
}
