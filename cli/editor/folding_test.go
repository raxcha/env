package editor

import (
	"env/settings"
	"env/types"
	"slices"
	"strings"
	"testing"
)

func TestBlockFoldEligibility(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lines []string
		y, x  int
		fold  bool
	}{
		{"block", []string{"one", "two"}, 0, 0, true},
		{"single", []string{"one", "", "two"}, 0, 0, false},
		{"middle", []string{"one", "two", "three"}, 1, 0, false},
		{"column", []string{"one", "two"}, 0, 1, false},
		{"blank", []string{" ", "one", "two"}, 0, 0, false},
		{"dashes", []string{"---", "one", "two"}, 0, 0, false},
		{"after dashes", []string{" --- ", "one", "two"}, 1, 0, true},
		{"timestamp", []string{"12:34", "one", "two"}, 0, 0, false},
		{"after timestamp", []string{"12:34", "one", "two"}, 1, 0, true},
		{"timestamp not counted", []string{"one", "12:34", "two"}, 0, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := Editor{Content: slices.Clone(tc.lines), Cursor: [3]int{tc.x, tc.y}}
			e.InputVisual(&types.Input{Description: "left"})
			if (len(e.FoldedBlocks) == 1) != tc.fold {
				t.Fatalf("unexpected folds: %v", e.FoldedBlocks)
			}
			if !slices.Equal(e.Content, tc.lines) {
				t.Fatal("fold changed content")
			}
			if tc.fold {
				e.InputVisual(&types.Input{Description: "left"})
				if len(e.FoldedBlocks) != 0 {
					t.Fatal("second left did not unfold")
				}
			}
		})
	}
}

func TestFoldNavigationAndEditing(t *testing.T) {
	e := Editor{Content: []string{"one", "two", "three", "", "last"}}
	e.InputVisual(&types.Input{Description: "left"})
	e.InputVisual(&types.Input{Description: "down"})
	if e.Cursor[1] != 3 {
		t.Fatal("down entered hidden block")
	}
	e.InputVisual(&types.Input{Description: "up"})
	if e.Cursor[1] != 0 {
		t.Fatal("up entered hidden block")
	}
	e.InputVisual(&types.Input{Description: "char", Char: 'x'})
	if len(e.FoldedBlocks) != 0 || e.Content[1] != "two" {
		t.Fatal("edit must reveal preserved block")
	}
	e.InputVisual(&types.Input{Description: "ctrl+z"})
	if e.Content[0] != "one" {
		t.Fatal("fold interfered with undo")
	}
	e = Editor{Content: []string{"one", "two"}}
	e.InputVisual(&types.Input{Description: "left"})
	e.InputVisual(&types.Input{Description: "down"})
	if e.Cursor[1] != 0 {
		t.Fatal("down entered folded end of file")
	}
}

func TestFoldRendering(t *testing.T) {
	for _, wrap := range []bool{false, true} {
		sizes := &types.Dimensions{Full: types.Size{24, 5}, Size: types.Size{24, 5}}
		e := Editor{Content: []string{"head", strings.Repeat("hidden", 20), "hidden", "", "tail"}, Settings: &settings.Settings{}, VisualSizes: sizes, Numbers: true, Wrap: wrap}
		e.InputVisual(&types.Input{Description: "left"})
		e.InputVisual(&types.Input{Description: "down"})
		e.InputVisual(&types.Input{Description: "down"})
		frame := e.drawVisual().Frames[0]
		row := func(y int) string {
			var s strings.Builder
			for _, c := range frame.Cells[y*24 : (y+1)*24] {
				s.WriteRune(c.Char)
			}
			return s.String()
		}
		if !strings.HasPrefix(row(1), " …1 head …") || !strings.Contains(row(2), "4") || !strings.HasPrefix(row(3), "  5 tail") {
			t.Fatalf("wrap=%v: incorrect visible rows: %q / %q / %q", wrap, row(1), row(2), row(3))
		}
	}
}
