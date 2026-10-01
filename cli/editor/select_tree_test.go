package editor

import (
	"env/filesystem"
	"env/types"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestSelectIncludesContainingTree(t *testing.T) {
	for _, draft := range []bool{false, true} {
		e := Editor{Modules: make([]*entry, 2)}
		opened := &types.Page{Path: "log/2026.09.30", Name: "2026.09.30", Type: "shallow", Content: []string{"opened"}}
		root := &types.Page{Path: "log", Name: "log", Type: "deep"}
		if draft {
			opened.Stage = "draft"
		} else {
			root.Children = []*types.Page{{Path: opened.Path, Name: opened.Name, Type: "shallow"}}
		}
		e.newPageEntry(opened, 0)
		e.newPageEntry(root, 1)
		for range 2 {
			e.generateAllEntries()
			if len(e.Entries) != 3 || e.Entries[0].Page.Path != "log" || e.Entries[1].Page != opened {
				t.Fatalf("draft=%v: expected root then opened page exactly once", draft)
			}
			if e.Entries[0].descendantPages() != 1 {
				t.Fatal("incorrect descendant count")
			}
		}
		e.Focus = 0
		e.collapse(true)
		e.generateAllEntries()
		if len(e.Entries) != 2 || e.Entries[0].Page.Path != "log" {
			t.Fatal("collapsed child still visible")
		}
		e.collapse(false)
		if len(e.Entries) != 3 {
			t.Fatal("child not restored")
		}
	}
}

func TestLoadMoreSelect(t *testing.T) {
	root := t.TempDir()
	fs, err := filesystem.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		path := filepath.Join(root, "log", fmt.Sprintf("page%02d", i))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("text"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	req := &types.Loading{Path: "log", Sort: "normal", Depth: -1, Latest: 5}
	ok, page := fs.LoadLocal(req)
	if !ok || page == nil || len(page.Children) != 5 {
		t.Fatal("initial page limit not applied")
	}
	e := Editor{Filesystem: fs, Request: req, Modules: make([]*entry, 2), Content: []string{"unsaved"}}
	e.newPageEntry(page, 0)
	e.generateAllEntries()
	for _, step := range []struct {
		key    string
		latest int
	}{{"enter", 6}, {"ctrl+enter", 10}} {
		e.Focus = len(e.Entries) - 1
		if e.Entries[e.Focus].Label != "load more!" {
			t.Fatal("last entry is not load more")
		}
		e.InputSelect(&types.Input{Description: step.key})
		if req.Latest != step.latest || len(e.Modules[0].Children) != step.latest {
			t.Fatalf("%s: latest=%d children=%d", step.key, req.Latest, len(e.Modules[0].Children))
		}
		if e.Focus != len(e.Entries)-1 {
			t.Fatal("load more lost focus")
		}
		if !slices.Equal(e.Content, []string{"unsaved"}) {
			t.Fatal("loading changed editor content")
		}
	}
}
