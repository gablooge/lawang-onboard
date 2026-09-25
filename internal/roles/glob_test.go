package roles

import "testing"

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		// Empty cases.
		{"", "", true},
		{"", "a", false},
		{"a", "", false},

		// Literal segments.
		{"README.md", "README.md", true},
		{"README.md", "readme.md", false},
		{"SECURITY.md", "SECURITY.md", true},
		{"SECURITY.md", "security.md", false},
		{"docs/adr/0001.md", "docs/adr/0001.md", true},
		{"docs/adr/0001.md", "docs/adr/0002.md", false},

		// Single-star: exactly one segment.
		{"*", "README.md", true},
		{"*", "foo", true},
		{"*", "foo/bar", false},
		{"docs/*", "docs/architecture.md", true},
		{"docs/*", "docs/adr/0001.md", false},
		{"internal/ingress/*", "internal/ingress/ingress.go", true},
		{"internal/ingress/*", "internal/ingress/sub/file.go", false},

		// Double-star: zero or more segments.
		{"**", "", true},
		{"**", "README.md", true},
		{"**", "docs/adr/0001.md", true},
		{"growth/**", "growth/README.md", true},
		{"growth/**", "growth/drafts/launch.md", true},
		{"growth/**", "other/README.md", false},
		{"docs/**", "docs/adr/0001.md", true},
		{"docs/**", "docs/adr/sub/deep.md", true},
		{"docs/**", "other/architecture.md", false},
		{"docs/adr/**", "docs/adr/0001.md", true},
		{"docs/adr/**", "docs/adr/sub/deep.md", true},
		{"docs/adr/**", "docs/other.md", false},
		{"internal/**", "internal/config/config.go", true},
		{"internal/**", "internal/ingress/ingress.go", true},
		{"internal/**", "other/file.go", false},
		{"cmd/**", "cmd/onboard/main.go", true},
		{"cmd/**", "other/main.go", false},
		{"migrations/**", "migrations/00001_tenancy.sql", true},

		// Patterns with ** matching zero segments.
		{"a/**/b", "a/b", true},
		{"a/**/b", "a/x/b", true},
		{"a/**/b", "a/x/y/b", true},
		{"a/**/b", "a/x/y/c", false},

		// Mixed.
		{"internal/*/config.go", "internal/config/config.go", true},
		{"internal/*/config.go", "internal/ingress/config.go", true},
		{"internal/*/config.go", "internal/a/b/config.go", false},

		// The catch-all pattern from roles.yaml.
		{"**", "anything/at/all", true},
	}

	for _, tt := range tests {
		got := matchGlob(tt.pattern, tt.path)
		if got != tt.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
		}
	}
}
