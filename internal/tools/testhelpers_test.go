package tools

import (
	"testing"

	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// rolesYAML is the canonical demo roles for tool tests.
const rolesYAML = `
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
    token_env: TEST_TOKEN_MAINTAINER
    scopes: ["*"]
  employee:
    token_env: TEST_TOKEN_EMPLOYEE
    scopes: ["path:*", "docs:*", "adr:*", "backlog:*"]
  contractor:
    token_env: TEST_TOKEN_CONTRACTOR
    scopes: ["path:internal/ingress", "path:internal/provider", "docs:public", "adr:public", "backlog:public"]
`

// testCfg parses the embedded rolesYAML and returns the Config.
func testCfg(t *testing.T) *roles.Config {
	t.Helper()
	cfg, err := roles.ParseBytes([]byte(rolesYAML))
	if err != nil {
		t.Fatalf("parse roles: %v", err)
	}
	return cfg
}

// testItems returns a small synthetic corpus for unit tests.
func testItems() []corpus.Item {
	return []corpus.Item{
		// file visible to all path:* holders and maintainer
		{
			ID:    "file:internal/ingress/ingress.go",
			Kind:  "file",
			Title: "ingress.go",
			Paths: []string{"internal/ingress/ingress.go"},
			Text:  "package ingress",
			Scopes: []string{"path:internal/ingress"},
		},
		// file only maintainer and employee see
		{
			ID:    "file:SECURITY.md",
			Kind:  "file",
			Title: "SECURITY.md",
			Paths: []string{"SECURITY.md"},
			Text:  "security policy",
			Scopes: []string{"private:security"},
		},
		// doc all see
		{
			ID:    "doc:docs/architecture.md",
			Kind:  "doc",
			Title: "architecture.md",
			Paths: []string{"docs/architecture.md"},
			Text:  "architecture overview",
			Scopes: []string{"docs:public"},
		},
		// adr visible to employee and contractor (adr:public)
		{
			ID:    "adr:docs/adr/0001.md",
			Kind:  "adr",
			Title: "0001.md",
			Paths: []string{"docs/adr/0001.md"},
			Text:  "decision record",
			Scopes: []string{"adr:public"},
		},
		// growth item only maintainer sees
		{
			ID:    "doc:growth/README.md",
			Kind:  "doc",
			Title: "growth README",
			Paths: []string{"growth/README.md"},
			Text:  "growth content",
			Scopes: []string{"private:growth"},
		},
		// issue visible to employee and contractor
		{
			ID:    "issue:1",
			Kind:  "issue",
			Title: "B01: Skeleton",
			Text:  "backlog item",
			Scopes: []string{"backlog:public"},
		},
		// item with no scopes: denied for all
		{
			ID:    "file:noscope",
			Kind:  "file",
			Title: "no scope item",
			Text:  "no scopes",
			Scopes: nil,
		},
	}
}
