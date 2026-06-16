package tabs

import (

	// external dependencies ...
	"env/cli"
	"env/engine"
	"env/utilities"
)

// ...
type Tabs struct {

	// a reference to the master struct ...
	Parent cli.Parent

	Utilities *utilities.Utilities
	// direct reference to a very useful module ...

	Bounds engine.Boundaries // for size management ...

	// basic info ...
	Name string
	Spec string

	// basic boolean states ...
	Switch bool
	On     bool
	Float  bool
}
