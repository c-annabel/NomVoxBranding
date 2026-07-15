package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/c-annabel/NomVoxBranding/internal/ai"
	"github.com/c-annabel/NomVoxBranding/internal/models"
)

// debugVisualsResponse is the JSON body returned by GET /api/debug-visuals.
type debugVisualsResponse struct {
	// Gemini image generation (optional — may fail if credits depleted)
	GeminiImageOK    bool   `json:"gemini_image_ok"`
	GeminiImageError string `json:"gemini_image_error,omitempty"`
	DataURILen       int    `json:"data_uri_len,omitempty"`

	// watsonx SVG logo generation (primary fallback — free tier)
	SVGLogoOK    bool   `json:"svg_logo_ok"`
	SVGLogoError string `json:"svg_logo_error,omitempty"`
	SVGDataURI   string `json:"svg_data_uri_preview,omitempty"` // first 120 chars

	// watsonx text (Granite/Llama) — used for mockup, persona, SVG
	GraniteOK    bool   `json:"granite_ok"`
	GraniteError string `json:"granite_error,omitempty"`
}

// DebugVisualsHandler handles GET /api/debug-visuals.
// Tests Gemini image gen, watsonx SVG logo gen, and watsonx text. DEV ONLY.
func DebugVisualsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	resp := debugVisualsResponse{}

	// ── 1. Test Gemini image generation ──────────────────────────────────────
	imagenClient, err := ai.NewImagenClient()
	if err != nil {
		resp.GeminiImageError = "client init: " + err.Error()
	} else {
		imgs, err := imagenClient.GenerateImages(ctx, "a simple blue circle on a white background", 1, "1:1")
		if err != nil {
			resp.GeminiImageError = err.Error()
		} else if len(imgs) == 0 {
			resp.GeminiImageError = "no images returned"
		} else {
			uri := ai.Base64ToDataURI(imgs[0], "image/png")
			if uri == "" {
				resp.GeminiImageError = "base64 decode failed (len=" + itoa(len(imgs[0])) + ")"
			} else {
				resp.GeminiImageOK = true
				resp.DataURILen = len(uri)
			}
		}
	}

	// ── 2. Test watsonx SVG logo generation ──────────────────────────────────
	graniteClient, graniteErr := ai.NewGraniteClient()
	if graniteErr != nil {
		resp.SVGLogoError = "granite client: " + graniteErr.Error()
		resp.GraniteError = graniteErr.Error()
	} else {
		// Test SVG logo
		testCard := models.NameCard{Name: "TestBrand", Tagline: "Test tagline"}
		testIntake := models.IntakePayload{
			Industry:    "technology",
			ColorMood:   "deep indigo and cyan",
			Personality: "modern, clean",
		}
		sys := ai.BuildSVGLogoSystemPrompt()
		usr := ai.BuildSVGLogoUserPrompt(testCard, testIntake, "profile")
		raw, err := graniteClient.Generate(ctx, sys, usr)
		if err != nil {
			resp.SVGLogoError = err.Error()
		} else {
			uri := ai.SVGToDataURI(raw)
			if uri == "" {
				resp.SVGLogoError = "SVGToDataURI returned empty (raw len=" + itoa(len(raw)) + ")"
			} else {
				resp.SVGLogoOK = true
				if len(uri) > 120 {
					resp.SVGDataURI = uri[:120] + "…"
				} else {
					resp.SVGDataURI = uri
				}
			}
		}

		// Test plain text (Granite/Llama)
		out, err := graniteClient.Generate(ctx, "You are a test assistant.", "Reply with exactly: OK")
		if err != nil {
			resp.GraniteError = err.Error()
		} else if len(out) == 0 {
			resp.GraniteError = "empty response"
		} else {
			resp.GraniteOK = true
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// itoa is a minimal int-to-string helper to avoid importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := [20]byte{}
	pos := 20
	for n > 0 {
		pos--
		digits[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(digits[pos:])
}
