package sidebar

import (
	"env/cli"
	"env/engine"
)

func CreateSidebar(path string, parent cli.Parent) *Sidebar {
	s := Sidebar {
		Parent: parent,
		Bounds: engine.Boundaries{},
		Utilities: parent.GetUtilities(),

		Name: "sidebar",
		Path: "",
		CurrentPath: path,
		Page: nil,

		Sort: "basic",

		Switch: false,
		On: true,
		Float: false,
	}
	return &s
}