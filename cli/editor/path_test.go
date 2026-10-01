package editor

import (
	"env/filesystem"
	"env/settings"
	"env/types"
	"env/utils"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestPathFieldEditing(t *testing.T) {
	e := Editor{Page: &types.Page{Path: "old"}, Content: []string{"body"}}
	e.InputVisual(&types.Input{Description: "up"})
	if !e.EditingPath {
		t.Fatal("up must enter path field")
	}
	e.InputVisual(&types.Input{Description: "char", Char: 'x'})
	if e.PathEditor.Content[0] != "xold" || !slices.Equal(e.Content, []string{"body"}) {
		t.Fatal("path editing changed body")
	}
	e.InputVisual(&types.Input{Description: "ctrl+z"})
	if e.PathEditor.Content[0] != "old" {
		t.Fatal("path undo failed")
	}
	e.InputVisual(&types.Input{Description: "enter"})
	if e.EditingPath || len(e.Content) != 1 {
		t.Fatal("enter must return to body without inserting a line")
	}
}

func TestSavePathMovesPage(t *testing.T) {
	for _, draft := range []bool{false, true} {
		t.Run(map[bool]string{true: "draft", false: "existing"}[draft], func(t *testing.T) {
			dir := t.TempDir()
			if !draft {
				if err := os.WriteFile(filepath.Join(dir, "old"), []byte("before"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			fs, err := filesystem.Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			ok, page := fs.LoadLocal(&types.Loading{Path: "old", CreateDraft: true, Cont: true, Latest: -1})
			if !ok {
				t.Fatal("load failed")
			}
			t.Cleanup(func() {
				for _, patch := range fs.Stack {
					fs.CancelPatch(patch)
				}
			})
			e := Editor{Page: page, Filesystem: fs, Content: []string{"after"}, Spec: "old"}
			e.initPathEditor()
			e.PathEditor.Content[0] = "new/place"
			e.savePage()
			if e.PathError != "" {
				t.Fatal(e.PathError)
			}
			if fs.Cache["new/place"] != page || fs.Cache["old"] != nil || e.Spec != "new/place" {
				t.Fatal("cache or editor path not updated")
			}
			if !draft && page.Og.Path != "old" {
				t.Fatal("original path lost before applying move")
			}
			if _, err := os.Stat(filepath.Join(dir, "new/place")); !os.IsNotExist(err) {
				t.Fatal("move happened before patch")
			}
			select {
			case patch := <-fs.Patch:
				fs.ApplyPatch(patch)
			case <-time.After(time.Second):
				t.Fatal("no patch")
			}
			data, err := os.ReadFile(filepath.Join(dir, "new/place"))
			if err != nil || string(data) != "after" {
				t.Fatalf("saved body=%q err=%v", data, err)
			}
			if _, err := os.Stat(filepath.Join(dir, "old")); !os.IsNotExist(err) {
				t.Fatal("old file still exists")
			}
		})
	}
}

func TestPathHeaderStaysAtTop(t *testing.T) {
	sizes := &types.Dimensions{Full: types.Size{30, 4}, Size: types.Size{30, 4}}
	e := Editor{Page: &types.Page{Path: "log/day", Stage: "draft"}, Content: []string{"a", "b", "c", "d", "e", "f"}, Cursor: [3]int{0, 5, 0}, Settings: &settings.Settings{}, VisualSizes: sizes}
	e.initPathEditor()
	frame := e.drawPath(*sizes)
	for i, r := range []rune(" " + utils.StageIcons["draft"] + " " + "log/day") {
		if frame.Cells[i].Char != r {
			t.Fatal("path header missing")
		}
	}
	if e.visualViewport(e.Content, 30) == 0 {
		t.Fatal("body did not scroll")
	}
	queue := e.drawVisual()
	for i, r := range []rune(" " + utils.StageIcons["draft"] + " " + "log/day") {
		if queue.Frames[0].Cells[i].Char != r {
			t.Fatal("scrolling body covered fixed header")
		}
	}
}

func TestSavePathRejectsExistingDestination(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"old", "taken"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	fs, err := filesystem.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	_, page := fs.LoadLocal(&types.Loading{Path: "old", Cont: true, Latest: -1})
	e := Editor{Page: page, Filesystem: fs, Content: []string{"changed"}}
	e.initPathEditor()
	for _, path := range []string{"taken", "../outside", ""} {
		e.PathEditor.Content[0] = path
		e.savePage()
		if e.PathError == "" || page.Path != "old" {
			t.Fatalf("invalid move accepted: %q", path)
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "taken"))
	if err != nil || string(data) != "taken" {
		t.Fatal("destination overwritten")
	}
}
