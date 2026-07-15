package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// GraniteClient calls the IBM watsonx.ai Granite LLM endpoint.
type GraniteClient struct {
	apiKey     string
	projectID  string
	baseURL    string
	httpClient *http.Client
}

// NewGraniteClient creates a client from environment variables.
func NewGraniteClient() (*GraniteClient, error) {
	apiKey := os.Getenv("WATSONX_API_KEY")
	projectID := os.Getenv("WATSONX_PROJECT_ID")
	baseURL := os.Getenv("WATSONX_URL")
	if apiKey == "" || projectID == "" {
		return nil, fmt.Errorf("granite: WATSONX_API_KEY and WATSONX_PROJECT_ID must be set")
	}
	if baseURL == "" {
		baseURL = "https://us-south.ml.cloud.ibm.com"
	}
	return &GraniteClient{
		apiKey:    apiKey,
		projectID: projectID,
		baseURL:   baseURL,
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}, nil
}

// ── Request / Response types ──────────────────────────────────────────────────

// watsonxTextRequest is for the /ml/v1/text/generation endpoint (Granite base models).
type watsonxTextRequest struct {
	ModelID    string        `json:"model_id"`
	ProjectID  string        `json:"project_id"`
	Input      string        `json:"input"`
	Parameters watsonxParams `json:"parameters"`
}

// watsonxChatRequest is for the /ml/v1/text/chat endpoint (Llama instruct / Granite instruct).
// Using the chat endpoint avoids the repetition-loop issue seen when concatenating
// system+user into a single "input" field on instruction-tuned models.
type watsonxChatRequest struct {
	ModelID    string         `json:"model_id"`
	ProjectID  string         `json:"project_id"`
	Messages   []chatMessage  `json:"messages"`
	Parameters watsonxParams  `json:"parameters"`
}

type chatMessage struct {
	Role    string `json:"role"`    // "system" | "user" | "assistant"
	Content string `json:"content"`
}

type watsonxParams struct {
	DecodingMethod    string  `json:"decoding_method"`
	MaxNewTokens      int     `json:"max_new_tokens"`
	Temperature       float64 `json:"temperature"`
	RepetitionPenalty float64 `json:"repetition_penalty"`
}

// ── Model selection ────────────────────────────────────────────────────────────

// endpointCaps describes what API format and model a given endpoint supports.
type endpointCaps struct {
	modelID    string // model to use
	useChatAPI bool   // true = /ml/v1/text/chat, false = /ml/v1/text/generation
}

// capsForEndpoint returns the model and API format for the given base URL.
// ca-tor endpoint catalog (verified 2026-07-15):
//   - meta/llama-3-3-70b-instruct (instruct → use chat API)
// us-south endpoint catalog:
//   - ibm/granite-3-8b-instruct (instruct → use chat API)
func capsForEndpoint(baseURL string) endpointCaps {
	if strings.Contains(baseURL, "ca-tor") {
		return endpointCaps{
			modelID:    "meta-llama/llama-3-3-70b-instruct",
			useChatAPI: true,
		}
	}
	return endpointCaps{
		modelID:    "ibm/granite-3-8b-instruct",
		useChatAPI: true,
	}
}

// ── Generate ──────────────────────────────────────────────────────────────────

// Generate sends a prompt to the LLM endpoint and returns the raw text response.
// Uses the chat completions endpoint to avoid the repetition-loop issue with
// instruction-tuned models when system+user are concatenated into a single input.
func (c *GraniteClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Get IAM token
	tokenClient := NewIBMTokenClient(c.apiKey)
	iamToken, err := tokenClient.FetchToken(ctx)
	if err != nil {
		return "", fmt.Errorf("granite: IAM token exchange: %w", err)
	}

	caps := capsForEndpoint(c.baseURL)
	params := watsonxParams{
		DecodingMethod:    "greedy",
		MaxNewTokens:      1800, // 1800 tokens — fits 4 cards reliably; lower = faster inference
		Temperature:       0.7,
		RepetitionPenalty: 1.05,
	}

	if caps.useChatAPI {
		return c.generateChat(ctx, iamToken, caps.modelID, systemPrompt, userPrompt, params)
	}
	return c.generateText(ctx, iamToken, caps.modelID, systemPrompt+"\n\n"+userPrompt, params)
}

// generateChat uses POST /ml/v1/text/chat (OpenAI-compatible messages format).
func (c *GraniteClient) generateChat(ctx context.Context, token, modelID, systemPrompt, userPrompt string, params watsonxParams) (string, error) {
	reqBody := watsonxChatRequest{
		ModelID:   modelID,
		ProjectID: c.projectID,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Parameters: params,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("granite: chat marshal: %w", err)
	}

	endpoint := fmt.Sprintf("%s/ml/v1/text/chat?version=2023-05-29", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("granite: chat build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("granite: chat http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("granite: chat status %d: %s", resp.StatusCode, string(b))
	}

	// Chat response envelope: choices[0].message.content
	var chatEnvelope struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		// Also handle watsonx-style results[] fallback
		Results []struct {
			GeneratedText string `json:"generated_text"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatEnvelope); err != nil {
		return "", fmt.Errorf("granite: chat decode: %w", err)
	}
	if len(chatEnvelope.Choices) > 0 && chatEnvelope.Choices[0].Message.Content != "" {
		return chatEnvelope.Choices[0].Message.Content, nil
	}
	if len(chatEnvelope.Results) > 0 && chatEnvelope.Results[0].GeneratedText != "" {
		return chatEnvelope.Results[0].GeneratedText, nil
	}
	return "", fmt.Errorf("granite: chat empty response")
}

// generateText uses POST /ml/v1/text/generation (raw input string).
func (c *GraniteClient) generateText(ctx context.Context, token, modelID, fullInput string, params watsonxParams) (string, error) {
	reqBody := watsonxTextRequest{
		ModelID:    modelID,
		ProjectID:  c.projectID,
		Input:      fullInput,
		Parameters: params,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("granite: text marshal: %w", err)
	}

	endpoint := fmt.Sprintf("%s/ml/v1/text/generation?version=2023-05-29", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("granite: text build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("granite: text http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("granite: text status %d: %s", resp.StatusCode, string(b))
	}

	var envelope struct {
		Results []struct {
			GeneratedText string `json:"generated_text"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return "", fmt.Errorf("granite: text decode: %w", err)
	}
	if len(envelope.Results) == 0 || envelope.Results[0].GeneratedText == "" {
		return "", fmt.Errorf("granite: text empty response")
	}
	return envelope.Results[0].GeneratedText, nil
}

// ExtractJSONObject strips markdown fences and extracts a top-level JSON object {…}.
func ExtractJSONObject(text string) string {
	text = strings.TrimSpace(text)
	// Strip markdown fences
	if idx := strings.Index(text, "```json"); idx != -1 {
		text = text[idx+7:]
	} else if idx := strings.Index(text, "```"); idx != -1 {
		text = text[idx+3:]
	}
	if idx := strings.LastIndex(text, "```"); idx != -1 {
		text = text[:idx]
	}
	text = strings.TrimSpace(text)

	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}
	end := strings.LastIndex(text, "}")
	if end != -1 && end > start {
		return strings.TrimSpace(text[start : end+1])
	}
	return ""
}

// parseJSON is a thin wrapper around json.Unmarshal for use within this package.
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ExtractJSON strips markdown fences and extracts the JSON array.
// If the model was cut off before closing the array, it attempts to repair
// the truncated JSON by closing any open object and the array.
func ExtractJSON(text string) string {
	// Remove markdown fences if present
	text = strings.TrimSpace(text)
	if idx := strings.Index(text, "```json"); idx != -1 {
		text = text[idx+7:]
	} else if idx := strings.Index(text, "```"); idx != -1 {
		text = text[idx+3:]
	}
	if idx := strings.LastIndex(text, "```"); idx != -1 {
		text = text[:idx]
	}
	text = strings.TrimSpace(text)

	// Find the outermost JSON array bounds
	start := strings.Index(text, "[")
	if start == -1 {
		return text
	}

	// If there's a clean closing ], use it
	end := strings.LastIndex(text, "]")
	if end != -1 && end > start {
		return strings.TrimSpace(text[start : end+1])
	}

	// Array was truncated — attempt repair:
	// 1. Take everything from [ onward
	// 2. Find the last complete object (last "}")
	// 3. Close the array after it
	partial := text[start:]
	lastObj := strings.LastIndex(partial, "}")
	if lastObj != -1 {
		repaired := strings.TrimSpace(partial[:lastObj+1]) + "\n]"
		return repaired
	}

	return strings.TrimSpace(text)
}

// IBMTokenClient fetches a bearer token from either:
//   - Public IBM Cloud IAM  (baseURL starts with https://us-south / ca-tor .ml.cloud.ibm.com)
//   - Cloud Pak for Data    (baseURL contains .dai.cloud.ibm.com or CPD_URL env var is set)
//
// Detection is automatic: if WATSONX_CPD_URL is set, CPD auth is used;
// otherwise public IBM Cloud IAM is used.
type IBMTokenClient struct {
	apiKey     string
	cpdURL     string // e.g. https://ca-tor.dai.cloud.ibm.com  (empty = public cloud)
	httpClient *http.Client
}

// tokenCache avoids re-fetching the IAM token on every LLM call.
// Tokens are valid for 3600s; we cache for 3000s to be safe.
var (
	cachedToken     string
	cachedTokenExp  time.Time
	cachedTokenLock sync.Mutex
)

func NewIBMTokenClient(apiKey string) *IBMTokenClient {
	return &IBMTokenClient{
		apiKey:     apiKey,
		cpdURL:     os.Getenv("WATSONX_CPD_URL"), // set this for Cloud Pak for Data
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *IBMTokenClient) FetchToken(ctx context.Context) (string, error) {
	// For CPD, never cache — CPD tokens are short-lived per-user
	if t.cpdURL != "" {
		return t.fetchCPDToken(ctx)
	}
	// For public IBM Cloud IAM, cache the token for 50 minutes
	cachedTokenLock.Lock()
	defer cachedTokenLock.Unlock()
	if cachedToken != "" && time.Now().Before(cachedTokenExp) {
		return cachedToken, nil
	}
	tok, err := t.fetchIAMToken(ctx)
	if err != nil {
		return "", err
	}
	cachedToken    = tok
	cachedTokenExp = time.Now().Add(50 * time.Minute)
	return tok, nil
}

// fetchIAMToken exchanges an IBM Cloud API key for a short-lived IAM bearer token.
// Used for public watsonx.ai (ml.cloud.ibm.com endpoints).
func (t *IBMTokenClient) fetchIAMToken(ctx context.Context) (string, error) {
	body := "grant_type=urn%3Aibm%3Aparams%3Aoauth%3Agrant-type%3Aapikey&apikey=" + t.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://iam.cloud.ibm.com/identity/token",
		strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("IAM token exchange returned empty token — check WATSONX_API_KEY")
	}
	return result.AccessToken, nil
}

// fetchCPDToken exchanges a Cloud Pak for Data API key for a CPD bearer token.
// Used when WATSONX_CPD_URL is set (e.g. https://ca-tor.dai.cloud.ibm.com).
func (t *IBMTokenClient) fetchCPDToken(ctx context.Context) (string, error) {
	payload, _ := json.Marshal(map[string]string{"apikey": t.apiKey})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		t.cpdURL+"/icp4d-api/v1/authorize",
		bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Token == "" {
		return "", fmt.Errorf("CPD token exchange returned empty token — check WATSONX_API_KEY and WATSONX_CPD_URL")
	}
	return result.Token, nil
}

var _ = io.Discard
