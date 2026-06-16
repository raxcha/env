package projects

import (
	"env/cli"
	"env/engine"
	"env/filesystem"
	"env/utilities"
)

type Projects struct {
	Parent    cli.Parent
	Bounds    engine.Boundaries
	Utilities *utilities.Utilities

	Name string
	Path string

	ProjectsPage *filesystem.Page
	EventsPage   *filesystem.Page

	Items     []*ProjectItem
	Selected  int
	Prompt    string
	PanelMode string
	Context   ProjectContext

	On    bool
	Float bool
}

type ProjectContext struct {
	Kind        string // root, project, templates, events
	ProjectPath string
	ProjectName string
}

type ProjectItem struct {
	Page     *filesystem.Page
	Kind     string // back, new-project, project, virtual-templates, virtual-events, add-template, template, event
	Name     string
	Path     string
	Priority int
	Depth    int
}

type projectRect struct {
	X int
	Y int
	W int
	H int
}
