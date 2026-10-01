package editor

import (
	"env/filesystem"
	"env/types"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestLogEventOnSave(t *testing.T) {
	for _, marker := range []string{"`", ">", "=", "~"} {
		for _, existing := range []string{"none", "time", "id", "both"} {
			t.Run(marker+existing, func(t *testing.T) {
				content := []string{marker + " event"}
				y := 0
				if existing == "time" || existing == "both" {
					content = append([]string{"08:32"}, content...)
					y++
				}
				if existing == "id" || existing == "both" {
					content = append(content, "id: abcdef123456")
				}
				e := Editor{Page: &types.Page{Path: "log/2026/09/30"}, Content: content, Cursor: [3]int{2, y, 0}}
				e.ensureLogEvent()
				if len(e.Content) != 3 || e.Cursor != [3]int{2, 1, 0} || e.Content[1] != marker+" event" {
					t.Fatalf("content=%q cursor=%v", e.Content, e.Cursor)
				}
				if _, err := time.Parse("15:04", e.Content[0]); err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(e.Content[2], "id: ") || len(e.Content[2]) != 16 {
					t.Fatalf("id=%q", e.Content[2])
				}
				if (existing == "time" || existing == "both") && e.Content[0] != "08:32" {
					t.Fatal("existing time changed")
				}
				if (existing == "id" || existing == "both") && e.Content[2] != "id: abcdef123456" {
					t.Fatal("existing id changed")
				}
				before := slices.Clone(e.Content)
				e.ensureLogEvent()
				if !slices.Equal(before, e.Content) {
					t.Fatal("repeated save changed event")
				}
			})
		}
	}
}

func TestLogEventIgnoresOtherLines(t *testing.T) {
	for _, tc := range []struct{ path, line string }{
		{"proj/page", "> event"}, {"log", "> event"}, {"logbook/page", "> event"}, {"log/page", "ordinary"}, {"log/page", ""},
	} {
		e := Editor{Page: &types.Page{Path: tc.path}, Content: []string{tc.line}}
		e.ensureLogEvent()
		if !slices.Equal(e.Content, []string{tc.line}) {
			t.Fatalf("unexpected change: %q", e.Content)
		}
	}
}

func TestSaveLogEventCreatesCachedDraft(t *testing.T) {
	for _, tc := range []struct{ marker, root string }{{"`", "rand"}, {">", "rec"}, {"=", "rec"}, {"~", "rec"}} {
		t.Run(tc.marker, func(t *testing.T) {
			dir := t.TempDir()
			fs, err := filesystem.Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				for _, p := range fs.Stack {
					fs.CancelPatch(p)
				}
			})
			e := Editor{Filesystem: fs, Page: fs.NewDraft("log/day", nil), Content: []string{tc.marker + " O que está escrito / na linha!"}}
			e.savePage()
			id := strings.TrimPrefix(e.Content[2], "id: ")
			path := tc.root + "/>" + id + "$o-que-está-escrito-na-linha"
			draft := fs.Cache[path]
			if draft == nil || draft.Stage != "draft" || draft.Name != filepath.Base(path) {
				t.Fatalf("missing event draft at %q: %#v", path, draft)
			}
			if _, err := os.Stat(filepath.Join(dir, tc.root)); !os.IsNotExist(err) {
				t.Fatalf("event directory created: %v", err)
			}
			e.savePage()
			if fs.Cache[path] != draft {
				t.Fatal("repeat save replaced draft")
			}
			e.Content[e.Cursor[1]] = tc.marker + " changed title"
			e.savePage()
			count := 0
			for key := range fs.Cache {
				if strings.HasPrefix(key, tc.root+"/>") {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("created %d drafts for one ID", count)
			}
		})
	}
}
