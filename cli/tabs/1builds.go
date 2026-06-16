package tabs

import (
	"env/cli"
	"env/engine"
)

func CreateTabs (parent cli.Parent) *Tabs {
	
	t := Tabs {
		Parent: parent,
		Utilities: parent.GetUtilities(),
		Bounds: engine.Boundaries{},

		Name: "tabs",
		Spec: ".",
		
		Switch: false,
		On: true,
		Float: false,
	}
	return &t
}
