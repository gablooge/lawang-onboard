package tools

import (
	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/filter"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// WithheldResult is the response from the withheld tool.
type WithheldResult struct {
	// ByScope is the count of hidden items per missing scope across the whole corpus.
	ByScope map[string]int `json:"by_scope"`
	// Total is the number of corpus items the role cannot see.
	Total int `json:"total"`
	// Withheld is always empty for this tool (it reports on others).
	Withheld map[string]int `json:"withheld"`
}

// Withheld scans all corpus items for the role and returns, for the whole
// corpus, the number of hidden items per missing scope and a total of hidden
// items. It gives a session-wide summary of what the role cannot see without
// revealing any item contents or IDs.
func Withheld(role roles.Role, items []corpus.Item) WithheldResult {
	w := filter.Withheld{}
	total := 0
	for _, item := range items {
		if !filter.Visible(role, item) {
			w.Record(role, item)
			total++
		}
	}
	return WithheldResult{
		ByScope:  w.Summary(),
		Total:    total,
		Withheld: map[string]int{},
	}
}
