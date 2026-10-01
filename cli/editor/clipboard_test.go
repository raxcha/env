package editor

import (
	"env/cli"
	"env/types"
	"slices"
	"testing"
)

type clipboardParent struct {
	cli.Parent
	clipboard string
}

func (p *clipboardParent) GetClipboard() string { return p.clipboard }
func (p *clipboardParent) SetClipboard(content string) { p.clipboard = content }

func TestLineClipboard(t *testing.T) {

	parent := &clipboardParent{}
	e := Editor{Parent: parent, Content: []string{"ação", "fim"}, Cursor: [3]int{2, 0}}

	press := func(input string, content []string, cursor [3]int) {
		t.Helper()
		e.InputVisual(&types.Input{Description: input})
		if !slices.Equal(e.Content, content) || e.Cursor != cursor {
			t.Fatalf("%s: got %q %v, want %q %v", input, e.Content, e.Cursor, content, cursor)
		}
	}

	press("ctrl+C", []string{"ação", "fim"}, [3]int{2, 0})
	press("ctrl+d", []string{"ação", "ação", "fim"}, [3]int{2, 1})
	press("ctrl+z", []string{"ação", "fim"}, [3]int{2, 0})
	press("ctrl+y", []string{"ação", "ação", "fim"}, [3]int{2, 1})
	press("ctrl+X", []string{"ação", "fim"}, [3]int{2, 1})
	press("ctrl+V", []string{"ação", "fim", "ação"}, [3]int{0, 2})
	press("ctrl+z", []string{"ação", "fim"}, [3]int{2, 1})
	press("ctrl+y", []string{"ação", "fim", "ação"}, [3]int{0, 2})
	press("ctrl+x", []string{"ação", "fim"}, [3]int{0, 1})

	other := Editor{Parent: parent, Content: []string{"outro"}}
	other.InputVisual(&types.Input{Description: "ctrl+v"})
	if !slices.Equal(other.Content, []string{"outro", "ação"}) {
		t.Fatalf("shared clipboard: %q", other.Content)
	}
}

func TestCutOnlyLine(t *testing.T) {

	parent := &clipboardParent{}
	e := Editor{Parent: parent, Content: []string{"texto"}}
	e.InputVisual(&types.Input{Description: "ctrl+x"})
	if !slices.Equal(e.Content, []string{""}) || parent.clipboard != "texto\n" {
		t.Fatalf("cut: %q, clipboard %q", e.Content, parent.clipboard)
	}
	e.InputVisual(&types.Input{Description: "ctrl+z"})
	if !slices.Equal(e.Content, []string{"texto"}) {
		t.Fatalf("undo cut: %q", e.Content)
	}

	e = Editor{Parent: parent, Content: []string{""}}
	e.InputVisual(&types.Input{Description: "ctrl+c"})
	e.InputVisual(&types.Input{Description: "ctrl+v"})
	if !slices.Equal(e.Content, []string{"", ""}) {
		t.Fatalf("paste empty line: %q", e.Content)
	}

	parent.clipboard = ""
	e.InputVisual(&types.Input{Description: "ctrl+v"})
	if !slices.Equal(e.Content, []string{"", ""}) {
		t.Fatalf("empty clipboard: %q", e.Content)
	}
}
