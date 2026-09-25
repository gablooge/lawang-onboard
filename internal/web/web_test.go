package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gablooge/lawang-onboard/internal/audit"
	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// rolesYAML mirrors the canonical demo roles for web tests.
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
    token_env: TEST_WEB_TOKEN_MAINTAINER
    scopes: ["*"]
  employee:
    token_env: TEST_WEB_TOKEN_EMPLOYEE
    scopes: ["path:*", "docs:*", "adr:*", "backlog:*"]
  contractor:
    token_env: TEST_WEB_TOKEN_CONTRACTOR
    scopes: ["path:internal/ingress", "path:internal/provider", "docs:public", "adr:public", "backlog:public"]
`

// testCfg parses rolesYAML.
func testCfg(t *testing.T) *roles.Config {
	t.Helper()
	cfg, err := roles.ParseBytes([]byte(rolesYAML))
	if err != nil {
		t.Fatalf("parse roles: %v", err)
	}
	return cfg
}

// testItems returns a small synthetic corpus.
func testItems() []corpus.Item {
	return []corpus.Item{
		{
			ID: "file:internal/ingress/ingress.go", Kind: "file",
			Title: "ingress.go", Paths: []string{"internal/ingress/ingress.go"},
			Text: "package ingress", Scopes: []string{"path:internal/ingress"},
		},
		{
			ID: "file:SECURITY.md", Kind: "file",
			Title: "SECURITY.md", Paths: []string{"SECURITY.md"},
			Text: "security policy", Scopes: []string{"private:security"},
		},
		{
			ID: "doc:docs/architecture.md", Kind: "doc",
			Title: "architecture.md", Paths: []string{"docs/architecture.md"},
			Text: "architecture overview", Scopes: []string{"docs:public"},
		},
		{
			ID: "doc:growth/README.md", Kind: "doc",
			Title: "growth README", Paths: []string{"growth/README.md"},
			Text: "growth content", Scopes: []string{"private:growth"},
		},
		{
			ID: "issue:1", Kind: "issue",
			Title: "B01: Skeleton", Text: "backlog item",
			Scopes: []string{"backlog:public"},
		},
	}
}

// do performs a GET to path on the given handler.
func do(h http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// TestGetRoles_OK tests that /api/roles returns all three role names.
func TestGetRoles_OK(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems())
	rec := do(h, "/api/roles")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var names []string
	if err := json.NewDecoder(rec.Body).Decode(&names); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("len(names) = %d, want 3", len(names))
	}
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	for _, want := range []string{"maintainer", "employee", "contractor"} {
		if !found[want] {
			t.Errorf("role %q not found in response", want)
		}
	}
}

// TestGetTour_UnknownRole verifies that an unknown role returns 400.
func TestGetTour_UnknownRole(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems())
	rec := do(h, "/api/tour?role=ghost")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// TestGetTour_MissingRole verifies that omitting role returns 400.
func TestGetTour_MissingRole(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems())
	rec := do(h, "/api/tour")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// tourRole calls /api/tour?role=<name> and returns the decoded nodes.
func tourRole(t *testing.T, h http.Handler, role string) []map[string]any {
	t.Helper()
	rec := do(h, "/api/tour?role="+role)
	if rec.Code != http.StatusOK {
		t.Fatalf("[%s] tour status = %d, want 200", role, rec.Code)
	}
	var body struct {
		Nodes []map[string]any `json:"nodes"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("[%s] decode: %v", role, err)
	}
	return body.Nodes
}

// TestGetTour_EachRole verifies that each role sees a different set of items.
func TestGetTour_EachRole(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()
	h := Handler(cfg, items)

	maintainerNodes := tourRole(t, h, "maintainer")
	employeeNodes := tourRole(t, h, "employee")
	contractorNodes := tourRole(t, h, "contractor")

	// Maintainer sees everything (including SECURITY.md and growth items).
	if len(maintainerNodes) <= len(employeeNodes) {
		t.Errorf("maintainer should see more nodes than employee: %d vs %d",
			len(maintainerNodes), len(employeeNodes))
	}

	// Employee should not see growth README (private:growth).
	for _, n := range employeeNodes {
		if n["id"] == "doc:growth/README.md" {
			t.Error("employee should not see growth/README.md")
		}
	}

	// Contractor cannot see SECURITY.md (private:security) or growth items.
	for _, n := range contractorNodes {
		id, _ := n["id"].(string)
		if id == "file:SECURITY.md" {
			t.Error("contractor should not see SECURITY.md")
		}
		if id == "doc:growth/README.md" {
			t.Error("contractor should not see growth/README.md")
		}
	}

	// Contractor can see the ingress file.
	found := false
	for _, n := range contractorNodes {
		if n["id"] == "file:internal/ingress/ingress.go" {
			found = true
		}
	}
	if !found {
		t.Error("contractor should see file:internal/ingress/ingress.go")
	}
}

// TestGetWithheld_UnknownRole verifies 400 for unknown role.
func TestGetWithheld_UnknownRole(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems())
	rec := do(h, "/api/withheld?role=ghost")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// withheldRole calls /api/withheld?role=<name> and returns the decoded result.
func withheldRole(t *testing.T, h http.Handler, role string) map[string]any {
	t.Helper()
	rec := do(h, "/api/withheld?role="+role)
	if rec.Code != http.StatusOK {
		t.Fatalf("[%s] withheld status = %d, want 200", role, rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("[%s] decode: %v", role, err)
	}
	return body
}

// TestGetWithheld_EachRole checks that withheld counts differ per role.
func TestGetWithheld_EachRole(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()
	h := Handler(cfg, items)

	maintainerBody := withheldRole(t, h, "maintainer")
	contractorBody := withheldRole(t, h, "contractor")

	maintainerTotal := int(maintainerBody["total"].(float64))
	contractorTotal := int(contractorBody["total"].(float64))

	// Maintainer (scope "*") should see everything, so total withheld = 0.
	if maintainerTotal != 0 {
		t.Errorf("maintainer withheld total = %d, want 0", maintainerTotal)
	}
	// Contractor cannot see SECURITY.md and growth items, so total > 0.
	if contractorTotal == 0 {
		t.Errorf("contractor withheld total = 0, want > 0")
	}
}

// TestGetSearch_UnknownRole verifies 400.
func TestGetSearch_UnknownRole(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems())
	rec := do(h, "/api/search?role=ghost&q=ingress")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// searchRole calls /api/search?role=<name>&q=<q> and returns the hit IDs.
func searchRole(t *testing.T, h http.Handler, role, q string) []string {
	t.Helper()
	rec := do(h, "/api/search?role="+role+"&q="+q)
	if rec.Code != http.StatusOK {
		t.Fatalf("[%s] search status = %d, want 200", role, rec.Code)
	}
	var body struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("[%s] decode: %v", role, err)
	}
	ids := make([]string, len(body.Items))
	for i, it := range body.Items {
		ids[i] = it.ID
	}
	return ids
}

// TestGetSearch_EachRole ensures that a word only in a private item is hidden from unprivileged roles.
func TestGetSearch_EachRole(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()
	h := Handler(cfg, items)

	// "security" appears only in SECURITY.md (private:security). Contractor cannot see it.
	contractorIDs := searchRole(t, h, "contractor", "security")
	for _, id := range contractorIDs {
		if id == "file:SECURITY.md" {
			t.Error("contractor should not receive SECURITY.md in search results")
		}
	}

	// Maintainer can see SECURITY.md (scope "*").
	maintainerIDs := searchRole(t, h, "maintainer", "security")
	found := false
	for _, id := range maintainerIDs {
		if id == "file:SECURITY.md" {
			found = true
		}
	}
	if !found {
		t.Error("maintainer should see SECURITY.md in search for 'security'")
	}

	// Employee cannot see SECURITY.md (private:security not in employee scopes).
	employeeIDs := searchRole(t, h, "employee", "security")
	for _, id := range employeeIDs {
		if id == "file:SECURITY.md" {
			t.Error("employee should not see SECURITY.md in search results")
		}
	}
}

// TestGetAudit_UnknownRole verifies 400.
func TestGetAudit_UnknownRole(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems())
	rec := do(h, "/api/audit?role=ghost")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// TestGetAudit_OK tests that audit entries are returned after a tour call.
func TestGetAudit_OK(t *testing.T) {
	// Replace global audit ring to isolate this test.
	prev := audit.Global
	audit.Global = &audit.Ring{}
	t.Cleanup(func() { audit.Global = prev })

	cfg := testCfg(t)
	h := Handler(cfg, testItems())

	// Call tour to produce an audit entry.
	do(h, "/api/tour?role=maintainer")

	rec := do(h, "/api/audit?role=maintainer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var entries []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one audit entry after tour call")
	}
	// Check that the entry has the right role and tool.
	e := entries[0]
	if e["role"] != "maintainer" {
		t.Errorf("role = %v, want maintainer", e["role"])
	}
	if e["tool"] != "tour" {
		t.Errorf("tool = %v, want tour", e["tool"])
	}
}

// TestIndexHTML verifies that GET / returns the embedded HTML.
func TestIndexHTML(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems())
	rec := do(h, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct == "" {
		t.Error("Content-Type should not be empty")
	}
	body := rec.Body.String()
	if len(body) < 100 {
		t.Errorf("response body too short (%d bytes), expected HTML content", len(body))
	}
	if !contains(body, "Lawang Onboard") {
		t.Error("expected 'Lawang Onboard' in the HTML response")
	}
}

// TestGetSearch_MissingRole verifies 400 for missing role param.
func TestGetSearch_MissingRole(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems())
	rec := do(h, "/api/search?q=ingress")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// TestGetWithheld_MissingRole verifies 400 for missing role param.
func TestGetWithheld_MissingRole(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems())
	rec := do(h, "/api/withheld")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// TestGetAudit_MissingRole verifies 400 for missing role param.
func TestGetAudit_MissingRole(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems())
	rec := do(h, "/api/audit")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
