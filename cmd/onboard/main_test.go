package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gablooge/lawang-onboard/internal/roles"
)

// okHandler is a trivial http.Handler that always returns 200.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// minimalRolesYAML is the embedded roles config used by cmd/onboard tests.
const minimalRolesYAML = `
scopes:
  - glob: "internal/ingress/**"
    scope: "path:internal/ingress"
  - glob: "**"
    scope: "path:repo"

kind_scopes:
  issue: "backlog:public"
  issue_comment: "backlog:public"

roles:
  maintainer:
    token_env: TEST_ONBOARD_TOKEN_MAINTAINER
    scopes: ["*"]
  contractor:
    token_env: TEST_ONBOARD_TOKEN_CONTRACTOR
    scopes: ["path:internal/ingress"]
`

// testSetup parses the roles YAML, sets token env vars, and returns the config.
// It installs t.Cleanup to restore the env vars after the test.
func testSetup(t *testing.T) *roles.Config {
	t.Helper()
	cfg, err := roles.ParseBytes([]byte(minimalRolesYAML))
	if err != nil {
		t.Fatalf("parse roles: %v", err)
	}

	// Set deterministic test tokens.
	tokens := map[string]string{
		"TEST_ONBOARD_TOKEN_MAINTAINER": "tok-maintainer-test",
		"TEST_ONBOARD_TOKEN_CONTRACTOR": "tok-contractor-test",
	}
	for k, v := range tokens {
		t.Setenv(k, v)
	}

	return cfg
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	cfg := testSetup(t)
	handler := authMiddleware(cfg, okHandler)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_WrongToken(t *testing.T) {
	cfg := testSetup(t)
	handler := authMiddleware(cfg, okHandler)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer wrong-token-that-nobody-has")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	cfg := testSetup(t)

	// okHandler always returns 200. The authMiddleware should pass through for a valid token.
	handler := authMiddleware(cfg, okHandler)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("TEST_ONBOARD_TOKEN_MAINTAINER"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Errorf("valid token should not return 401")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestBearerRole_MissingHeader(t *testing.T) {
	cfg := testSetup(t)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	_, ok := bearerRole(cfg, req)
	if ok {
		t.Error("bearerRole: expected false for request with no Authorization header")
	}
}

func TestBearerRole_BadScheme(t *testing.T) {
	cfg := testSetup(t)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Basic somevalue")
	_, ok := bearerRole(cfg, req)
	if ok {
		t.Error("bearerRole: expected false for Basic auth scheme")
	}
}

func TestBearerRole_UnknownToken(t *testing.T) {
	cfg := testSetup(t)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	_, ok := bearerRole(cfg, req)
	if ok {
		t.Error("bearerRole: expected false for unknown token")
	}
}

func TestBearerRole_KnownToken(t *testing.T) {
	cfg := testSetup(t)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	token := os.Getenv("TEST_ONBOARD_TOKEN_MAINTAINER")
	req.Header.Set("Authorization", "Bearer "+token)
	role, ok := bearerRole(cfg, req)
	if !ok {
		t.Fatal("bearerRole: expected true for known token")
	}
	if role.Name != "maintainer" {
		t.Errorf("role.Name = %q, want maintainer", role.Name)
	}
}

// TestToolsNeverReachWithBadAuth verifies that with a bad token the handler
// returns 401 before any tool could run. We instrument the inner handler to
// record whether it was called.
func TestToolsNeverReachWithBadAuth(t *testing.T) {
	cfg := testSetup(t)

	for _, tc := range []struct {
		name  string
		token string
	}{
		{"no-token", ""},
		{"wrong-token", "wrong-token-value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reached := false
			inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				reached = true
				w.WriteHeader(http.StatusOK)
			})
			handler := authMiddleware(cfg, inner)

			req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("[%s] status = %d, want 401", tc.name, rec.Code)
			}
			if reached {
				t.Errorf("[%s] inner handler should not have been called with bad token", tc.name)
			}
		})
	}
}

// TestToolAnnotations verifies that every registered tool has the expected
// read-only, non-destructive, idempotent, closed-world annotations.
func TestToolAnnotations(t *testing.T) {
	cfg := testSetup(t)
	role, ok := cfg.RoleForToken(os.Getenv("TEST_ONBOARD_TOKEN_MAINTAINER"))
	if !ok {
		t.Fatal("could not resolve maintainer role")
	}

	ctx := context.Background()
	srv := buildServer(cfg, role, nil)

	ct, st := mcp.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer ss.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer cs.Close()

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	wantNames := []string{
		"whoami", "map_system", "search", "get",
		"trace_feature", "why", "setup_guide", "starter_tasks", "withheld",
	}

	// Index returned tools by name for easy lookup.
	byName := make(map[string]*mcp.Tool, len(res.Tools))
	for _, tool := range res.Tools {
		byName[tool.Name] = tool
	}

	// Verify every expected tool is present and has the right annotations.
	for _, name := range wantNames {
		tool, found := byName[name]
		if !found {
			t.Errorf("tool %q not found in ListTools result", name)
			continue
		}
		ann := tool.Annotations
		if ann == nil {
			t.Errorf("tool %q: Annotations is nil", name)
			continue
		}
		if !ann.ReadOnlyHint {
			t.Errorf("tool %q: ReadOnlyHint = false, want true", name)
		}
		if ann.DestructiveHint == nil || *ann.DestructiveHint != false {
			t.Errorf("tool %q: DestructiveHint = %v, want *false", name, ann.DestructiveHint)
		}
		if !ann.IdempotentHint {
			t.Errorf("tool %q: IdempotentHint = false, want true", name)
		}
		if ann.OpenWorldHint == nil || *ann.OpenWorldHint != false {
			t.Errorf("tool %q: OpenWorldHint = %v, want *false", name, ann.OpenWorldHint)
		}
	}

	// Fail if the server advertises tools we did not account for.
	for _, tool := range res.Tools {
		found := false
		for _, name := range wantNames {
			if tool.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unexpected tool %q in ListTools result; add it to wantNames", tool.Name)
		}
	}
}
