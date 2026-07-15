package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ImagenClient wraps Google AI image-generation and vision endpoints.
// Image generation uses Gemini 2.0 Flash (free tier, native image output).
// Vision analysis uses Gemini 2.0 Flash (text output).
// Both share the same GOOGLE_AI_API_KEY.
type ImagenClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewImagenClient() (*ImagenClient, error) {
	key := os.Getenv("GOOGLE_AI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("imagen: GOOGLE_AI_API_KEY not set")
	}
	return &ImagenClient{
		apiKey:     key,
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}, nil
}

// ── Gemini 2.0 Flash — native image generation ────────────────────────────────
//
// Endpoint: /v1beta/models/gemini-2.5-flash-image:generateContent
// The model is prompted with responseModalities ["IMAGE","TEXT"].
// Each response Part may be either text or inlineData (base64 PNG/JPEG).
// We collect every inlineData part and return the raw base64 strings.
//
// Free-tier limits (as of 2026): 15 req/min, 1500 req/day.
// Stable model (no -preview suffix). Replaces Imagen 3 which required paid credits.
// Fallback: gemini-3.1-flash-image (newer) if gemini-2.5-flash-image is unavailable.

const geminiImageGenEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/" +
	"gemini-2.5-flash-image:generateContent"

type geminiImageGenRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiImageGenConfig   `json:"generationConfig"`
}

type geminiContent struct {
	Role  string        `json:"role"`
	Parts []geminiPart  `json:"parts"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64
}

type geminiImageGenConfig struct {
	ResponseModalities []string `json:"responseModalities"`
	// CandidateCount is not used here — we loop and call once per image for reliability.
}

type geminiImageGenResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text       string `json:"text,omitempty"`
				InlineData *struct {
					MimeType string `json:"mimeType"`
					Data     string `json:"data"`
				} `json:"inlineData,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// GenerateImages sends a prompt to Gemini 2.0 Flash image generation.
// count controls how many images to return (1–4). Each image is fetched
// in a separate request to avoid quota issues.
// Returns a slice of raw base64-encoded PNG strings.
func (c *ImagenClient) GenerateImages(ctx context.Context, prompt string, count int, aspectRatio string) ([]string, error) {
	if count < 1 {
		count = 1
	}
	if count > 4 {
		count = 4
	}
	// Embed aspect ratio hint into the prompt since the API doesn't accept it as a parameter.
	fullPrompt := prompt
	if aspectRatio != "" && aspectRatio != "1:1" {
		fullPrompt = fmt.Sprintf("%s\n\nImage aspect ratio: %s.", prompt, aspectRatio)
	}

	out := make([]string, 0, count)
	for i := 0; i < count; i++ {
		b64, err := c.generateOneImage(ctx, fullPrompt)
		if err != nil {
			// Tolerate partial failures — return whatever we got so far.
			if len(out) == 0 {
				return nil, err // fail only if we got nothing at all
			}
			break
		}
		out = append(out, b64)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("gemini-image: no images returned")
	}
	return out, nil
}

// generateOneImage calls the Gemini image-generation endpoint once and returns
// the first inlineData (base64) part from the response.
func (c *ImagenClient) generateOneImage(ctx context.Context, prompt string) (string, error) {
	reqBody := geminiImageGenRequest{
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: prompt}},
			},
		},
		GenerationConfig: geminiImageGenConfig{
			ResponseModalities: []string{"IMAGE", "TEXT"},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("gemini-image: marshal: %w", err)
	}

	url := geminiImageGenEndpoint + "?key=" + c.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("gemini-image: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini-image: http: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gemini-image: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini-image: status %d: %s",
			resp.StatusCode, TruncateStr(string(respBytes), 400))
	}

	var result geminiImageGenResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("gemini-image: decode: %w", err)
	}

	// Surface API-level errors embedded in a 200 response.
	if result.Error != nil {
		return "", fmt.Errorf("gemini-image: API error %d %s: %s",
			result.Error.Code, result.Error.Status, result.Error.Message)
	}

	// Walk all candidates + parts, return first image found.
	for _, cand := range result.Candidates {
		for _, part := range cand.Content.Parts {
			if part.InlineData != nil && part.InlineData.Data != "" {
				return part.InlineData.Data, nil
			}
		}
	}
	return "", fmt.Errorf("gemini-image: no image part in response (candidates=%d)", len(result.Candidates))
}

// ── Gemini 2.0 Flash — vision analysis ───────────────────────────────────────

const geminiVisionEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

// AnalyseImage sends a base64-encoded image to Gemini 2.0 Flash and returns
// a structured VisionContext (colours, mood, style, textures) as raw JSON text.
func (c *ImagenClient) AnalyseImage(ctx context.Context, base64Image, mimeType string) (string, error) {
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []interface{}{
					map[string]interface{}{
						"inlineData": map[string]string{
							"mimeType": mimeType,
							"data":     base64Image,
						},
					},
					map[string]string{
						"text": `Analyse this image for brand identity inspiration. Return ONLY a JSON object with these fields:
{"colours":["#hex1","#hex2","#hex3"],"mood":["word1","word2","word3"],"style":["style1","style2"],"textures":["texture1","texture2"]}
- colours: 3 dominant hex colours
- mood: 3 single-word mood descriptors
- style: 2 style descriptors (e.g. "minimalist","earthy")
- textures: 2 texture descriptors (e.g. "matte","linen")
Return ONLY the JSON object, no other text.`,
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("gemini vision: marshal: %w", err)
	}

	url := geminiVisionEndpoint + "?key=" + c.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("gemini vision: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini vision: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini vision: status %d: %s", resp.StatusCode, TruncateStr(string(b), 300))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("gemini vision: decode: %w", err)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini vision: empty response")
	}
	return result.Candidates[0].Content.Parts[0].Text, nil
}

// Base64ToDataURI converts a raw base64 string to a data URI for embedding in HTML/JSON.
// Gemini returns standard base64 (with padding). We try all variants for robustness.
func Base64ToDataURI(b64 string, mimeType string) string {
	if b64 == "" {
		return ""
	}
	if mimeType == "" {
		mimeType = "image/png"
	}
	// Try standard (with padding) — Gemini primary format
	if _, err := base64.StdEncoding.DecodeString(b64); err == nil {
		return "data:" + mimeType + ";base64," + b64
	}
	// Try URL-safe (no padding) — Imagen 3 / older callers
	if _, err := base64.RawURLEncoding.DecodeString(b64); err == nil {
		return "data:" + mimeType + ";base64," + b64
	}
	// Try standard (no padding) — some proxies strip padding
	if _, err := base64.RawStdEncoding.DecodeString(b64); err == nil {
		return "data:" + mimeType + ";base64," + b64
	}
	return ""
}
