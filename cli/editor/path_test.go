package editor

import (
	"env/cli"
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
	for i, r := range []rune("  " + utils.StageIcons["draft"] + " " + "log/day") {
		if frame.Cells[i].Char != r {
			t.Fatal("path header missing")
		}
	}
	if e.visualViewport(e.Content, 30) == 0 {
		t.Fatal("body did not scroll")
	}
	queue := e.drawVisual()
	for i, r := range []rune("log/day") {
		if queue.Frames[0].Cells[i].Char != r {
			t.Fatal("scrolling body covered fixed header")
		}
	}
}

func TestPathAlignsWithContent(t *testing.T) {
	for _, stage := range []string{"draft", "*api"} {
		for _, editing := range []bool{false, true} {
			for _, zen := range []bool{false, true} {
				sizes := &types.Dimensions{Full: types.Size{70, 6}, Pos: types.Pos{5, 1}, Size: types.Size{60, 4}}
				e := Editor{Page: &types.Page{Path: "path", Stage: stage}, Content: []string{"body"}, Settings: &settings.Settings{}, VisualSizes: sizes, Numbers: true, Zen: zen, EditingPath: editing}
				frame := e.drawVisual().Frames[0]
				x := 9
				if zen {
					x = 14
				}
				if frame.Cells[70+x].Char != 'p' || frame.Cells[140+x].Char != 'b' {
					t.Fatalf("path and body misaligned: stage=%s editing=%v zen=%v", stage, editing, zen)
				}
				if frame.Cells[70+x-1].Char != ' ' {
					t.Fatal("stage must have one space before path")
				}
				if frame.Cells[70+x-2].Char != []rune(utils.StageIcons[stage])[len([]rune(utils.StageIcons[stage]))-1] {
					t.Fatal("stage icon missing")
				}
			}
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
	for _, path := range []string{"taken", "../outside"} {
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

type blankPathParent struct {
	cli.Parent
	clients []cli.Client
	closed  int
}

func (p *blankPathParent) GetClients() []cli.Client { return p.clients }
func (p *blankPathParent) CloseClient(index int)    { p.closed = index }

func TestSaveBlankPathStagesDoomAndClosesEditor(t *testing.T) {
	for _, stage := range []string{"draft", "edit", "local", "*api", "api", "move"} {
		for _, blank := range []string{"", "   "} {
			for _, editingPath := range []bool{false, true} {
				t.Run(stage+"/"+blank+"/"+map[bool]string{true: "path", false: "body"}[editingPath], func(t *testing.T) {
					dir := t.TempDir()
					if err := os.WriteFile(filepath.Join(dir, "note"), []byte("original"), 0600); err != nil {
						t.Fatal(err)
					}
					fs, err := filesystem.Open(dir)
					if err != nil {
						t.Fatal(err)
					}
					_, page := fs.LoadLocal(&types.Loading{Path: "note", Cont: true, Latest: -1})
					fs.Cache[page.Path] = page
					page.Stage = stage
					if stage == "draft" {
						page.Og = nil
					}
					original := page.Og
					parent := &blankPathParent{closed: -1}
					e := &Editor{Parent: parent, Filesystem: fs, Page: page, Content: []string{"changed"}, EditingPath: editingPath}
					parent.clients = []cli.Client{&Editor{}, e}
					e.initPathEditor()
					e.PathEditor.Content[0] = blank
					e.InputVisual(&types.Input{Description: "ctrl+s"})
					if page.Stage != "doom" || parent.closed != 1 || e.PathError != "" {
						t.Fatalf("stage=%q closed=%d error=%q", page.Stage, parent.closed, e.PathError)
					}
					if page.Path != "note" || page.Og != original || fs.Cache["note"] != page || len(fs.Stack) != 0 {
						t.Fatal("staging deletion changed the path, snapshot, cache, or scheduled a sync")
					}
					select {
					case patch := <-fs.Patch:
						t.Cleanup(func() { fs.CancelPatch(patch) })
						if len(patch.Description) != 1 || !slices.Equal(patch.Description[0], []string{"note", stage, "doom"}) {
							t.Fatalf("unexpected notification: %v", patch.Description)
						}
						if len(patch.Commands) != 0 || patch.Deadline.IsZero() {
							t.Fatal("notification must expire without scheduling deletion")
						}
						fs.ApplyPatch(patch)
						if page.Stage != "doom" {
							t.Fatal("notification changed the staged page")
						}
					case <-time.After(time.Second):
						t.Fatal("missing doom notification")
					}
					data, err := os.ReadFile(filepath.Join(dir, "note"))
					if err != nil || string(data) != "original" {
						t.Fatalf("content=%q err=%v", data, err)
					}
				})
			}
		}
	}
}
