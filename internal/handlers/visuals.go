package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/c-annabel/NomVoxBranding/internal/ai"
	"github.com/c-annabel/NomVoxBranding/internal/models"
)

// visualsRequest is the incoming payload from the frontend.
// The frontend sends the selected NameCard + original intake + optional session_id.
// SelectedLogoKey ("profile"|"app"|"business") and SelectedLogoStyle are sent after the
// user picks a logo on step 1 so the mood board can reference the chosen visual style.
type visualsRequest struct {
	SessionID         string               `json:"session_id"`
	Card              models.NameCard      `json:"card"`
	Intake            models.IntakePayload `json:"intake"`
	SelectedLogoKey   string               `json:"selected_logo_key"`   // "profile"|"app"|"business"
	SelectedLogoStyle string               `json:"selected_logo_style"` // human-readable style description
}

// visualsResponse is returned to the frontend with base64-encoded data URIs.
type visualsResponse struct {
	MoodBoard    []string `json:"mood_board"`     // 4 data URIs  (or fewer if some failed)
	LogoProfile  string   `json:"logo_profile"`   // data URI
	LogoApp      string   `json:"logo_app"`       // data URI
	LogoBusiness string   `json:"logo_business"`  // data URI
	MockupHTML   string   `json:"mockup_html"`    // raw HTML string
	Persona      *models.BrandPersona `json:"persona,omitempty"`
}

// VisualsHandler handles POST /api/visuals.
// It fans out concurrently to:
//   1. Imagen 3 — 4 mood board images (1 call, sampleCount=4)
//   2. Imagen 3 — 3 logo concepts (3 parallel calls)
//   3. Granite — landing page HTML mockup
//   4. Granite — brand persona JSON
//
// All four tasks run in parallel; partial failures are tolerated.
func VisualsHandler(w http.ResponseWriter, r *http.Request) {
	var req visualsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Card.Name) == "" {
		jsonError(w, "card.name is required", http.StatusBadRequest)
		return
	}

	imagenClient, imagenErr := ai.NewImagenClient()
	graniteClient, graniteErr := ai.NewGraniteClient()

	// If both services are unavailable, fail fast.
	if imagenErr != nil && graniteErr != nil {
		jsonError(w, "both AI services unavailable: "+imagenErr.Error(), http.StatusServiceUnavailable)
		return
	}

	// Load vision context from session if available
	var visionCtx *models.VisionContext
	if req.SessionID != "" {
		if store, err := getStore(); err == nil {
			if sess, err := store.Get(r.Context(), req.SessionID); err == nil && sess != nil {
				visionCtx = sess.VisionCtx
			}
		}
	}

	var (
		mu          sync.Mutex
		resp        visualsResponse
		wg          sync.WaitGroup
	)

	// ── Task 1: Mood board (4 images) ────────────────────────────────
	if imagenErr == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prompt := ai.MoodBoardPrompt(req.Card, req.Intake, visionCtx, req.SelectedLogoKey, req.SelectedLogoStyle)
			imgs, err := imagenClient.GenerateImages(r.Context(), prompt, 4, "1:1")
			if err != nil {
				log.Printf("visuals: mood board: %v", err)
				return
			}
			uris := make([]string, 0, len(imgs))
			for _, b64 := range imgs {
				uri := ai.Base64ToDataURI(b64, "image/png")
				if uri != "" {
					uris = append(uris, uri)
				}
			}
			mu.Lock()
			resp.MoodBoard = uris
			mu.Unlock()
		}()
	}

	// ── Task 2a: Logo — profile ───────────────────────────────────────
	if imagenErr == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prompt := ai.LogoConceptPrompt(req.Card, req.Intake, "profile")
			imgs, err := imagenClient.GenerateImages(r.Context(), prompt, 1, "1:1")
			if err != nil {
				log.Printf("visuals: logo profile: %v", err)
				return
			}
			if len(imgs) > 0 {
				uri := ai.Base64ToDataURI(imgs[0], "image/png")
				mu.Lock()
				resp.LogoProfile = uri
				mu.Unlock()
			}
		}()
	}

	// ── Task 2b: Logo — app icon ──────────────────────────────────────
	if imagenErr == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prompt := ai.LogoConceptPrompt(req.Card, req.Intake, "app")
			imgs, err := imagenClient.GenerateImages(r.Context(), prompt, 1, "1:1")
			if err != nil {
				log.Printf("visuals: logo app: %v", err)
				return
			}
			if len(imgs) > 0 {
				uri := ai.Base64ToDataURI(imgs[0], "image/png")
				mu.Lock()
				resp.LogoApp = uri
				mu.Unlock()
			}
		}()
	}

	// ── Task 2c: Logo — business card ────────────────────────────────
	if imagenErr == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prompt := ai.LogoConceptPrompt(req.Card, req.Intake, "business")
			imgs, err := imagenClient.GenerateImages(r.Context(), prompt, 1, "16:9")
			if err != nil {
				log.Printf("visuals: logo business: %v", err)
				return
			}
			if len(imgs) > 0 {
				uri := ai.Base64ToDataURI(imgs[0], "image/png")
				mu.Lock()
				resp.LogoBusiness = uri
				mu.Unlock()
			}
		}()
	}

	// ── Task 3: Landing page HTML mockup (Granite) ───────────────────
	if graniteErr == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sysPrompt := ai.BuildMockupSystemPrompt()
			userPrompt := ai.BuildMockupUserPrompt(req.Card, req.Intake)
			html, err := graniteClient.Generate(r.Context(), sysPrompt, userPrompt)
			if err != nil {
				log.Printf("visuals: mockup html: %v", err)
				return
			}
			// Strip any markdown fences
			html = stripMDFences(html)
			mu.Lock()
			resp.MockupHTML = html
			mu.Unlock()
		}()
	}

	// ── Task 4: Brand persona (Granite) ──────────────────────────────
	if graniteErr == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sysPrompt := ai.BuildPersonaSystemPrompt()
			userPrompt := ai.BuildPersonaUserPrompt(req.Card, req.Intake)
			raw, err := graniteClient.Generate(r.Context(), sysPrompt, userPrompt)
			if err != nil {
				log.Printf("visuals: persona: %v", err)
				return
			}
			persona, err := ai.ParsePersona(raw)
			if err != nil {
				log.Printf("visuals: persona parse: %v", err)
				return
			}
			mu.Lock()
			resp.Persona = persona
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Persist visuals to session
	if req.SessionID != "" {
		if store, err := getStore(); err == nil {
			if sess, err := store.Get(r.Context(), req.SessionID); err == nil && sess != nil {
				sess.Visuals = &models.VisualPack{
					MoodBoard:    resp.MoodBoard,
					LogoProfile:  resp.LogoProfile,
					LogoApp:      resp.LogoApp,
					LogoBusiness: resp.LogoBusiness,
				}
				sess.Persona = resp.Persona
				if serr := store.Set(r.Context(), sess); serr != nil {
					log.Printf("visuals: session save: %v", serr)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// stripMDFences removes markdown code fences from a string.
func stripMDFences(s string) string {
	s = strings.TrimSpace(s)
	for _, fence := range []string{"```html", "```"} {
		if idx := strings.Index(s, fence); idx != -1 {
			s = s[idx+len(fence):]
		}
	}
	if idx := strings.LastIndex(s, "```"); idx != -1 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}
