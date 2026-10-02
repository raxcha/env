package editor

import (
	"env/types"
	"testing"
)

func TestSidebarContextsAndNavigation(t *testing.T) {
	makeEditor := func(context string) *Editor {
		page := &types.Page{Path: context + "/one", Name: "one"}
		e := &Editor{Spec: page.Path, Requests: []*types.Loading{{Path: context}}, Modules: make([]*entry, 1)}
		e.newPageEntry(page, 0)
		return e
	}
	rec, log, duplicate, proj := makeEditor("rec"), makeEditor("log"), makeEditor("rec"), makeEditor("proj")
	s := &Sidebar{}
	area := types.Dimensions{Full: types.Size{100, 30}, Size: types.Size{100, 30}}
	content := s.Prepare([]*Editor{rec, log, duplicate, proj}, area)
	contexts := []string{}
	for _, row := range s.rows {
		if row.index < 0 {
			contexts = append(contexts, row.context)
		}
	}
	if len(contexts) != 3 || contexts[0] != "rec" || contexts[1] != "log" || contexts[2] != "proj" {
		t.Fatalf("unexpected contexts: %v", contexts)
	}
	if content.Pos[0] != s.Sizes.Size[0] || content.Size[0]+s.Sizes.Size[0] != 100 {
		t.Fatal("sidebar must reserve one shared strip")
	}
	for range 4 {
		s.Input(&types.Input{Description: "down"})
	}
	if s.rows[s.focus].owner != log || s.rows[s.focus].index != 0 {
		t.Fatal("navigation must cross context headers")
	}
	s.Prepare([]*Editor{proj}, area)
	for _, row := range s.rows {
		if row.owner != proj {
			t.Fatal("monocle retained another context")
		}
	}
}

func TestSidebarShortArea(t *testing.T) {
	s := &Sidebar{}
	e := &Editor{Spec: "rec", Modules: make([]*entry, 1)}
	for _, size := range []types.Size{{0, 0}, {1, 1}, {10, 20}} {
		area := types.Dimensions{Full: size, Size: size}
		content := s.Prepare([]*Editor{e}, area)
		if content.Size[0] < 0 || content.Size[1] < 0 || s.Sizes.Size[0] < 0 || s.Sizes.Size[1] < 0 {
			t.Fatal("negative geometry")
		}
	}
}

func TestSidebarFoldContexts(t *testing.T) {
	area := types.Dimensions{Full: types.Size{100, 30}, Size: types.Size{100, 30}}
	s := &Sidebar{}
	editors := []*Editor{}
	for _, context := range []string{"log", "rec", "rand", "fami", "proj"} {
		e := &Editor{Spec: context, Modules: make([]*entry, 1)}
		e.newPageEntry(&types.Page{Path: context + "/one", Name: "one"}, 0)
		editors = append(editors, e)
	}
	s.Prepare(editors, area)
	for i, e := range editors {
		if s.rows[s.focus].context != e.Spec || s.rows[s.focus].index != -1 {
			t.Fatal("context header must be selectable")
		}
		s.Input(&types.Input{Description: "left"})
		s.Prepare(editors, area)
		if s.rows[s.focus].context != e.Spec || s.rows[s.focus].index != -1 {
			t.Fatal("folding lost focus")
		}
		for _, row := range s.rows {
			if row.context == e.Spec && row.index >= 0 {
				t.Fatal("folded context still has children")
			}
		}
		if i < len(editors)-1 {
			s.Input(&types.Input{Description: "down"})
		}
	}
	if len(s.rows) != 5 {
		t.Fatal("expected only five headers")
	}
	s.Prepare(editors[4:], area)
	s.Prepare(editors, area)
	if len(s.rows) != 5 || s.rows[s.focus].context != "proj" {
		t.Fatal("fold state or focus lost after switching contexts")
	}
	s.Input(&types.Input{Description: "right"})
	s.Prepare(editors, area)
	if len(s.rows) != 7 {
		t.Fatal("expanding must restore children and load more")
	}
	s.Input(&types.Input{Description: "enter"})
	s.Prepare(editors, area)
	if len(s.rows) != 5 {
		t.Fatal("enter must toggle folding")
	}
}
