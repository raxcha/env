package tabs

import "env/engine"

// straight out of master orchestration ...
func (t *Tabs) Resize(newbounds *engine.Boundaries) {

	t.mergeBoundaries(newbounds)
}

// generate actual boundaries ...
func (t *Tabs) mergeBoundaries(bounds *engine.Boundaries) {
	// based on suggested ...

	bounds.ActualPos = bounds.Pos
	bounds.ActualSize = bounds.Size
	t.Bounds = *bounds
}
