package launcher

import (
	"env/cli"
	"env/filesystem"
	"env/settings"
	"env/types"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type testParent struct {
	cli.Parent
	settings *settings.Settings
	fs       *filesystem.Filesystem
	opened   string
	follow   bool
	template string
}

func (p *testParent) GetSettings() *settings.Settings          { return p.settings }
func (p *testParent) GetFilesystem() *filesystem.Filesystem    { return p.fs }
func (p *testParent) AddClients(arg, mode string, follow bool) { p.opened = arg; p.follow = follow }

func (p *testParent) PrepareTemplate(reference string) bool { p.template = reference; return true }

func TestLauncherActions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, name := range []string{
		"proj/@low$2/.one", "proj/@high$99/.z", "proj/@high$99/.a/index",
		"proj/@high$99/ordinary", "proj/@high$99/nested/.excluded",
		"proj/.@project/.excluded", "proj/other/.excluded",
	} {
		path := filepath.Join(home, "prsnlspc", name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	parent := &testParent{fs: &filesystem.Filesystem{}}
	launcher := CreateLauncher(parent)
	launcher.Open("default")
	launcher.enter(0)
	if parent.opened != "today" || !parent.follow || launcher.On {
		t.Fatal("today did not open and close the launcher")
	}
	launcher.Open("default")
	launcher.move(1)
	launcher.enter(0)
	var labels []string
	for _, item := range launcher.Items {
		labels = append(labels, item.Label)
	}
	want := []string{"@high$99/.a", "@high$99/.z", "@low$2/.one"}
	if !reflect.DeepEqual(labels, want) {
		t.Fatalf("templates = %v, want %v", labels, want)
	}
	if launcher.Focus != 0 || launcher.Spec != "templates" {
		t.Fatal("submenu did not reset focus")
	}
	launcher.enter(0)
	if parent.template != "@high$99/.a" || launcher.On {
		t.Fatal("template selection did not prepare and close")
	}
	launcher.Open("templates")
	launcher.char('é')
	if len(launcher.Items) != 0 {
		t.Fatal("filter should have no matches")
	}
	launcher.enter(0) // Empty lists must be safe.
	launcher.backspace()
	if launcher.Prompt != "" || len(launcher.Items) != 3 {
		t.Fatal("backspace did not restore templates")
	}
}

func TestThemePreviewAndSelection(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, "prsnlspc")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".options")
	original := "username: user\ntheme: Alpha\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	a := types.Theme{Background: types.Color{R: 1}}
	b := types.Theme{Background: types.Color{R: 2}}
	s := &settings.Settings{Options: map[string]any{"theme": "Alpha"}, Themes: map[string]types.Theme{"Beta": b, "Alpha": a}, Theme: a}
	l := CreateLauncher(&testParent{fs: &filesystem.Filesystem{}, settings: s})
	l.Open("default")
	l.Focus = 6
	l.enter(0)
	if len(l.Items) != 2 || l.Items[0].Label != "Alpha" {
		t.Fatal("missing or unsorted themes")
	}
	l.move(1)
	if s.Theme != b {
		t.Fatal("navigation did not preview theme")
	}
	data, _ := os.ReadFile(path)
	if string(data) != original {
		t.Fatal("preview saved prematurely")
	}
	l.Close()
	if s.Theme != a {
		t.Fatal("cancel did not restore theme")
	}
	l.Open("themes")
	l.char('b')
	if s.Theme != b {
		t.Fatal("filter did not preview theme")
	}
	l.enter(0)
	data, _ = os.ReadFile(path)
	if l.On || s.Options["theme"] != "Beta" || string(data) != "username: user\ntheme: Beta\n" {
		t.Fatalf("selection not saved: %s", data)
	}
	l.Open("themes")
	l.char('z')
	l.enter(0)
	l.backspace()
	l.Close()
}
