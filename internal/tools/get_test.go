package tools

import (
	"testing"
)

func TestGet(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()

	tests := []struct {
		role         string
		id           string
		wantFound    bool
		wantWithheld bool
	}{
		// maintainer sees everything.
		{"maintainer", "file:internal/ingress/ingress.go", true, false},
		{"maintainer", "file:SECURITY.md", true, false},
		{"maintainer", "doc:growth/README.md", true, false},

		// employee sees path:*, docs:*, adr:*, backlog:* but not private:*.
		{"employee", "file:internal/ingress/ingress.go", true, false},
		{"employee", "doc:docs/architecture.md", true, false},
		{"employee", "adr:docs/adr/0001.md", true, false},
		{"employee", "issue:1", true, false},
		{"employee", "file:SECURITY.md", false, true},
		{"employee", "doc:growth/README.md", false, true},

		// contractor sees specific path scopes, docs:public, adr:public, backlog:public.
		{"contractor", "file:internal/ingress/ingress.go", true, false},
		{"contractor", "doc:docs/architecture.md", true, false},
		{"contractor", "adr:docs/adr/0001.md", true, false},
		{"contractor", "issue:1", true, false},
		{"contractor", "file:SECURITY.md", false, true},
		{"contractor", "doc:growth/README.md", false, true},

		// Non-existent item returns neither found nor withheld.
		{"maintainer", "file:does-not-exist", false, false},
		{"contractor", "file:does-not-exist", false, false},
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
			if result.Withheld != tt.wantWithheld {
				t.Errorf("Withheld = %v, want %v", result.Withheld, tt.wantWithheld)
			}
			if result.WithheldCounts == nil {
				t.Error("WithheldCounts must not be nil")
			}
			if tt.wantWithheld && len(result.WithheldCounts) == 0 {
				t.Error("expected non-empty WithheldCounts for denied item")
			}
		})
	}
}

// TestGetNoScopeItem verifies that a no-scope item is always denied.
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
			t.Errorf("role %q: no-scope item should be withheld, got %v", roleName, result.Item)
		}
		if !result.Withheld {
			t.Errorf("role %q: Withheld should be true for no-scope item", roleName)
		}
	}
}
