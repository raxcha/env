package editor

import (
	"env/engine"
	"env/filesystem"
	"path/filepath"
	"strings"
)

func (e *Editor) GetName() string {
	return e.Name
}

func (e *Editor) GetPath() string {
	return e.Path
}

func (e *Editor) GetSpec() string {
	return e.Path
}

func (e *Editor) GetKind() string {
	return "editor"
}

func (e *Editor) GetFloating() bool {
	return e.Float
}

func (e *Editor) GetBounds() *engine.Boundaries {
	return e.Bounds
}

func (e *Editor) IsSidebarOn() bool {
	return false
}

func (e *Editor) GetSidebarBounds() *engine.Boundaries {
	if e.Sidebar == nil {
		return nil
	}
	return &e.Sidebar.Bounds
}

func (e *Editor) SetPage(page *filesystem.Page) {
	if page == nil {
		return
	}

	if page.Path == "error" {
		if e.Sidebar != nil && e.Sidebar.Path != "" {
			e.Sidebar.SetPage(&filesystem.Page{
				Name:     filepath.Base(e.Sidebar.Path),
				Path:     e.Sidebar.Path,
				Type:     "deep",
				Options:  map[string]any{},
				Content:  []string{},
				Metadata: map[string]any{},
				Children: []*filesystem.Page{},
				Stage:    "draft",
				Sorting:  e.Sidebar.Sort,
			})
		}

		return
	}

	if e.Sidebar != nil &&
		strings.HasPrefix(e.Sidebar.Path, "e.proj/") &&
		page.Path == "b.rec" {

		e.Sidebar.SetEvents(page)

		if e.Sidebar.Switch {
			e.Sidebar.RequestSelectedPreview()
		}

		return
	}

	// Preview do sidebar.
	if e.Sidebar != nil && page.Path == e.Sidebar.PreviewPath && page.Path != e.Path {
		e.Sidebar.SetPreview(page)
		return
	}

	// resto continua igual...

	// Preview do sidebar.
	if e.Sidebar != nil && page.Path == e.Sidebar.PreviewPath && page.Path != e.Path {
		e.Sidebar.SetPreview(page)
		return
	}

	if page.Path == e.Path {
		e.applyPage(page, false)
		e.saveState()
		return
	}

	if e.pageHasSameID(page) {
		e.applyPage(page, true)
		e.saveState()
		return
	}

	if e.Sidebar != nil && page.Path == e.Sidebar.Path {
		e.Sidebar.SetPage(page)

		if e.Sidebar.Switch {
			e.Sidebar.RequestSelectedPreview()
		}

		return
	}
}

func (e *Editor) AcceptsPage(page *filesystem.Page) bool {
	if e == nil || page == nil {
		return false
	}

	if page.Path == "error" || page.Path == e.Path || e.pageHasSameID(page) {
		return true
	}

	if e.Sidebar == nil {
		return false
	}

	if strings.HasPrefix(e.Sidebar.Path, "e.proj/") && page.Path == "b.rec" {
		return true
	}

	return page.Path == e.Sidebar.Path || page.Path == e.Sidebar.PreviewPath
}

func (e *Editor) applyPage(page *filesystem.Page, adoptPath bool) {
	e.Page = page

	if adoptPath {
		e.Path = page.Path
		e.Name = filepath.Base(page.Path)
	}

	if len(page.Content) > 0 {
		e.Content = ensureEditableContent(page.Content)
		e.Page.Content = e.Content
	}
	e.AppliedContent = copyLines(e.Content)

	if e.Sidebar != nil {
		e.Sidebar.CurrentPath = page.Path
		if adoptPath {
			e.Sidebar.Path = sidebarRootPath(page.Path)
		}

		// Caso e.Path == e.Sidebar.Path, como em e.proj/projeto,
		// usa a própria leitura como conteúdo do sidebar,
		// mas NÃO pede outra leitura de sidebar aqui.
		if e.Sidebar.Path == page.Path {
			e.Sidebar.SetPage(page)

			if e.Sidebar.Switch {
				e.Sidebar.RequestSelectedPreview()
			}
		}
	}
}

func (e *Editor) pageHasSameID(page *filesystem.Page) bool {
	if page == nil {
		return false
	}

	if !editorPathIsDerivedBlockPage(e.Path) {
		return false
	}

	currentID := firstContentID(e.Content)
	if currentID == "" && e.Page != nil {
		currentID = firstContentID(e.Page.Content)
	}

	return currentID != "" && currentID == firstContentID(page.Content)
}

func editorPathIsDerivedBlockPage(path string) bool {
	if !strings.HasPrefix(path, "b.rec/") && !strings.HasPrefix(path, "c.rand/") {
		return false
	}

	if strings.HasSuffix(path, "/.reminders") || path == "b.rec/.reminders" || path == "c.rand/.reminders" {
		return false
	}

	return true
}

func firstContentID(content []string) string {
	for _, line := range content {
		if isIdLine(line) {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "id:"))
		}
	}

	return ""
}

func (e *Editor) GetTitle() string {
	if e.Page != nil && e.Page.Name != "" {
		return e.Page.Name
	}
	return e.Name
}

func (e Editor) GetDisplayName() string {
	path := e.Path
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		path = strings.Join(parts[:2], "/") + "/.../" + parts[len(parts)-1]
	}
	return path
}
