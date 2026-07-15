// cmd/diagnose-imagen/main.go
//
// Standalone Imagen 3 + watsonx diagnostic tool — no server needed.
// Run:   go run ./cmd/diagnose-imagen
//
// Reads GOOGLE_AI_API_KEY, WATSONX_API_KEY, WATSONX_PROJECT_ID from .env
// (if present) or the process environment.
// Writes results to diagnose-imagen.log AND prints to stdout.
//
// Exit codes:
//   0 — all checked services are OK
//   1 — one or more checks failed (see log for details + fix advice)
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/c-annabel/NomVoxBranding/internal/ai"
	"github.com/joho/godotenv"
)

func main() {
	// ── Load .env (silent no-op when file absent) ────────────────────────────
	if err := godotenv.Load(); err != nil {
		println("note: no .env file — relying on process environment")
	}

	logFile, _ := os.OpenFile("diagnose-imagen.log",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	logLine := func(format string, args ...interface{}) {
		line := fmt.Sprintf("[%s] %s\n",
			time.Now().Format("15:04:05"),
			fmt.Sprintf(format, args...))
		fmt.Print(line)
		if logFile != nil {
			logFile.WriteString(line) //nolint:errcheck
		}
	}

	allOK := true

	// ══════════════════════════════════════════════════════════════════════════
	// 1. Imagen 3
	// ══════════════════════════════════════════════════════════════════════════
	logLine("=== Imagen 3 (Google AI Studio) ===")
	googleKey := os.Getenv("GOOGLE_AI_API_KEY")
	if googleKey == "" {
		logLine("FAIL — GOOGLE_AI_API_KEY is not set")
		allOK = false
	} else {
		prefix := googleKey
		if len(prefix) > 8 {
			prefix = prefix[:8]
		}
		logLine("GOOGLE_AI_API_KEY present (len=%d, prefix=%s…)", len(googleKey), prefix)

		imagenClient, err := ai.NewImagenClient()
		if err != nil {
			logLine("FAIL — NewImagenClient: %v", err)
			allOK = false
		} else {
			logLine("ImagenClient created OK")
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()

			logLine("Sending test prompt → Imagen 3 (sampleCount=1, aspect=1:1)…")
			t0 := time.Now()
			imgs, err := imagenClient.GenerateImages(
				ctx, "a simple blue circle on a white background", 1, "1:1")
			elapsed := time.Since(t0)

			if err != nil {
				logLine("FAIL — GenerateImages: %v  (elapsed=%s)", err, elapsed)
				allOK = false
				diagnoseImagenError(logLine, err.Error())
			} else if len(imgs) == 0 {
				logLine("FAIL — GenerateImages returned 0 images (elapsed=%s)", elapsed)
				allOK = false
			} else {
				logLine("Received %d image(s) in %s", len(imgs), elapsed)
				uri := ai.Base64ToDataURI(imgs[0], "image/png")
				if uri == "" {
					logLine("FAIL — Base64ToDataURI empty (raw b64 len=%d)", len(imgs[0]))
					allOK = false
				} else {
					pfx := uri
					if len(pfx) > 60 {
						pfx = pfx[:60]
					}
					logLine("OK — data URI len=%d (preview: %s…)", len(uri), pfx)
					if saveErr := savePNG(imgs[0], "diagnose-imagen-output.png"); saveErr != nil {
						logLine("note: could not save test image: %v", saveErr)
					} else {
						logLine("Test image saved → diagnose-imagen-output.png")
					}
				}
			}
		}
	}

	// ══════════════════════════════════════════════════════════════════════════
	// 2. watsonx / Granite-Llama
	// ══════════════════════════════════════════════════════════════════════════
	logLine("")
	logLine("=== watsonx.ai (IBM Granite / Llama) ===")
	watsonxKey := os.Getenv("WATSONX_API_KEY")
	watsonxPID := os.Getenv("WATSONX_PROJECT_ID")
	if watsonxKey == "" || watsonxPID == "" {
		logLine("FAIL — WATSONX_API_KEY and/or WATSONX_PROJECT_ID not set")
		logLine("  WATSONX_API_KEY set:    %v", watsonxKey != "")
		logLine("  WATSONX_PROJECT_ID set: %v", watsonxPID != "")
		allOK = false
	} else {
		logLine("WATSONX_API_KEY present (len=%d)", len(watsonxKey))
		logLine("WATSONX_PROJECT_ID: %s", watsonxPID)
		gc, err := ai.NewGraniteClient()
		if err != nil {
			logLine("FAIL — NewGraniteClient: %v", err)
			allOK = false
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			logLine("Sending test prompt…")
			t0 := time.Now()
			out, err := gc.Generate(ctx, "You are a test assistant.", "Reply with exactly: PONG")
			elapsed := time.Since(t0)
			if err != nil {
				logLine("FAIL — Generate: %v  (elapsed=%s)", err, elapsed)
				allOK = false
			} else if strings.TrimSpace(out) == "" {
				logLine("FAIL — empty response (elapsed=%s)", elapsed)
				allOK = false
			} else {
				logLine("OK — response: %q  (elapsed=%s)", strings.TrimSpace(out), elapsed)
			}
		}
	}

	// ══════════════════════════════════════════════════════════════════════════
	// 3. Redis (informational only)
	// ══════════════════════════════════════════════════════════════════════════
	logLine("")
	logLine("=== Redis / Upstash ===")
	if redisURL := os.Getenv("REDIS_URL"); redisURL == "" {
		logLine("SKIP — REDIS_URL not set (sessions won't persist, app still functions)")
	} else {
		masked := redisURL
		if len(masked) > 24 {
			masked = masked[:24] + "…"
		}
		logLine("REDIS_URL present: %s", masked)
	}

	// ── Summary ──────────────────────────────────────────────────────────────
	logLine("")
	logLine("=== Summary ===")
	if allOK {
		logLine("ALL CHECKS PASSED ✓")
		os.Exit(0)
	} else {
		logLine("ONE OR MORE CHECKS FAILED — see details above and in diagnose-imagen.log")
		os.Exit(1)
	}
}

// diagnoseImagenError prints actionable fix advice based on the HTTP error string.
func diagnoseImagenError(logLine func(string, ...interface{}), errStr string) {
	logLine("--- Diagnosis ---")
	switch {
	case strings.Contains(errStr, "403"):
		logLine("HTTP 403 — API key rejected OR Imagen not enabled for this project.")
		logLine("  Fix A: https://console.cloud.google.com → APIs & Services → Enable 'Generative Language API'")
		logLine("  Fix B: Imagen 3 may require Vertex AI (not free Google AI Studio tier).")
		logLine("         Alternative: switch to Gemini 2.0 Flash image generation (responseModalities IMAGE)")
		logLine("         See: https://ai.google.dev/gemini-api/docs/image-generation")
	case strings.Contains(errStr, "404"):
		logLine("HTTP 404 — Model 'imagen-3.0-generate-002' not found on this endpoint/tier.")
		logLine("  Fix: Try 'imagen-3.0-fast-generate-001' or use Gemini image generation instead.")
		logLine("       See: https://ai.google.dev/gemini-api/docs/imagen")
	case strings.Contains(errStr, "400"):
		logLine("HTTP 400 — Malformed request. Check sampleCount (must be 1–4) and aspectRatio.")
		logLine("  Raw error: %s", errStr)
	case strings.Contains(errStr, "429"):
		if strings.Contains(errStr, "credits") || strings.Contains(errStr, "RESOURCE_EXHAUSTED") || strings.Contains(errStr, "prepayment") {
			logLine("HTTP 429 RESOURCE_EXHAUSTED — Prepaid credits depleted (NOT a rate limit).")
			logLine("  Fix A (free): Use Gemini 2.0 Flash image generation (gemini-2.0-flash-preview-image-generation).")
			logLine("         The code has already been updated. Re-run this tool to confirm.")
			logLine("  Fix B (paid): Top up credits at https://ai.studio/projects → Billing.")
		} else {
			logLine("HTTP 429 — Rate limit exceeded. Wait 60 seconds and retry.")
		}
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline"):
		logLine("TIMEOUT — Request exceeded 90s. Check network; first Imagen call can be slow.")
		logLine("  Fix: increase httpClient.Timeout in imagen.go to 120s")
	default:
		logLine("Unknown error — full: %s", errStr)
	}
}

// savePNG tries all base64 variants and writes raw bytes to path.
func savePNG(b64str, path string) error {
	encs := []*base64.Encoding{
		base64.RawURLEncoding,
		base64.RawStdEncoding,
		base64.StdEncoding,
	}
	var data []byte
	var lastErr error
	for _, enc := range encs {
		if d, err := enc.DecodeString(b64str); err == nil {
			data = d
			break
		} else {
			lastErr = err
		}
	}
	if data == nil {
		return fmt.Errorf("base64 decode failed (len=%d): %v", len(b64str), lastErr)
	}
	return os.WriteFile(path, data, 0644)
}
