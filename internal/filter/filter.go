// Package filter decides whether a corpus item is visible to a role.
// It is pure: no I/O, no state, no access to tokens or request context.
package filter

import (
	"strings"

	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// Visible returns true when role is permitted to see item.
//
// Rules applied in order:
//  1. item.Scopes is empty: deny (record "(no-scope)").
//  2. item.Scopes contains "(unmapped)": deny (record "(unmapped)"); maintainer's
//     "*" shortcut never fires because this check precedes it.
//  3. role.Scopes contains "*": allow (maintainer shortcut).
//  4. For each scope in item.Scopes: the role must hold it exactly, or hold a
//     wildcard "prefix:*" that covers it. If any scope is uncovered: deny.
//  5. Allow.
func Visible(role roles.Role, item corpus.Item) bool {
	// Rule 1: no scopes at all.
	if len(item.Scopes) == 0 {
		return false
	}

	// Rule 2: contains "(unmapped)" - denied for every role including maintainer.
	for _, s := range item.Scopes {
		if s == "(unmapped)" {
			return false
		}
	}

	// Rule 3: maintainer shortcut ("*" covers everything that is not denied above).
	for _, s := range role.Scopes {
		if s == "*" {
			return true
		}
	}

	// Rule 4: every item scope must be covered by the role.
	for _, need := range item.Scopes {
		if !roleCoversScope(role.Scopes, need) {
			return false
		}
	}

	// Rule 5: allow.
	return true
}

// roleCoversScope reports whether the role's scope list covers the required scope.
// Coverage means either an exact match or a "prefix:*" wildcard match.
func roleCoversScope(roleScopes []string, need string) bool {
	for _, have := range roleScopes {
		if have == need {
			return true
		}
		// Wildcard: "prefix:*" covers any scope that starts with "prefix:".
		if strings.HasSuffix(have, ":*") {
			prefix := have[:len(have)-1] // "prefix:" (keep the colon, drop the *)
			if strings.HasPrefix(need, prefix) {
				return true
			}
		}
	}
	return false
}

// Withheld accumulates counts of denied scope names across a batch of filter decisions.
// The key is the scope name that caused the denial (or the sentinel "(no-scope)").
// The value is never a token, item title, or any user-visible text.
type Withheld map[string]int

// Record increments the withheld counts for item. It must only be called when
// Visible returned false for the same item.
func (w Withheld) Record(item corpus.Item) {
	if len(item.Scopes) == 0 {
		w["(no-scope)"]++
		return
	}
	for _, s := range item.Scopes {
		w[s]++
	}
}

// Summary returns a copy of the withheld map safe for returning to callers.
func (w Withheld) Summary() map[string]int {
	out := make(map[string]int, len(w))
	for k, v := range w {
		out[k] = v
	}
	return out
}
