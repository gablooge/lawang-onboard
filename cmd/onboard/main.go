// Command onboard is the Lawang Onboard MCP server.
// It serves permission-filtered context about the Lawang codebase over the
// Model Context Protocol.
//
// For stdio transport (-stdio flag) the bearer token comes from the
// ONBOARD_TOKEN environment variable; for HTTP it comes from the
// Authorization: Bearer header on each request.
//
// Tokens are never logged or returned.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/roles"
	"github.com/gablooge/lawang-onboard/internal/tools"
	"github.com/gablooge/lawang-onboard/internal/web"
)

func main() {
	stdioFlag := flag.Bool("stdio", false, "run as stdio MCP server")
	addrFlag := flag.String("addr", ":8080", "HTTP listen address")
	corpusDir := flag.String("corpus", "corpus", "corpus directory")
	rolesFile := flag.String("roles", "roles.yaml", "roles YAML file")
	webOnly := flag.Bool("web-only", false, "serve / and /api without /mcp")
	demoRoles := flag.Bool("demo-roles", false, "allow ?role= param on /api/* instead of requiring a Bearer token (local demo only, never enable in production)")
	flag.Parse()

	cfg, err := roles.Load(*rolesFile)
	if err != nil {
		log.Fatalf("load roles: %v", err)
	}

	if *demoRoles {
		log.Println("WARNING: -demo-roles is on; /api/* accepts ?role= without a token. Do not use in production.")
	}

	items, err := corpus.Load(*corpusDir, cfg)
	if err != nil {
		log.Fatalf("load corpus: %v", err)
	}

	if *stdioFlag {
		runStdio(cfg, items)
		return
	}
	runHTTP(*addrFlag, *webOnly, *demoRoles, cfg, items)
}

// runStdio resolves the bearer token from ONBOARD_TOKEN once at startup and
// exits with a named error if it is missing or unknown.
func runStdio(cfg *roles.Config, items []corpus.Item) {
	const envVar = "ONBOARD_TOKEN"
	token := os.Getenv(envVar)
	if token == "" {
		fmt.Fprintf(os.Stderr, "onboard: %s is not set\n", envVar)
		os.Exit(1)
	}
	role, ok := cfg.RoleForToken(token)
	if !ok {
		fmt.Fprintf(os.Stderr, "onboard: %s is set but does not match any known role\n", envVar)
		os.Exit(1)
	}

	srv := buildServer(cfg, role, items)
	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("stdio server: %v", err)
	}
}

// runHTTP starts the HTTP server. It always mounts the web UI (/ and /api).
// Unless webOnly is true it also mounts /mcp with bearer-token auth.
func runHTTP(addr string, webOnly bool, demoRoles bool, cfg *roles.Config, items []corpus.Item) {
	// Mount the web UI on a fresh mux so we can return it for testing.
	mux := buildMux(webOnly, demoRoles, cfg, items)

	log.Printf("onboard: listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

// buildMux constructs the http.ServeMux for the server. It is a separate
// function so that tests can call it without starting a listener.
func buildMux(webOnly bool, demoRoles bool, cfg *roles.Config, items []corpus.Item) *http.ServeMux {
	mux := http.NewServeMux()

	// Mount the web UI (/, /api/...).
	webHandler := web.Handler(cfg, items, web.Options{DemoRoles: demoRoles})
	mux.Handle("/", webHandler)
	mux.Handle("/api/", webHandler)

	if !webOnly {
		mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
			role, ok := bearerRole(cfg, r)
			if !ok {
				return nil // causes 400; the outer wrapper handles 401
			}
			return buildServer(cfg, role, items)
		}, &mcp.StreamableHTTPOptions{Stateless: true})
		// Wrap handler to reject missing/unknown tokens with 401 before MCP sees it.
		mux.Handle("/mcp", authMiddleware(cfg, mcpHandler))
	}

	return mux
}

// authMiddleware rejects requests whose Authorization header is missing or
// whose bearer token does not match a known role with 401.
func authMiddleware(cfg *roles.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := bearerRole(cfg, r); !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// bearerRole extracts the bearer token from the Authorization header and
// resolves it to a role. Returns false if the token is absent or unknown.
// The token value is never logged or stored.
func bearerRole(cfg *roles.Config, r *http.Request) (roles.Role, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return roles.Role{}, false
	}
	token := auth[len("Bearer "):]
	return cfg.RoleForToken(token)
}

// buildServer constructs an MCP server for the given role with all tools wired.
func buildServer(cfg *roles.Config, role roles.Role, items []corpus.Item) *mcp.Server {
	impl := &mcp.Implementation{
		Name:    "lawang-onboard",
		Version: "0.1.0",
	}
	srv := mcp.NewServer(impl, nil)
	registerTools(srv, cfg, role, items)
	return srv
}

// --- Tool argument types ---

type getArgs struct {
	ID string `json:"id" mcp:"item ID to retrieve (e.g. file:internal/ingress/ingress.go)"`
}

type searchArgs struct {
	Query      string `json:"query" mcp:"search query"`
	MaxResults int    `json:"max_results,omitempty" mcp:"maximum results to return (default 20)"`
}

type mapSystemArgs struct {
	Depth int `json:"depth,omitempty" mcp:"path depth limit: 0 for no limit"`
}

type traceFeatureArgs struct {
	Term string `json:"term" mcp:"term to trace across files, commits, reviews, and ADRs"`
}

type whyArgs struct {
	Path string `json:"path" mcp:"repository path to explain (e.g. internal/ingress/ingress.go)"`
}

type starterTasksArgs struct{}

// readOnlyAnnotations are the annotations applied to every tool: all tools
// are read-only, non-destructive, idempotent, and operate on a closed local
// corpus with no external side effects.
var readOnlyAnnotations = &mcp.ToolAnnotations{
	ReadOnlyHint:    true,
	DestructiveHint: boolPtr(false),
	IdempotentHint:  true,
	OpenWorldHint:   boolPtr(false),
}

// boolPtr returns a pointer to b, for use with *bool annotation fields.
func boolPtr(b bool) *bool { return &b }

// registerTools adds all tools to srv bound to the given role and items.
func registerTools(srv *mcp.Server, cfg *roles.Config, role roles.Role, items []corpus.Item) {
	// whoami
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "whoami",
		Description: "Return the caller's role name and scope list.",
		Annotations: readOnlyAnnotations,
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, tools.WhoamiResult, error) {
		result := tools.Whoami(role)
		return nil, result, nil
	})

	// get
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get",
		Description: "Retrieve a single corpus item by its ID. Returns an empty result if the item is not found or not visible to the caller; both cases are indistinguishable.",
		Annotations: readOnlyAnnotations,
	}, func(_ context.Context, _ *mcp.CallToolRequest, args getArgs) (*mcp.CallToolResult, tools.GetResult, error) {
		result := tools.Get(role, items, args.ID)
		return nil, result, nil
	})

	// search
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search",
		Description: "Search corpus items visible to the caller using BM25 ranking. Hidden items never affect results or scores.",
		Annotations: readOnlyAnnotations,
	}, func(_ context.Context, _ *mcp.CallToolRequest, args searchArgs) (*mcp.CallToolResult, tools.SearchResult, error) {
		result := tools.Search(role, items, args.Query, args.MaxResults)
		return nil, result, nil
	})

	// map_system
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "map_system",
		Description: "List all files, docs, and ADRs visible to the caller, optionally limited to a path depth.",
		Annotations: readOnlyAnnotations,
	}, func(_ context.Context, _ *mcp.CallToolRequest, args mapSystemArgs) (*mcp.CallToolResult, tools.MapSystemResult, error) {
		result := tools.MapSystem(role, items, args.Depth)
		return nil, result, nil
	})

	// withheld
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "withheld",
		Description: "Return the number of hidden items per missing scope plus a total of hidden items, across the whole corpus.",
		Annotations: readOnlyAnnotations,
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, tools.WithheldResult, error) {
		result := tools.Withheld(role, items)
		return nil, result, nil
	})

	// trace_feature
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "trace_feature",
		Description: "Find files, commits, reviews, and ADRs related to a term, grouped by kind.",
		Annotations: readOnlyAnnotations,
	}, func(_ context.Context, _ *mcp.CallToolRequest, args traceFeatureArgs) (*mcp.CallToolResult, tools.TraceFeatureResult, error) {
		result := tools.TraceFeature(role, items, args.Term)
		return nil, result, nil
	})

	// why
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "why",
		Description: "For a repository path, return commits that touched it, review comments on it, and ADRs or docs that mention it.",
		Annotations: readOnlyAnnotations,
	}, func(_ context.Context, _ *mcp.CallToolRequest, args whyArgs) (*mcp.CallToolResult, tools.WhyResult, error) {
		result := tools.Why(role, items, args.Path)
		return nil, result, nil
	})

	// setup_guide
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "setup_guide",
		Description: "Return the visible parts of README.md, Makefile, and .github/workflows files.",
		Annotations: readOnlyAnnotations,
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, tools.SetupGuideResult, error) {
		result := tools.SetupGuide(role, items)
		return nil, result, nil
	})

	// starter_tasks
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "starter_tasks",
		Description: "Return open issues whose label areas are within the caller's scopes.",
		Annotations: readOnlyAnnotations,
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ starterTasksArgs) (*mcp.CallToolResult, tools.StarterTasksResult, error) {
		result := tools.StarterTasks(role, items, cfg.LabelAreas)
		return nil, result, nil
	})
}

