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

// ImagenClient calls Google AI Studio's Imagen 3 endpoint for image generation
// and Gemini 2.0 Flash for vision analysis.
// Requires GOOGLE_AI_API_KEY in the environment.
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
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// ── Imagen 3 — image generation ───────────────────────────────────────────────

const imagenEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/imagen-3.0-generate-002:predict"

type imagenRequest struct {
	Instances  []imagenInstance  `json:"instances"`
	Parameters imagenParameters  `json:"parameters"`
}

type imagenInstance struct {
	Prompt string `json:"prompt"`
}

type imagenParameters struct {
	SampleCount      int    `json:"sampleCount"`
	AspectRatio      string `json:"aspectRatio"`
	SafetyFilterLevel string `json:"safetFilterLevel,omitempty"`
}

type imagenResponse struct {
	Predictions []struct {
		BytesBase64Encoded string `json:"bytesBase64Encoded"`
		MimeType           string `json:"mimeType"`
	} `json:"predictions"`
}

// GenerateImages sends a prompt to Imagen 3 and returns base64-encoded PNG images.
// count must be 1–4.
func (c *ImagenClient) GenerateImages(ctx context.Context, prompt string, count int, aspectRatio string) ([]string, error) {
	if count < 1 {
		count = 1
	}
	if count > 4 {
		count = 4
	}
	if aspectRatio == "" {
		aspectRatio = "1:1"
	}

	reqBody := imagenRequest{
		Instances:  []imagenInstance{{Prompt: prompt}},
		Parameters: imagenParameters{SampleCount: count, AspectRatio: aspectRatio},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("imagen: marshal: %w", err)
	}

	url := imagenEndpoint + "?key=" + c.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("imagen: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("imagen: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("imagen: status %d: %s", resp.StatusCode, TruncateStr(string(b), 300))
	}

	var result imagenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("imagen: decode: %w", err)
	}

	out := make([]string, 0, len(result.Predictions))
	for _, p := range result.Predictions {
		out = append(out, p.BytesBase64Encoded)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("imagen: no images returned")
	}
	return out, nil
}

// ── Gemini 2.0 Flash — vision analysis ───────────────────────────────────────

const geminiVisionEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

// AnalyseImage sends a base64-encoded image to Gemini 2.0 Flash and returns
// a structured VisionContext (colours, mood, style, textures).
func (c *ImagenClient) AnalyseImage(ctx context.Context, base64Image, mimeType string) (string, error) {
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	type inlinePart struct {
		InlineData struct {
			MimeType string `json:"mimeType"`
			Data     string `json:"data"`
		} `json:"inlineData"`
	}
	type textPart struct {
		Text string `json:"text"`
	}

	// Build multipart request
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
// Imagen 3 returns URL-safe base64 (RFC 4648 §5, uses - and _ instead of + and /).
// We try RawURLEncoding first, then RawStdEncoding, then StdEncoding so the function
// works regardless of which variant the upstream API returns.
func Base64ToDataURI(b64 string, mimeType string) string {
	if b64 == "" {
		return ""
	}
	if mimeType == "" {
		mimeType = "image/png"
	}
	// Try URL-safe (no padding) — Imagen 3 primary format
	if _, err := base64.RawURLEncoding.DecodeString(b64); err == nil {
		return "data:" + mimeType + ";base64," + b64
	}
	// Try standard (no padding) — some proxies strip padding
	if _, err := base64.RawStdEncoding.DecodeString(b64); err == nil {
		return "data:" + mimeType + ";base64," + b64
	}
	// Try standard with padding — legacy fallback
	if _, err := base64.StdEncoding.DecodeString(b64); err == nil {
		return "data:" + mimeType + ";base64," + b64
	}
	return ""
}
