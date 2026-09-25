package tools

import (
	"testing"
)

func TestWithheld(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()

	tests := []struct {
		role           string
		wantCounts     bool
		hiddenScopes   []string // at least one of these must appear in the summary
	}{
		// maintainer sees everything mappable; only the no-scope item is denied.
		{
			role:        "maintainer",
			wantCounts:  true,
			hiddenScopes: []string{"(no-scope)"},
		},
		// employee cannot see private:growth, private:security and the no-scope item.
		{
			role:        "employee",
			wantCounts:  true,
			hiddenScopes: []string{"private:growth", "private:security", "(no-scope)"},
		},
		// contractor cannot see private:growth, private:security and the no-scope item.
		{
			role:        "contractor",
			wantCounts:  true,
			hiddenScopes: []string{"private:growth", "private:security", "(no-scope)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			role, ok := cfg.Roles[tt.role]
			if !ok {
				t.Fatalf("role %q not found", tt.role)
			}
			result := Withheld(role, items)

			if result.Summary == nil {
				t.Error("Summary must not be nil")
			}
			if result.Withheld == nil {
				t.Error("Withheld must not be nil")
			}
			if len(result.Withheld) != 0 {
				t.Errorf("withheld.Withheld should be empty (it reports on others), got %v", result.Withheld)
			}

			if tt.wantCounts && len(result.Summary) == 0 {
				t.Error("expected non-empty Summary")
			}

			for _, scope := range tt.hiddenScopes {
				if _, ok := result.Summary[scope]; !ok {
					t.Errorf("expected scope %q in Summary, got %v", scope, result.Summary)
				}
			}
		})
	}
}
