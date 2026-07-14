package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/c-annabel/NomVoxBranding/internal/handlers"
)

func main() {
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
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware)

	// ── Health check ────────────────────────────────────────────────
	r.Get("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok","service":"nomvox"}`)
	})

	// ── API routes ──────────────────────────────────────────────────
	r.Post("/api/generate", handlers.GenerateHandler)
	// r.Post("/api/session",        handlers.CreateSession)      // ST-04
	// r.Patch("/api/session/react", handlers.React)              // ST-04
	// r.Post("/api/availability",   handlers.CheckAvailability)  // ST-05
	// r.Post("/api/vision-context", handlers.VisionContext)      // ST-08
	// r.Post("/api/visuals",        handlers.Visuals)            // ST-08
	// r.Post("/api/persona",        handlers.Persona)            // ST-07
	// r.Post("/api/mockup",         handlers.Mockup)             // ST-09
	// r.Post("/api/export",         handlers.Export)             // ST-10

	log.Printf("NomVox API server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// corsMiddleware allows requests from the Next.js frontend (localhost:3000 in
// dev, Vercel URL in production). Set ALLOWED_ORIGIN in the environment.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("ALLOWED_ORIGIN")
		if origin == "" {
			origin = "http://localhost:3000"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
