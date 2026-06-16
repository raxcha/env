package picker

import (
	"env/engine"
	"env/filesystem"
)

func (p *Picker) GetName() string {
	return p.Name
}

func (p *Picker) GetPath() string {
	return p.Path
}

func (p *Picker) GetKind() string {
	return "picker"
}

func (p *Picker) GetFloating() bool {
	return p.Float
}

func (p *Picker) GetBounds() *engine.Boundaries {
	return &p.Bounds
}

func (p *Picker) SetPage(page *filesystem.Page) {

	if p.Parent == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil || fs.Cache == nil {
		return
	}

	selectedPath := ""

	if len(p.Items) > 0 && p.Selected >= 0 && p.Selected < len(p.Items) {
		if p.Items[p.Selected] != nil && p.Items[p.Selected].Page != nil {
			selectedPath = p.Items[p.Selected].Page.Path
		}
	}

	p.Cache = fs.Cache

	if selectedPath != "" {
		p.StartPath = selectedPath
	}

	p.updateItems()
}

func (p *Picker) AcceptsPage(page *filesystem.Page) bool {
	return p != nil && page != nil
}

func (p *Picker) GetSpec() string {
	return p.Path
}

func (p *Picker) GetTitle() string                     { return p.Name }
func (p *Picker) IsSidebarOn() bool                    { return false }
func (p *Picker) GetSidebarBounds() *engine.Boundaries { return nil }
