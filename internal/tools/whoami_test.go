package tools

import (
	"slices"
	"testing"

	"github.com/gablooge/lawang-onboard/internal/roles"
)

func TestWhoami(t *testing.T) {
	cfg := testCfg(t)

	tests := []struct {
		role       string
		wantScopes []string
	}{
		{"maintainer", []string{"*"}},
		{"employee", []string{"path:*", "docs:*", "adr:*", "backlog:*"}},
		{"contractor", []string{"path:internal/ingress", "path:internal/provider", "docs:public", "adr:public", "backlog:public"}},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			role, ok := cfg.Roles[tt.role]
			if !ok {
				t.Fatalf("role %q not found", tt.role)
			}
			result := Whoami(role)

			if result.Role != tt.role {
				t.Errorf("Role = %q, want %q", result.Role, tt.role)
			}
			if result.Scopes == nil {
				t.Error("Scopes must not be nil")
			}
			for _, want := range tt.wantScopes {
				if !slices.Contains(result.Scopes, want) {
					t.Errorf("scope %q missing from result scopes %v", want, result.Scopes)
				}
			}
			if result.Withheld == nil {
				t.Error("Withheld must not be nil")
			}
			if len(result.Withheld) != 0 {
				t.Errorf("whoami should have empty withheld, got %v", result.Withheld)
			}
		})
	}
}

// TestWhoamiZeroRole ensures the function works with a zero Role (no panic).
func TestWhoamiZeroRole(t *testing.T) {
	result := Whoami(roles.Role{})
	if result.Scopes == nil {
		t.Error("Scopes must not be nil even for zero role")
	}
	if result.Withheld == nil {
		t.Error("Withheld must not be nil")
	}
}
