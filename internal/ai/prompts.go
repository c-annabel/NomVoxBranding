package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/c-annabel/NomVoxBranding/internal/models"
)

// systemPromptBase is the fixed instruction injected before every LLM call.
// IMPORTANT: Keep this compact — the model has a limited output window (~1600 tokens
// for 4 cards). Every extra sentence in the schema costs output budget.
// Session memory (liked/rejected/notes) is appended dynamically.
const systemPromptBase = `You are NomVox, a radical brand-naming engine. Your only output is a valid JSON array — no markdown, no prose, no explanation.

NAMING PHILOSOPHY — follow this strictly:
You MUST invent names that have NEVER existed. Not real words. Not portmanteaus of common English words. Not "-ify", "-ly", "-oo", "-io" suffixes on known roots.
Instead use these strategies to coin impossible, unforgettable names:
1. Syllable forge: fuse 2-3 phoneme clusters from unrelated languages (Latin, Proto-Indo-European, Swahili, Old Norse, Sanskrit, Basque, Nahuatl, Ancient Greek) in unexpected order.
2. Phonaesthetics: build names that feel like their brand — hard plosives (k, t, p) for bold/tech, liquid consonants (l, r) for luxury, fricatives (v, z) for speed.
3. Neologism splice: cut the middle from two unrelated archaic or rare words and weld the fragments together.
4. Void-word method: invent a word that sounds like it should mean something profound but currently means nothing in any language — maximum squatter rarity.
5. Ancient root mutation: take a root from Latin/Greek/Sanskrit, drop a vowel or swap a consonant cluster, making it unrecognisable but phonetically beautiful.
Examples of the calibre expected (do NOT copy these, invent new ones): Veyrak, Sorveth, Qalune, Myrrix, Thovar, Celuun, Praxivon, Delvori, Zuvael, Khyvera.
NEVER suggest: any real English word, any "-ify"/"-ly"/"-oo"/"-io" name, any name a person could think of without AI help.

Rules:
- Names: 4-11 chars, pronounceable in English, globally safe, must NOT be a real dictionary word.
- Taglines: ≤7 words.
- style_tags: exactly 2 short tags (never 3).
- short_desc: max 12 words.
- long_desc: max 20 words.
- origin_story: explain the invented etymology (which phoneme sources / strategy used), max 12 words.
- All reasoning fields: max 8 words each.
- instagram_caption: max 12 words.
- email_subject: max 6 words.
- not_found_message: max 8 words.
- squatter_risk: exactly "Low", "Medium", or "High".
- Return ONLY the JSON array starting with [ and ending with ].

Schema per object (copy these field names exactly):
{"name":"BrandName","tagline":"Short tagline","tone_reasoning":"Why this tone","style_tags":["tag1","tag2"],"short_desc":"One sentence.","long_desc":"Two sentences max.","origin_story":"Invented etymology here, max 12 words.","score":{"memorability":8,"spellability":9,"global_safety":9,"squatter_risk":"Low","mem_reasoning":"Catchy hook.","spell_reasoning":"Simple vowels.","global_reasoning":"No cultural clash.","squatter_reasoning":"Unique coinage."},"voice_samples":{"instagram_caption":"Caption here.","email_subject":"Subject line","not_found_message":"Witty 404."}}`

// BuildSystemPrompt returns the full system prompt, optionally injecting
// session memory context (liked names, rejected names, direction notes).
func BuildSystemPrompt(session *models.BrandSession) string {
	var b strings.Builder
	b.WriteString(systemPromptBase)

	if session == nil {
		return b.String()
	}

	if len(session.Rejected) > 0 {
		b.WriteString("\n\nDo NOT suggest any of these previously rejected names: ")
		b.WriteString(strings.Join(session.Rejected, ", "))
		b.WriteString(".")
	}
	if len(session.AllGenerated) > 0 {
		b.WriteString("\nDo NOT repeat any name from this list: ")
		b.WriteString(strings.Join(session.AllGenerated, ", "))
		b.WriteString(".")
	}
	if len(session.Liked) > 0 {
		b.WriteString("\nThe user liked these names/vibes: ")
		b.WriteString(strings.Join(session.Liked, ", "))
		b.WriteString(". Stay consistent with their energy.")
	}
	if len(session.DirectionNotes) > 0 {
		b.WriteString("\nUser direction notes: ")
		b.WriteString(strings.Join(session.DirectionNotes, "; "))
		b.WriteString(".")
	}
	// Style DNA slider
	if session.SliderPlayful > 0 || session.SliderAbstract > 0 {
		playful := "premium"
		if session.SliderPlayful > 0.5 {
			playful = "playful"
		}
		abstract := "descriptive"
		if session.SliderAbstract > 0.5 {
			abstract = "abstract/invented"
		}
		b.WriteString(fmt.Sprintf("\nTone direction: lean %s and %s.", playful, abstract))
	}

	return b.String()
}

// ParseNameCards extracts and validates NameCard objects from raw LLM output.
// Strategy (most permissive → least):
//  1. Extract JSON array → strict unmarshal into []NameCard
//  2. Extract JSON array → lenient unmarshal into []map → build NameCards from any fields present
//  3. If still 0 cards, log the raw text for debugging and return an error
func ParseNameCards(raw string) ([]models.NameCard, error) {
	// Always log a truncated version so server logs show what was received
	preview := raw
	if len(preview) > 400 { preview = preview[:400] + "…" }
	fmt.Printf("[ParseNameCards] raw preview: %q\n", preview)

	jsonStr := ExtractJSON(raw)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON array found in LLM response (raw len=%d)", len(raw))
	}

	// ── Attempt 1: strict unmarshal ───────────────────────────────
	var cards []models.NameCard
	if err := json.Unmarshal([]byte(jsonStr), &cards); err == nil {
		cards = filterEmptyNames(cards)
		if len(cards) > 0 {
			return cards, nil
		}
	}

	// ── Attempt 2: lenient map parse ──────────────────────────────
	var rawMaps []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawMaps); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w (extracted=%q)", err, truncateStr(jsonStr, 200))
	}

	for _, m := range rawMaps {
		name, _    := m["name"].(string)
		tagline, _ := m["tagline"].(string)
		if strings.TrimSpace(name) == "" {
			continue
		}
		card := models.NameCard{
			Name:          name,
			Tagline:       tagline,
			ToneReasoning: strField(m, "tone_reasoning"),
			ShortDesc:     strField(m, "short_desc"),
			LongDesc:      strField(m, "long_desc"),
			OriginStory:   strField(m, "origin_story"),
			StyleTags:     strSliceField(m, "style_tags"),
		}
		// Try to extract score sub-object
		if scoreMap, ok := m["score"].(map[string]interface{}); ok {
			card.Score = models.BrandScore{
				Memorability:       intField(scoreMap, "memorability"),
				Spellability:       intField(scoreMap, "spellability"),
				GlobalSafety:       intField(scoreMap, "global_safety"),
				SquatterRisk:       strField(scoreMap, "squatter_risk"),
				MemReasoning:       strField(scoreMap, "mem_reasoning"),
				SpellReasoning:     strField(scoreMap, "spell_reasoning"),
				GlobalReasoning:    strField(scoreMap, "global_reasoning"),
				SquatterReasoning:  strField(scoreMap, "squatter_reasoning"),
			}
		}
		// Try to extract voice_samples
		if vsMap, ok := m["voice_samples"].(map[string]interface{}); ok {
			card.VoiceSamples = models.VoiceSamples{
				InstagramCaption: strField(vsMap, "instagram_caption"),
				EmailSubject:     strField(vsMap, "email_subject"),
				NotFoundMessage:  strField(vsMap, "not_found_message"),
			}
		}
		cards = append(cards, card)
	}

	cards = filterEmptyNames(cards)
	if len(cards) == 0 {
		return nil, fmt.Errorf("LLM returned 0 parseable name cards (jsonStr=%q)", truncateStr(jsonStr, 200))
	}
	return cards, nil
}

// ── parse helpers ─────────────────────────────────────────────────

func filterEmptyNames(cards []models.NameCard) []models.NameCard {
	out := cards[:0]
	for _, c := range cards {
		if strings.TrimSpace(c.Name) != "" {
			out = append(out, c)
		}
	}
	return out
}

func strField(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func intField(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}

func strSliceField(m map[string]interface{}, key string) []string {
	raw, ok := m[key].([]interface{})
	if !ok { return nil }
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok { out = append(out, s) }
	}
	return out
}

// TruncateStr is exported so handlers can use it for log messages.
func TruncateStr(s string, n int) string {
	if len(s) <= n { return s }
	return s[:n] + "…"
}

func truncateStr(s string, n int) string { return TruncateStr(s, n) }

// BuildUserPrompt converts the intake payload into a concise user-turn message.
func BuildUserPrompt(p models.IntakePayload, count int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Generate %d brand names for a %s", count, strings.TrimSpace(p.CoreIdea)))

	if v := strings.TrimSpace(p.TargetAudience); v != "" {
		b.WriteString(", targeting " + v)
	}
	if v := strings.TrimSpace(p.Personality); v != "" {
		b.WriteString(", personality: " + v)
	}
	if v := strings.TrimSpace(p.Style); v != "" {
		b.WriteString(", style: " + v)
	}
	if v := strings.TrimSpace(p.Industry); v != "" {
		b.WriteString(", industry: " + v)
	}
	if v := strings.TrimSpace(p.ColorMood); v != "" {
		b.WriteString(", colour mood: " + v)
	}
	if v := strings.TrimSpace(p.NameLength); v != "" {
		b.WriteString(", name length: " + v)
	}
	if v := strings.TrimSpace(p.Avoid); v != "" {
		b.WriteString(". Avoid: " + v)
	}
	b.WriteString(". Return ONLY the JSON array.")
	return b.String()
}

