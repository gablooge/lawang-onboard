package filter

import (
	"testing"

	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// maintainerRole holds the "*" wildcard - can see anything that is not denied early.
var maintainerRole = roles.Role{Name: "maintainer", Scopes: []string{"*"}}

// employeeRole holds path:*, docs:*, adr:*, backlog:* wildcards.
var employeeRole = roles.Role{
	Name:   "employee",
	Scopes: []string{"path:*", "docs:*", "adr:*", "backlog:*"},
}

// contractorRole holds specific scopes only.
var contractorRole = roles.Role{
	Name:   "contractor",
	Scopes: []string{"path:internal/ingress", "path:internal/provider", "docs:public", "adr:public", "backlog:public", "setup:public"},
}

func TestVisible(t *testing.T) {
	tests := []struct {
		name    string
		role    roles.Role
		item    corpus.Item
		want    bool
	}{
		// ----------------------------------------------------------------
		// Rule 1: empty Scopes -> deny for every role including maintainer
		// ----------------------------------------------------------------
		{
			name: "no-scope nil: deny maintainer",
			role: maintainerRole,
			item: corpus.Item{ID: "x", Scopes: nil},
			want: false,
		},
		{
			name: "no-scope empty: deny maintainer",
			role: maintainerRole,
			item: corpus.Item{ID: "x", Scopes: []string{}},
			want: false,
		},
		{
			name: "no-scope nil: deny employee",
			role: employeeRole,
			item: corpus.Item{ID: "x", Scopes: nil},
			want: false,
		},
		{
			name: "no-scope nil: deny contractor",
			role: contractorRole,
			item: corpus.Item{ID: "x", Scopes: nil},
			want: false,
		},

		// ----------------------------------------------------------------
		// Rule 2: (unmapped) -> deny for every role including maintainer
		// ----------------------------------------------------------------
		{
			name: "(unmapped) alone: deny maintainer",
			role: maintainerRole,
			item: corpus.Item{ID: "x", Scopes: []string{"(unmapped)"}},
			want: false,
		},
		{
			name: "(unmapped) alone: deny employee",
			role: employeeRole,
			item: corpus.Item{ID: "x", Scopes: []string{"(unmapped)"}},
			want: false,
		},
		{
			name: "(unmapped) alone: deny contractor",
			role: contractorRole,
			item: corpus.Item{ID: "x", Scopes: []string{"(unmapped)"}},
			want: false,
		},
		{
			name: "(unmapped) mixed with valid scope: deny maintainer",
			role: maintainerRole,
			item: corpus.Item{ID: "x", Scopes: []string{"docs:public", "(unmapped)"}},
			want: false,
		},

		// ----------------------------------------------------------------
		// Rule 3: role has "*" -> allow (after rules 1 and 2)
		// ----------------------------------------------------------------
		{
			name: "maintainer: allow private:security item",
			role: maintainerRole,
			item: corpus.Item{ID: "security", Scopes: []string{"private:security"}},
			want: true,
		},
		{
			name: "maintainer: allow private:growth item",
			role: maintainerRole,
			item: corpus.Item{ID: "growth", Scopes: []string{"private:growth"}},
			want: true,
		},
		{
			name: "maintainer: allow multi-scope item",
			role: maintainerRole,
			item: corpus.Item{ID: "multi", Scopes: []string{"path:internal/core", "docs:public"}},
			want: true,
		},

		// ----------------------------------------------------------------
		// Rule 4: exact scope match
		// ----------------------------------------------------------------
		{
			name: "contractor: exact match docs:public",
			role: contractorRole,
			item: corpus.Item{ID: "doc", Scopes: []string{"docs:public"}},
			want: true,
		},
		{
			name: "contractor: exact match backlog:public",
			role: contractorRole,
			item: corpus.Item{ID: "issue", Scopes: []string{"backlog:public"}},
			want: true,
		},
		{
			name: "contractor: exact match path:internal/ingress",
			role: contractorRole,
			item: corpus.Item{ID: "ingress", Scopes: []string{"path:internal/ingress"}},
			want: true,
		},
		{
			name: "contractor: exact match path:internal/provider",
			role: contractorRole,
			item: corpus.Item{ID: "provider", Scopes: []string{"path:internal/provider"}},
			want: true,
		},

		// ----------------------------------------------------------------
		// Rule 4: wildcard "prefix:*" match
		// ----------------------------------------------------------------
		{
			name: "employee: path:* covers path:internal/core",
			role: employeeRole,
			item: corpus.Item{ID: "core", Scopes: []string{"path:internal/core"}},
			want: true,
		},
		{
			name: "employee: path:* covers path:internal/ingress",
			role: employeeRole,
			item: corpus.Item{ID: "ingress", Scopes: []string{"path:internal/ingress"}},
			want: true,
		},
		{
			name: "employee: docs:* covers docs:public",
			role: employeeRole,
			item: corpus.Item{ID: "doc", Scopes: []string{"docs:public"}},
			want: true,
		},
		{
			name: "employee: adr:* covers adr:public",
			role: employeeRole,
			item: corpus.Item{ID: "adr", Scopes: []string{"adr:public"}},
			want: true,
		},
		{
			name: "employee: backlog:* covers backlog:public",
			role: employeeRole,
			item: corpus.Item{ID: "backlog", Scopes: []string{"backlog:public"}},
			want: true,
		},

		// ----------------------------------------------------------------
		// Rule 4: denial - missing scope
		// ----------------------------------------------------------------
		{
			name: "contractor: deny private:security",
			role: contractorRole,
			item: corpus.Item{ID: "sec", Scopes: []string{"private:security"}},
			want: false,
		},
		{
			name: "contractor: deny private:growth",
			role: contractorRole,
			item: corpus.Item{ID: "growth", Scopes: []string{"private:growth"}},
			want: false,
		},
		{
			name: "contractor: deny path:internal/core (not in exact list)",
			role: contractorRole,
			item: corpus.Item{ID: "core", Scopes: []string{"path:internal/core"}},
			want: false,
		},
		{
			name: "employee: deny private:security",
			role: employeeRole,
			item: corpus.Item{ID: "sec", Scopes: []string{"private:security"}},
			want: false,
		},
		{
			name: "employee: deny private:growth",
			role: employeeRole,
			item: corpus.Item{ID: "growth", Scopes: []string{"private:growth"}},
			want: false,
		},

		// ----------------------------------------------------------------
		// Rule 4: multi-scope item - ALL must be covered
		// ----------------------------------------------------------------
		{
			name: "contractor: deny multi-scope item requiring path:internal/core",
			role: contractorRole,
			item: corpus.Item{ID: "multi", Scopes: []string{"path:internal/ingress", "path:internal/core"}},
			want: false,
		},
		{
			name: "employee: allow multi-scope item with path:* and docs:*",
			role: employeeRole,
			item: corpus.Item{ID: "multi", Scopes: []string{"path:internal/ingress", "docs:public"}},
			want: true,
		},
		{
			name: "contractor: allow multi-scope item with both held exactly",
			role: contractorRole,
			item: corpus.Item{ID: "multi2", Scopes: []string{"path:internal/ingress", "path:internal/provider"}},
			want: true,
		},

		// ----------------------------------------------------------------
		// Wildcard "prefix:*" does not match unrelated prefixes
		// ----------------------------------------------------------------
		{
			name: "employee: path:* does not cover docs:public",
			role: roles.Role{Name: "t", Scopes: []string{"path:*"}},
			item: corpus.Item{ID: "doc", Scopes: []string{"docs:public"}},
			want: false,
		},
		{
			name: "employee: docs:* does not cover backlog:public",
			role: roles.Role{Name: "t", Scopes: []string{"docs:*"}},
			item: corpus.Item{ID: "bl", Scopes: []string{"backlog:public"}},
			want: false,
		},

		// ----------------------------------------------------------------
		// Edge: unknown scope name is denied (fail closed)
		// ----------------------------------------------------------------
		{
			name: "employee: unknown scope denied",
			role: employeeRole,
			item: corpus.Item{ID: "unk", Scopes: []string{"unknown:thing"}},
			want: false,
		},
		{
			name: "contractor: unknown scope denied",
			role: contractorRole,
			item: corpus.Item{ID: "unk", Scopes: []string{"unknown:thing"}},
			want: false,
		},

		// ----------------------------------------------------------------
		// setup:public - contractor sees Makefile and CI, not private/path:repo
		// ----------------------------------------------------------------
		{
			name: "contractor: allow setup:public (Makefile)",
			role: contractorRole,
			item: corpus.Item{ID: "file-Makefile", Scopes: []string{"setup:public"}},
			want: true,
		},
		{
			name: "contractor: allow setup:public (.github/workflows/ci.yml)",
			role: contractorRole,
			item: corpus.Item{ID: "file-.github-workflows-ci.yml", Scopes: []string{"setup:public"}},
			want: true,
		},
		{
			name: "contractor: deny path:repo",
			role: contractorRole,
			item: corpus.Item{ID: "file-go.sum", Scopes: []string{"path:repo"}},
			want: false,
		},
		{
			name: "contractor: deny path:cmd",
			role: contractorRole,
			item: corpus.Item{ID: "file-cmd-main", Scopes: []string{"path:cmd"}},
			want: false,
		},
		{
			name: "contractor: deny private:growth",
			role: contractorRole,
			item: corpus.Item{ID: "growth2", Scopes: []string{"private:growth"}},
			want: false,
		},
		{
			name: "contractor: deny private:security",
			role: contractorRole,
			item: corpus.Item{ID: "sec2", Scopes: []string{"private:security"}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Visible(tt.role, tt.item)
			if got != tt.want {
				t.Errorf("Visible(%q, scopes=%v) = %v, want %v",
					tt.role.Name, tt.item.Scopes, got, tt.want)
			}
		})
	}
}

func TestWithheld(t *testing.T) {
	t.Run("no-scope item records (no-scope)", func(t *testing.T) {
		w := make(Withheld)
		item := corpus.Item{ID: "x", Scopes: nil}
		if Visible(contractorRole, item) {
			t.Fatal("expected denial")
		}
		w.Record(contractorRole, item)
		s := w.Summary()
		if s["(no-scope)"] != 1 {
			t.Errorf("want (no-scope)=1, got %v", s)
		}
	})

	t.Run("denied item records missing scopes only", func(t *testing.T) {
		w := make(Withheld)
		item := corpus.Item{ID: "sec", Scopes: []string{"private:security"}}
		if Visible(contractorRole, item) {
			t.Fatal("expected denial")
		}
		w.Record(contractorRole, item)
		s := w.Summary()
		if s["private:security"] != 1 {
			t.Errorf("want private:security=1, got %v", s)
		}
		// contractor holds docs:public; a denied item with that scope must NOT
		// appear in the withheld summary.
		itemHeld := corpus.Item{ID: "multi", Scopes: []string{"docs:public", "private:security"}}
		w2 := make(Withheld)
		if Visible(contractorRole, itemHeld) {
			t.Fatal("expected denial for multi-scope item")
		}
		w2.Record(contractorRole, itemHeld)
		s2 := w2.Summary()
		if _, ok := s2["docs:public"]; ok {
			t.Errorf("docs:public is held by contractor but appeared in withheld summary: %v", s2)
		}
		if s2["private:security"] != 1 {
			t.Errorf("want private:security=1 in withheld summary, got %v", s2)
		}
	})

	t.Run("accumulates across multiple items", func(t *testing.T) {
		w := make(Withheld)
		items := []corpus.Item{
			{ID: "a", Scopes: []string{"private:security"}},
			{ID: "b", Scopes: []string{"private:security"}},
			{ID: "c", Scopes: []string{"private:growth"}},
		}
		for _, item := range items {
			if !Visible(contractorRole, item) {
				w.Record(contractorRole, item)
			}
		}
		s := w.Summary()
		if s["private:security"] != 2 {
			t.Errorf("want private:security=2, got %v", s)
		}
		if s["private:growth"] != 1 {
			t.Errorf("want private:growth=1, got %v", s)
		}
	})

	t.Run("Summary returns a copy", func(t *testing.T) {
		w := make(Withheld)
		w["x"] = 5
		s := w.Summary()
		s["x"] = 99 // mutating the copy must not affect w
		if w["x"] != 5 {
			t.Error("Summary did not return a copy")
		}
	})

	t.Run("(unmapped) item records (unmapped)", func(t *testing.T) {
		w := make(Withheld)
		item := corpus.Item{ID: "u", Scopes: []string{"(unmapped)"}}
		if Visible(maintainerRole, item) {
			t.Fatal("expected denial for maintainer too")
		}
		w.Record(maintainerRole, item)
		s := w.Summary()
		if s["(unmapped)"] != 1 {
			t.Errorf("want (unmapped)=1, got %v", s)
		}
	})

	t.Run("held scopes not counted for employee", func(t *testing.T) {
		// employee holds path:*, so path:internal/core is covered.
		// A denied item with only path:internal/core should NOT appear
		// in the withheld summary because employee holds path:*.
		// (In practice employee IS allowed such items, but this verifies
		// the Record logic in isolation with a role that cannot see the item.)
		limitedRole := roles.Role{
			Name:   "limited",
			Scopes: []string{"path:internal/ingress"},
		}
		item := corpus.Item{ID: "core", Scopes: []string{"path:internal/core"}}
		if Visible(limitedRole, item) {
			t.Fatal("expected denial")
		}
		w := make(Withheld)
		w.Record(limitedRole, item)
		s := w.Summary()
		if _, ok := s["path:internal/ingress"]; ok {
			t.Errorf("path:internal/ingress is held by role but appeared in withheld: %v", s)
		}
		if s["path:internal/core"] != 1 {
			t.Errorf("want path:internal/core=1, got %v", s)
		}
	})
}
