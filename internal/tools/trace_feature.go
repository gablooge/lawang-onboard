package tools

import (
	"strings"

	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/filter"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// TraceFeatureResult is the response from the trace_feature tool.
type TraceFeatureResult struct {
	Files    []FeatureItem  `json:"files"`
	Commits  []FeatureItem  `json:"commits"`
	Reviews  []FeatureItem  `json:"reviews"`
	ADRs     []FeatureItem  `json:"adrs"`
	Withheld map[string]int `json:"withheld"`
}

// FeatureItem is one result entry for trace_feature.
type FeatureItem struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
}

// TraceFeature returns corpus items matching term, filtered by role visibility,
// grouped and ordered: files, commits, reviews, ADRs. A hidden item never
// appears in any group. The withheld map counts scope denials for this call.
func TraceFeature(role roles.Role, items []corpus.Item, term string) TraceFeatureResult {
	w := filter.Withheld{}
	result := TraceFeatureResult{
		Files:    []FeatureItem{},
		Commits:  []FeatureItem{},
		Reviews:  []FeatureItem{},
		ADRs:     []FeatureItem{},
		Withheld: map[string]int{},
	}
	if term == "" {
		return result
	}

	lower := strings.ToLower(term)

	for _, item := range items {
		switch item.Kind {
		case "file", "commit", "review", "adr":
			// only these four kinds
		default:
			continue
		}

		if !filter.Visible(role, item) {
			w.Record(item)
			continue
		}

		if !itemMatchesTerm(item, lower) {
			continue
		}

		fi := FeatureItem{ID: item.ID, Kind: item.Kind, Title: item.Title}
		switch item.Kind {
		case "file":
			result.Files = append(result.Files, fi)
		case "commit":
			result.Commits = append(result.Commits, fi)
		case "review":
			result.Reviews = append(result.Reviews, fi)
		case "adr":
			result.ADRs = append(result.ADRs, fi)
		}
	}

	result.Withheld = w.Summary()
	return result
}

// itemMatchesTerm reports whether the item's title, text, or any path contains term (case-insensitive).
func itemMatchesTerm(item corpus.Item, lowerTerm string) bool {
	if strings.Contains(strings.ToLower(item.Title), lowerTerm) {
		return true
	}
	if strings.Contains(strings.ToLower(item.Text), lowerTerm) {
		return true
	}
	for _, p := range item.Paths {
		if strings.Contains(strings.ToLower(p), lowerTerm) {
			return true
		}
	}
	return false
}
