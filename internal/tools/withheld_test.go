package tools

import (
	"testing"

	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/filter"
)

func TestWithheld(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()

	tests := []struct {
		role         string
		wantCounts   bool
		hiddenScopes []string // at least one of these must appear in by_scope
	}{
		// maintainer sees everything mappable; only the no-scope item is denied.
		{
			role:         "maintainer",
			wantCounts:   true,
			hiddenScopes: []string{"(no-scope)"},
		},
		// employee cannot see private:growth, private:security and the no-scope item.
		{
			role:         "employee",
			wantCounts:   true,
			hiddenScopes: []string{"private:growth", "private:security", "(no-scope)"},
		},
		// contractor cannot see private:growth, private:security and the no-scope item.
		{
			role:         "contractor",
			wantCounts:   true,
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

			if result.ByScope == nil {
				t.Error("ByScope must not be nil")
			}
			if result.Withheld == nil {
				t.Error("Withheld must not be nil")
			}
			if len(result.Withheld) != 0 {
				t.Errorf("withheld.Withheld should be empty (it reports on others), got %v", result.Withheld)
			}

			if tt.wantCounts && len(result.ByScope) == 0 {
				t.Error("expected non-empty ByScope")
			}

			for _, scope := range tt.hiddenScopes {
				if _, ok := result.ByScope[scope]; !ok {
					t.Errorf("expected scope %q in ByScope, got %v", scope, result.ByScope)
				}
			}
		})
	}
}

// TestWithheldNoHeldScopeLeaks verifies that for every role no scope in the
// withheld by_scope is one the role holds, and that the contractor's Total
// equals 610 minus the number of items the contractor can see in the real corpus.
func TestWithheldNoHeldScopeLeaks(t *testing.T) {
	cfg := testCfg(t)

	// Load the real corpus. The test file lives at internal/tools/ so the
	// corpus directory is two levels up.
	allItems, err := corpus.Load("../../corpus", cfg)
	if err != nil {
		t.Fatalf("corpus.Load: %v", err)
	}
	if len(allItems) == 0 {
		t.Fatal("no corpus items loaded")
	}

	const totalCorpus = 610

	for roleName, role := range cfg.Roles {
		t.Run(roleName, func(t *testing.T) {
			result := Withheld(role, allItems)

			// No scope in by_scope may be one the role holds.
			for scope := range result.ByScope {
				if scope == "(no-scope)" || scope == "(unmapped)" {
					continue
				}
				if withheldTestRoleCoversScope(role.Scopes, scope) {
					t.Errorf("role %q holds scope %q but it appears in withheld by_scope", roleName, scope)
				}
			}

			// Contractor-specific: Total must equal 610 minus visible count.
			if roleName == "contractor" {
				visibleCount := 0
				for _, item := range allItems {
					if filter.Visible(role, item) {
						visibleCount++
					}
				}
				wantHidden := totalCorpus - visibleCount
				if result.Total != wantHidden {
					t.Errorf("contractor: Total=%d want %d (610 - %d visible)",
						result.Total, wantHidden, visibleCount)
				}
			}
		})
	}
}

// withheldTestRoleCoversScope reports whether roleScopes covers the given scope
// (exact match or wildcard "prefix:*"). Used only in the withheld test.
func withheldTestRoleCoversScope(roleScopes []string, need string) bool {
	for _, have := range roleScopes {
		if have == "*" || have == need {
			return true
		}
		if len(have) >= 2 && have[len(have)-1] == '*' && have[len(have)-2] == ':' {
			prefix := have[:len(have)-1] // "prefix:" without the *
			if len(need) >= len(prefix) && need[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

