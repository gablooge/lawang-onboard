package tools

import (
	"testing"

	"github.com/gablooge/lawang-onboard/internal/corpus"
)

// setupItems returns a corpus with setup-relevant files plus private items.
func setupItems() []corpus.Item {
	return []corpus.Item{
		// README.md: path:repo scope (visible to all roles via path:* or specific)
		{
			ID:     "file:README.md",
			Kind:   "file",
			Title:  "README.md",
			Paths:  []string{"README.md"},
			Text:   "# Project README\nHow to set up the project.",
			Scopes: []string{"docs:public"},
		},
		// Makefile: path:repo scope
		{
			ID:     "file:Makefile",
			Kind:   "file",
			Title:  "Makefile",
			Paths:  []string{"Makefile"},
			Text:   "check:\n\tgo vet ./...\n\tgo test -race ./...",
			Scopes: []string{"path:repo"},
		},
		// .github/workflows/ci.yml: path:repo scope
		{
			ID:     "file:.github/workflows/ci.yml",
			Kind:   "file",
			Title:  ".github/workflows/ci.yml",
			Paths:  []string{".github/workflows/ci.yml"},
			Text:   "name: ci\non:\n  push:",
			Scopes: []string{"path:repo"},
		},
		// SECURITY.md: private:security (only maintainer)
		{
			ID:     "file:SECURITY.md",
			Kind:   "file",
			Title:  "SECURITY.md",
			Paths:  []string{"SECURITY.md"},
			Text:   "security policy",
			Scopes: []string{"private:security"},
		},
		// growth doc: private:growth (only maintainer)
		{
			ID:     "doc:growth/README.md",
			Kind:   "doc",
			Title:  "growth README",
			Paths:  []string{"growth/README.md"},
			Text:   "growth content",
			Scopes: []string{"private:growth"},
		},
		// ingress source file (not a setup item)
		{
			ID:     "file:internal/ingress/ingress.go",
			Kind:   "file",
			Title:  "ingress.go",
			Paths:  []string{"internal/ingress/ingress.go"},
			Text:   "package ingress",
			Scopes: []string{"path:internal/ingress"},
		},
	}
}

func TestSetupGuideMaintainer(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["maintainer"]
	result := SetupGuide(role, setupItems())

	ids := make(map[string]bool)
	for _, item := range result.Items {
		ids[item.ID] = true
	}

	// Maintainer sees README, Makefile, and CI workflow.
	if !ids["file:README.md"] {
		t.Error("maintainer: README.md missing from setup_guide")
	}
	if !ids["file:Makefile"] {
		t.Error("maintainer: Makefile missing from setup_guide")
	}
	if !ids["file:.github/workflows/ci.yml"] {
		t.Error("maintainer: ci.yml missing from setup_guide")
	}

	// Non-setup items must not appear.
	if ids["file:internal/ingress/ingress.go"] {
		t.Error("maintainer: ingress.go should not appear in setup_guide")
	}
}

func TestSetupGuideEmployee(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["employee"]
	result := SetupGuide(role, setupItems())

	ids := make(map[string]bool)
	for _, item := range result.Items {
		ids[item.ID] = true
	}

	if !ids["file:README.md"] {
		t.Error("employee: README.md missing from setup_guide")
	}
	if !ids["file:.github/workflows/ci.yml"] {
		t.Error("employee: ci.yml missing from setup_guide")
	}
}

func TestSetupGuideContractor(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["contractor"]
	result := SetupGuide(role, setupItems())

	ids := make(map[string]bool)
	for _, item := range result.Items {
		ids[item.ID] = true
	}

	// Contractor holds docs:public so sees README.md (docs:public scoped).
	if !ids["file:README.md"] {
		t.Error("contractor: README.md missing from setup_guide")
	}

	// Makefile is path:repo, not held by contractor, so it should be withheld.
	if ids["file:Makefile"] {
		t.Error("contractor: Makefile should be withheld (path:repo not in contractor scopes)")
	}
}

func TestSetupGuideWithheldSummary(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["contractor"]
	result := SetupGuide(role, setupItems())

	if result.Withheld == nil {
		t.Error("withheld map must not be nil")
	}
}
