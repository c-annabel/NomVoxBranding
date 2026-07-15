package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/c-annabel/NomVoxBranding/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env in development. In production (Fly.io / Vercel) env vars are
	// injected by the platform so this is a silent no-op when .env is absent.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — using platform environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()

	// ── Global middleware ───────────────────────────────────────────
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// ── Health check ────────────────────────────────────────────────
	r.Get("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok","service":"nomvox"}`)
	})

	// ── API routes ──────────────────────────────────────────────────
	// /api/generate has its own timeout (120s) to allow full LLM responses
	r.With(middleware.Timeout(120 * time.Second)).Post("/api/generate",     handlers.GenerateHandler)
	r.With(middleware.Timeout(30 * time.Second)).Get("/api/diagnose",        handlers.DiagnoseHandler)   // DEV ONLY
	r.With(middleware.Timeout(60 * time.Second)).Get("/api/debug-llm",       handlers.DebugLLMHandler)   // DEV ONLY
	r.With(middleware.Timeout(90 * time.Second)).Get("/api/debug-visuals",   handlers.DebugVisualsHandler) // DEV ONLY
	r.With(middleware.Timeout(30 * time.Second)).Post("/api/session",        handlers.CreateSession)
	r.With(middleware.Timeout(30 * time.Second)).Patch("/api/session/react", handlers.React)
	r.With(middleware.Timeout(60 * time.Second)).Post("/api/availability",   handlers.CheckAvailability)
	// ST-08: visual identity — mood board, logos, persona, landing-page mockup
	// All image generation calls fan out concurrently; 120s covers worst-case Imagen latency
	r.With(middleware.Timeout(120 * time.Second)).Post("/api/visuals", handlers.VisualsHandler)
	// ST-10: ZIP export
	r.With(middleware.Timeout(30 * time.Second)).Post("/api/export",  handlers.ExportHandler)

	log.Printf("NomVox API server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// corsMiddleware allows requests from the Next.js frontend.
//
// Configuration (environment variables):
//   ALLOWED_ORIGIN   — single origin, e.g. "https://nomvox.app" (default: localhost:3000)
//   ALLOWED_ORIGINS  — comma-separated list of origins, e.g. "https://nomvox.app,https://nomvox.vercel.app"
//   CORS_WILDCARD    — set to "true" to allow all origins with * (use only for public APIs)
//
// In development no variable is needed — localhost:3000 is always allowed.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wildcard mode (opt-in)
		if os.Getenv("CORS_WILDCARD") == "true" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// Build allowed-origin set
			allowedOrigins := map[string]bool{
				"http://localhost:3000":  true,
				"http://localhost:3001":  true,
				"http://127.0.0.1:3000": true,
			}
			// Single ALLOWED_ORIGIN env var
			if single := os.Getenv("ALLOWED_ORIGIN"); single != "" {
				allowedOrigins[single] = true
			}
			// Comma-separated ALLOWED_ORIGINS env var
			if multi := os.Getenv("ALLOWED_ORIGINS"); multi != "" {
				for _, o := range strings.Split(multi, ",") {
					if t := strings.TrimSpace(o); t != "" {
						allowedOrigins[t] = true
					}
				}
			}
			reqOrigin := r.Header.Get("Origin")
			if allowedOrigins[reqOrigin] {
				w.Header().Set("Access-Control-Allow-Origin", reqOrigin)
				w.Header().Set("Vary", "Origin")
			} else if reqOrigin == "" {
				// Same-origin or non-browser request — allow
				w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			}
			// else: unknown origin — no ACAO header (browser will block)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID, Authorization")
		w.Header().Set("Access-Control-Max-Age", "600")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
