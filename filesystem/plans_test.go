package filesystem

import (
	"env/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanCopiesPersistAndUpdate(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "proj/@demo$42"), 0755); err != nil {
		t.Fatal(err)
	}
	f, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	save := func(path, content string) {
		t.Helper()
		page := f.NewDraft(path, nil)
		page.Content = []string{content}
		_, patch := f.SyncLocal(&types.Syncing{Branch: page})
		f.ApplyPatch(patch)
	}
	save("rec/>abcdef123456$demo", "first")
	save("rec/>fedcba654321$demo", "second")
	f, err = Open(root)
	if err != nil {
		t.Fatal(err)
	}
	save("rec/>abcdef123456$demo", "updated")
	for name, want := range map[string]string{
		">0.1": "plan-source: rec/>abcdef123456$demo\nupdated",
		">0.2": "plan-source: rec/>fedcba654321$demo\nsecond",
	} {
		data, err := os.ReadFile(filepath.Join(root, "proj/@demo$42", name))
		if err != nil || string(data) != want {
			t.Fatalf("%s = %q, %v; want %q", name, data, err, want)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "proj/@demo$42/>0.3")); !os.IsNotExist(err) {
		t.Fatalf("repeat save created another copy: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "rec/>abcdef123456$demo"))
	if err != nil || string(data) != "updated" {
		t.Fatalf("original = %q, %v", data, err)
	}
}

func TestPlanCopyNumberingIncludesDiskAndPendingCopies(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "proj/@demo/>0.9"), 0755); err != nil {
		t.Fatal(err)
	}
	f, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	for i, id := range []string{"aaaaaaaaaaaa", "bbbbbbbbbbbb"} {
		page := f.NewDraft("rec/>"+id+"$demo", nil)
		page.Content = []string{"body"}
		copy := f.copyPlanDraft(page)
		want := []string{"proj/@demo/>0.10", "proj/@demo/>0.11"}[i]
		if copy == nil || copy.Path != want {
			t.Fatalf("copy = %#v; want %s", copy, want)
		}
		page.Content[0] = "changed"
		if strings.Contains(strings.Join(copy.Content, "\n"), "changed") {
			t.Fatal("copy shares source content")
		}
		if again := f.copyPlanDraft(page); again != copy || again.Content[1] != "changed" {
			t.Fatal("pending copy not updated")
		}
	}
	for _, path := range []string{"rec/=aaaaaaaaaaaa$demo", "rand/>aaaaaaaaaaaa$demo", "rec/>missing-project", "proj/@demo/>0.1"} {
		if copy := f.copyPlanDraft(f.NewDraft(path, nil)); copy != nil {
			t.Fatalf("unexpected copy for %s", path)
		}
	}
}
