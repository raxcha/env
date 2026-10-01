package engine

import (
	"env/settings"
	"env/types"
	"io"
	"os"
	"strings"
	"testing"
)

func TestWrapClearsPreviousCells(t *testing.T) {

	s := &settings.Settings{}
	sizes := types.Dimensions{Full: types.Size{5, 3}, Size: types.Size{5, 3}}
	e := &Engine{
		oldframe: s.GenerateFrame(sizes, []string{"§xy999 abcdefghijk"}, 0, []int{0, 0, 0, 0}),
		newframe: s.GenerateFrame(sizes, []string{"§xy999 abc"}, 0, []int{0, 0, 0, 0}),
	}

	output, err := os.CreateTemp(t.TempDir(), "render")
	if err != nil { t.Fatal(err) }
	defer output.Close()
	stdout := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = stdout }()

	e.manageFrame()

	if _, err := output.Seek(0, 0); err != nil { t.Fatal(err) }
	data, err := io.ReadAll(output)
	if err != nil { t.Fatal(err) }

	if strings.Contains(string(data), "\033[2J") {
		t.Fatal("unchanged dimensions must keep incremental rendering")
	}

	for _, pos := range []string{"\033[2;1H", "\033[3;1H"} {
		_, after, ok := strings.Cut(string(data), pos)
		if !ok { t.Fatalf("missing update at %q", pos) }
		end := strings.Index(after, "\033[0m")
		if end < 0 || !strings.HasSuffix(after[:end], " ") {
			t.Fatalf("expected a space at %q", pos)
		}
	}

	for _, width := range []int{4, 6} {

		if err := output.Truncate(0); err != nil { t.Fatal(err) }
		if _, err := output.Seek(0, 0); err != nil { t.Fatal(err) }

		e.oldframe = e.newframe
		sizes.Full[0] = width
		sizes.Size[0] = width
		e.newframe = s.GenerateFrame(sizes, []string{"§xy999 abc"}, 0, []int{0, 0, 0, 0})
		e.manageFrame()

		if _, err := output.Seek(0, 0); err != nil { t.Fatal(err) }
		data, err := io.ReadAll(output)
		if err != nil { t.Fatal(err) }

		if !strings.HasPrefix(string(data), "\033[0m\033[2J") {
			t.Fatalf("width %d: expected screen clearing before redraw", width)
		}
		if len(e.diff.Cells) != width*3 {
			t.Fatalf("width %d: expected a complete redraw", width)
		}
	}
}
