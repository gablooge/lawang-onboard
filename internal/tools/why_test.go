package tools

import (
	"testing"

	"github.com/gablooge/lawang-onboard/internal/corpus"
)

// whyItems extends testItems with items for why tests.
func whyItems() []corpus.Item {
	base := testItems()
	return append(base,
		// commit touching internal/ingress (visible to contractor)
		corpus.Item{
			ID:     "commit:why001",
			Kind:   "commit",
			Title:  "Fix ingress timeout handling",
			Paths:  []string{"internal/ingress/ingress.go", "internal/ingress/ingress_test.go"},
			Text:   "fix timeout logic",
			Scopes: []string{"path:internal/ingress"},
		},
		// review comment on internal/ingress (visible to contractor)
		corpus.Item{
			ID:     "review_comment:50:222",
			Kind:   "review",
			Title:  "Review comment on PR #50, internal/ingress/ingress.go",
			Paths:  []string{"internal/ingress/ingress.go"},
			Text:   "nit: rename this variable",
			Scopes: []string{"path:internal/ingress"},
		},
		// ADR mentioning ingress (visible to contractor)
		corpus.Item{
			ID:     "adr:docs/adr/0003.md",
			Kind:   "adr",
			Title:  "0003: ingress timeout strategy",
			Paths:  []string{"docs/adr/0003.md"},
			Text:   "We chose internal/ingress as the timeout boundary.",
			Scopes: []string{"adr:public"},
		},
		// commit only touching growth (not visible to contractor)
		corpus.Item{
			ID:     "commit:grow002",
			Kind:   "commit",
			Title:  "Growth analytics ingress hook",
			Paths:  []string{"growth/hook.go"},
			Text:   "ingress hook for growth",
			Scopes: []string{"private:growth"},
		},
		// review comment on growth path (not visible to contractor)
		corpus.Item{
			ID:     "review_comment:51:333",
			Kind:   "review",
			Title:  "Review comment on PR #51, growth/hook.go",
			Paths:  []string{"growth/hook.go"},
			Text:   "review of growth ingress hook",
			Scopes: []string{"private:growth"},
		},
		// security doc mentioning ingress (not visible to contractor)
		corpus.Item{
			ID:     "file:SECURITY.md",
			Kind:   "doc",
			Title:  "SECURITY.md",
			Paths:  []string{"SECURITY.md"},
			Text:   "The internal/ingress layer enforces rate limits.",
			Scopes: []string{"private:security"},
		},
	)
}

func TestWhyContractorIngressReturnsReviews(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["contractor"]
	result := Why(role, whyItems(), "internal/ingress")

	// Must return review comments on internal/ingress.
	if len(result.Reviews) == 0 {
		t.Fatal("contractor: why internal/ingress must return review comments")
	}
	found := false
	for _, r := range result.Reviews {
		if r.ID == "review_comment:50:222" {
			found = true
		}
	}
	if !found {
		t.Error("contractor: review_comment:50:222 not found in why results")
	}

	// Must not return growth or security items.
	for _, c := range result.Commits {
		if c.ID == "commit:grow002" {
			t.Error("contractor: growth commit leaked via why")
		}
	}
	for _, r := range result.Reviews {
		if r.ID == "review_comment:51:333" {
			t.Error("contractor: growth review leaked via why")
		}
	}
	for _, a := range result.ADRsDocs {
		if a.ID == "file:SECURITY.md" {
			t.Error("contractor: security doc leaked via why")
		}
	}
}

func TestWhyContractorIngressReturnsCommits(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["contractor"]
	result := Why(role, whyItems(), "internal/ingress")

	if len(result.Commits) == 0 {
		t.Fatal("contractor: why internal/ingress must return commits")
	}
	found := false
	for _, c := range result.Commits {
		if c.ID == "commit:why001" {
			found = true
		}
	}
	if !found {
		t.Error("contractor: commit:why001 not found in why results")
	}
}

func TestWhyContractorIngressReturnsADRs(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["contractor"]
	result := Why(role, whyItems(), "internal/ingress")

	if len(result.ADRsDocs) == 0 {
		t.Fatal("contractor: why internal/ingress must return ADRs/docs")
	}
	found := false
	for _, a := range result.ADRsDocs {
		if a.ID == "adr:docs/adr/0003.md" {
			found = true
		}
	}
	if !found {
		t.Error("contractor: adr:docs/adr/0003.md not found in why results")
	}
}

func TestWhyMaintainerSeesAll(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["maintainer"]
	result := Why(role, whyItems(), "internal/ingress")

	// Maintainer should see commits and reviews for ingress.
	if len(result.Commits) == 0 {
		t.Error("maintainer: expected commits for internal/ingress")
	}
	if len(result.Reviews) == 0 {
		t.Error("maintainer: expected reviews for internal/ingress")
	}
}

func TestWhyEmployeeDoesNotLeakGrowth(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["employee"]
	result := Why(role, whyItems(), "growth")

	// Employee has path:* but not private:growth, so growth items are withheld.
	for _, c := range result.Commits {
		if c.ID == "commit:grow002" {
			t.Error("employee: private growth commit leaked via why")
		}
	}
	for _, r := range result.Reviews {
		if r.ID == "review_comment:51:333" {
			t.Error("employee: private growth review leaked via why")
		}
	}
}

func TestWhyEmptyPath(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["maintainer"]
	result := Why(role, whyItems(), "")
	if len(result.Commits)+len(result.Reviews)+len(result.ADRsDocs) != 0 {
		t.Error("empty path should return no results")
	}
}
