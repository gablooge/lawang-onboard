package tools

import (
	"testing"
)

func TestSearch(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()

	tests := []struct {
		role        string
		query       string
		wantIDs     []string
		forbiddenID string
	}{
		// maintainer sees all matching items.
		{
			role:    "maintainer",
			query:   "ingress",
			wantIDs: []string{"file:internal/ingress/ingress.go"},
		},
		{
			role:    "maintainer",
			query:   "growth",
			wantIDs: []string{"doc:growth/README.md"},
		},
		// employee cannot see private:growth or private:security.
		{
			role:        "employee",
			query:       "growth",
			forbiddenID: "doc:growth/README.md",
		},
		{
			role:        "employee",
			query:       "security",
			forbiddenID: "file:SECURITY.md",
		},
		// contractor cannot see private items.
		{
			role:        "contractor",
			query:       "growth",
			forbiddenID: "doc:growth/README.md",
		},
		// contractor can see ingress.
		{
			role:    "contractor",
			query:   "ingress",
			wantIDs: []string{"file:internal/ingress/ingress.go"},
		},
		// empty query returns no results.
		{
			role:    "maintainer",
			query:   "",
			wantIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.role+"/"+tt.query, func(t *testing.T) {
			role, ok := cfg.Roles[tt.role]
			if !ok {
				t.Fatalf("role %q not found", tt.role)
			}
			result := Search(role, items, tt.query, 20)

			if result.Withheld == nil {
				t.Error("Withheld must not be nil")
			}

			// Check that forbidden IDs are absent.
			if tt.forbiddenID != "" {
				for _, hit := range result.Items {
					if hit.ID == tt.forbiddenID {
						t.Errorf("LEAK: role %q should not see %q in search results",
							tt.role, tt.forbiddenID)
					}
				}
			}

			// Check that expected IDs are present.
			for _, wantID := range tt.wantIDs {
				found := false
				for _, hit := range result.Items {
					if hit.ID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %q in results but not found; items: %v",
						wantID, result.Items)
				}
			}
		})
	}
}

// TestSearchHiddenItemsNeverAffectScores verifies that a word that appears only
// in hidden items returns zero results for a role that cannot see those items.
func TestSearchHiddenItemsNeverAffectScores(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()

	// "growth" only appears in the growth item (private:growth).
	contractor := cfg.Roles["contractor"]
	result := Search(contractor, items, "growth content", 20)
	if len(result.Items) != 0 {
		t.Errorf("contractor: search for growth content returned %d items, want 0: %v",
			len(result.Items), result.Items)
	}
}
