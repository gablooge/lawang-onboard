package tools

import (
	"strings"

	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/filter"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// SetupGuideResult is the response from the setup_guide tool.
type SetupGuideResult struct {
	Items    []SetupItem    `json:"items"`
	Withheld map[string]int `json:"withheld"`
}

// SetupItem is one entry from setup_guide.
type SetupItem struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

// setupPaths is the set of paths that setup_guide considers.
var setupPaths = map[string]bool{
	"readme.md": true,
	"makefile":  true,
}

// SetupGuide returns the visible portions of README.md, Makefile, and
// .github/workflows files. The withheld map counts scope denials for this call.
func SetupGuide(role roles.Role, items []corpus.Item) SetupGuideResult {
	w := filter.Withheld{}
	result := SetupGuideResult{
		Items:    []SetupItem{},
		Withheld: map[string]int{},
	}

	for _, item := range items {
		if item.Kind != "file" && item.Kind != "doc" {
			continue
		}
		if !isSetupItem(item) {
			continue
		}
		if !filter.Visible(role, item) {
			w.Record(role, item)
			continue
		}
		result.Items = append(result.Items, SetupItem{
			ID:    item.ID,
			Kind:  item.Kind,
			Title: item.Title,
			Text:  item.Text,
		})
	}

	result.Withheld = w.Summary()
	return result
}

// isSetupItem reports whether item is a README, Makefile, or .github/workflows file.
func isSetupItem(item corpus.Item) bool {
	for _, p := range item.Paths {
		lower := strings.ToLower(p)
		if setupPaths[lower] {
			return true
		}
		if strings.HasPrefix(lower, ".github/workflows/") {
			return true
		}
	}
	return false
}
