package tools

import (
	"testing"

	"github.com/gablooge/lawang-onboard/internal/corpus"
)

// traceItems extends testItems with items suited to trace_feature tests.
func traceItems() []corpus.Item {
	base := testItems()
	return append(base,
		// commit touching ingress (visible to contractor)
		corpus.Item{
			ID:     "commit:abc1234",
			Kind:   "commit",
			Title:  "Add ingress rate limiting",
			Paths:  []string{"internal/ingress/ingress.go"},
			Text:   "rate limit ingress requests",
			Scopes: []string{"path:internal/ingress"},
		},
		// review on ingress (visible to contractor)
		corpus.Item{
			ID:     "review_comment:99:111",
			Kind:   "review",
			Title:  "Review comment on PR #99, internal/ingress/ingress.go",
			Paths:  []string{"internal/ingress/ingress.go"},
			Text:   "please add a test for the rate limit path",
			Scopes: []string{"path:internal/ingress"},
		},
		// adr mentioning ingress (visible to contractor via adr:public)
		corpus.Item{
			ID:     "adr:docs/adr/0002.md",
			Kind:   "adr",
			Title:  "0002: ingress rate limiting decision",
			Paths:  []string{"docs/adr/0002.md"},
			Text:   "We decided to rate limit ingress per provider.",
			Scopes: []string{"adr:public"},
		},
		// growth commit only maintainer sees
		corpus.Item{
			ID:     "commit:grow001",
			Kind:   "commit",
			Title:  "Add growth analytics",
			Paths:  []string{"growth/analytics.go"},
			Text:   "growth analytics feature",
			Scopes: []string{"private:growth"},
		},
	)
}

func TestTraceFeatureMaintainer(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["maintainer"]
	items := traceItems()

	result := TraceFeature(role, items, "ingress")

	// Maintainer sees file, commit, review, and adr for ingress.
	if len(result.Files) == 0 {
		t.Error("maintainer: expected at least one file for 'ingress'")
	}
	if len(result.Commits) == 0 {
		t.Error("maintainer: expected at least one commit for 'ingress'")
	}
	if len(result.Reviews) == 0 {
		t.Error("maintainer: expected at least one review for 'ingress'")
	}
	if len(result.ADRs) == 0 {
		t.Error("maintainer: expected at least one ADR for 'ingress'")
	}
}

func TestTraceFeatureContractor(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["contractor"]
	items := traceItems()

	result := TraceFeature(role, items, "ingress")

	// Contractor sees ingress items.
	if len(result.Files) == 0 {
		t.Error("contractor: expected file results for 'ingress'")
	}
	if len(result.Commits) == 0 {
		t.Error("contractor: expected commit results for 'ingress'")
	}
	if len(result.Reviews) == 0 {
		t.Error("contractor: expected review results for 'ingress'")
	}
	if len(result.ADRs) == 0 {
		t.Error("contractor: expected ADR results for 'ingress'")
	}

	// Contractor must not see growth items.
	for _, f := range result.Files {
		if f.ID == "doc:growth/README.md" {
			t.Error("contractor: growth file leaked into trace_feature results")
		}
	}
	for _, c := range result.Commits {
		if c.ID == "commit:grow001" {
			t.Error("contractor: growth commit leaked into trace_feature results")
		}
	}
}

func TestTraceFeatureEmployee(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["employee"]
	items := traceItems()

	// Employee has path:* so sees all path items, but not private:growth.
	result := TraceFeature(role, items, "growth")
	for _, c := range result.Commits {
		if c.ID == "commit:grow001" {
			t.Error("employee: private growth commit leaked")
		}
	}
	// Employee should not see the private growth doc.
	for _, f := range result.Files {
		if f.ID == "doc:growth/README.md" {
			t.Error("employee: private growth doc leaked")
		}
	}
}

func TestTraceFeatureEmptyTerm(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["maintainer"]
	result := TraceFeature(role, traceItems(), "")
	if len(result.Files)+len(result.Commits)+len(result.Reviews)+len(result.ADRs) != 0 {
		t.Error("empty term should return no results")
	}
}
