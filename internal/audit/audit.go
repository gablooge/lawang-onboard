// Package audit records every tool call in an in-memory ring buffer (last 200
// entries). Both the MCP tools and the /api handlers write here. No tokens are
// ever stored; only role names are recorded.
package audit

import (
	"sync"
	"time"
)

const ringSize = 200

// Record is one audit log entry.
type Record struct {
	Time         time.Time      `json:"time"`
	Role         string         `json:"role"`
	Tool         string         `json:"tool"`
	ReturnedIDs  []string       `json:"returned_ids"`
	WithheldCounts map[string]int `json:"withheld_counts"`
}

// Ring is a fixed-capacity in-memory ring buffer. It is safe for concurrent use.
type Ring struct {
	mu   sync.Mutex
	buf  [ringSize]Record
	head int // index of the next write slot
	n    int // number of valid entries (0..ringSize)
}

// Append adds a record to the ring, overwriting the oldest entry when full.
func (r *Ring) Append(rec Record) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.head] = rec
	r.head = (r.head + 1) % ringSize
	if r.n < ringSize {
		r.n++
	}
}

// Entries returns a snapshot of the records in chronological order (oldest first).
func (r *Ring) Entries() []Record {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.n == 0 {
		return []Record{}
	}
	out := make([]Record, r.n)
	// When the buffer is not yet full the oldest entry is at index 0.
	// When the buffer is full the oldest entry is at r.head (the slot about to be overwritten).
	start := 0
	if r.n == ringSize {
		start = r.head
	}
	for i := 0; i < r.n; i++ {
		out[i] = r.buf[(start+i)%ringSize]
	}
	return out
}

// Global is the process-wide audit ring used by MCP tools and /api handlers.
var Global = &Ring{}

// Log appends a record to Global.
func Log(role, tool string, returnedIDs []string, withheldCounts map[string]int) {
	if returnedIDs == nil {
		returnedIDs = []string{}
	}
	if withheldCounts == nil {
		withheldCounts = map[string]int{}
	}
	Global.Append(Record{
		Time:           time.Now().UTC(),
		Role:           role,
		Tool:           tool,
		ReturnedIDs:    returnedIDs,
		WithheldCounts: withheldCounts,
	})
}
