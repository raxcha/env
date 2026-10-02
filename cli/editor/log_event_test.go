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
	for _, tc := range []struct {
		marker, root, prefix string
		timed                bool
	}{{"`", "rand", "`", false}, {">", "rec", ">", false}, {"=", "rec", "=", false}, {"=", "rec", "=", true}, {"~", "rec", "~", false}} {
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
			if tc.timed {
				e.Content = append([]string{"08:32"}, e.Content...)
				e.Cursor[1] = 1
			}
			e.savePage()
			id := strings.TrimPrefix(e.Content[2], "id: ")
			title := "o-que-está-escrito-na-linha"
			path := tc.root + "/" + tc.prefix + id + "$" + title
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
				if strings.HasPrefix(key, tc.root+"/"+tc.prefix) {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("created %d drafts for one ID", count)
			}
		})
	}
}

func TestLogEventDraftNames(t *testing.T) {
	for _, tc := range []struct{ line, path string }{
		{"~ Nome do título", "rec/~abcdef123456$nome-do-título"},
		{"~ um título com muitas palavras extras", "rec/~abcdef123456$um-título-com-muitas-palavras-extras"},
		{"` Nome do título", "rand/`abcdef123456$nome-do-título"},
		{"` um título com muitas palavras extras", "rand/`abcdef123456$um-título-com-muitas-palavras-extras"},
		{"` extraordinariamentecomprida depois", "rand/`abcdef123456$extraordinariamentecomprida-depois"},
		{"> trabalhar no @Meu-Projeto/tarefa", "rec/>abcdef123456$trabalhar-no-meu-projeto-tarefa"},
		{"> trabalhar no @Meu-Projeto$referencia", "rec/>abcdef123456$trabalhar-no-meu-projeto-referencia"},
		{"> @um-projeto-com-nome-muito-longo tarefa", "rec/>abcdef123456$um-projeto-com-nome-muito-longo-tarefa"},
		{"= @proj-name.template-name", "rec/=abcdef123456$proj-name.template-name"},
		{"= @culture.watch", "rec/=abcdef123456$culture.watch"},
	} {
		t.Run(tc.line, func(t *testing.T) {
			fs, err := filesystem.Open(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			e := Editor{Filesystem: fs, Page: &types.Page{Path: "log/day"}, Content: []string{"08:32", tc.line, "id: abcdef123456"}, Cursor: [3]int{0, 1, 0}}
			e.ensureLogEvent()
			if fs.Cache[tc.path] == nil {
				t.Fatalf("expected draft %q, got %v", tc.path, fs.Cache)
			}
		})
	}
}
