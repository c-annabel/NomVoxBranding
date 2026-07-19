package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/c-annabel/NomVoxBranding/internal/ai"
	"github.com/c-annabel/NomVoxBranding/internal/models"
	"github.com/c-annabel/NomVoxBranding/internal/session"
	"github.com/google/uuid"
)

// fileLog writes a timestamped line to nomvox.log (append mode).
// Non-fatal — if log file can't be written, prints to stderr instead.
func fileLog(format string, args ...interface{}) {
	msg := fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
	f, err := os.OpenFile("nomvox.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("(fileLog unavailable) " + msg)
		return
	}
	defer f.Close()
	f.WriteString(msg)
	log.Print(msg) // also print to stderr so terminal shows it live
}

// generateResponse is what the frontend receives.
type generateResponse struct {
	SessionID string            `json:"session_id"`
	Cards     []models.NameCard `json:"cards"`
}

// generateRequest is the full incoming payload from the frontend.
type generateRequest struct {
	SessionID string               `json:"session_id"` // empty on first call
	Intake    models.IntakePayload `json:"intake"`
	Action    string               `json:"action"` // "init" | "reproduce"
}

// lazyStore holds the session store singleton initialised on first use.
// We defer initialisation so the server can start without Redis configured.
var lazyStore *session.Store

func getStore() (*session.Store, error) {
	if lazyStore != nil {
		return lazyStore, nil
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL not set")
	}
	s, err := session.New(redisURL)
	if err != nil {
		return nil, err
	}
	lazyStore = s
	return lazyStore, nil
}

// GenerateHandler handles POST /api/generate.
// It loads session memory from Redis, calls IBM Granite with full context,
// persists updated session, and returns enriched NameCards.
//
// PROTOTYPE RESILIENCE: if the LLM is unavailable for any reason (missing keys,
// token quota exhausted, rate limits, unparseable output), the handler silently
// serves curated fallback NameCards instead of an error, so the demo always
// completes end-to-end.
func GenerateHandler(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Intake.CoreIdea) == "" {
		jsonError(w, "core_idea is required", http.StatusBadRequest)
		return
	}

	// ── Session management ─────────────────────────────────────────
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = uuid.NewString()
	}

	store, storeErr := getStore()
	var sess *models.BrandSession

	if storeErr == nil {
		loaded, err := store.Get(r.Context(), sessionID)
		if err != nil {
			log.Printf("generate: session load: %v", err)
		}
		if loaded != nil {
			sess = loaded
		}
	}

	if sess == nil {
		sess = &models.BrandSession{
			SessionID: sessionID,
			Intake:    req.Intake,
		}
	}

	// Always generate 4 names. The new radical-naming prompt is longer than the
	// original, so 5 names frequently hits the 3000-token ceiling and truncates JSON.
	// 4 names reliably fit within budget with the enriched system prompt.
	nameCount := 4

	// serveCards finalises the response: updates session memory and encodes JSON.
	serveCards := func(cards []models.NameCard) {
		for _, c := range cards {
			sess.AllGenerated = appendUnique(sess.AllGenerated, c.Name)
		}
		sess.Intake = req.Intake // update intake in case user changed fields

		if storeErr == nil {
			if err := store.Set(r.Context(), sess); err != nil {
				log.Printf("generate: session save: %v", err)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(generateResponse{ //nolint:errcheck
			SessionID: sessionID,
			Cards:     cards,
		})
	}

	// ── Build prompts with full session context ────────────────────
	client, err := ai.NewGraniteClient()
	if err != nil {
		fileLog("generate: granite client unavailable (%v) — serving fallback names", err)
		serveCards(fallbackNameCards(req.Intake, nameCount))
		return
	}

	systemPrompt := ai.BuildSystemPrompt(sess)
	userPrompt := ai.BuildUserPrompt(req.Intake, nameCount)

	fileLog("generate: calling LLM (nameCount=%d, sessionID=%s)", nameCount, sessionID)
	start := time.Now()

	// ── Call LLM — use request context so chi timeout cancels cleanly ──
	raw, err := client.Generate(r.Context(), systemPrompt, userPrompt)
	elapsed := time.Since(start)

	if err != nil {
		fileLog("generate: LLM call failed after %s: %v", elapsed, err)
		// Retry once with an even tighter prompt
		userPrompt2 := fmt.Sprintf("Generate %d brand names. Return ONLY a JSON array [ ... ]. No other text.", nameCount)
		raw, err = client.Generate(r.Context(), systemPrompt, userPrompt2)
		if err != nil {
			fileLog("generate: retry failed (%v) — serving fallback names", err)
			serveCards(fallbackNameCards(req.Intake, nameCount))
			return
		}
	}

	fileLog("generate: LLM responded in %s, raw len=%d, preview=%q", elapsed, len(raw), ai.TruncateStr(raw, 120))

	// ── Parse + validate cards ─────────────────────────────────────
	cards, parseErr := ai.ParseNameCards(raw)
	if parseErr != nil || len(cards) == 0 {
		fileLog("generate: parse failed (%v) — serving fallback names", parseErr)
		serveCards(fallbackNameCards(req.Intake, nameCount))
		return
	}
	fileLog("generate: parsed %d cards OK", len(cards))

	serveCards(cards)
}

// appendUnique appends s to slice only if not already present.
func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

// assembleLLMPrompt is kept for backward compatibility — now delegates to ai package.
func assembleLLMPrompt(p models.IntakePayload) string {
	return ai.BuildUserPrompt(p, 8)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
