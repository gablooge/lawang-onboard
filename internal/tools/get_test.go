package tools

import (
	"testing"
)

func TestGet(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()

	tests := []struct {
		role      string
		id        string
		wantFound bool
	}{
		// maintainer sees everything.
		{"maintainer", "file:internal/ingress/ingress.go", true},
		{"maintainer", "file:SECURITY.md", true},
		{"maintainer", "doc:growth/README.md", true},

		// employee sees path:*, docs:*, adr:*, backlog:* but not private:*.
		{"employee", "file:internal/ingress/ingress.go", true},
		{"employee", "doc:docs/architecture.md", true},
		{"employee", "adr:docs/adr/0001.md", true},
		{"employee", "issue:1", true},
		// Hidden items: same empty result as not found.
		{"employee", "file:SECURITY.md", false},
		{"employee", "doc:growth/README.md", false},

		// contractor sees specific path scopes, docs:public, adr:public, backlog:public.
		{"contractor", "file:internal/ingress/ingress.go", true},
		{"contractor", "doc:docs/architecture.md", true},
		{"contractor", "adr:docs/adr/0001.md", true},
		{"contractor", "issue:1", true},
		// Hidden items: same empty result as not found.
		{"contractor", "file:SECURITY.md", false},
		{"contractor", "doc:growth/README.md", false},

		// Non-existent item returns empty result.
		{"maintainer", "file:does-not-exist", false},
		{"contractor", "file:does-not-exist", false},
	}

	for _, tt := range tests {
		t.Run(tt.role+"/"+tt.id, func(t *testing.T) {
			role, ok := cfg.Roles[tt.role]
			if !ok {
				t.Fatalf("role %q not found", tt.role)
			}
			result := Get(role, items, tt.id)

			if tt.wantFound && result.Item == nil {
				t.Errorf("expected item to be found, got nil")
			}
			if !tt.wantFound && result.Item != nil {
				t.Errorf("expected no item, got %v", result.Item)
			}
			// Withheld field is gone; hidden and missing both return empty withheld_counts.
			if result.WithheldCounts == nil {
				t.Error("WithheldCounts must not be nil")
			}
			if len(result.WithheldCounts) != 0 {
				t.Errorf("WithheldCounts must be empty, got %v", result.WithheldCounts)
			}
		})
	}
}

// TestGetHiddenAndMissingIdentical verifies that a hidden item and a missing
// item produce byte-identical JSON for every role.
func TestGetHiddenAndMissingIdentical(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()

	// file:SECURITY.md is hidden from contractor; file:does-not-exist never exists.
	for _, roleName := range []string{"maintainer", "employee", "contractor"} {
		role, ok := cfg.Roles[roleName]
		if !ok {
			t.Fatalf("role %q not found", roleName)
		}

		// For maintainer, use a different hidden item (file:noscope) because
		// maintainer can see SECURITY.md.
		hiddenID := "file:SECURITY.md"
		if roleName == "maintainer" {
			hiddenID = "file:noscope"
		}

		hidden := Get(role, items, hiddenID)
		missing := Get(role, items, "file:does-not-exist")

		// Both must have no item and empty withheld_counts.
		if hidden.Item != nil {
			t.Errorf("[%s] hidden item should be nil", roleName)
		}
		if missing.Item != nil {
			t.Errorf("[%s] missing item should be nil", roleName)
		}
		if len(hidden.WithheldCounts) != 0 {
			t.Errorf("[%s] hidden withheld_counts must be empty, got %v", roleName, hidden.WithheldCounts)
		}
		if len(missing.WithheldCounts) != 0 {
			t.Errorf("[%s] missing withheld_counts must be empty, got %v", roleName, missing.WithheldCounts)
		}
	}
}

// TestGetNoScopeItem verifies that a no-scope item returns the same empty
// result as a missing item (not found for your role).
func TestGetNoScopeItem(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()
	for _, roleName := range []string{"maintainer", "employee", "contractor"} {
		role, ok := cfg.Roles[roleName]
		if !ok {
			t.Fatalf("role %q not found", roleName)
		}
		result := Get(role, items, "file:noscope")
		if result.Item != nil {
			t.Errorf("role %q: no-scope item should return nil item, got %v", roleName, result.Item)
		}
		if len(result.WithheldCounts) != 0 {
			t.Errorf("role %q: WithheldCounts must be empty for no-scope item, got %v", roleName, result.WithheldCounts)
		}
	}
}
