package filesystem

import (
	"env/types"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// copyPlanDraft reserves the next plan number, including pending saves.
func (f *Filesystem) copyPlanDraft(page *types.Page) *types.Page {
	path := filepath.ToSlash(filepath.Clean(page.Path))
	if filepath.Dir(path) != "rec" || !strings.HasPrefix(filepath.Base(path), ">") {
		return nil
	}
	switch page.Stage {
	case "draft", "edit", "local", "*api", "api":
	default:
		return nil
	}
	_, project, ok := strings.Cut(filepath.Base(path), "$")
	if !ok || project == "" || strings.ContainsAny(project, "/\\\x00") {
		return nil
	}
	f.cacheMutex.Lock()
	defer f.cacheMutex.Unlock()
	name := "@" + project
	candidates := map[string]bool{}
	addProject := func(candidate string) {
		base, suffix, hasSuffix := strings.Cut(candidate, "$")
		if !strings.EqualFold(base, name) {
			return
		}
		if hasSuffix {
			if _, err := strconv.Atoi(suffix); err != nil {
				return
			}
		}
		candidates[candidate] = true
	}
	entries, _ := os.ReadDir(filepath.Join(f.whereIsRoot(), "proj"))
	for _, entry := range entries {
		addProject(entry.Name())
	}
	for path, cached := range f.Cache {
		parts := strings.Split(filepath.ToSlash(path), "/")
		if len(parts) >= 2 && parts[0] == "proj" && cached != nil && cached.Stage != "doom" && cached.Stage != "ghost" {
			addProject(parts[1])
		}
	}
	names := make([]string, 0, len(candidates))
	for candidate := range candidates {
		names = append(names, candidate)
	}
	sort.Strings(names)
	if len(names) > 0 {
		name = names[0]
	}
	directory := "proj/" + name
	source := "plan-source: " + path
	content := append([]string{source}, page.Content...)
	update := func(path string, existing []string, deep bool) *types.Page {
		if len(existing) == 0 || existing[0] != source {
			return nil
		}
		copy := f.Cache[path]
		if copy == nil {
			copy = f.NewDraft(path, nil)
			copy.Stage = "edit"
			if deep {
				copy.Type = "deep"
			}
			if f.Cache == nil {
				f.Cache = map[string]*types.Page{}
			}
			f.Cache[path] = copy
		}
		f.EditPage(copy, content)
		return copy
	}
	for path, cached := range f.Cache {
		if filepath.ToSlash(filepath.Dir(path)) == directory && strings.HasPrefix(filepath.Base(path), ">0.") && cached != nil && cached.Stage != "doom" && cached.Stage != "ghost" {
			if copy := update(path, cached.Content, cached.Type == "deep"); copy != nil {
				return copy
			}
		}
	}
	next := 1
	count := func(name string) {
		if suffix, ok := strings.CutPrefix(name, ">0."); ok {
			if n, err := strconv.Atoi(suffix); err == nil && n >= next {
				next = n + 1
			}
		}
	}
	entries, _ = os.ReadDir(filepath.Join(f.whereIsRoot(), directory))
	for _, entry := range entries {
		count(entry.Name())
		if strings.HasPrefix(entry.Name(), ">0.") {
			path := directory + "/" + entry.Name()
			absolute := filepath.Join(f.whereIsRoot(), path)
			if entry.IsDir() {
				absolute = filepath.Join(absolute, "index")
			}
			if copy := update(path, f.getContent(absolute), entry.IsDir()); copy != nil {
				return copy
			}
		}
	}
	for path := range f.Cache {
		if filepath.ToSlash(filepath.Dir(path)) == directory {
			count(filepath.Base(path))
		}
	}
	copy := f.NewDraft(directory+"/>0."+strconv.Itoa(next), nil)
	copy.Content = content
	if f.Cache == nil {
		f.Cache = map[string]*types.Page{}
	}
	f.Cache[copy.Path] = copy
	return copy
}
