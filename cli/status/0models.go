package status

import (
	"env/cli"
	"env/engine"
	"env/utilities"
)

type Status struct {
	Parent    cli.Parent
	Bounds    *engine.Boundaries
	Utilities *utilities.Utilities

	Name string
	Path string

	On    bool
	Float bool
}