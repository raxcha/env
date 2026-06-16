package projects

import (
	"env/engine"
	"env/filesystem"
	"path/filepath"
	"strings"
)

func (p *Projects) GetName() string               { return p.Name }
func (p *Projects) GetPath() string               { return p.Path }
func (p *Projects) GetKind() string               { return "projects" }
func (p *Projects) GetFloating() bool             { return p.Float }
func (p *Projects) GetBounds() *engine.Boundaries { return &p.Bounds }

func (p *Projects) SetPage(page *filesystem.Page) {
	if p == nil || page == nil {
		return
	}

	selectedPath := p.selectedPath()

	if page.Path == "e.proj" {
		p.ProjectsPage = page
		p.refreshProjectNameFromCache()
		p.updateItems()
		p.selectPath(selectedPath)
		return
	}

	if page.Path == "b.rec" {
		p.EventsPage = page
		p.updateItems()
		p.selectPath(selectedPath)
		return
	}

	if strings.HasPrefix(page.Path, "e.proj/") && p.ProjectsPage != nil {
		mergeCachedProjectPage(p.ProjectsPage, page)
		p.refreshProjectNameFromCache()
		p.updateItems()
		p.selectPath(selectedPath)
		return
	}
}

func (p *Projects) AcceptsPage(page *filesystem.Page) bool {
	if p == nil || page == nil {
		return false
	}

	return page.Path == "e.proj" ||
		page.Path == "b.rec" ||
		strings.HasPrefix(page.Path, "e.proj/")
}

func (p *Projects) refreshProjectNameFromCache() {
	if p.Context.ProjectPath == "" || p.ProjectsPage == nil {
		return
	}

	project := findProjectPage(p.ProjectsPage, p.Context.ProjectPath)
	if project == nil {
		return
	}

	p.Context.ProjectName = projectDisplayName(project)
}

func mergeCachedProjectPage(root *filesystem.Page, page *filesystem.Page) {
	if root == nil || page == nil {
		return
	}

	for i, child := range root.Children {
		if child == nil {
			continue
		}

		if filepath.Clean(child.Path) == filepath.Clean(page.Path) {
			root.Children[i] = page
			return
		}

		mergeCachedProjectPage(child, page)
	}
}

func (p *Projects) GetSpec() string {
	return p.Path
}

func (p *Projects) GetTitle() string                     { return p.Name }
func (p *Projects) IsSidebarOn() bool                    { return false }
func (p *Projects) GetSidebarBounds() *engine.Boundaries { return nil }
