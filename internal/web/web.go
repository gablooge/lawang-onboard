// Package web provides the /api HTTP handlers for the demo page.
// The role comes from the "role" query parameter; it is looked up in cfg.Roles.
// No token ever reaches the browser.
package web

import (
	"embed"
	"encoding/json"
	"net/http"

	"github.com/gablooge/lawang-onboard/internal/audit"
	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/roles"
	"github.com/gablooge/lawang-onboard/internal/tools"
)

//go:embed index.html
var staticFiles embed.FS

// Handler returns an http.Handler that serves:
//
//	GET /             -> index.html
//	GET /api/roles    -> list of role names
//	GET /api/tour     -> map_system result for ?role=
//	GET /api/withheld -> withheld result for ?role=
//	GET /api/search   -> search result for ?role= &q=
//	GET /api/audit    -> audit ring entries for ?role=
//
// /mcp is NOT mounted here; the caller mounts it separately.
func Handler(cfg *roles.Config, items []corpus.Item) http.Handler {
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

	mux.Handle("/api/roles", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		names := make([]string, 0, len(cfg.Roles))
		for name := range cfg.Roles {
			names = append(names, name)
		}
		writeJSON(w, names)
	}))

	mux.Handle("/api/tour", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := roleParam(cfg, w, r)
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
		role, ok := roleParam(cfg, w, r)
		if !ok {
			return
		}
		result := tools.Withheld(role, items)
		audit.Log(role.Name, "withheld", []string{}, result.ByScope)
		writeJSON(w, result)
	}))

	mux.Handle("/api/search", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := roleParam(cfg, w, r)
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
		// role param is required for consistency and to enforce unknown-role 400,
		// but the audit log is not filtered by role (it is already role-name data).
		_, ok := roleParam(cfg, w, r)
		if !ok {
			return
		}
		writeJSON(w, audit.Global.Entries())
	}))

	return mux
}

// roleParam extracts and validates the "role" query parameter.
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

// writeJSON encodes v as JSON and writes it to w with Content-Type application/json.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
