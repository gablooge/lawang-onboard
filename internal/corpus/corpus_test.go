package corpus

import (
	"testing"

	"github.com/gablooge/lawang-onboard/internal/roles"
)

// testConfig returns a Config for tests, equivalent to the real roles.yaml.
func testConfig(t *testing.T) *roles.Config {
	t.Helper()
	const yaml = `
scopes:
  - glob: "growth/**"
    scope: "private:growth"
  - glob: "SECURITY.md"
    scope: "private:security"
  - glob: "docs/adr/**"
    scope: "adr:public"
  - glob: "docs/**"
    scope: "docs:public"
  - glob: "README.md"
    scope: "docs:public"
  - glob: "internal/ingress/**"
    scope: "path:internal/ingress"
  - glob: "internal/provider/**"
    scope: "path:internal/provider"
  - glob: "internal/**"
    scope: "path:internal/core"
  - glob: "cmd/**"
    scope: "path:cmd"
  - glob: "migrations/**"
    scope: "path:migrations"
  - glob: "**"
    scope: "path:repo"

label_scopes:
  growth: "private:growth"

kind_scopes:
  issue: "backlog:public"
  issue_comment: "backlog:public"

label_areas:
  provider: "path:internal/provider"

roles:
  maintainer:
    token_env: ONBOARD_TOKEN_MAINTAINER
    scopes: ["*"]
  employee:
    token_env: ONBOARD_TOKEN_EMPLOYEE
    scopes: ["path:*", "docs:*", "adr:*", "backlog:*"]
  contractor:
    token_env: ONBOARD_TOKEN_CONTRACTOR
    scopes: ["path:internal/ingress", "path:internal/provider", "docs:public", "adr:public", "backlog:public"]
`
	// Use the exported parse helper via roles package.
	cfg, err := roles.ParseBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("testConfig: %v", err)
	}
	return cfg
}

// TestLoadRealCorpus loads the actual corpus directory and checks that no item has nil Scopes.
func TestLoadRealCorpus(t *testing.T) {
	cfg := testConfig(t)
	items, err := Load("../../corpus", cfg)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("no items loaded")
	}
	for _, item := range items {
		if item.Scopes == nil {
			t.Errorf("item %q has nil Scopes", item.ID)
		}
	}
}

// TestScopeAssignmentByKind tests scope assignment for one item of each kind.
func TestScopeAssignmentByKind(t *testing.T) {
	cfg := testConfig(t)

	// We'll build a small in-memory corpus with one item per kind.
	issueLabels := map[int][]string{
		10: {"backlog"},            // issue 10: no label_scopes match -> kind scope
		11: {"backlog", "provider"}, // issue 11: no label_scopes match -> kind scope
		99: {"growth"},             // issue 99: matches label_scopes
	}

	tests := []struct {
		name       string
		item       Item
		wantScopes []string
	}{
		{
			name: "file: single path -> path scope",
			item: Item{Kind: "file", Paths: []string{"internal/config/config.go"}},
			wantScopes: []string{"path:internal/core"},
		},
		{
			name: "doc: README.md -> docs:public",
			item: Item{Kind: "doc", Paths: []string{"README.md"}},
			wantScopes: []string{"docs:public"},
		},
		{
			name: "adr: docs/adr/ path -> adr:public",
			item: Item{Kind: "adr", Paths: []string{"docs/adr/0001.md"}},
			wantScopes: []string{"adr:public"},
		},
		{
			name: "commit: multiple paths -> union of scopes",
			item: Item{Kind: "commit", Paths: []string{
				"internal/config/config.go",
				"docs/architecture.md",
			}},
			// both paths produce different scopes
			wantScopes: []string{"path:internal/core", "docs:public"},
		},
		{
			name: "pr: paths spanning scopes -> union",
			item: Item{Kind: "pr", Paths: []string{
				"internal/ingress/ingress.go",
				"internal/provider/provider.go",
			}},
			wantScopes: []string{"path:internal/ingress", "path:internal/provider"},
		},
		{
			name: "issue: no label match -> kind scope backlog:public",
			item: Item{Kind: "issue", Labels: []string{"backlog"}},
			wantScopes: []string{"backlog:public"},
		},
		{
			name: "issue: growth label -> private:growth",
			item: Item{Kind: "issue", Labels: []string{"growth"}},
			wantScopes: []string{"private:growth"},
		},
		{
			name: "issue_comment: inherits parent issue labels (backlog -> backlog:public)",
			item: Item{Kind: "issue_comment", Issue: 10},
			wantScopes: []string{"backlog:public"},
		},
		{
			name: "issue_comment: inherits parent issue labels (growth -> private:growth)",
			item: Item{Kind: "issue_comment", Issue: 99},
			wantScopes: []string{"private:growth"},
		},
		{
			name: "review: scoped by its own path only",
			item: Item{Kind: "review", Paths: []string{"internal/config/config.go"}},
			wantScopes: []string{"path:internal/core"},
		},
		{
			name: "review: internal/provider path",
			item: Item{Kind: "review", Paths: []string{"internal/provider/provider.go"}},
			wantScopes: []string{"path:internal/provider"},
		},
		{
			name: "doc: SECURITY.md -> private:security",
			item: Item{Kind: "doc", Paths: []string{"SECURITY.md"}},
			wantScopes: []string{"private:security"},
		},
		{
			name: "doc: growth/README.md -> private:growth",
			item: Item{Kind: "doc", Paths: []string{"growth/README.md"}},
			wantScopes: []string{"private:growth"},
		},
		{
			name: "file: unmapped path -> (unmapped)",
			item: Item{Kind: "file", Paths: []string{"something/unknown/path.go"}},
			wantScopes: []string{"path:repo"}, // ** catch-all
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := assignScopes(tt.item, issueLabels, cfg)
			if !scopesEqual(got, tt.wantScopes) {
				t.Errorf("assignScopes = %v, want %v", got, tt.wantScopes)
			}
		})
	}
}

// TestIssueCommentInheritsParentLabels verifies that an issue_comment's scope is
// derived from its parent issue's labels, not from the comment's own (empty) labels.
func TestIssueCommentInheritsParentLabels(t *testing.T) {
	cfg := testConfig(t)

	// Issue 5 has the "growth" label.
	issueLabels := map[int][]string{
		5: {"growth"},
	}

	comment := Item{Kind: "issue_comment", Issue: 5, Labels: nil}
	got := assignScopes(comment, issueLabels, cfg)
	if len(got) != 1 || got[0] != "private:growth" {
		t.Errorf("issue_comment inherited wrong scope: %v, want [private:growth]", got)
	}
}

// TestReviewCommentScopedByOwnPath verifies that a review comment uses only its own
// single path for scope, not the paths of its pull request.
func TestReviewCommentScopedByOwnPath(t *testing.T) {
	cfg := testConfig(t)

	issueLabels := map[int][]string{}

	// This review comment is on internal/provider/provider.go within a PR that also
	// touches many other paths. Scope must come from the comment's own path only.
	comment := Item{
		Kind:  "review",
		Paths: []string{"internal/provider/provider.go"},
	}
	got := assignScopes(comment, issueLabels, cfg)
	if len(got) != 1 || got[0] != "path:internal/provider" {
		t.Errorf("review scope = %v, want [path:internal/provider]", got)
	}
}

// TestUnmappedPathDenied verifies that a path with no matching glob rule
// results in "(unmapped)" scope, which is denied by filter.Visible.
func TestUnmappedPathDenied(t *testing.T) {
	// A config with no catch-all.
	cfg, err := roles.ParseBytes([]byte(`
scopes:
  - glob: "docs/**"
    scope: "docs:public"
label_scopes: {}
kind_scopes: {}
label_areas: {}
roles: {}
`))
	if err != nil {
		t.Fatalf("ParseBytes: %v", err)
	}

	item := Item{Kind: "file", Paths: []string{"internal/secret/secret.go"}}
	got := assignScopes(item, nil, cfg)
	if len(got) != 1 || got[0] != "(unmapped)" {
		t.Errorf("unmapped path scope = %v, want [(unmapped)]", got)
	}
}

// scopesEqual returns true if a and b contain the same elements (order-independent).
func scopesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for _, s := range a {
		m[s]++
	}
	for _, s := range b {
		m[s]--
		if m[s] < 0 {
			return false
		}
	}
	return true
}
