package editor

import (
	"env/cli"
	"env/types"
	"slices"
	"testing"
)

type projectLinkParent struct {
	cli.Parent
	spec   string
	follow bool
}

func (p *projectLinkParent) AddClients(spec, mode string, follow bool) {
	p.spec, p.follow = spec, follow
}

func TestProjectLinkEnter(t *testing.T) {
	const line = "ação @meu-projeto_2 @outro"
	for _, key := range []string{"enter", "ctrl+enter"} {
		for _, tc := range []struct {
			x    int
			want string
		}{
			{5, "editor:proj/@meu-projeto_2"},
			{9, "editor:proj/@meu-projeto_2"},
			{18, "editor:proj/@meu-projeto_2"},
			{20, "editor:proj/@outro"},
			{25, "editor:proj/@outro"},
			{26, ""},
			{19, ""},
			{0, ""},
		} {
			parent := &projectLinkParent{}
			e := Editor{Parent: parent, Content: []string{line}, Cursor: [3]int{tc.x, 0, 0}}
			cursor := e.Cursor
			e.InputVisual(&types.Input{Description: key})
			if parent.spec != tc.want {
				t.Fatalf("%s at %d: opened %q, want %q", key, tc.x, parent.spec, tc.want)
			}
			if tc.want != "" {
				if !slices.Equal(e.Content, []string{line}) || e.Cursor != cursor || len(e.Undo) != 0 || parent.follow != (key == "enter") {
					t.Fatalf("%s at %d: link changed editing state or focus behavior", key, tc.x)
				}
			} else if len(e.Content) != 2 {
				t.Fatalf("%s at %d: normal action did not insert a line", key, tc.x)
			}
		}
	}
}
