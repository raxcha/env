package editor

import (
	"env/filesystem"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestAppendTemplateThenSave(t *testing.T) {
	root := t.TempDir()
	fs, err := filesystem.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, patch := range fs.Stack {
			fs.CancelPatch(patch)
		}
	})
	page := fs.NewDraft("log/2026.10.01", nil)
	e := &Editor{Filesystem: fs, Page: page, Content: []string{"pending edits"}, EditingPath: true}
	e.AppendTemplate("@proj$42/.template", time.Date(2026, 10, 1, 14, 32, 0, 0, time.Local))
	want := []string{"pending edits", "", "14:32", "= @proj.template"}
	if !slices.Equal(e.Content[:4], want) || !strings.HasPrefix(e.Content[4], "id: ") || len(e.Content[4]) != 16 {
		t.Fatalf("unexpected log syntax: %q", e.Content)
	}
	if e.Cursor[1] != 3 || e.EditingPath {
		t.Fatal("cursor must be on the event line")
	}
	if len(fs.Stack) != 0 || len(fs.Cache) != 0 || len(page.Content) != 0 {
		t.Fatal("selection must only edit the buffer")
	}
	if _, err := os.Stat(filepath.Join(root, "log")); !os.IsNotExist(err) {
		t.Fatal("selection wrote the log")
	}
	prepared := slices.Clone(e.Content)
	e.undo()
	if !slices.Equal(e.Content, []string{"pending edits"}) {
		t.Fatal("append must undo as one edit")
	}
	e.redo()
	e.savePage()
	if !slices.Equal(e.Content, prepared) {
		t.Fatal("saving duplicated the generated syntax")
	}
	id := strings.TrimPrefix(e.Content[4], "id: ")
	if fs.Cache["rec/="+id+"$proj.template"] == nil {
		t.Fatal("Ctrl+S did not create the session draft")
	}
}

func TestAppendTemplateExpandsDates(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "proj", "@proj$42", ".template")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	source := "date: .{yyyy.mm.dd (weekday)}\ntime: .{yyyy.mm.dd hh:mm }"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	fs, err := filesystem.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	e := &Editor{Filesystem: fs}
	e.AppendTemplate("@proj$42/.template", time.Date(2026, 10, 1, 22, 7, 0, 0, time.Local))
	want := []string{"date: 2026.10.01 thursday", "time: 2026.10.01 22:07"}
	if len(e.Content) != 5 || !slices.Equal(e.Content[3:], want) {
		t.Fatalf("unexpected expanded template: %q", e.Content)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != source {
		t.Fatal("source template changed")
	}
}
