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
// SelectedLogoKey ("profile"|"app"|"business") and SelectedLogoStyle are sent
// after the user picks a logo so the mood board can reference the chosen visual style.
type visualsRequest struct {
	SessionID         string               `json:"session_id"`
	Card              models.NameCard      `json:"card"`
	Intake            models.IntakePayload `json:"intake"`
	SelectedLogoKey   string               `json:"selected_logo_key"`   // "profile"|"app"|"business"
	SelectedLogoStyle string               `json:"selected_logo_style"` // human-readable style description
}

// visualsResponse is returned to the frontend.
// Logos are data URIs — either SVG (data:image/svg+xml;base64,…) or PNG
// (data:image/png;base64,…) depending on which generator succeeded.
// MockupHTML is raw HTML for the landing page iframe.
type visualsResponse struct {
	MoodBoard    []string             `json:"mood_board"`           // SVG data URIs (or empty)
	LogoProfile  string               `json:"logo_profile"`         // data URI
	LogoApp      string               `json:"logo_app"`             // data URI
	LogoBusiness string               `json:"logo_business"`        // data URI
	MockupHTML   string               `json:"mockup_html"`          // raw HTML string
	Persona      *models.BrandPersona `json:"persona,omitempty"`
}

// VisualsHandler handles POST /api/visuals.
//
// Generation strategy (priority order):
//   1. Gemini image generation → PNG data URIs (requires credits; skipped if 429/403)
//   2. watsonx SVG generation → SVG data URIs via Granite/Llama (free, always tried)
//   3. CSS art placeholders → rendered by the frontend (no server call needed)
//
// All tasks fan out concurrently; partial failures are tolerated.
// The frontend CSS art placeholders are the guaranteed final fallback.
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

	// Both clients — failures are tolerated individually.
	imagenClient, imagenErr := ai.NewImagenClient()
	graniteClient, graniteErr := ai.NewGraniteClient()

	if imagenErr != nil && graniteErr != nil {
		jsonError(w, "both AI services unavailable: "+imagenErr.Error(), http.StatusServiceUnavailable)
		return
	}

	// Load vision context from session if available.
	var visionCtx *models.VisionContext
	if req.SessionID != "" {
		if store, err := getStore(); err == nil {
			if sess, err := store.Get(r.Context(), req.SessionID); err == nil && sess != nil {
				visionCtx = sess.VisionCtx
			}
		}
	}

	var (
		mu   sync.Mutex
		resp visualsResponse
		wg   sync.WaitGroup
	)

	// ── Helper: generate one SVG logo via watsonx ────────────────────────────
	generateSVGLogo := func(logoType string) string {
		if graniteErr != nil {
			return ""
		}
		sys := ai.BuildSVGLogoSystemPrompt()
		usr := ai.BuildSVGLogoUserPrompt(req.Card, req.Intake, logoType)
		raw, err := graniteClient.Generate(r.Context(), sys, usr)
		if err != nil {
			log.Printf("visuals: svg logo %s: %v", logoType, err)
			return ""
		}
		// Strip any markdown fences around SVG output
		raw = stripMDFences(raw)
		// If LLM wrapped in ```svg or ```xml fences, strip those too
		for _, fence := range []string{"```svg", "```xml"} {
			if idx := strings.Index(raw, fence); idx != -1 {
				raw = raw[idx+len(fence):]
			}
		}
		raw = strings.TrimSpace(raw)
		uri := ai.SVGToDataURI(raw)
		if uri == "" {
			log.Printf("visuals: svg logo %s: SVGToDataURI returned empty (raw len=%d)", logoType, len(raw))
		}
		return uri
	}

	// ── Task 1: Mood board ───────────────────────────────────────────────────
	// Attempt 1a: Gemini image generation (4 PNG images).
	// Attempt 1b: watsonx SVG mood board grid (4 SVG tiles, concatenated as one data URI each).
	wg.Add(1)
	go func() {
		defer wg.Done()

		// 1a — Gemini image generation
		if imagenErr == nil {
			prompt := ai.MoodBoardPrompt(req.Card, req.Intake, visionCtx, req.SelectedLogoKey, req.SelectedLogoStyle)
			imgs, err := imagenClient.GenerateImages(r.Context(), prompt, 4, "1:1")
			if err == nil && len(imgs) > 0 {
				uris := make([]string, 0, len(imgs))
				for _, b64 := range imgs {
					if uri := ai.Base64ToDataURI(b64, "image/png"); uri != "" {
						uris = append(uris, uri)
					}
				}
				if len(uris) > 0 {
					mu.Lock()
					resp.MoodBoard = uris
					mu.Unlock()
					return
				}
			}
			log.Printf("visuals: mood board (gemini): %v — falling back to SVG", err)
		}

		// 1b — SVG mood board tiles via watsonx (4 distinct brand tiles)
		if graniteErr != nil {
			return
		}
		moodTiles := ai.BuildSVGMoodBoardPrompts(req.Card, req.Intake)
		svgURIs := make([]string, 0, len(moodTiles))
		for i, tilePrompt := range moodTiles {
			sys := ai.BuildSVGTileSystemPrompt()
			raw, err := graniteClient.Generate(r.Context(), sys, tilePrompt)
			if err != nil {
				log.Printf("visuals: svg mood tile %d: %v", i, err)
				continue
			}
			raw = stripMDFences(raw)
			for _, fence := range []string{"```svg", "```xml"} {
				if idx := strings.Index(raw, fence); idx != -1 {
					raw = raw[idx+len(fence):]
				}
			}
			raw = strings.TrimSpace(raw)
			if uri := ai.SVGToDataURI(raw); uri != "" {
				svgURIs = append(svgURIs, uri)
			}
		}
		if len(svgURIs) > 0 {
			mu.Lock()
			resp.MoodBoard = svgURIs
			mu.Unlock()
		}
	}()

	// ── Task 2a: Logo — profile ──────────────────────────────────────────────
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Try Gemini PNG first
		if imagenErr == nil {
			prompt := ai.LogoConceptPrompt(req.Card, req.Intake, "profile")
			imgs, err := imagenClient.GenerateImages(r.Context(), prompt, 1, "1:1")
			if err == nil && len(imgs) > 0 {
				if uri := ai.Base64ToDataURI(imgs[0], "image/png"); uri != "" {
					mu.Lock()
					resp.LogoProfile = uri
					mu.Unlock()
					return
				}
			}
			log.Printf("visuals: logo profile (gemini): %v — falling back to SVG", err)
		}
		// Fallback: SVG via watsonx
		if uri := generateSVGLogo("profile"); uri != "" {
			mu.Lock()
			resp.LogoProfile = uri
			mu.Unlock()
		}
	}()

	// ── Task 2b: Logo — app icon ─────────────────────────────────────────────
	wg.Add(1)
	go func() {
		defer wg.Done()
		if imagenErr == nil {
			prompt := ai.LogoConceptPrompt(req.Card, req.Intake, "app")
			imgs, err := imagenClient.GenerateImages(r.Context(), prompt, 1, "1:1")
			if err == nil && len(imgs) > 0 {
				if uri := ai.Base64ToDataURI(imgs[0], "image/png"); uri != "" {
					mu.Lock()
					resp.LogoApp = uri
					mu.Unlock()
					return
				}
			}
			log.Printf("visuals: logo app (gemini): %v — falling back to SVG", err)
		}
		if uri := generateSVGLogo("app"); uri != "" {
			mu.Lock()
			resp.LogoApp = uri
			mu.Unlock()
		}
	}()

	// ── Task 2c: Logo — business card ────────────────────────────────────────
	wg.Add(1)
	go func() {
		defer wg.Done()
		if imagenErr == nil {
			prompt := ai.LogoConceptPrompt(req.Card, req.Intake, "business")
			imgs, err := imagenClient.GenerateImages(r.Context(), prompt, 1, "16:9")
			if err == nil && len(imgs) > 0 {
				if uri := ai.Base64ToDataURI(imgs[0], "image/png"); uri != "" {
					mu.Lock()
					resp.LogoBusiness = uri
					mu.Unlock()
					return
				}
			}
			log.Printf("visuals: logo business (gemini): %v — falling back to SVG", err)
		}
		if uri := generateSVGLogo("business"); uri != "" {
			mu.Lock()
			resp.LogoBusiness = uri
			mu.Unlock()
		}
	}()

	// ── Task 3: Landing page HTML mockup (Granite) ───────────────────────────
	if graniteErr == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sysPrompt := ai.BuildMockupSystemPrompt()
			userPrompt := ai.BuildMockupUserPrompt(req.Card, req.Intake, req.SelectedLogoKey, req.SelectedLogoStyle)
			html, err := graniteClient.Generate(r.Context(), sysPrompt, userPrompt)
			if err != nil {
				log.Printf("visuals: mockup html: %v", err)
				return
			}
			html = stripMDFences(html)
			mu.Lock()
			resp.MockupHTML = html
			mu.Unlock()
		}()
	}

	// ── Task 4: Brand persona (Granite) ──────────────────────────────────────
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

	// Persist visuals to session.
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
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
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
