package tools

import (
	"testing"

	"github.com/gablooge/lawang-onboard/internal/corpus"
)

// starterItems returns a corpus with issues suited for starter_tasks tests.
func starterItems() []corpus.Item {
	return []corpus.Item{
		// Open issue with label "provider" -> area scope path:internal/provider
		{
			ID:     "issue:10",
			Kind:   "issue",
			Title:  "Add provider health check",
			Labels: []string{"provider"},
			State:  "open",
			Scopes: []string{"backlog:public"},
		},
		// Open issue with label "backlog" (no area mapping -> not a starter task)
		{
			ID:     "issue:11",
			Kind:   "issue",
			Title:  "General backlog item",
			Labels: []string{"backlog"},
			State:  "open",
			Scopes: []string{"backlog:public"},
		},
		// Closed issue with label "provider" (excluded because not open)
		{
			ID:     "issue:12",
			Kind:   "issue",
			Title:  "Closed provider issue",
			Labels: []string{"provider"},
			State:  "closed",
			Scopes: []string{"backlog:public"},
		},
		// Open issue with label "growth" (only maintainer)
		{
			ID:     "issue:13",
			Kind:   "issue",
			Title:  "Growth campaign issue",
			Labels: []string{"growth"},
			State:  "open",
			Scopes: []string{"private:growth"},
		},
		// item with no scopes
		{
			ID:     "issue:noscope",
			Kind:   "issue",
			Title:  "No scope issue",
			Labels: []string{"provider"},
			State:  "open",
			Scopes: nil,
		},
	}
}

func TestStarterTasksMaintainer(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["maintainer"]
	result := StarterTasks(role, starterItems(), cfg.LabelAreas)

	// Maintainer sees issue:10 (provider area, open).
	found := false
	for _, task := range result.Tasks {
		if task.ID == "issue:10" {
			found = true
		}
	}
	if !found {
		t.Error("maintainer: issue:10 (provider) should be a starter task")
	}

	// issue:12 is closed, must not appear.
	for _, task := range result.Tasks {
		if task.ID == "issue:12" {
			t.Error("maintainer: closed issue:12 must not be a starter task")
		}
	}
}

func TestStarterTasksContractor(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["contractor"]
	result := StarterTasks(role, starterItems(), cfg.LabelAreas)

	// Contractor holds path:internal/provider, so issue:10 with label "provider" is a starter task.
	found := false
	for _, task := range result.Tasks {
		if task.ID == "issue:10" {
			found = true
		}
	}
	if !found {
		t.Error("contractor: issue:10 (provider) should be a starter task")
	}

	// issue:13 has label "growth", which resolves to private:growth, contractor does not hold it.
	for _, task := range result.Tasks {
		if task.ID == "issue:13" {
			t.Error("contractor: growth issue:13 must not appear (private:growth not in contractor scopes)")
		}
	}

	// closed issue must not appear
	for _, task := range result.Tasks {
		if task.ID == "issue:12" {
			t.Error("contractor: closed issue:12 must not be a starter task")
		}
	}
}

func TestStarterTasksEmployee(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["employee"]
	result := StarterTasks(role, starterItems(), cfg.LabelAreas)

	// Employee has path:* which covers path:internal/provider.
	found := false
	for _, task := range result.Tasks {
		if task.ID == "issue:10" {
			found = true
		}
	}
	if !found {
		t.Error("employee: issue:10 should be a starter task (path:* covers path:internal/provider)")
	}

	// Employee does not hold private:growth.
	for _, task := range result.Tasks {
		if task.ID == "issue:13" {
			t.Error("employee: private growth issue:13 must not appear")
		}
	}
}

func TestStarterTasksNoAreaLabelNotIncluded(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["contractor"]
	result := StarterTasks(role, starterItems(), cfg.LabelAreas)

	// issue:11 has label "backlog" which has no label_areas mapping -> not a starter task.
	for _, task := range result.Tasks {
		if task.ID == "issue:11" {
			t.Error("contractor: issue:11 (backlog label, no area mapping) must not be a starter task")
		}
	}
}

func TestStarterTasksWithheldSummary(t *testing.T) {
	cfg := testCfg(t)
	role := cfg.Roles["contractor"]
	result := StarterTasks(role, starterItems(), cfg.LabelAreas)
	if result.Withheld == nil {
		t.Error("withheld map must not be nil")
	}
}
