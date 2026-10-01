package filesystem

import (
	"env/types"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestGetContentLineEndings(t *testing.T) {

	for _, text := range []string{"first\n\nlast\n", "first\r\n\r\nlast\r\n", "first\r\n\nlast\r\n"} {

		path := filepath.Join(t.TempDir(), "note")
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}

		f := &Filesystem{}
		content := f.getContent(path)
		if !slices.Equal(content, []string{"first", "", "last", ""}) {
			t.Fatalf("unexpected content: %q", content)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != text {
			t.Fatal("reading must not change the file")
		}
	}
}

func TestLoadEmptyFilteredResult(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "proj"), 0700); err != nil {
		t.Fatal(err)
	}
	f := &Filesystem{root: root, Cache: map[string]*types.Page{}}
	req := &types.Loading{Path: "proj", Mode: "fresh", Depth: 1, Filters: []string{"shallow"}, Latest: 5, Token: "test"}
	ok, page := f.LoadLocal(req)
	if !ok || page != nil {
		t.Fatalf("expected successful empty result, got ok=%v page=%v", ok, page)
	}
	select {
	case page, open := <-f.Load(req):
		if open || page != nil {
			t.Fatalf("expected closed channel for empty result, got open=%v page=%v", open, page)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("empty load did not finish")
	}
}

func TestLoadMissingPathCreatesCachedDraft(t *testing.T) {
	root := t.TempDir()
	f := &Filesystem{root: root}
	req := &types.Loading{Path: "proj/../proj/new", Mode: "fresh", Latest: -1, Token: "new-page", CreateDraft: true}
	var page *types.Page
	select {
	case page = <-f.Load(req):
	case <-time.After(5 * time.Second):
		t.Fatal("load did not finish")
	}
	if page == nil || page.Stage != "draft" || page.Path != "proj/new" || page.Name != "new" || page.Token != req.Token {
		t.Fatalf("unexpected draft: %+v", page)
	}
	if !slices.Equal(page.Content, []string{""}) || f.Cache["proj/new"] != page {
		t.Fatal("expected a blank draft in the cache")
	}
	if _, err := os.Stat(filepath.Join(root, "proj")); !os.IsNotExist(err) {
		t.Fatalf("loading a draft must not create directories: %v", err)
	}
	page.Content = []string{"unsaved"}
	ok, again := f.LoadLocal(req)
	if !ok || again != page || !slices.Equal(again.Content, []string{"unsaved"}) {
		t.Fatal("reloading must preserve the cached draft")
	}
}

func TestProjectPriorityFromName(t *testing.T) {
	for _, tc := range []struct {
		path string
		want int
	}{
		{"proj/@nome$99", 99},
		{"proj/@nome$0", 0},
		{"proj/@nome$99/index", 99},
		{"proj/@nome$invalid", 7},
		{"proj/@nome$", 7},
		{"proj/@nome", 7},
		{"proj/nome$99", 7},
		{"other/@nome$99", 7},
	} {
		t.Run(tc.path, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, tc.path)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("priority:7\n"), 0600); err != nil {
				t.Fatal(err)
			}
			f := &Filesystem{root: root, Standards: map[string][]string{".{int}": {"priority"}}}
			got, ok := f.getMetadata(path)["priority"].(int)
			if !ok || got != tc.want {
				t.Fatalf("priority = %v (int: %v), want %d", got, ok, tc.want)
			}
		})
	}
}

func TestSyncAppliesAfterDeadline(t *testing.T) {
	f := &Filesystem{root: t.TempDir(), Patch: make(chan *types.Patch, 1)}
	page := f.NewDraft("note", nil)
	page.Content = []string{"saved"}
	patch := <-f.Sync(&types.Syncing{Branch: page})
	t.Cleanup(func() { f.CancelPatch(patch) })
	if page.Stage != "draft" {
		t.Fatal("patch applied immediately")
	}
	select {
	case <-patch.Done:
	case <-time.After(7 * time.Second):
		t.Fatal("patch not applied after five seconds")
	}
	if time.Now().Before(patch.Deadline) {
		t.Fatal("patch applied before deadline")
	}
	data, err := os.ReadFile(filepath.Join(f.root, "note"))
	if err != nil || string(data) != "saved" || page.Stage != "edit" {
		t.Fatalf("content=%q stage=%q err=%v", data, page.Stage, err)
	}
}

func TestPatchApplyAndCancelOnlyOnce(t *testing.T) {
	f := &Filesystem{}
	count := 0
	patch := f.newPatch()
	patch.Commands = []func(){func() { count++ }}
	f.ApplyPatch(patch)
	f.ApplyPatch(patch)
	f.CancelPatch(patch)
	if count != 1 {
		t.Fatalf("patch applied %d times", count)
	}
	cancelled := f.newPatch()
	cancelled.Commands = patch.Commands
	f.CancelPatch(cancelled)
	f.ApplyPatch(cancelled)
	if count != 1 {
		t.Fatal("cancelled patch applied")
	}
}

func TestTimeMetadataFromTitle(t *testing.T) {
	for _, tc := range []struct {
		path  string
		valid bool
	}{
		{"log/2026.09.30", true},
		{"log/2026.09.30*", true},
		{"log/2026.09.30 notes", true},
		{"log/2026.09.30/index", true},
		{"log/2026.02.30", false},
		{"log/note-2026.09.30", false},
		{"log/2026.9.30", false},
	} {
		t.Run(tc.path, func(t *testing.T) {
			f := &Filesystem{}
			metadata := f.getMetadata(filepath.Join(t.TempDir(), tc.path))
			got, ok := metadata["time"].(time.Time)
			if ok != tc.valid {
				t.Fatalf("time metadata present=%v, want %v", ok, tc.valid)
			}
			if ok && !got.Equal(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)) {
				t.Fatalf("unexpected date %v", got)
			}
		})
	}
}
