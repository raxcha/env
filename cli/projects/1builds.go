package projects

import (
	"env/cli"
	"env/engine"
	"env/filesystem"
	"path/filepath"
	"strings"
)

func CreateProjects(path string, parent cli.Parent) *Projects {
	path = normalizeProjectsPath(path)

	p := Projects{
		Parent:    parent,
		Bounds:    engine.Boundaries{},
		Utilities: parent.GetUtilities(),

		Name: "projects",
		Path: path,

		Items:    []*ProjectItem{},
		Selected: 0,
		Prompt:   "",
		Context:  ProjectContext{Kind: "root"},

		On:    true,
		Float: false,
	}

	p.requestProjects()
	p.requestBaseTemplates()
	p.requestEvents()
	p.setInitialContext(path)
	p.updateItems()

	return &p
}

func normalizeProjectsPath(path string) string {
	path = strings.TrimSpace(strings.Trim(path, "/"))
	path = filepath.ToSlash(filepath.Clean(path))

	if path == "" || path == "." || path == "..." {
		return "e.proj"
	}

	if path == "e.proj" || strings.HasPrefix(path, "e.proj/") {
		return path
	}

	return "e.proj/" + path
}

func (p *Projects) setInitialContext(path string) {
	path = normalizeProjectsPath(path)
	p.Path = path

	if path == "e.proj" {
		p.Context = ProjectContext{Kind: "root"}
		return
	}

	p.Context = ProjectContext{
		Kind:        "project",
		ProjectPath: path,
		ProjectName: filepath.Base(path),
	}
}

func (p *Projects) requestProjects() {
	if p == nil || p.Parent == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	req.Path = "e.proj"
	req.Depth = -1
	req.Cont = true
	req.Meta = true
	req.Opts = true
	req.Sort = "priority"
	fs.Load(req)
}

func (p *Projects) requestEvents() {
	if p == nil || p.Parent == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	req.Path = "b.rec"
	req.Depth = 1
	req.Cont = true
	req.Meta = true
	req.Opts = true
	req.Sort = "time"
	fs.Load(req)
}

func (p *Projects) requestBaseTemplates() {
	if p == nil || p.Parent == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	req.Path = "e.proj/.templates"
	req.Depth = 1
	req.Cont = true
	req.Meta = true
	req.Opts = true
	req.Sort = "basic"
	fs.Load(req)
}
