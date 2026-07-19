package handlers

import (
	"math/rand"
	"strings"

	"github.com/c-annabel/NomVoxBranding/internal/models"
)

// ── Prototype fallback name generation ───────────────────────────────────────
// When the LLM is unavailable (quota exhausted, rate limited, network down),
// the prototype must still demo end-to-end. fallbackNameCards returns curated,
// invented brand names lightly adapted to the user's intake. These are clearly
// plausible brand names with full card data so the rest of the flow
// (availability, visuals, export) works unchanged.

type fallbackSeed struct {
	Name    string
	Tagline string
	Origin  string
	Tags    []string
}

// Pool of invented, brandable names. None are real dictionary words, which
// keeps them personality-neutral enough to fit most intake briefs.
var fallbackPool = []fallbackSeed{
	{"Zhilakai", "Grow with us", "Basque and Sanskrit fusion", []string{"approachable", "playful"}},
	{"Thaluki", "Nurture their curiosity", "Sanskrit and Old Norse fusion", []string{"playful", "rounded"}},
	{"Volkori", "Discover and thrive", "Finnish and Māori fusion", []string{"friendly", "warm"}},
	{"Nimbara", "Bright minds, open doors", "Latin nimbus with African -ara suffix", []string{"optimistic", "rounded"}},
	{"Soluvia", "Light the way forward", "Latin sol (sun) + via (path)", []string{"warm", "aspirational"}},
	{"Kindero", "Where growing feels good", "Germanic kinder with Esperanto ending", []string{"approachable", "playful"}},
	{"Braventa", "Courage meets curiosity", "Brave + adventure blend", []string{"bold", "encouraging"}},
	{"Lumizen", "Calm, clear, curious", "Latin lumen + Japanese zen", []string{"calm", "modern"}},
	{"Orivelle", "Every start is golden", "French or (gold) + -velle (new)", []string{"premium", "warm"}},
	{"Tessarix", "Building blocks of brilliance", "Greek tessera (tile) + -ix", []string{"structured", "smart"}},
	{"Favoro", "In favour of you", "Latin favor with rounded -o ending", []string{"friendly", "simple"}},
	{"Miravent", "Look ahead, wonder more", "Latin mirari (wonder) + advent", []string{"aspirational", "modern"}},
}

// fallbackNameCards returns n curated NameCards adapted to the intake.
// Names containing any user-avoided word are filtered out. Selection is
// shuffled so Regenerate produces a different batch each click.
func fallbackNameCards(intake models.IntakePayload, n int) []models.NameCard {
	avoid := strings.ToLower(strings.TrimSpace(intake.Avoid))
	avoidWords := []string{}
	for _, w := range strings.FieldsFunc(avoid, func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		if w != "" {
			avoidWords = append(avoidWords, w)
		}
	}

	// Filter pool against avoid list.
	eligible := make([]fallbackSeed, 0, len(fallbackPool))
	for _, s := range fallbackPool {
		nameLower := strings.ToLower(s.Name)
		blocked := false
		for _, w := range avoidWords {
			if strings.Contains(nameLower, w) {
				blocked = true
				break
			}
		}
		if !blocked {
			eligible = append(eligible, s)
		}
	}
	if len(eligible) == 0 {
		eligible = fallbackPool
	}

	// Shuffle so each regenerate shows a fresh batch.
	rand.Shuffle(len(eligible), func(i, j int) { eligible[i], eligible[j] = eligible[j], eligible[i] })
	if n > len(eligible) {
		n = len(eligible)
	}

	// Contextual copy derived from intake.
	idea := strings.TrimSpace(intake.CoreIdea)
	if idea == "" {
		idea = "your brand"
	}
	audience := strings.TrimSpace(intake.TargetAudience)
	if audience == "" {
		audience = "your audience"
	}
	personality := strings.TrimSpace(intake.Personality)
	if personality == "" {
		personality = "friendly, modern"
	}

	cards := make([]models.NameCard, 0, n)
	for _, s := range eligible[:n] {
		cards = append(cards, models.NameCard{
			Name:          s.Name,
			Tagline:       s.Tagline,
			ToneReasoning: "Curated to match a " + personality + " tone.",
			StyleTags:     s.Tags,
			ShortDesc:     upperFirst(idea) + ".",
			LongDesc:      upperFirst(idea) + " — crafted for " + audience + " with a " + personality + " personality.",
			OriginStory:   s.Origin,
			Score: models.BrandScore{
				Memorability:      8,
				Spellability:      7,
				GlobalSafety:      9,
				SquatterRisk:      "Low",
				MemReasoning:      "Short, rhythmic, and distinctive.",
				SpellReasoning:    "Phonetic spelling with familiar letter patterns.",
				GlobalReasoning:   "No known negative meanings in major languages.",
				SquatterReasoning: "Invented word — low squatting likelihood.",
			},
			VoiceSamples: models.VoiceSamples{
				InstagramCaption: "Meet " + s.Name + " — " + strings.ToLower(s.Tagline) + ".",
				EmailSubject:     s.Name + ": " + s.Tagline,
				NotFoundMessage:  "Oops — this page wandered off. Let's get you home.",
			},
		})
	}
	return cards
}

// upperFirst capitalises the first rune of a string.
func upperFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}
