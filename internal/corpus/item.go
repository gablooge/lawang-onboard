// Package corpus loads the *.jsonl files from the corpus directory and assigns
// scopes to each item at load time. After Load returns, every Item is read-only.
package corpus

// Item is one unit of corpus content. Scopes is populated once at load time and
// never changed. Text is data: it is never executed or interpreted.
type Item struct {
	ID     string   // "commit:04459d8", "issue:10", etc.
	Kind   string   // file | doc | adr | commit | pr | review | issue | issue_comment
	Title  string
	Paths  []string // may be empty for issues and issue comments
	Text   string
	Scopes []string // assigned at load time, never changed

	// Fields used by specific items.
	Labels []string
	Links  []string
	Date   string
	Issue  int // for issue_comment: parent issue number
	State  string // for issues and PRs: "open", "closed", "merged"
}
