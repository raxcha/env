package filesystem

import (
	"env/utils"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// dailyDraftContent runs only when a missing page is first loaded as a draft.
// The caller holds cacheMutex, so cached source pages can be read directly.
func (f *Filesystem) dailyDraftContent(path string) []string {
	date, err := time.Parse("2006.01.02", filepath.Base(path))
	if filepath.Dir(path) != "log" || err != nil {
		return []string{""}
	}
	read := func(path string) []string {
		if page := f.Cache[path]; page != nil && page.Stage != "ghost" && page.Stage != "doom" {
			return append([]string(nil), page.Content...)
		}
		_, _, absolute := f.standardizePaths(path)
		if f.checkPath(absolute) == "dir" {
			absolute = filepath.Join(absolute, "index")
		}
		return f.getContent(absolute)
	}
	content := utils.ExpandTemplate(read(f.dailyTemplatePath()), time.Now())
	previous := read("log/" + date.AddDate(0, 0, -1).Format("2006.01.02"))
	for i := 0; i < len(previous); i++ {
		line := strings.TrimSpace(previous[i])
		if dailyBoundary(line) || len(line) < 2 || line[0] != line[1] || !strings.ContainsRune("`->~=", rune(line[0])) {
			continue
		}
		end := i + 1
		for end < len(previous) && !dailyBoundary(previous[end]) {
			end++
		}
		if len(content) > 0 && strings.TrimSpace(content[len(content)-1]) != "" {
			content = append(content, "")
		}
		content = append(content, previous[i:end]...)
		i = end - 1
	}
	if len(content) == 0 {
		return []string{""}
	}
	return content
}

// dailyTemplatePath resolves the main project's integer priority suffix.
// The caller holds cacheMutex.
func (f *Filesystem) dailyTemplatePath() string {
	projects := map[string]bool{}
	add := func(name string) {
		if suffix, ok := strings.CutPrefix(name, "@main$"); ok {
			if _, err := strconv.Atoi(suffix); err == nil {
				projects[name] = true
			}
		}
	}
	entries, _ := os.ReadDir(filepath.Join(f.whereIsRoot(), "proj"))
	for _, entry := range entries {
		if entry.IsDir() {
			add(entry.Name())
		}
	}
	for path, page := range f.Cache {
		if page == nil || page.Stage == "ghost" || page.Stage == "doom" {
			continue
		}
		parts := strings.Split(filepath.ToSlash(path), "/")
		if len(parts) >= 2 && parts[0] == "proj" {
			add(parts[1])
		}
	}
	names := make([]string, 0, len(projects))
	for name := range projects {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) > 0 {
		return "proj/" + names[0] + "/.basics"
	}
	return "proj/@main/.basics"
}

func dailyBoundary(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" || utils.DashedLine(line) {
		return true
	}
	for _, layout := range []string{"15:04", "2006.01.02", "2006.01.02 15:04", "15:04 2006.01.02"} {
		if _, err := time.Parse(layout, line); err == nil {
			return true
		}
	}
	return false
}
