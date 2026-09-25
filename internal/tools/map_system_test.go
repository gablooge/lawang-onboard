package tools

import (
	"testing"
)

func TestMapSystem(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()

	tests := []struct {
		role        string
		depth       int
		wantKinds   []string
		forbiddenID string
		minNodes    int
	}{
		// maintainer sees everything including private items.
		{
			role:      "maintainer",
			depth:     0,
			wantKinds: []string{"file", "doc", "adr"},
			minNodes:  4, // ingress, SECURITY.md, architecture.md, adr, growth
		},
		// employee sees path:*, docs:*, adr:* but not private:growth or private:security.
		{
			role:        "employee",
			depth:       0,
			forbiddenID: "doc:growth/README.md",
			minNodes:    2,
		},
		// contractor sees specific files.
		{
			role:     "contractor",
			depth:    0,
			minNodes: 1,
		},
		// depth=1 groups paths to their first segment.
		{
			role:  "maintainer",
			depth: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			role, ok := cfg.Roles[tt.role]
			if !ok {
				t.Fatalf("role %q not found", tt.role)
			}
			result := MapSystem(role, items, tt.depth)

			if result.Nodes == nil {
				t.Error("Nodes must not be nil")
			}
			if result.Withheld == nil {
				t.Error("Withheld must not be nil")
			}

			if len(result.Nodes) < tt.minNodes {
				t.Errorf("expected at least %d nodes, got %d", tt.minNodes, len(result.Nodes))
			}

			// Check forbidden IDs are absent.
			if tt.forbiddenID != "" {
				for _, n := range result.Nodes {
					if n.ID == tt.forbiddenID {
						t.Errorf("LEAK: role %q should not see %q in map_system", tt.role, tt.forbiddenID)
					}
				}
			}

			// Check only file/doc/adr kinds appear.
			for _, n := range result.Nodes {
				if n.Kind != "file" && n.Kind != "doc" && n.Kind != "adr" {
					t.Errorf("unexpected kind %q in map_system result", n.Kind)
				}
			}
		})
	}
}

// TestMapSystemDepthTrim checks that depth trimming deduplicates correctly.
func TestMapSystemDepthTrim(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()

	// With depth=1, internal/ingress/ingress.go -> "internal" and
	// docs/architecture.md -> "docs", adr/0001.md -> "docs".
	// Duplicates should be removed.
	role := cfg.Roles["employee"]
	result := MapSystem(role, items, 1)

	seen := make(map[string]int)
	for _, n := range result.Nodes {
		key := n.Kind + ":" + n.Path
		seen[key]++
	}
	for key, count := range seen {
		if count > 1 {
			t.Errorf("duplicate node %q at depth 1 (count %d)", key, count)
		}
	}
}

// TestMapSystemIssuesExcluded verifies that issues are never returned.
func TestMapSystemIssuesExcluded(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()
	role := cfg.Roles["maintainer"]
	result := MapSystem(role, items, 0)
	for _, n := range result.Nodes {
		if n.Kind == "issue" || n.Kind == "issue_comment" || n.Kind == "commit" {
			t.Errorf("map_system returned unexpected kind %q", n.Kind)
		}
	}
}
