package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/c-annabel/NomVoxBranding/internal/ai"
)

// debugVisualsResponse is the JSON body returned by GET /api/debug-visuals.
type debugVisualsResponse struct {
	ImagenOK    bool   `json:"imagen_ok"`
	ImagenError string `json:"imagen_error,omitempty"`
	DataURILen  int    `json:"data_uri_len,omitempty"` // length of first data URI if generation succeeded
	GraniteOK   bool   `json:"granite_ok"`
	GraniteError string `json:"granite_error,omitempty"`
}

// DebugVisualsHandler handles GET /api/debug-visuals.
// It fires a minimal test prompt at Imagen 3 and watsonx and reports success/failure.
// DEV ONLY — not exposed in production.
func DebugVisualsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	resp := debugVisualsResponse{}

	// ── Test Imagen 3 ─────────────────────────────────────────────────────────
	imagenClient, err := ai.NewImagenClient()
	if err != nil {
		resp.ImagenError = "client init: " + err.Error()
	} else {
		imgs, err := imagenClient.GenerateImages(ctx, "a simple blue circle on a white background", 1, "1:1")
		if err != nil {
			resp.ImagenError = err.Error()
		} else if len(imgs) == 0 {
			resp.ImagenError = "no images returned"
		} else {
			uri := ai.Base64ToDataURI(imgs[0], "image/png")
			if uri == "" {
				resp.ImagenError = "base64 decode failed (len=" + itoa(len(imgs[0])) + ")"
			} else {
				resp.ImagenOK = true
				resp.DataURILen = len(uri)
			}
		}
	}

	// ── Test watsonx (Granite/Llama) ──────────────────────────────────────────
	graniteClient, err := ai.NewGraniteClient()
	if err != nil {
		resp.GraniteError = "client init: " + err.Error()
	} else {
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
	json.NewEncoder(w).Encode(resp)
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
