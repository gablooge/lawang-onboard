package tools

import (
	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/filter"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// GetResult is the response from the get tool.
// If the item exists but the role cannot see it, Withheld is true and
// WithheldScopes records which scopes caused the denial.
type GetResult struct {
	// Item is the found item, or nil when denied or not found.
	Item     *corpus.Item   `json:"item,omitempty"`
	Withheld bool           `json:"withheld,omitempty"`
	// WithheldCounts is the per-scope count of items withheld in this call (0 or 1).
	WithheldCounts map[string]int `json:"withheld_counts"`
}

// Get returns the single corpus item with the given ID if the role can see it.
// If the item exists but is denied, it returns {withheld: true}.
// If the item does not exist at all, it returns an empty result.
// The withheld_counts field always reflects the scope(s) that caused any denial.
func Get(role roles.Role, items []corpus.Item, id string) GetResult {
	w := filter.Withheld{}
	for _, item := range items {
		if item.ID != id {
			continue
		}
		if filter.Visible(role, item) {
			return GetResult{
				Item:           &item,
				WithheldCounts: map[string]int{},
			}
		}
		// Denied: record which scopes withheld it.
		w.Record(role, item)
		return GetResult{
			Withheld:       true,
			WithheldCounts: w.Summary(),
		}
	}
	// Not found.
	return GetResult{
		WithheldCounts: map[string]int{},
	}
}
