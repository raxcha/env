package editor

import "env/engine"

func (e *Editor) Resize(newbounds *engine.Boundaries) {
	
	e.mergeBoundaries(newbounds)

	
	e.Sidebar.Resize(newbounds)
}

func (e *Editor) mergeBoundaries (bounds *engine.Boundaries) {
	
	bounds.ActualPos = bounds.Pos
	bounds.ActualSize = bounds.Size
	e.Bounds = bounds
} 