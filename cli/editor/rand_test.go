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

func TestRandBlockSyncOnlyWhileDraft(t *testing.T) {
	root := t.TempDir()
	fs, err := filesystem.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	e := &Editor{Filesystem: fs, Page: fs.NewDraft("log/day", nil), Content: []string{"08:32", "` idea", "id: abcdef123456", "first", "", "other block"}, Cursor: [3]int{0, 1, 0}}
	e.ensureLogEvent()
	path := "rand/`abcdef123456$idea"
	page := fs.Cache[path]
	if page == nil || !slices.Equal(page.Content, e.Content[3:4]) {
		t.Fatalf("missing block content: %+v", page)
	}
	e.Cursor[1] = 3
	e.Content[3] = "updated"
	e.syncEventBlocks()
	if page.Stage != "draft" || !slices.Equal(page.Content, e.Content[3:4]) {
		t.Fatalf("draft not updated: %+v", page)
	}
	if _, err := os.Stat(filepath.Join(root, "rand")); !os.IsNotExist(err) {
		t.Fatal("sync persisted the draft")
	}
	for _, stage := range []string{"edit", "local", "api", "move"} {
		page.Stage = stage
		page.Content = []string{"independent page"}
		e.syncEventBlocks()
		if !slices.Equal(page.Content, []string{"independent page"}) {
			t.Fatalf("overwrote %s page", stage)
		}
	}
}

func TestRandBlockExpandsIdeaTemplate(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "rand"), 0700); err != nil {
		t.Fatal(err)
	}
	source := "kind: idea\nid: .{abcdef123456}\naliases: \nrelease-time: .{yyyy.mm.dd (weekday)}\nfamilies:\ntags: \nstatus:\n---\n.{?content}"
	path := filepath.Join(root, "rand", ".`abcdef123456$idea")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	fs, err := filesystem.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	e := &Editor{Filesystem: fs, Page: fs.NewDraft("log/day", nil), Content: []string{"` new idea", "id: 123456abcdef", "body .{abcdef123456}", "", "unrelated"}}
	e.ensureLogEvent()
	page := fs.Cache["rand/`123456abcdef$new-idea"]
	now := time.Now()
	wantHeader := []string{"kind: idea", "id: 123456abcdef", "aliases: ", "release-time: " + now.Format("2006.01.02") + " " + strings.ToLower(now.Weekday().String()), "families:", "tags: ", "status:", "---"}
	want := append(slices.Clone(wantHeader), e.Content[3:4]...)
	if page == nil || !slices.Equal(page.Content, want) {
		t.Fatalf("template not expanded: %+v; want %q", page, want)
	}
	e.Content[3] = "updated body"
	e.syncEventBlocks()
	want = append(slices.Clone(wantHeader), e.Content[3:4]...)
	if !slices.Equal(page.Content, want) {
		t.Fatalf("draft not synchronized: %q", page.Content)
	}
	page.Stage = "edit"
	e.Content[3] = "later change"
	e.syncEventBlocks()
	if !slices.Equal(page.Content, want) {
		t.Fatal("saved idea changed")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != source {
		t.Fatal("source template changed")
	}
}

func TestRandBlockDoesNotOverwriteSavedPageAfterRestart(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "rand"), 0700); err != nil {
		t.Fatal(err)
	}
	path := "rand/`abcdef123456$original-title"
	if err := os.WriteFile(filepath.Join(root, path), []byte("saved content"), 0600); err != nil {
		t.Fatal(err)
	}
	fs, err := filesystem.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	e := &Editor{Filesystem: fs, Page: fs.NewDraft("log/day", nil), Content: []string{"` changed title", "id: abcdef123456", "log content"}}
	e.syncEventBlocks()
	page := fs.Cache[path]
	if page == nil || page.Stage == "draft" || !slices.Equal(page.Content, []string{"saved content"}) || len(fs.Cache) != 1 {
		t.Fatalf("saved page was recreated or overwritten: %+v", page)
	}
}

func TestIdeaIntegerTemplate(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "rand"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "rand", ".`abcdef123456$idea"), []byte("priority: .{int}\n---\n.{?content}"), 0600); err != nil {
		t.Fatal(err)
	}
	fs, err := filesystem.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	e := &Editor{Filesystem: fs, Page: fs.NewDraft("log/day", nil)}
	for _, tc := range []struct{ body, want string }{
		{"text !3 .{int}", "3"},
		{"!0", "0"},
		{"!-12", "-12"},
		{"!+7", "+7"},
		{"!001", "001"},
		{"!999999999999999999999999", "999999999999999999999999"},
		{"!2 !5", "2"},
		{"! !abc !1.5 !+-2 word!4", ".{int}"},
		{"no marker", ".{int}"},
	} {
		e.Content = []string{"` idea !99", "id: abcdef123456", "body", tc.body, "", "!88"}
		e.syncEventBlocks()
		page := fs.Cache["rand/`abcdef123456$idea-99"]
		want := []string{"priority: " + tc.want, "---", "body", tc.body}
		if page == nil || !slices.Equal(page.Content, want) {
			t.Fatalf("body %q: got %+v, want %q", tc.body, page, want)
		}
	}
}

func TestRecordTemplates(t *testing.T) {
	for _, tc := range []struct{ marker, kind, title, suffix string }{
		{"~", "task", "uma tarefa", "uma-tarefa"},
		{">", "plan", "trabalho @Projeto$42/parte", "trabalho-projeto-42-parte"},
		{"=", "session", "trabalho @Projeto/parte", "trabalho-projeto-parte"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, "rec"), 0700); err != nil {
				t.Fatal(err)
			}
			source := "kind: " + tc.kind + "\nid: .{abcdef123456}\npriority: .{int}\ndate: .{yyyy.mm.dd}\n---\n.{?content}"
			template := filepath.Join(root, "rec", "."+tc.marker+"abcdef123456$"+tc.kind)
			if err := os.WriteFile(template, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			fs, err := filesystem.Open(root)
			if err != nil {
				t.Fatal(err)
			}
			e := &Editor{Filesystem: fs, Page: fs.NewDraft("log/day", nil), Content: []string{tc.marker + " " + tc.title, "id: 123456abcdef", "body !7 .{abcdef123456}", "", "unrelated !99"}}
			e.ensureLogEvent()
			page := fs.Cache["rec/"+tc.marker+"123456abcdef$"+tc.suffix]
			want := []string{"kind: " + tc.kind, "id: 123456abcdef", "priority: 7", "date: " + time.Now().Format("2006.01.02"), "---", "body !7 .{abcdef123456}"}
			if page == nil || !slices.Equal(page.Content, want) {
				t.Fatalf("got %+v, want %q", page, want)
			}
			e.Cursor[1] = 3
			e.Content[3] = "updated !2"
			e.syncEventBlocks()
			want[2], want[5] = "priority: 2", "updated !2"
			if !slices.Equal(page.Content, want) {
				t.Fatalf("draft not updated: %q", page.Content)
			}
			for _, stage := range []string{"edit", "local", "api", "move"} {
				page.Stage = stage
				e.Content[3] = "later !9"
				e.syncEventBlocks()
				if !slices.Equal(page.Content, want) {
					t.Fatalf("overwrote %s page", stage)
				}
			}
			data, err := os.ReadFile(template)
			if err != nil || string(data) != source {
				t.Fatal("source template changed")
			}
		})
	}
}
