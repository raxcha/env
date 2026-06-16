package status

import (
	"env/cli"
	"env/engine"
)

func CreateStatus(parent cli.Parent) *Status {
	return &Status{
		Parent:    parent,
		Bounds:    &engine.Boundaries{},
		Utilities: parent.GetUtilities(),

		Name: "status",
		Path: "status",

		On:    true,
		Float: true,
	}
}