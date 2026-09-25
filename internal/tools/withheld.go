package tools

import (
	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/filter"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// WithheldResult is the response from the withheld tool.
type WithheldResult struct {
	// Summary is the accumulated withheld counts from the current call.
	Summary map[string]int `json:"summary"`
	// Withheld is always empty for this tool (it reports on others).
	Withheld map[string]int `json:"withheld"`
}

// Withheld scans all corpus items for the role and returns the total per-scope
// count of items denied. It gives a session-wide summary of what the role
// cannot see without revealing any item contents or IDs.
func Withheld(role roles.Role, items []corpus.Item) WithheldResult {
	w := filter.Withheld{}
	for _, item := range items {
		if !filter.Visible(role, item) {
			w.Record(item)
		}
	}
	return WithheldResult{
		Summary:  w.Summary(),
		Withheld: map[string]int{},
	}
}
