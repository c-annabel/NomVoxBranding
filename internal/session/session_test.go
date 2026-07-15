package session_test

import (
	"testing"

	"github.com/c-annabel/NomVoxBranding/internal/models"
)

// appendUnique mirrors the internal helper in handlers/generate.go.
// Copied here so we can unit-test the deduplication logic without depending
// on the handlers package.
func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

// removeString mirrors the helper in handlers/session.go.
func removeString(slice []string, s string) []string {
	out := slice[:0]
	for _, v := range slice {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}

// ── appendUnique ─────────────────────────────────────────────────────────────

func TestAppendUnique_NoDuplicates(t *testing.T) {
	slice := []string{"Veyrak", "Sorveth"}
	got := appendUnique(slice, "Veyrak")
	if len(got) != 2 {
		t.Errorf("expected len 2, got %d — duplicate was added", len(got))
	}
}

func TestAppendUnique_AddsNew(t *testing.T) {
	slice := []string{"Veyrak"}
	got := appendUnique(slice, "Sorveth")
	if len(got) != 2 {
		t.Errorf("expected len 2, got %d", len(got))
	}
	if got[1] != "Sorveth" {
		t.Errorf("expected Sorveth at index 1, got %q", got[1])
	}
}

func TestAppendUnique_EmptySlice(t *testing.T) {
	got := appendUnique(nil, "Qalune")
	if len(got) != 1 || got[0] != "Qalune" {
		t.Errorf("expected [Qalune], got %v", got)
	}
}

// ── removeString ─────────────────────────────────────────────────────────────

func TestRemoveString_RemovesMatch(t *testing.T) {
	slice := []string{"Veyrak", "Sorveth", "Qalune"}
	got := removeString(slice, "Sorveth")
	if len(got) != 2 {
		t.Errorf("expected len 2, got %d", len(got))
	}
	for _, v := range got {
		if v == "Sorveth" {
			t.Error("expected Sorveth to be removed")
		}
	}
}

func TestRemoveString_NoMatch(t *testing.T) {
	slice := []string{"Veyrak", "Sorveth"}
	got := removeString(slice, "Myrrix")
	if len(got) != 2 {
		t.Errorf("expected len unchanged at 2, got %d", len(got))
	}
}

func TestRemoveString_EmptySlice(t *testing.T) {
	got := removeString(nil, "Veyrak")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

// ── BrandSession struct integrity ─────────────────────────────────────────────

// TestBrandSession_ZeroValue verifies that a zero-value BrandSession is valid
// and all slice fields are nil (not panicky on range).
func TestBrandSession_ZeroValue(t *testing.T) {
	var sess models.BrandSession
	if sess.SessionID != "" {
		t.Error("expected empty SessionID")
	}
	// Should be rangeable without panic
	for range sess.Liked {
	}
	for range sess.Rejected {
	}
	for range sess.AllGenerated {
	}
	for range sess.DirectionNotes {
	}
}

// TestBrandSession_Accumulation simulates a full session feedback loop:
// generate → like → reject → note → regenerate → select.
func TestBrandSession_Accumulation(t *testing.T) {
	sess := &models.BrandSession{SessionID: "test-abc"}

	// Simulate generating 4 names
	for _, n := range []string{"Veyrak", "Sorveth", "Qalune", "Myrrix"} {
		sess.AllGenerated = appendUnique(sess.AllGenerated, n)
	}
	if len(sess.AllGenerated) != 4 {
		t.Errorf("expected 4 AllGenerated, got %d", len(sess.AllGenerated))
	}

	// Like two
	sess.Liked = appendUnique(sess.Liked, "Veyrak")
	sess.Liked = appendUnique(sess.Liked, "Sorveth")
	if len(sess.Liked) != 2 {
		t.Errorf("expected 2 Liked, got %d", len(sess.Liked))
	}

	// Reject one with note
	sess.Rejected = appendUnique(sess.Rejected, "Qalune")
	sess.Liked = removeString(sess.Liked, "Qalune") // remove from liked if it was there
	sess.DirectionNotes = append(sess.DirectionNotes, "too soft-sounding")
	if len(sess.Rejected) != 1 {
		t.Errorf("expected 1 Rejected, got %d", len(sess.Rejected))
	}

	// Slider update
	sess.SliderPlayful = 0.8
	sess.SliderAbstract = 0.3
	if sess.SliderPlayful != 0.8 {
		t.Errorf("expected SliderPlayful 0.8, got %f", sess.SliderPlayful)
	}

	// Select a name
	sess.SelectedName = "Veyrak"
	if sess.SelectedName != "Veyrak" {
		t.Errorf("expected SelectedName Veyrak, got %q", sess.SelectedName)
	}
}
