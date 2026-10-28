// Package onboard is the module root. It exists so that the Bob skills under .bob/ can be
// embedded into the binary from the one place they are written: an embed pattern cannot reach
// above its own package directory, and a copy under internal/ would be a second source that
// drifts from the first.
package onboard

import "embed"

// Skills holds the four skill files that Bob's Onboard mode reads, and the mode's rules. The
// server serves the same text as MCP prompts, so a client that is not Bob gets the same guidance.
//
//go:embed .bob/skills/tour/SKILL.md
//go:embed .bob/skills/trace/SKILL.md
//go:embed .bob/skills/why/SKILL.md
//go:embed .bob/skills/first-week/SKILL.md
//go:embed .bob/rules-onboard/01-rules.md
var Skills embed.FS
