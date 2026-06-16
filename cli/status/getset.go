package status

import (
	"env/engine"
	"env/filesystem"
)

func (s *Status) GetName() string {
	return s.Name
}

func (s *Status) GetPath() string {
	return s.Path
}

func (s *Status) GetKind() string {
	return "status"
}

func (s *Status) GetFloating() bool {
	return s.Float
}

func (s *Status) GetBounds() *engine.Boundaries {
	return s.Bounds
}

func (s *Status) SetPage(page *filesystem.Page) {}

func (s *Status) GetSpec() string                      { return "status" }
func (s *Status) IsSidebarOn() bool                    { return false }
func (s *Status) GetSidebarBounds() *engine.Boundaries { return nil }