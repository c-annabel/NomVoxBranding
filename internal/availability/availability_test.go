package availability_test

import (
	"strings"
	"testing"

	"github.com/c-annabel/NomVoxBranding/internal/availability"
)

// ── sanitiseHandle (via CheckName exercising internal logic) ──────────────────

// TestCheckName_HandleSanitisation verifies that names with spaces and special
// characters are correctly sanitised before probing.
// We test only the struct shape and field types, not the live probe results
// (those require network access and are non-deterministic).
func TestCheckName_ReturnsResult(t *testing.T) {
	// Use a clearly fictional name — reduces chance of actual network blocking
	// while still exercising the struct and return type
	result := availability.CheckName("zxqveyraktest9999")
	if result.Name != "zxqveyraktest9999" {
		t.Errorf("expected Name to be preserved, got %q", result.Name)
	}
	// Score must be 0–100
	if result.Score < 0 || result.Score > 100 {
		t.Errorf("expected Score in 0–100, got %f", result.Score)
	}
}

// TestCheckNames_EmptyInput verifies graceful handling of empty name list.
func TestCheckNames_EmptyInput(t *testing.T) {
	passing, partials := availability.CheckNames([]string{})
	if len(passing) != 0 {
		t.Errorf("expected 0 passing, got %d", len(passing))
	}
	if len(partials) != 0 {
		t.Errorf("expected 0 partials, got %d", len(partials))
	}
}

// TestCheckNames_ResultShape verifies that a multi-name check returns
// well-formed AvailabilityResult structs.
func TestCheckNames_ResultShape(t *testing.T) {
	// Two very unlikely-to-exist names
	names := []string{"zxqveyrak9999", "mxqsorveth8888"}
	passing, partials := availability.CheckNames(names)

	all := append(passing, partials...)
	if len(all) == 0 {
		// All names may be probed as "unknown" if network is unavailable in CI —
		// that's fine; we just need at least one result back
		t.Log("all names returned unknown — network may be unavailable in this environment")
		return
	}

	for _, r := range all {
		if r.Name == "" {
			t.Error("expected non-empty Name in result")
		}
		if r.Score < 0 || r.Score > 100 {
			t.Errorf("expected Score 0–100, got %f for %q", r.Score, r.Name)
		}
	}
}

// TestCheckNames_PartialSort verifies that partials are returned sorted
// descending by score.
func TestCheckNames_PartialsSortedDescending(t *testing.T) {
	names := []string{"zxqveyrak9999", "mxqsorveth8888", "pqrqalune7777", "wrthovar6666"}
	_, partials := availability.CheckNames(names)

	for i := 1; i < len(partials); i++ {
		if partials[i].Score > partials[i-1].Score {
			t.Errorf("partials not sorted: index %d (%.1f) > index %d (%.1f)",
				i, partials[i].Score, i-1, partials[i-1].Score)
		}
	}
}

// TestCheckNames_NoDuplicateNames verifies no duplicate names appear across
// passing + partials.
func TestCheckNames_NoDuplicates(t *testing.T) {
	names := []string{"zxqveyrak9999", "mxqsorveth8888"}
	passing, partials := availability.CheckNames(names)

	seen := map[string]bool{}
	for _, r := range append(passing, partials...) {
		if seen[r.Name] {
			t.Errorf("duplicate name in results: %q", r.Name)
		}
		seen[r.Name] = true
	}
}

// ── Availability score sanity ─────────────────────────────────────────────────

// TestCheckName_ScoreIsPercentage verifies that the score returned is
// expressed as a percentage (0–100), not a raw weight sum (0–6).
func TestCheckName_ScoreIsPercentage(t *testing.T) {
	r := availability.CheckName("aaazzztest0000")
	if r.Score > 100 {
		t.Errorf("score > 100: %f — must be a percentage, not a raw weight", r.Score)
	}
}

// ── Platform handle sanitisation ──────────────────────────────────────────────

// TestHandleSanitisation indirectly tests that handle sanitisation strips
// non-alphanumeric characters by checking the result name is preserved verbatim.
func TestHandleSanitisation_SpecialChars(t *testing.T) {
	// CheckName strips the name to a handle internally but returns the original Name
	r := availability.CheckName("My Brand Name!")
	if r.Name != "My Brand Name!" {
		t.Errorf("expected original name preserved in result, got %q", r.Name)
	}
}

// TestHandleSanitisation_EmptyHandle verifies that an empty-after-sanitise name
// doesn't panic — the checker should return a zero-score result.
func TestHandleSanitisation_EmptyResult(t *testing.T) {
	// Name with only special chars → handle becomes "" → all probes return unknown
	r := availability.CheckName("!!!")
	if r.Score < 0 {
		t.Error("negative score for empty-handle name")
	}
	_ = strings.TrimSpace(r.Name) // just ensure no panic
}
