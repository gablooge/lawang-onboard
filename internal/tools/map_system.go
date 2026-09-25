package tools

import (
	"strings"

	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/filter"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// MapSystemResult is the response from the map_system tool.
type MapSystemResult struct {
	// Nodes is the list of visible file/doc/adr entries.
	Nodes    []MapNode      `json:"nodes"`
	Withheld map[string]int `json:"withheld"`
}

// MapNode is a single entry in the system map.
type MapNode struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Path string `json:"path"`
}

// MapSystem returns all files, docs, and ADRs visible to role with paths
// limited to the given depth. Depth 0 means no limit; depth N trims paths
// to N segments and deduplicates. Items without paths are skipped.
func MapSystem(role roles.Role, items []corpus.Item, depth int) MapSystemResult {
	w := filter.Withheld{}
	seen := make(map[string]bool)
	var nodes []MapNode

	for _, item := range items {
		if item.Kind != "file" && item.Kind != "doc" && item.Kind != "adr" {
			continue
		}
		if !filter.Visible(role, item) {
			w.Record(item)
			continue
		}
		// Use the first path of the item (items with multiple paths are rare
		// for file/doc/adr kinds).
		path := ""
		if len(item.Paths) > 0 {
			path = item.Paths[0]
		}
		if path == "" {
			continue
		}
		if depth > 0 {
			path = trimDepth(path, depth)
		}
		key := item.Kind + ":" + path
		if seen[key] {
			continue
		}
		seen[key] = true
		nodes = append(nodes, MapNode{
			ID:   item.ID,
			Kind: item.Kind,
			Path: path,
		})
	}

	if nodes == nil {
		nodes = []MapNode{}
	}
	return MapSystemResult{Nodes: nodes, Withheld: w.Summary()}
}

// trimDepth returns the first n path segments of p (slash-separated).
// If p has fewer segments, it is returned unchanged.
func trimDepth(p string, n int) string {
	parts := strings.Split(p, "/")
	if len(parts) <= n {
		return p
	}
	return strings.Join(parts[:n], "/")
}
