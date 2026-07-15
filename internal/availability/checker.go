// Package availability implements the NomVox handle/domain probing engine.
//
// Scoring weights (total = 6pts, 80% threshold = 4.8pts):
//   Domain     2.0 pts  (non-negotiable)
//   Instagram  1.0 pt
//   X          1.0 pt
//   TikTok     1.0 pt
//   Threads    0.5 pt
//   YouTube    0.5 pt
//
// Unknown probes (429 / 5xx) are treated as "taken" to be conservative.
package availability

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/c-annabel/NomVoxBranding/internal/models"
)

const (
	totalWeight   = 6.0
	passThreshold = 3.6 // 60%
)

type platformWeight struct {
	name   string
	weight float64
	probe  func(handle string, client *http.Client) (available bool, unknown bool)
}

var platforms = []platformWeight{
	{"domain", 2.0, probeDomain},
	{"instagram", 1.0, probeInstagram},
	{"x", 1.0, probeX},
	{"tiktok", 1.0, probeTikTok},
	{"threads", 0.5, probeThreads},
	{"youtube", 0.5, probeYouTube},
}

// CheckName probes all platforms concurrently and returns an AvailabilityResult.
func CheckName(name string) models.AvailabilityResult {
	handle := sanitiseHandle(name)
	client := &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	type result struct {
		platform  string
		available bool
		unknown   bool
		weight    float64
	}

	results := make([]result, len(platforms))
	var wg sync.WaitGroup

	for i, p := range platforms {
		i, p := i, p
		wg.Add(1)
		go func() {
			defer wg.Done()
			avail, unknown := p.probe(handle, client)
			results[i] = result{p.name, avail, unknown, p.weight}
		}()
	}
	wg.Wait()

	// Build PlatformResult struct
	pr := models.PlatformResult{}
	var score float64

	for _, res := range results {
		// Unknown → treated as taken (conservative), but flag it
		effectiveAvail := res.available && !res.unknown
		if effectiveAvail {
			score += res.weight
		}
		switch res.platform {
		case "domain":
			pr.Domain = effectiveAvail
			pr.DomainUnknown = res.unknown
		case "instagram":
			pr.Instagram = effectiveAvail
			pr.InstagramUnknown = res.unknown
		case "x":
			pr.X = effectiveAvail
			pr.XUnknown = res.unknown
		case "tiktok":
			pr.TikTok = effectiveAvail
			pr.TikTokUnknown = res.unknown
		case "threads":
			pr.Threads = effectiveAvail
			pr.ThreadsUnknown = res.unknown
		case "youtube":
			pr.YouTube = effectiveAvail
			pr.YouTubeUnknown = res.unknown
		}
	}

	pct := (score / totalWeight) * 100
	return models.AvailabilityResult{
		Name:   name,
		Probes: pr,
		Score:  pct,
		Passes: pct >= (passThreshold/totalWeight)*100,
	}
}

// CheckNames probes a list of names concurrently and returns only those passing the 80% gate.
// If zero names pass, returns the closest matches (top 3 by score) as a fallback.
func CheckNames(names []string) (passing []models.AvailabilityResult, partials []models.AvailabilityResult) {
	all := make([]models.AvailabilityResult, len(names))
	var wg sync.WaitGroup
	for i, n := range names {
		i, n := i, n
		wg.Add(1)
		go func() {
			defer wg.Done()
			all[i] = CheckName(n)
		}()
	}
	wg.Wait()

	for _, r := range all {
		if r.Passes {
			passing = append(passing, r)
		} else {
			partials = append(partials, r)
		}
	}

	// Sort partials by score descending (bubble-sort is fine for ≤12 items)
	for i := 0; i < len(partials); i++ {
		for j := i + 1; j < len(partials); j++ {
			if partials[j].Score > partials[i].Score {
				partials[i], partials[j] = partials[j], partials[i]
			}
		}
	}
	// Cap partials fallback at 3
	if len(partials) > 3 {
		partials = partials[:3]
	}
	return
}

// ── Platform probers ──────────────────────────────────────────────────────────
//
// Strategy: HEAD/GET the public profile URL. 
//   200 → profile exists → taken
//   404 → no profile → available
//   429/5xx → rate-limited / server error → unknown (treated conservatively as taken)
//   3xx → follow up to 3 redirects; if final is 200 → taken

func probeURL(rawURL string, client *http.Client) (available bool, unknown bool) {
	resp, err := client.Get(rawURL)
	if err != nil {
		// Connection refused or DNS failure — treat as unknown
		return false, true
	}
	defer resp.Body.Close()
	// Drain a small amount to allow connection reuse
	io.CopyN(io.Discard, resp.Body, 512)

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return true, false // Available
	case resp.StatusCode == http.StatusOK:
		return false, false // Taken
	case resp.StatusCode == 429 || resp.StatusCode >= 500:
		return false, true // Unknown
	default:
		// 301/302/403 etc. — conservative: treat as taken
		return false, false
	}
}

func sanitiseHandle(name string) string {
	// Lowercase, remove spaces and special chars, keep alphanumeric + underscore
	h := strings.ToLower(name)
	h = strings.ReplaceAll(h, " ", "")
	var b strings.Builder
	for _, r := range h {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func probeDomain(handle string, client *http.Client) (bool, bool) {
	// Use RDAP (Registration Data Access Protocol) — free, no API key, no scraping.
	// RDAP lookup for .com domain: https://rdap.verisign.com/com/v1/domain/{name}.com
	// 200 = registered/taken; 404 = available
	rdapURL := fmt.Sprintf("https://rdap.verisign.com/com/v1/domain/%s.com", url.PathEscape(handle))
	resp, err := client.Get(rdapURL)
	if err != nil {
		return false, true // unknown
	}
	defer resp.Body.Close()
	io.CopyN(io.Discard, resp.Body, 512)
	switch resp.StatusCode {
	case 404:
		return true, false  // available
	case 200:
		return false, false // taken
	default:
		return false, true  // unknown
	}
}

func probeInstagram(handle string, client *http.Client) (bool, bool) {
	return probeURL("https://www.instagram.com/"+handle+"/", client)
}

func probeX(handle string, client *http.Client) (bool, bool) {
	return probeURL("https://x.com/"+handle, client)
}

func probeTikTok(handle string, client *http.Client) (bool, bool) {
	return probeURL("https://www.tiktok.com/@"+handle, client)
}

func probeThreads(handle string, client *http.Client) (bool, bool) {
	return probeURL("https://www.threads.net/@"+handle, client)
}

func probeYouTube(handle string, client *http.Client) (bool, bool) {
	return probeURL("https://www.youtube.com/@"+handle, client)
}
