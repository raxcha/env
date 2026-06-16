package picker

import (
	"env/cli"
	"env/engine"
	"env/filesystem"
)

func CreatePicker(path string, parent cli.Parent) *Picker {

	if path == "" {
		path = "."
	}

	cache := &filesystem.Page{}
	if parent != nil && parent.GetFilesystem() != nil && parent.GetFilesystem().Cache != nil {
		cache = parent.GetFilesystem().Cache
	}

	p := Picker{
		Parent: parent,
		Bounds: engine.Boundaries{},
		Utilities: parent.GetUtilities(),

		Cache: cache,
		Flat: []*filesystem.Page{},
		Items: []*Match{},
		Selected: 0,
		Prompt: "",
		Mode: "literal",
		Sorting: "auto",
		Scope: "context",

		Name: "picker",
		Path: path,
		StartPath: path,

		Switch: false,
		On: true,
		Float: false,
	}

	p.RefreshFlat()
	return &p
}
