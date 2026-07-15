package handlers

import (
	"fmt"
	"net/http"

	"github.com/c-annabel/NomVoxBranding/internal/ai"
)

// DebugLLMHandler handles GET /api/debug-llm
// Returns the RAW text the LLM produces for a fixed tiny prompt.
// No JSON parsing — shows exactly what the model outputs through the server stack.
// Remove before production.
func DebugLLMHandler(w http.ResponseWriter, r *http.Request) {
	client, err := ai.NewGraniteClient()
	if err != nil {
		http.Error(w, "granite client: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	system := `Return ONLY a JSON array with exactly 2 objects. Each: {"name":string,"tagline":string}. No other text.`
	user   := `Two brand names for an eco coffee brand. JSON array only.`

	raw, err := client.Generate(r.Context(), system, user)
	if err != nil {
		http.Error(w, "LLM error: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "=== RAW LLM OUTPUT ===\n%s\n\n=== ExtractJSON result ===\n%s",
		raw, ai.ExtractJSON(raw))
}
