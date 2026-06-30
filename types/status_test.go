package types

import "testing"

// ── Testing constants / iota values ─────────────────────────────
//
// iota values are part of your public API (they're stored in SQLite!).
// If someone reorders the constants, the DB breaks. This test catches that.

func TestStatusValues(t *testing.T) {
	// These values are stored in the database, so they must never change.
	if Todo != 0 {
		t.Errorf("Todo = %d, want 0", Todo)
	}
	if InProgress != 1 {
		t.Errorf("InProgress = %d, want 1", InProgress)
	}
	if Done != 2 {
		t.Errorf("Done = %d, want 2", Done)
	}
}
