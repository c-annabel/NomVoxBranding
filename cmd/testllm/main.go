// cmd/testllm/main.go
// Standalone LLM connectivity test — runs without the HTTP server.
// Tests: IAM token → chat API → JSON parsing → name card output.
// Usage:  go run ./cmd/testllm/
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	apiKey    := os.Getenv("WATSONX_API_KEY")
	projectID := os.Getenv("WATSONX_PROJECT_ID")
	baseURL   := os.Getenv("WATSONX_URL")
	if baseURL == "" { baseURL = "https://us-south.ml.cloud.ibm.com" }

	fmt.Println("=== NomVox LLM Test ===")
	fmt.Printf("Base URL   : %s\n", baseURL)
	fmt.Printf("API Key    : %s...%s\n", apiKey[:4], apiKey[len(apiKey)-4:])
	fmt.Printf("Project ID : %s...%s\n\n", projectID[:4], projectID[len(projectID)-4:])

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 25 * time.Second}

	// ── Step 1: IAM token ─────────────────────────────────────────
	fmt.Print("Step 1 — IAM token exchange... ")
	token, err := fetchToken(ctx, client, apiKey)
	if err != nil {
		fmt.Printf("FAIL\n  %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OK (%d chars)\n", len(token))

	// ── Step 2: Tiny chat call — one sentence ─────────────────────
	fmt.Print("Step 2 — Chat API (1 name, tiny prompt)... ")
	modelID := "meta-llama/llama-3-3-70b-instruct"
	if !strings.Contains(baseURL, "ca-tor") {
		modelID = "ibm/granite-3-8b-instruct"
	}

	tinySystem := `Return ONLY a JSON array with exactly 1 object: {"name":"...","tagline":"..."}. No other text.`
	tinyUser   := `One brand name for an eco coffee brand. JSON array only.`

	raw, err := chatCall(ctx, client, token, baseURL, modelID, projectID, tinySystem, tinyUser, 80)
	if err != nil {
		fmt.Printf("FAIL\n  %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OK\n  Raw: %q\n", truncate(raw, 200))

	// ── Step 3: JSON parse ────────────────────────────────────────
	fmt.Print("Step 3 — JSON parse... ")
	jsonStr := extractJSON(raw)
	var cards []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &cards); err != nil {
		fmt.Printf("FAIL\n  raw=%q\n  err=%v\n", jsonStr, err)
		os.Exit(1)
	}
	fmt.Printf("OK — %d card(s)\n", len(cards))
	for _, c := range cards {
		fmt.Printf("  name=%v  tagline=%v\n", c["name"], c["tagline"])
	}

	// ── Step 4: Full NomVox prompt (4 names, 30s timeout) ─────────
	fmt.Print("\nStep 4 — Full NomVox prompt (4 names, 30s)... ")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	fullSystem := buildSystemPrompt()
	fullUser   := `Generate 4 brand names for an eco-friendly coffee brand targeting young professionals, personality: minimal warm honest, color mood: green cream. Return ONLY the JSON array.`

	raw2, err := chatCall(ctx2, client, token, baseURL, modelID, projectID, fullSystem, fullUser, 1600)
	if err != nil {
		fmt.Printf("FAIL\n  %v\n", err)
		fmt.Println("\nDiagnosis: Full prompt likely too large or model timed out.")
		fmt.Println("Fix: reduce max_new_tokens or simplify the system prompt.")
		os.Exit(1)
	}
	fmt.Printf("OK (%d chars)\n", len(raw2))

	jsonStr2 := extractJSON(raw2)
	var cards2 []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr2), &cards2); err != nil {
		fmt.Printf("  JSON parse FAIL: %v\n", err)
		fmt.Printf("  Raw (first 300): %q\n", truncate(raw2, 300))
	} else {
		fmt.Printf("  %d cards parsed OK\n", len(cards2))
		for _, c := range cards2 {
			fmt.Printf("    [%v] %v\n", c["name"], c["tagline"])
		}
	}

	fmt.Println("\n=== ALL STEPS PASSED ===")
}

// ── Helpers ───────────────────────────────────────────────────────

func fetchToken(ctx context.Context, client *http.Client, apiKey string) (string, error) {
	body := "grant_type=urn%3Aibm%3Aparams%3Aoauth%3Agrant-type%3Aapikey&apikey=" + apiKey
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://iam.cloud.ibm.com/identity/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil { return "", fmt.Errorf("IAM request: %w", err) }
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 { return "", fmt.Errorf("IAM HTTP %d: %s", resp.StatusCode, b) }
	var r struct{ AccessToken string `json:"access_token"` }
	json.Unmarshal(b, &r)
	if r.AccessToken == "" { return "", fmt.Errorf("empty access_token: %s", b) }
	return r.AccessToken, nil
}

func chatCall(ctx context.Context, client *http.Client, token, baseURL, modelID, projectID, system, user string, maxTokens int) (string, error) {
	payload := map[string]interface{}{
		"model_id":   modelID,
		"project_id": projectID,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user",   "content": user},
		},
		"parameters": map[string]interface{}{
			"decoding_method":    "greedy",
			"max_new_tokens":     maxTokens,
			"temperature":        0.7,
			"repetition_penalty": 1.1,
		},
	}
	b, _ := json.Marshal(payload)
	endpoint := baseURL + "/ml/v1/text/chat?version=2023-05-29"
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil { return "", fmt.Errorf("chat HTTP: %w", err) }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 { return "", fmt.Errorf("chat HTTP %d: %s", resp.StatusCode, body) }

	var env struct {
		Choices []struct {
			Message struct{ Content string `json:"content"` } `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(body, &env)
	if len(env.Choices) > 0 { return env.Choices[0].Message.Content, nil }
	return "", fmt.Errorf("no choices in response: %s", truncate(string(body), 200))
}

func extractJSON(text string) string {
	text = strings.TrimSpace(text)
	if i := strings.Index(text, "```json"); i != -1 { text = text[i+7:] } else
	if i := strings.Index(text, "```"); i != -1 { text = text[i+3:] }
	if i := strings.LastIndex(text, "```"); i != -1 { text = text[:i] }
	start := strings.Index(text, "[")
	end   := strings.LastIndex(text, "]")
	if start != -1 && end > start { return strings.TrimSpace(text[start : end+1]) }
	return strings.TrimSpace(text)
}

func truncate(s string, n int) string {
	if len(s) <= n { return s }
	return s[:n] + "..."
}

func buildSystemPrompt() string {
	return `You are NomVox, an AI brand-naming partner. Return ONLY a valid JSON array — no markdown, no extra text.

Rules: Names ≤12 chars. Taglines ≤7 words. style_tags: 2-3 tags. All strings concise. squatter_risk: "Low","Medium","High". Return ONLY [ ... ].

Schema: {"name":string,"tagline":string,"tone_reasoning":string,"style_tags":[string],"short_desc":string,"long_desc":string,"origin_story":string,"score":{"memorability":int,"spellability":int,"global_safety":int,"squatter_risk":string,"mem_reasoning":string,"spell_reasoning":string,"global_reasoning":string,"squatter_reasoning":string},"voice_samples":{"instagram_caption":string,"email_subject":string,"not_found_message":string}}`
}
