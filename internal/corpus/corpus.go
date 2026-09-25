package corpus

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gablooge/lawang-onboard/internal/roles"
)

// rawItem mirrors the JSON shape of a corpus record.
type rawItem struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Title    string   `json:"title"`
	Paths    []string `json:"paths"`
	Text     string   `json:"text"`
	Labels   []string `json:"labels"`
	Links    []string `json:"links"`
	Date     string   `json:"date"`
	Issue    int      `json:"issue"`
	State    string   `json:"state"`
}

// Load reads all *.jsonl files from dir and returns the corpus with Scopes populated.
// The cfg is used to assign scopes at load time. Issue comments inherit their parent
// issue's labels for scope assignment.
func Load(dir string, cfg *roles.Config) ([]Item, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("corpus: glob %s: %w", dir, err)
	}

	// First pass: collect all items.
	var raws []rawItem
	for _, path := range matches {
		got, err := readJSONL(path)
		if err != nil {
			return nil, err
		}
		raws = append(raws, got...)
	}

	// Build an index of issue labels by issue number so that issue comments can
	// inherit them.
	issueLabels := make(map[int][]string)
	for _, r := range raws {
		if r.Kind == "issue" && r.Issue == 0 {
			// Extract the issue number from the ID "issue:NN".
			var n int
			if _, err := fmt.Sscanf(r.ID, "issue:%d", &n); err == nil && n > 0 {
				issueLabels[n] = r.Labels
			}
		}
	}

	items := make([]Item, 0, len(raws))
	for _, r := range raws {
		item := Item{
			ID:     r.ID,
			Kind:   r.Kind,
			Title:  r.Title,
			Paths:  r.Paths,
			Text:   r.Text,
			Labels: r.Labels,
			Links:  r.Links,
			Date:   r.Date,
			Issue:  r.Issue,
			State:  r.State,
		}

		// Assign scopes.
		item.Scopes = assignScopes(item, issueLabels, cfg)
		items = append(items, item)
	}

	return items, nil
}

// readJSONL reads one *.jsonl file and returns all items parsed from it.
func readJSONL(path string) ([]rawItem, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("corpus: open %s: %w", path, err)
	}
	defer f.Close()

	var items []rawItem
	sc := bufio.NewScanner(f)
	// Some corpus items have very long text fields.
	sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" {
			continue
		}
		var r rawItem
		if err := json.Unmarshal([]byte(text), &r); err != nil {
			return nil, fmt.Errorf("corpus: %s line %d: %w", path, line, err)
		}
		items = append(items, r)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("corpus: read %s: %w", path, err)
	}
	return items, nil
}

// assignScopes returns the scopes for item according to the rules in cfg.
// The result is a deduplicated, non-nil slice. An empty result means the item
// has no scopes and will be denied by filter.Visible.
func assignScopes(item Item, issueLabels map[int][]string, cfg *roles.Config) []string {
	switch item.Kind {
	case "review":
		// A review comment uses only its own paths, one scope per path via the path rule.
		// It does not inherit the PR's full path list.
		return pathsToScopes(item.Paths, cfg.ScopeRules)

	case "issue_comment":
		// An issue comment carries its parent issue's labels, not its own paths.
		parentLabels := issueLabels[item.Issue]
		return labelOrKindScope(parentLabels, item.Kind, cfg)

	case "issue":
		return labelOrKindScope(item.Labels, item.Kind, cfg)
	}

	// For items with paths (file, doc, adr, commit, pr): union of path scopes.
	if len(item.Paths) > 0 {
		return pathsToScopes(item.Paths, cfg.ScopeRules)
	}

	// Fallback: kind scope.
	return labelOrKindScope(item.Labels, item.Kind, cfg)
}

// pathsToScopes returns the deduplicated union of path scopes for all paths.
func pathsToScopes(paths []string, rules []roles.ScopeRule) []string {
	if len(paths) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	var out []string
	for _, p := range paths {
		scope := roles.PathScope(rules, p)
		if !seen[scope] {
			seen[scope] = true
			out = append(out, scope)
		}
	}
	return out
}

// labelOrKindScope returns the scope for an item without paths (issue, issue_comment).
// It checks labels first, then falls back to the kind scope.
func labelOrKindScope(labels []string, kind string, cfg *roles.Config) []string {
	for _, label := range labels {
		if scope, ok := cfg.LabelScopes[label]; ok {
			return []string{scope}
		}
	}
	if scope, ok := cfg.KindScopes[kind]; ok {
		return []string{scope}
	}
	return nil
}
