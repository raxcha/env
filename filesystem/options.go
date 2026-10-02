package filesystem

import (
	"os"
	"path/filepath"
	"strings"

	"env/utils"
)

// SaveTheme updates only the theme option, preserving the other file contents.
func (f *Filesystem) SaveTheme(name string) error {
	path := filepath.Join(f.whereIsRoot(), ".options")
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	newline := "\n"
	if strings.Contains(string(data), "\r\n") {
		newline = "\r\n"
	}
	lines := strings.Split(string(data), newline)
	found := false
	for i, line := range lines {
		if utils.DashedLine(line) {
			break
		}
		key, _ := utils.SplitTwo(line, ":")
		if key == "theme" {
			lines[i] = "theme: " + name
			found = true
		}
	}
	if !found {
		lines = append([]string{"theme: " + name}, lines...)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, newline)), 0600); err != nil {
		return err
	}
	if f.Options == nil {
		f.Options = map[string]any{}
	}
	f.Options["theme"] = name
	return nil
}
