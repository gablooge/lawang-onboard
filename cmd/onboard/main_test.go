package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
