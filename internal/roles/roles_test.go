package roles

import (
	"slices"
	"testing"
)

// minimalYAML is a small but valid roles.yaml for unit tests.
const minimalYAML = `
scopes:
  - glob: "growth/**"
    scope: "private:growth"
  - glob: "SECURITY.md"
    scope: "private:security"
  - glob: "docs/adr/**"
    scope: "adr:public"
  - glob: "docs/**"
    scope: "docs:public"
  - glob: "internal/ingress/**"
    scope: "path:internal/ingress"
  - glob: "internal/provider/**"
    scope: "path:internal/provider"
  - glob: "internal/**"
    scope: "path:internal/core"
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

func TestParseScopes(t *testing.T) {
	cfg, err := parse([]byte(minimalYAML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(cfg.ScopeRules) == 0 {
		t.Fatal("no scope rules parsed")
	}

	// First rule: growth/** -> private:growth
	if cfg.ScopeRules[0].Glob != "growth/**" || cfg.ScopeRules[0].Scope != "private:growth" {
		t.Errorf("ScopeRules[0] = {%q, %q}, want {growth/**, private:growth}",
			cfg.ScopeRules[0].Glob, cfg.ScopeRules[0].Scope)
	}

	// Catch-all last: ** -> path:repo
	last := cfg.ScopeRules[len(cfg.ScopeRules)-1]
	if last.Glob != "**" || last.Scope != "path:repo" {
		t.Errorf("last rule = {%q, %q}, want {**, path:repo}", last.Glob, last.Scope)
	}
}

func TestParseRoles(t *testing.T) {
	cfg, err := parse([]byte(minimalYAML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	for _, name := range []string{"maintainer", "employee", "contractor"} {
		r, ok := cfg.Roles[name]
		if !ok {
			t.Errorf("role %q missing", name)
			continue
		}
		if r.Name != name {
			t.Errorf("role %q: Name = %q", name, r.Name)
		}
		if len(r.Scopes) == 0 {
			t.Errorf("role %q: no scopes", name)
		}
	}

	if !slices.Contains(cfg.Roles["maintainer"].Scopes, "*") {
		t.Error("maintainer scopes do not contain *")
	}
}

func TestParseLabelScopes(t *testing.T) {
	cfg, err := parse([]byte(minimalYAML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := cfg.LabelScopes["growth"]; got != "private:growth" {
		t.Errorf("LabelScopes[growth] = %q, want private:growth", got)
	}
}

func TestParseLabelAreas(t *testing.T) {
	cfg, err := parse([]byte(minimalYAML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := cfg.LabelAreas["provider"]; got != "path:internal/provider" {
		t.Errorf("LabelAreas[provider] = %q, want path:internal/provider", got)
	}
}

func TestParseKindScopes(t *testing.T) {
	cfg, err := parse([]byte(minimalYAML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := cfg.KindScopes["issue"]; got != "backlog:public" {
		t.Errorf("KindScopes[issue] = %q, want backlog:public", got)
	}
}

func TestPathScope(t *testing.T) {
	cfg, err := parse([]byte(minimalYAML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	rules := cfg.ScopeRules

	tests := []struct {
		path string
		want string
	}{
		{"growth/README.md", "private:growth"},
		{"SECURITY.md", "private:security"},
		{"docs/adr/0001.md", "adr:public"},
		{"docs/architecture.md", "docs:public"},
		{"internal/ingress/ingress.go", "path:internal/ingress"},
		{"internal/provider/provider.go", "path:internal/provider"},
		{"internal/config/config.go", "path:internal/core"},
		{"README.md", "path:repo"},
		{"Makefile", "path:repo"},
		{"cmd/onboard/main.go", "path:repo"},
	}

	for _, tt := range tests {
		got := PathScope(rules, tt.path)
		if got != tt.want {
			t.Errorf("PathScope(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestPathScopeUnmapped(t *testing.T) {
	// A path that matches no rule returns "(unmapped)".
	rules := []ScopeRule{
		{Glob: "docs/**", Scope: "docs:public"},
	}
	got := PathScope(rules, "internal/config/config.go")
	if got != "(unmapped)" {
		t.Errorf("PathScope = %q, want (unmapped)", got)
	}
}
