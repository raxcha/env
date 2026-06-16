package sidebar

import (
	"env/cli"
	"env/engine"
	"env/filesystem"
	"env/utilities"
)

type Sidebar struct {

	Parent cli.Parent
	Bounds engine.Boundaries
	Utilities *utilities.Utilities

	Name string
	Path string
	CurrentPath string
	Page *filesystem.Page
	EventsPage *filesystem.Page
	Sort string

	Flat []*Node
	Cursor int
	ViewStart int

	PreviewPath string
	Preview *filesystem.Page
	
	Switch bool
	On bool
	Float bool
}

type Node struct {
	
	Page *filesystem.Page
	Parent *Node
	Children []*Node
	Depth int
	Expanded bool
	Virtual bool
}