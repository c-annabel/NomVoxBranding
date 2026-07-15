package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/c-annabel/NomVoxBranding/internal/ai"
	"github.com/c-annabel/NomVoxBranding/internal/models"
	"github.com/google/uuid"
)

// ── POST /api/session ──────────────────────────────────────────────────────────
// Creates a new brand session and persists the initial intake to Redis.

type createSessionRequest struct {
	Intake models.IntakePayload `json:"intake"`
}

type createSessionResponse struct {
	SessionID string `json:"session_id"`
}

func CreateSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	store, err := getStore()
	if err != nil {
		jsonError(w, "session store unavailable", http.StatusServiceUnavailable)
		return
	}

	sess := &models.BrandSession{
		SessionID: uuid.NewString(),
		Intake:    req.Intake,
	}
	if err := store.Set(r.Context(), sess); err != nil {
		log.Printf("session.create: %v", err)
		jsonError(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createSessionResponse{SessionID: sess.SessionID})
}

// ── PATCH /api/session/react ───────────────────────────────────────────────────
// Updates session with user reaction: like / reject / note / visual-note / slider / select.
// On reject with a non-empty note, also returns an AI clarifying question.

type reactRequest struct {
	SessionID      string  `json:"session_id"`
	Name           string  `json:"name,omitempty"`
	Action         string  `json:"action"` // like | reject | note | visual-note | slider | select
	Note           string  `json:"note,omitempty"`
	SliderPlayful  float64 `json:"slider_playful,omitempty"`
	SliderAbstract float64 `json:"slider_abstract,omitempty"`
}

type reactResponse struct {
	OK                bool   `json:"ok"`
	ClarifyingQuestion string `json:"clarifying_question,omitempty"` // only on reject+note
}

func React(w http.ResponseWriter, r *http.Request) {
	var req reactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.SessionID == "" {
		jsonError(w, "session_id is required", http.StatusBadRequest)
		return
	}

	store, err := getStore()
	if err != nil {
		jsonError(w, "session store unavailable", http.StatusServiceUnavailable)
		return
	}

	sess, err := store.Get(r.Context(), req.SessionID)
	if err != nil || sess == nil {
		// Create a minimal session on miss — graceful recovery
		sess = &models.BrandSession{SessionID: req.SessionID}
	}

	var clarifyingQ string

	switch req.Action {
	case "like":
		if req.Name != "" {
			sess.Liked = appendUnique(sess.Liked, req.Name)
		}

	case "reject":
		if req.Name != "" {
			sess.Rejected = appendUnique(sess.Rejected, req.Name)
			// Remove from liked if it was there
			sess.Liked = removeString(sess.Liked, req.Name)
		}
		if strings.TrimSpace(req.Note) != "" {
			sess.DirectionNotes = append(sess.DirectionNotes, req.Note)
			// Fire Anti-Name Reasoning: ask Granite a single clarifying question
			q, qErr := generateClarifyingQuestion(r.Context(), req.Name, req.Note, sess)
			if qErr != nil {
				log.Printf("react: clarifying Q: %v", qErr)
			} else {
				clarifyingQ = q
			}
		}

	case "note":
		if strings.TrimSpace(req.Note) != "" {
			sess.DirectionNotes = append(sess.DirectionNotes, req.Note)
		}

	case "visual-note":
		if strings.TrimSpace(req.Note) != "" {
			sess.VisualNotes = append(sess.VisualNotes, req.Note)
		}

	case "slider":
		if req.SliderPlayful != 0 {
			sess.SliderPlayful = req.SliderPlayful
		}
		if req.SliderAbstract != 0 {
			sess.SliderAbstract = req.SliderAbstract
		}

	case "select":
		if req.Name != "" {
			sess.SelectedName = req.Name
			sess.Liked = appendUnique(sess.Liked, req.Name)
		}
	}

	if saveErr := store.Set(r.Context(), sess); saveErr != nil {
		log.Printf("react: session save: %v", saveErr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reactResponse{
		OK:                 true,
		ClarifyingQuestion: clarifyingQ,
	})
}

// ── Anti-Name Reasoning ────────────────────────────────────────────────────────
// generateClarifyingQuestion asks Granite to respond to a rejection note with
// exactly one clarifying question — making the "creative partner" loop visible.

func generateClarifyingQuestion(ctx context.Context, rejectedName, note string, sess *models.BrandSession) (string, error) {
	client, err := ai.NewGraniteClient()
	if err != nil {
		return "", err
	}

	systemPrompt := `You are NomVox, an empathetic AI brand-naming partner. 
The user just rejected a name and explained why. 
Respond with EXACTLY ONE short, insightful clarifying question (≤15 words) that helps you understand their direction better.
Do NOT suggest new names. Do NOT explain. Output only the question itself — no quotes, no prefix.`

	userPrompt := buildAntiNamePrompt(rejectedName, note, sess)

	raw, err := client.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	// Clean up: take only first line
	q := strings.TrimSpace(strings.SplitN(raw, "\n", 2)[0])
	q = strings.Trim(q, `"'`)
	return q, nil
}

func buildAntiNamePrompt(rejectedName, note string, sess *models.BrandSession) string {
	var b strings.Builder
	if rejectedName != "" {
		b.WriteString("The user rejected the name \"" + rejectedName + "\".")
	}
	b.WriteString(" Their reason: " + note + ".")
	if len(sess.Liked) > 0 {
		b.WriteString(" Names they liked so far: " + strings.Join(sess.Liked, ", ") + ".")
	}
	b.WriteString(" Ask one clarifying question.")
	return b.String()
}

// removeString removes a string from a slice.
func removeString(slice []string, s string) []string {
	out := slice[:0]
	for _, v := range slice {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}
