// Package web provides the /api HTTP handlers for the demo page.
// By default the role for every /api/* endpoint (except /api/roles) is
// resolved from an "Authorization: Bearer <token>" header, using the same
// constant-time token lookup as /mcp.  Missing, unknown, or ambiguous token
// returns 401 {"error":"unauthorized"}.
//
// When the server is started with -demo-roles (DemoRoles: true) the role may
// instead come from the "role" query parameter; this is only for local
// demos and must never be enabled in compose.yaml.
package web

import (
	"embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gablooge/lawang-onboard/internal/audit"
	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/roles"
	"github.com/gablooge/lawang-onboard/internal/tools"
)

//go:embed index.html
var staticFiles embed.FS

// Options controls optional server behaviour.
type Options struct {
	// DemoRoles, when true, allows the role to be specified via the "role"
	// query parameter instead of an Authorization header.  Never set this in
	// production.
	DemoRoles bool
}

// Handler returns an http.Handler that serves:
//
//	GET /             -> index.html
//	GET /api/roles    -> list of role names (always public)
//	GET /api/tour     -> map_system result for the authenticated role
//	GET /api/withheld -> withheld result for the authenticated role
//	GET /api/search   -> search result for the authenticated role &q=
//	GET /api/audit    -> audit ring entries for the authenticated role
//
// /mcp is NOT mounted here; the caller mounts it separately.
func Handler(cfg *roles.Config, items []corpus.Item, opts Options) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := staticFiles.ReadFile("index.html")
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	}))

	// /api/roles is always public: only role names are returned, no item data.
	mux.Handle("/api/roles", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		names := make([]string, 0, len(cfg.Roles))
		for name := range cfg.Roles {
			names = append(names, name)
		}
		writeJSON(w, names)
	}))

	mux.Handle("/api/tour", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := resolveRole(cfg, w, r, opts)
		if !ok {
			return
		}
		result := tools.MapSystem(role, items, 0)
		ids := make([]string, len(result.Nodes))
		for i, n := range result.Nodes {
			ids[i] = n.ID
		}
		audit.Log(role.Name, "tour", ids, result.Withheld)
		writeJSON(w, result)
	}))

	mux.Handle("/api/withheld", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := resolveRole(cfg, w, r, opts)
		if !ok {
			return
		}
		result := tools.Withheld(role, items)
		audit.Log(role.Name, "withheld", []string{}, result.ByScope)
		writeJSON(w, result)
	}))

	mux.Handle("/api/search", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := resolveRole(cfg, w, r, opts)
		if !ok {
			return
		}
		q := r.URL.Query().Get("q")
		result := tools.Search(role, items, q, 20)
		ids := make([]string, len(result.Items))
		for i, hit := range result.Items {
			ids[i] = hit.ID
		}
		audit.Log(role.Name, "search", ids, result.Withheld)
		writeJSON(w, result)
	}))

	mux.Handle("/api/audit", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := resolveRole(cfg, w, r, opts)
		if !ok {
			return
		}
		all := audit.Global.Entries()
		filtered := make([]audit.Record, 0, len(all))
		for _, e := range all {
			if e.Role == role.Name {
				filtered = append(filtered, e)
			}
		}
		writeJSON(w, filtered)
	}))

	return mux
}

// resolveRole determines the role for a request.
// When opts.DemoRoles is true, the role may come from the "role" query
// parameter (returns 400 for missing/unknown role name).
// Otherwise (the default) the role must come from "Authorization: Bearer <token>"
// (returns 401 for missing, unknown, or ambiguous token).
func resolveRole(cfg *roles.Config, w http.ResponseWriter, r *http.Request, opts Options) (roles.Role, bool) {
	if opts.DemoRoles {
		return roleParam(cfg, w, r)
	}
	return bearerRole(cfg, w, r)
}

// bearerRole resolves the role from the Authorization header.
// Writes 401 and returns false on any failure.
func bearerRole(cfg *roles.Config, w http.ResponseWriter, r *http.Request) (roles.Role, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		writeUnauthorized(w)
		return roles.Role{}, false
	}
	token := auth[len("Bearer "):]
	role, ok := cfg.RoleForToken(token)
	if !ok {
		writeUnauthorized(w)
		return roles.Role{}, false
	}
	return role, true
}

// roleParam extracts and validates the "role" query parameter (demo-roles mode).
// On success it returns the Role and true.
// On failure it writes a 400 and returns false.
func roleParam(cfg *roles.Config, w http.ResponseWriter, r *http.Request) (roles.Role, bool) {
	name := r.URL.Query().Get("role")
	if name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "role parameter required"})
		return roles.Role{}, false
	}
	role, ok := cfg.Roles[name]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unknown role"})
		return roles.Role{}, false
	}
	return role, true
}

// writeUnauthorized sends a 401 with a JSON body.
func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}

// writeJSON encodes v as JSON and writes it to w with Content-Type application/json.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
