package utilities

import (
	"env/engine"
	"os"
	"path/filepath"
	"strings"
)

func CreateUtilities() *Utilities {
	u := Utilities{}
	rawThemes, _ := LoadRawThemes("themes.json")

	themeMap := make(map[string]*Theme, len(rawThemes))
	themes := make([]*Theme, 0, len(rawThemes))

	for _, raw := range rawThemes {
		t := mapTheme(raw)
		themes = append(themes, t)
		if t.Name != "" {
			themeMap[t.Name] = t
		}
	}

	u.Themes = themes
	u.Theme = selectStartupTheme(themeMap, "Spacemacs Dark")

	return &u
}

func selectStartupTheme(themeMap map[string]*Theme, fallback string) *Theme {
	if theme := themeByName(themeMap, rootOptionsThemeName()); theme != nil {
		return theme
	}
	if theme := themeByName(themeMap, fallback); theme != nil {
		return theme
	}
	for _, theme := range themeMap {
		return theme
	}
	return nil
}

func themeByName(themeMap map[string]*Theme, name string) *Theme {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	if theme := themeMap[name]; theme != nil {
		return theme
	}
	for themeName, theme := range themeMap {
		if strings.EqualFold(strings.TrimSpace(themeName), name) {
			return theme
		}
	}
	return nil
}

func rootOptionsThemeName() string {
	data, err := os.ReadFile(filepath.Join(personalRootPath(), ".options"))
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "---" {
			break
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok || !strings.EqualFold(strings.TrimSpace(key), "theme") {
			continue
		}
		return strings.TrimSpace(value)
	}

	return ""
}

func personalRootPath() string {
	if root := strings.TrimSpace(os.Getenv("ENV_ROOT")); root != "" {
		return filepath.Clean(root)
	}

	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		home = "/home/asdf"
	}

	return filepath.Join(home, "prsnl.spc")
}

func newStyle(t *Theme) *Style {
	return &Style{
		Bold:      false,
		Italic:    false,
		Underline: false,
		Fg:        &t.Foreground,
		Bg:        &t.Background,
		StdFg:     &t.Foreground,
		StdBg:     &t.Background,
		Uppercase: false,
		Almost:    false,
		All:       false,
		Visible:   true,
		Wrap:      0,
	}
}

func newCell(stl *Style) *engine.Cell {
	c := engine.Cell{
		Char:      ' ',
		Bold:      stl.Bold,
		Italic:    stl.Italic,
		Underline: stl.Underline,
		Fg:        stl.Fg,
		Bg:        stl.Bg,
		Uppercase: stl.Uppercase,
		Visible:   stl.Visible,
	}

	if stl.All {
		c.Bold = true
		c.Italic = true
		c.Underline = true
	}

	if stl.Almost {
		c.Bold = true
		c.Italic = false
		c.Underline = true
	}

	return &c

}

var invisibleCell = engine.Cell{Visible: false}

func newFrame() *engine.Frame {
	return &engine.Frame{
		Size:    []int{},
		Cells:   []engine.Cell{},
		Timeout: 1000,
	}
}

func NewQueue() *engine.Queue {
	return &engine.Queue{
		Size:   []int{},
		Frames: []engine.Frame{},
		Cycle:  false,
	}
}
