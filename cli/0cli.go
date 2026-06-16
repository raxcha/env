package cli

import (
	"env/engine"
	"env/filesystem"
	"env/routines"
)

type Cli interface {
	Resize(*engine.Boundaries)
	Input(*routines.Input)
	Draw() *engine.Queue

	GetName() string
	GetPath() string
	GetKind() string
	GetTitle() string
	GetFloating() bool
	GetBounds() *engine.Boundaries

	SetPage(*filesystem.Page)
	AcceptsPage(*filesystem.Page) bool

	GetSpec() string
	IsSidebarOn() bool
	GetSidebarBounds() *engine.Boundaries
}
