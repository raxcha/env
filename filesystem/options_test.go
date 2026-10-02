package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveThemeReload(t *testing.T) {
	for _, content := range []string{"", "username: user\n---\nbody\n", "theme: Old\r\nusername: user\r\n"} {
		t.Run(content, func(t *testing.T) {
			f := &Filesystem{root: t.TempDir()}
			path := filepath.Join(f.root, ".options")
			if content != "" {
				if err := os.WriteFile(path, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := f.SaveTheme("Tokyo Night"); err != nil {
				t.Fatal(err)
			}
			f.reload()
			if f.Options["theme"] != "Tokyo Night" {
				t.Fatalf("theme lost on reload: %v", f.Options)
			}
		})
	}
}
