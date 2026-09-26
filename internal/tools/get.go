package tools

import (
	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/filter"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// GetResult is the response from the get tool.
// Item is set when the role can see the item.
// When Item is nil the item was either not found or hidden; both cases are
// indistinguishable to the caller (not found for your role).
// WithheldCounts is always an empty map in this tool; use the withheld tool
// for per-scope counts across the whole corpus.
type GetResult struct {
	// Item is the found item, or nil when not found or denied.
	Item           *corpus.Item   `json:"item,omitempty"`
	WithheldCounts map[string]int `json:"withheld_counts"`
}

// Get returns the single corpus item with the given ID if the role can see it.
// If the item is hidden or does not exist, both cases return the same empty
// result so that the caller cannot distinguish between the two.
func Get(role roles.Role, items []corpus.Item, id string) GetResult {
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
		// Denied: return the same shape as not found.
		return GetResult{
			WithheldCounts: map[string]int{},
		}
	}
	// Not found.
	return GetResult{
		WithheldCounts: map[string]int{},
	}
}
