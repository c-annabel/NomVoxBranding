package handlers

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
)

// DiagnoseHandler handles GET /api/diagnose.
// Tests every credential in isolation and returns a structured report.
// Remove this endpoint before production deployment.
func DiagnoseHandler(w http.ResponseWriter, r *http.Request) {
	report := map[string]interface{}{}

	// ── 1. Environment variables ───────────────────────────────────
	apiKey    := os.Getenv("WATSONX_API_KEY")
	projectID := os.Getenv("WATSONX_PROJECT_ID")
	watsonURL := os.Getenv("WATSONX_URL")
	cpdURL    := os.Getenv("WATSONX_CPD_URL")
	redisURL  := os.Getenv("REDIS_URL")

	report["env"] = map[string]interface{}{
		"WATSONX_API_KEY":    maskSecret(apiKey),
		"WATSONX_PROJECT_ID": maskSecret(projectID),
		"WATSONX_URL":        watsonURL,
		"WATSONX_CPD_URL":    cpdURL,
		"REDIS_URL":          maskSecret(redisURL),
		"CPD_mode":           cpdURL != "",
	}

	// ── 2. Token exchange ──────────────────────────────────────────
	tokenReport := map[string]interface{}{}
	if apiKey == "" {
		tokenReport["status"] = "SKIP — WATSONX_API_KEY not set"
	} else {
		token, err := fetchTokenDiag(apiKey, cpdURL)
		if err != nil {
			tokenReport["status"] = "FAIL"
			tokenReport["error"]  = err.Error()
		} else {
			tokenReport["status"]       = "OK"
			tokenReport["token_prefix"] = token[:min(16, len(token))] + "..."
			tokenReport["token_length"] = len(token)
		}
	}
	report["token_exchange"] = tokenReport

	// ── 3. Granite API call (only if token OK) ─────────────────────
	graniteReport := map[string]interface{}{}
	if tokenStatus, ok := tokenReport["status"].(string); ok && tokenStatus == "OK" {
		token := tokenReport["token_prefix"].(string) // just for display; re-fetch below
		_ = token
		raw, err := testGraniteCall(apiKey, projectID, watsonURL, cpdURL)
		if err != nil {
			graniteReport["status"] = "FAIL"
			graniteReport["error"]  = err.Error()
		} else {
			graniteReport["status"]          = "OK"
			graniteReport["response_length"] = len(raw)
			graniteReport["response_prefix"] = truncate(raw, 200)
		}
	} else {
		graniteReport["status"] = "SKIP — token exchange failed"
	}
	report["granite_call"] = graniteReport

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(report)
}

// ── helpers ────────────────────────────────────────────────────────

func maskSecret(s string) string {
	if s == "" { return "(not set)" }
	if len(s) <= 8 { return "***" }
	return s[:4] + "..." + s[len(s)-4:]
}

func truncate(s string, n int) string {
	if len(s) <= n { return s }
	return s[:n] + "..."
}

func min(a, b int) int {
	if a < b { return a }
	return b
}

func fetchTokenDiag(apiKey, cpdURL string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	ctx    := context.Background()

	if cpdURL != "" {
		// CPD token
		payload, _ := json.Marshal(map[string]string{"apikey": apiKey})
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
			cpdURL+"/icp4d-api/v1/authorize", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil { return "", fmt.Errorf("CPD auth request failed: %w", err) }
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			return "", fmt.Errorf("CPD auth HTTP %d: %s", resp.StatusCode, string(b))
		}
		var result struct { Token string `json:"token"` }
		json.Unmarshal(b, &result)
		if result.Token == "" {
			return "", fmt.Errorf("CPD auth returned empty token. Raw: %s", truncate(string(b), 300))
		}
		return result.Token, nil
	}

	// IAM token
	body := "grant_type=urn%3Aibm%3Aparams%3Aoauth%3Agrant-type%3Aapikey&apikey=" + apiKey
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://iam.cloud.ibm.com/identity/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil { return "", fmt.Errorf("IAM request failed: %w", err) }
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("IAM HTTP %d: %s", resp.StatusCode, string(b))
	}
	var result struct { AccessToken string `json:"access_token"` }
	json.Unmarshal(b, &result)
	if result.AccessToken == "" {
		return "", fmt.Errorf("IAM returned empty access_token. Raw: %s", truncate(string(b), 300))
	}
	return result.AccessToken, nil
}

func testGraniteCall(apiKey, projectID, watsonURL, cpdURL string) (string, error) {
	token, err := fetchTokenDiag(apiKey, cpdURL)
	if err != nil { return "", err }

	if watsonURL == "" { watsonURL = "https://us-south.ml.cloud.ibm.com" }

	// Select correct model ID and endpoint based on region.
	// ca-tor: uses chat API + llama-3-3-70b-instruct
	// us-south: uses chat API + granite-3-8b-instruct
	modelID := "ibm/granite-3-8b-instruct"
	chatEndpoint := watsonURL + "/ml/v1/text/chat?version=2023-05-29"
	if strings.Contains(watsonURL, "ca-tor") {
		modelID = "meta-llama/llama-3-3-70b-instruct"
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model_id":   modelID,
		"project_id": projectID,
		"messages": []map[string]string{
			{"role": "system",  "content": "You are a helpful assistant. Reply in one word only."},
			{"role": "user",    "content": "Say hello."},
		},
		"parameters": map[string]interface{}{
			"decoding_method": "greedy",
			"max_new_tokens":  10,
		},
	})

	client := &http.Client{Timeout: 30 * time.Second}
	ctx    := context.Background()

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, chatEndpoint, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("POST %s: %w", chatEndpoint, err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode == 200 {
		return fmt.Sprintf("MODEL=%s URL=%s RESPONSE=%s", modelID, chatEndpoint, string(b)), nil
	}
	return "", fmt.Errorf("POST %s → HTTP %d: %s", chatEndpoint, resp.StatusCode, string(b))
}
