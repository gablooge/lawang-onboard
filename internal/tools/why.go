package tools

import (
	"strings"

	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/filter"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// WhyResult is the response from the why tool.
type WhyResult struct {
	// Commits are visible commits that touched the path.
	Commits []WhyItem `json:"commits"`
	// Reviews are visible review comments on the path.
	Reviews []WhyItem `json:"reviews"`
	// ADRsDocs are visible ADRs and docs that mention the path.
	ADRsDocs []WhyItem `json:"adrs_docs"`
	Withheld map[string]int `json:"withheld"`
}

// WhyItem is one result entry for why.
type WhyItem struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
}

// Why returns, for a given path, the visible commits that touched it, review
// comments whose path matches it, and ADRs or docs that mention it.
// The withheld map counts scope denials for this call.
func Why(role roles.Role, items []corpus.Item, path string) WhyResult {
	w := filter.Withheld{}
	result := WhyResult{
		Commits:  []WhyItem{},
		Reviews:  []WhyItem{},
		ADRsDocs: []WhyItem{},
		Withheld: map[string]int{},
	}
	if path == "" {
		return result
	}

	lowerPath := strings.ToLower(path)

	for _, item := range items {
		switch item.Kind {
		case "commit", "review", "adr", "doc":
			// only these kinds
		default:
			continue
		}

		if !filter.Visible(role, item) {
			w.Record(item)
			continue
		}

		switch item.Kind {
		case "commit":
			if pathInList(item.Paths, lowerPath) {
				result.Commits = append(result.Commits, WhyItem{ID: item.ID, Kind: item.Kind, Title: item.Title})
			}
		case "review":
			if pathInList(item.Paths, lowerPath) {
				result.Reviews = append(result.Reviews, WhyItem{ID: item.ID, Kind: item.Kind, Title: item.Title})
			}
		case "adr", "doc":
			if strings.Contains(strings.ToLower(item.Text), lowerPath) ||
				strings.Contains(strings.ToLower(item.Title), lowerPath) {
				result.ADRsDocs = append(result.ADRsDocs, WhyItem{ID: item.ID, Kind: item.Kind, Title: item.Title})
			}
		}
	}

	result.Withheld = w.Summary()
	return result
}

// pathInList reports whether path is a case-insensitive prefix-match against
// any element in paths. It matches the path itself and any file nested inside it.
func pathInList(paths []string, lowerPath string) bool {
	for _, p := range paths {
		lp := strings.ToLower(p)
		if lp == lowerPath || strings.HasPrefix(lp, lowerPath+"/") {
			return true
		}
	}
	return false
}
