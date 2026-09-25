package tools

import (
	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/filter"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// StarterTasksResult is the response from the starter_tasks tool.
type StarterTasksResult struct {
	Tasks    []StarterTask  `json:"tasks"`
	Withheld map[string]int `json:"withheld"`
}

// StarterTask is one open issue that is a starter task for the caller's role.
type StarterTask struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Labels []string `json:"labels"`
}

// StarterTasks returns open issues whose label_areas scope the role holds.
// An issue is included when the role holds the area scope of at least one of
// the issue's labels (from label_areas in roles.yaml). The issue must also
// pass filter.Visible, and its state must be "open".
// The withheld map counts scope denials for this call.
func StarterTasks(role roles.Role, items []corpus.Item, labelAreas map[string]string) StarterTasksResult {
	w := filter.Withheld{}
	result := StarterTasksResult{
		Tasks:    []StarterTask{},
		Withheld: map[string]int{},
	}

	for _, item := range items {
		if item.Kind != "issue" {
			continue
		}
		// Only open issues.
		if item.State != "open" {
			continue
		}
		if !filter.Visible(role, item) {
			w.Record(item)
			continue
		}
		// Check whether the role holds the area scope for at least one label.
		if !roleHasAreaForLabels(role, item.Labels, labelAreas) {
			continue
		}
		labels := item.Labels
		if labels == nil {
			labels = []string{}
		}
		result.Tasks = append(result.Tasks, StarterTask{
			ID:     item.ID,
			Title:  item.Title,
			Labels: labels,
		})
	}

	result.Withheld = w.Summary()
	return result
}

// roleHasAreaForLabels reports whether the role holds the area scope for any label.
// A role with the global "*" scope is treated as covering all area scopes.
func roleHasAreaForLabels(role roles.Role, labels []string, labelAreas map[string]string) bool {
	// Global wildcard: maintainer sees every area.
	for _, s := range role.Scopes {
		if s == "*" {
			// Only return true if there is at least one area-mapped label.
			for _, label := range labels {
				if _, ok := labelAreas[label]; ok {
					return true
				}
			}
			return false
		}
	}
	for _, label := range labels {
		areaScope, ok := labelAreas[label]
		if !ok {
			continue
		}
		if filter.RoleCoversScope(role.Scopes, areaScope) {
			return true
		}
	}
	return false
}
