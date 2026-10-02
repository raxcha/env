package utils

import (
	"slices"
	"testing"
	"time"
)

func TestExpandTemplate(t *testing.T) {
	now := time.Date(2026, 10, 1, 22, 7, 0, 0, time.FixedZone("BRT", -3*60*60))
	source := []string{
		"date: .{yyyy.mm.dd}",
		"date: .{yyyy.mm.dd (weekday)}",
		"time: .{yyyy.mm.dd hh:mm }",
		"time: .{yyyy.mm.dd  hh:mm (weekday)}",
		".{int} .{yyyy.mm.dd unknown}",
		".{yyyy.mm.dd} / .{yyyy.mm.dd}",
	}
	original := slices.Clone(source)
	want := []string{
		"date: 2026.10.01",
		"date: 2026.10.01 thursday",
		"time: 2026.10.01 22:07",
		"time: 2026.10.01 22:07 thursday",
		".{int} .{yyyy.mm.dd unknown}",
		"2026.10.01 / 2026.10.01",
	}
	if got := ExpandTemplate(source, now); !slices.Equal(got, want) {
		t.Fatalf("got %q; want %q", got, want)
	}
	if !slices.Equal(source, original) {
		t.Fatal("source template changed")
	}
}
