package editor

import (
	"env/filesystem"
	"env/types"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestSavePageStages(t *testing.T) {
	root := t.TempDir()
	fs, err := filesystem.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	ok, page := fs.LoadLocal(&types.Loading{Path: "new", CreateDraft: true, Latest: -1})
	if !ok || page == nil {
		t.Fatal("draft not loaded")
	}
	t.Cleanup(func() {
		for _, p := range fs.Stack {
			fs.CancelPatch(p)
		}
	})
	e := Editor{Filesystem: fs, Page: page, Content: []string{"first"}}
	for _, step := range []struct{ key, stage string }{
		{"ctrl+s", "edit"},
		{"ctrl+S", "local"},
	} {
		previous := page.Stage
		e.InputVisual(&types.Input{Description: step.key})
		select {
		case patch := <-fs.Patch:
			if page.Stage != previous || patch.Timer == nil || time.Until(patch.Deadline) <= 0 {
				t.Fatal("save must schedule a pending patch")
			}
			fs.ApplyPatch(patch)
		case <-time.After(time.Second):
			t.Fatal("save did not publish a patch")
		}
		if page.Stage != step.stage {
			t.Fatalf("stage = %q, want %q", page.Stage, step.stage)
		}
		data, err := os.ReadFile(filepath.Join(root, "new"))
		if err != nil || string(data) != "first" {
			t.Fatalf("saved content = %q, err = %v", data, err)
		}
	}
	e.Content = []string{"changed"}
	e.InputVisual(&types.Input{Description: "ctrl+s"})
	select {
	case patch := <-fs.Patch:
		fs.ApplyPatch(patch)
	case <-time.After(time.Second):
		t.Fatal("save did not publish a patch")
	}
	data, err := os.ReadFile(filepath.Join(root, "new"))
	if err != nil || string(data) != "changed" || page.Stage != "local" {
		t.Fatalf("edited save: content=%q stage=%q err=%v", data, page.Stage, err)
	}
	e.Content[0] = "not saved"
	if !slices.Equal(page.Content, []string{"changed"}) || !slices.Equal(page.Og.Content, []string{"changed"}) {
		t.Fatal("editor changes mutated the saved page or snapshot")
	}
}

func TestSavePageWithoutLoadedPage(t *testing.T) {
	e := Editor{Content: []string{"text"}}
	e.InputVisual(&types.Input{Description: "ctrl+s"})
}
