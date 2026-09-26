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

// testCfg parses rolesYAML and sets deterministic token env vars for this test run.
func testCfg(t *testing.T) *roles.Config {
	t.Helper()
	t.Setenv("TEST_WEB_TOKEN_MAINTAINER", "tok-web-maintainer")
	t.Setenv("TEST_WEB_TOKEN_EMPLOYEE", "tok-web-employee")
	t.Setenv("TEST_WEB_TOKEN_CONTRACTOR", "tok-web-contractor")
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

// do performs a GET to path on the given handler (no Authorization header).
func do(h http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// doWithToken performs a GET with an Authorization: Bearer header.
func doWithToken(h http.Handler, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// TestGetRoles_OK tests that /api/roles returns all three role names (public, no token needed).
func TestGetRoles_OK(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems(), Options{})
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

// TestAPIWithoutToken verifies that /api/* returns 401 with no token (default mode).
func TestAPIWithoutToken(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems(), Options{})
	for _, path := range []string{"/api/tour", "/api/withheld", "/api/search?q=x", "/api/audit"} {
		rec := do(h, path)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want 401", path, rec.Code)
		}
	}
}

// TestAPIRolesPublic verifies /api/roles is always public regardless of mode.
func TestAPIRolesPublic(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems(), Options{})
	rec := do(h, "/api/roles")
	if rec.Code != http.StatusOK {
		t.Errorf("/api/roles without token: status = %d, want 200", rec.Code)
	}
}

// TestAPIWithToken verifies that a valid Bearer token grants access and returns data.
func TestAPIWithToken(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems(), Options{})
	// Maintainer token should see the tour.
	rec := doWithToken(h, "/api/tour", "tok-web-maintainer")
	if rec.Code != http.StatusOK {
		t.Errorf("maintainer tour: status = %d, want 200", rec.Code)
	}
}

// TestAPIContractorTokenCannotSeePrivate verifies that a contractor Bearer token
// never exposes maintainer-only items, even if ?role=maintainer is supplied.
func TestAPIContractorTokenCannotSeePrivate(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems(), Options{})

	// Pass contractor token but try to request maintainer role via query param.
	// The handler must ignore the query param and use the token-resolved role.
	rec := doWithToken(h, "/api/tour?role=maintainer", "tok-web-contractor")
	if rec.Code != http.StatusOK {
		t.Fatalf("contractor tour: status = %d, want 200", rec.Code)
	}
	var body struct {
		Nodes []map[string]any `json:"nodes"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, n := range body.Nodes {
		id, _ := n["id"].(string)
		if id == "file:SECURITY.md" {
			t.Error("contractor token must not expose SECURITY.md even with ?role=maintainer")
		}
		if id == "doc:growth/README.md" {
			t.Error("contractor token must not expose growth/README.md even with ?role=maintainer")
		}
	}
}

// TestAPIWithWrongToken verifies that an unknown token returns 401.
func TestAPIWithWrongToken(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems(), Options{})
	rec := doWithToken(h, "/api/tour", "not-a-real-token")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong token: status = %d, want 401", rec.Code)
	}
}

// TestDemoRolesMode verifies that when DemoRoles is on, ?role= is accepted.
func TestDemoRolesMode(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems(), Options{DemoRoles: true})
	// Should accept role by query param without any token.
	rec := do(h, "/api/tour?role=contractor")
	if rec.Code != http.StatusOK {
		t.Errorf("demo-roles tour: status = %d, want 200", rec.Code)
	}
	// Missing role still returns 400.
	rec = do(h, "/api/tour")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("demo-roles missing role: status = %d, want 400", rec.Code)
	}
	// Unknown role still returns 400.
	rec = do(h, "/api/tour?role=ghost")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("demo-roles unknown role: status = %d, want 400", rec.Code)
	}
}

// tourWithToken calls /api/tour with a Bearer token and returns the decoded nodes.
func tourWithToken(t *testing.T, h http.Handler, token string) []map[string]any {
	t.Helper()
	rec := doWithToken(h, "/api/tour", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("tour status = %d, want 200", rec.Code)
	}
	var body struct {
		Nodes []map[string]any `json:"nodes"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body.Nodes
}

// TestGetTour_EachRole verifies that each role sees a different set of items.
func TestGetTour_EachRole(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()
	h := Handler(cfg, items, Options{})

	maintainerNodes := tourWithToken(t, h, "tok-web-maintainer")
	employeeNodes := tourWithToken(t, h, "tok-web-employee")
	contractorNodes := tourWithToken(t, h, "tok-web-contractor")

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

// withheldWithToken calls /api/withheld with a Bearer token and returns the decoded result.
func withheldWithToken(t *testing.T, h http.Handler, token string) map[string]any {
	t.Helper()
	rec := doWithToken(h, "/api/withheld", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("withheld status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body
}

// TestGetWithheld_EachRole checks that withheld counts differ per role.
func TestGetWithheld_EachRole(t *testing.T) {
	cfg := testCfg(t)
	items := testItems()
	h := Handler(cfg, items, Options{})

	maintainerBody := withheldWithToken(t, h, "tok-web-maintainer")
	contractorBody := withheldWithToken(t, h, "tok-web-contractor")

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

// searchWithToken calls /api/search with a Bearer token and returns the hit IDs.
func searchWithToken(t *testing.T, h http.Handler, token, q string) []string {
	t.Helper()
	rec := doWithToken(h, "/api/search?q="+q, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("search status = %d, want 200", rec.Code)
	}
	var body struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
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
	h := Handler(cfg, items, Options{})

	// "security" appears only in SECURITY.md (private:security). Contractor cannot see it.
	contractorIDs := searchWithToken(t, h, "tok-web-contractor", "security")
	for _, id := range contractorIDs {
		if id == "file:SECURITY.md" {
			t.Error("contractor should not receive SECURITY.md in search results")
		}
	}

	// Maintainer can see SECURITY.md (scope "*").
	maintainerIDs := searchWithToken(t, h, "tok-web-maintainer", "security")
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
	employeeIDs := searchWithToken(t, h, "tok-web-employee", "security")
	for _, id := range employeeIDs {
		if id == "file:SECURITY.md" {
			t.Error("employee should not see SECURITY.md in search results")
		}
	}
}

// TestGetAudit_OK tests that audit entries are returned after a tour call.
func TestGetAudit_OK(t *testing.T) {
	// Replace global audit ring to isolate this test.
	prev := audit.Global
	audit.Global = &audit.Ring{}
	t.Cleanup(func() { audit.Global = prev })

	cfg := testCfg(t)
	h := Handler(cfg, testItems(), Options{})

	// Call tour to produce an audit entry.
	doWithToken(h, "/api/tour", "tok-web-maintainer")

	rec := doWithToken(h, "/api/audit", "tok-web-maintainer")
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

// TestGetAudit_RoleFilter verifies that /api/audit returns only entries for the
// authenticated role. A contractor must not see maintainer entries, and must not
// receive any item ID that is outside the contractor's allowed scopes.
func TestGetAudit_RoleFilter(t *testing.T) {
	prev := audit.Global
	audit.Global = &audit.Ring{}
	t.Cleanup(func() { audit.Global = prev })

	cfg := testCfg(t)
	items := testItems()
	h := Handler(cfg, items, Options{})

	// Generate audit entries for two different roles.
	doWithToken(h, "/api/tour", "tok-web-maintainer")
	doWithToken(h, "/api/search?q=security", "tok-web-maintainer")
	doWithToken(h, "/api/tour", "tok-web-contractor")

	// Contractor's audit must contain only contractor entries.
	rec := doWithToken(h, "/api/audit", "tok-web-contractor")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var entries []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Items that a contractor cannot see.
	contractorForbidden := map[string]bool{
		"file:SECURITY.md":     true,
		"doc:growth/README.md": true,
	}

	for _, e := range entries {
		role, _ := e["role"].(string)
		if role != "contractor" {
			t.Errorf("contractor audit contains entry for role %q", role)
		}
		ids, _ := e["returned_ids"].([]any)
		for _, raw := range ids {
			id, _ := raw.(string)
			if contractorForbidden[id] {
				t.Errorf("contractor audit contains forbidden item ID %q", id)
			}
		}
	}

	// Sanity: audit must contain at least one contractor entry.
	if len(entries) == 0 {
		t.Error("expected at least one contractor audit entry")
	}
}

// TestIndexHTML verifies that GET / returns the embedded HTML.
func TestIndexHTML(t *testing.T) {
	cfg := testCfg(t)
	h := Handler(cfg, testItems(), Options{})
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
