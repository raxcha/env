package sidebar

import (
	"env/engine"
)

func (t *Sidebar) GetName() string {
	return "sidebar"
}

func (t *Sidebar) GetPath() string {
	return t.Path
}

func (t *Sidebar) GetKind() string {
	return "sidebar"
}

func (t *Sidebar) GetFloating() bool {
	return t.Float
}

func (t *Sidebar) GetBounds() *engine.Boundaries {
	return &t.Bounds
}
