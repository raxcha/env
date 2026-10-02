package filesystem

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestDailyDraftTemplateAndInheritance(t *testing.T) {
	for _, directory := range []bool{false, true} {
		t.Run(map[bool]string{false: "files", true: "directories"}[directory], func(t *testing.T) {
			root := t.TempDir()
			write := func(path, content string) {
				t.Helper()
				if directory {
					path += "/index"
				}
				path = filepath.Join(root, path)
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			template := []string{"title: daily", "---", "all template content", ""}
			write("proj/@main$42/.basics", strings.Join(template, "\n"))
			write("proj/@main$invalid/.basics", "invalid suffix")
			write("proj/@main/.basics", "legacy template")
			write("log/2026.09.30", "ordinary\n\n`` note\ndetail\n\n-- task\nchild\n09:00\nignored\n\n>> next\n---\n~~ later\n2026.09.30\n== session\nid: retained\n\n- single\n")
			f, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			req := f.NewLoading()
			req.Path, req.CreateDraft = "log/2026.10.01", true
			ok, page := f.LoadLocal(req)
			want := append(slices.Clone(template), "`` note", "detail", "", "-- task", "child", "", ">> next", "", "~~ later", "", "== session", "id: retained")
			if !ok || !slices.Equal(page.Content, want) {
				t.Fatalf("got %#v; want %#v", page, want)
			}
			page.Content = append(page.Content, "pending edit")
			_, again := f.LoadLocal(req)
			if again != page || !slices.Equal(again.Content, append(want, "pending edit")) {
				t.Fatal("reopening changed draft")
			}
			if _, err := os.Stat(filepath.Join(root, req.Path)); !os.IsNotExist(err) {
				t.Fatal("draft persisted before saving")
			}
			write(req.Path, "existing day")
			_, existing := f.LoadLocal(req)
			if !slices.Equal(existing.Content, []string{"existing day"}) {
				t.Fatal("existing page was regenerated")
			}
		})
	}
}

func TestDailyDraftMissingSources(t *testing.T) {
	f, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"log/2026.01.01", "log/invalid", "proj/example"} {
		req := f.NewLoading()
		req.Path, req.CreateDraft = path, true
		ok, page := f.LoadLocal(req)
		if !ok || !slices.Equal(page.Content, []string{""}) {
			t.Fatalf("unexpected draft for %s: %#v", path, page)
		}
	}
}
