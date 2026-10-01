package settings

import (
	"env/types"
	"testing"
)

func TestFrameBackgroundMargins(t *testing.T) {

	s := &Settings{Theme: types.Theme{
		Background: types.Color{R: 10},
		Foreground: types.Color{R: 200},
		Color02: types.Color{R: 100},
	}}
	sizes := types.Dimensions{Full: types.Size{16, 8}, Pos: types.Pos{2, 1}, Size: types.Size{12, 6}}
	base := s.GenerateFrame(types.Dimensions{Full: sizes.Full, Size: sizes.Full}, nil, 0, []int{0, 0, 0, 0})

	for _, text := range []string{"§yx999 abcdefghijk", "§yx999 abcdefghij¤bx k¤ "} {

		frame := s.GenerateFrame(sizes, []string{text, "§yx0 ", "§xy0 z"}, 0, []int{1, 1, 0, 1})
		if frame.Sizes != sizes { t.Fatal("frame must include its margins") }
		frame = s.MergeFrames(base, frame)

		for y := 0; y < 8; y++ {
			for x := 0; x < 16; x++ {

				expected := s.ChooseColor('x', "bg", 0)
				if y >= 2 && y <= 4 && x >= 2 && x < 14 {
					expected = s.ChooseColor('y', "bg", 0)
				}
				if text != "§yx999 abcdefghijk" && y == 3 && x == 3 {
					expected = s.ChooseColor('b', "bg", 0)
				}
				if frame.Cells[y*16+x].Ansi.Bg != expected {
					t.Fatalf("unexpected background at (%d, %d) for %q", x, y, text)
				}
			}
		}

		if frame.Cells[2*16+3].Char != 'a' || frame.Cells[3*16+3].Char != 'k' || frame.Cells[5*16+3].Char != 'z' {
			t.Fatal("margins must preserve text positioning and wrapping")
		}
	}
}

func TestFrameBackgroundWithoutContentWidth(t *testing.T) {

	s := &Settings{Theme: types.Theme{Foreground: types.Color{R: 200}}}
	sizes := types.Dimensions{Full: types.Size{2, 1}, Size: types.Size{2, 1}}
	frame := s.GenerateFrame(sizes, []string{"§yx999 abc"}, 0, []int{0, 1, 0, 1})

	for _, cell := range frame.Cells {
		if cell.Char != ' ' || cell.Ansi.Bg != s.ChooseColor('y', "bg", 0) {
			t.Fatal("margins must retain the line background even without room for text")
		}
	}
}
