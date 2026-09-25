// Package tools contains one file per MCP tool. Each tool takes a roles.Role
// and a corpus, filters items through filter.Visible, and returns a structured
// response that always includes a withheld summary for that call.
package tools

import (
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// WhoamiResult is the response from the whoami tool.
type WhoamiResult struct {
	Role     string         `json:"role"`
	Scopes   []string       `json:"scopes"`
	Withheld map[string]int `json:"withheld"`
}

// Whoami returns the role name and scope list for the caller. No corpus items
// are accessed, so the withheld map is always empty.
func Whoami(role roles.Role) WhoamiResult {
	scopes := role.Scopes
	if scopes == nil {
		scopes = []string{}
	}
	return WhoamiResult{
		Role:     role.Name,
		Scopes:   scopes,
		Withheld: map[string]int{},
	}
}
