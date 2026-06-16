package tabs

import (
	"env/engine"
	"env/filesystem"
)

func (t *Tabs) GetName() string {
	return "tabs"
}

func (t *Tabs) GetPath() string {
	return t.Spec
}

func (t *Tabs) GetKind() string {
	return "tabs"
}

func (t *Tabs) GetFloating() bool {
	return t.Float
}

func (t *Tabs) GetBounds() *engine.Boundaries {
	return &t.Bounds
}

func (t *Tabs) SetPage(page *filesystem.Page) {
}

func (t *Tabs) IsSidebarOn() bool             { return false }
func (t *Tabs) GetSidebarBounds() *engine.Boundaries { return nil }