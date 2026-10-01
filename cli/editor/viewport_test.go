package editor

import (
	"env/types"
	"env/utils"
	"testing"
)

func TestVisualViewport(t *testing.T) {

	for _, test := range []struct {
		name   string
		lines  []string
		width  int
		height int
		wrap   bool
	}{
		{"plain", []string{"a", "b", "c", "d", "e", "f"}, 5, 4, false},
		{"wrapped", []string{"12345678901", "b", "c", "d"}, 5, 4, true},
		{"cursor wraps", []string{"a", "b", "c", "123456"}, 5, 4, true},
		{"formatted", []string{"§xy999 12345¤bx 6¤ ", "b", "c", "d"}, 5, 4, true},
		{"preview", []string{"a", "b", "c", "d", "e", "f"}, 5, 3, true},
		{"one row", []string{"a", "b", "c"}, 5, 1, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			e := Editor{Content: []string{""}, Wrap: test.wrap, VisualSizes: &types.Dimensions{Size: types.Size{test.width, test.height + 1}}}
			for cursor := range test.lines {
				e.Cursor[1] = cursor
				first := e.visualViewport(test.lines, test.width)
				if first < 0 || first > cursor {
					t.Fatalf("invalid first line %d for cursor %d", first, cursor)
				}
				rows := 0
				for _, line := range test.lines[first : cursor+1] {
					parts := 1
					if test.wrap {
						parts = max(1, (utils.VisibleLength(line)+test.width-1)/test.width)
					}
					rows += parts
				}
				if rows > test.height {
					t.Fatalf("cursor %d hidden: first %d, rows %d, height %d", cursor, first, rows, test.height)
				}
			}
		})
	}
}
