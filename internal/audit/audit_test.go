package audit

import (
	"testing"
	"time"
)

func TestRing_Empty(t *testing.T) {
	r := &Ring{}
	entries := r.Entries()
	if len(entries) != 0 {
		t.Fatalf("empty ring: got %d entries, want 0", len(entries))
	}
}

func TestRing_FewEntries(t *testing.T) {
	r := &Ring{}
	for i := 0; i < 5; i++ {
		r.Append(Record{Time: time.Now(), Role: "r", Tool: "t"})
	}
	entries := r.Entries()
	if len(entries) != 5 {
		t.Fatalf("got %d entries, want 5", len(entries))
	}
}

func TestRing_CapacityExact(t *testing.T) {
	r := &Ring{}
	for i := 0; i < ringSize; i++ {
		r.Append(Record{Role: "r", Tool: "t"})
	}
	if len(r.Entries()) != ringSize {
		t.Fatalf("got %d entries, want %d", len(r.Entries()), ringSize)
	}
}

func TestRing_OverflowKeepsLast(t *testing.T) {
	r := &Ring{}
	// Append ringSize+10 entries; only the last ringSize should be kept.
	total := ringSize + 10
	for i := 0; i < total; i++ {
		r.Append(Record{Role: "role", Tool: itoa(i)})
	}
	entries := r.Entries()
	if len(entries) != ringSize {
		t.Fatalf("overflow: got %d entries, want %d", len(entries), ringSize)
	}
	// The first entry in the ring should be the one written at index 10.
	want := itoa(10)
	if entries[0].Tool != want {
		t.Errorf("first entry tool = %q, want %q", entries[0].Tool, want)
	}
	// The last entry should be the one written at index total-1.
	wantLast := itoa(total - 1)
	if entries[ringSize-1].Tool != wantLast {
		t.Errorf("last entry tool = %q, want %q", entries[ringSize-1].Tool, wantLast)
	}
}

func TestRing_OrderChronological(t *testing.T) {
	r := &Ring{}
	r.Append(Record{Tool: "first"})
	r.Append(Record{Tool: "second"})
	r.Append(Record{Tool: "third"})
	entries := r.Entries()
	if entries[0].Tool != "first" || entries[1].Tool != "second" || entries[2].Tool != "third" {
		t.Errorf("wrong order: %v %v %v", entries[0].Tool, entries[1].Tool, entries[2].Tool)
	}
}

func TestLog_WritesToGlobal(t *testing.T) {
	// Replace the global ring for this test to avoid cross-test pollution.
	prev := Global
	Global = &Ring{}
	t.Cleanup(func() { Global = prev })

	Log("testrole", "testtool", []string{"a", "b"}, map[string]int{"x": 1})
	entries := Global.Entries()
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	e := entries[0]
	if e.Role != "testrole" {
		t.Errorf("role = %q, want testrole", e.Role)
	}
	if e.Tool != "testtool" {
		t.Errorf("tool = %q, want testtool", e.Tool)
	}
	if len(e.ReturnedIDs) != 2 || e.ReturnedIDs[0] != "a" {
		t.Errorf("returned_ids = %v", e.ReturnedIDs)
	}
	if e.WithheldCounts["x"] != 1 {
		t.Errorf("withheld_counts[x] = %d, want 1", e.WithheldCounts["x"])
	}
}

func TestLog_NilSliceBecomesEmpty(t *testing.T) {
	prev := Global
	Global = &Ring{}
	t.Cleanup(func() { Global = prev })

	Log("r", "t", nil, nil)
	e := Global.Entries()[0]
	if e.ReturnedIDs == nil {
		t.Error("ReturnedIDs should not be nil")
	}
	if e.WithheldCounts == nil {
		t.Error("WithheldCounts should not be nil")
	}
}

// itoa converts an int to a string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
