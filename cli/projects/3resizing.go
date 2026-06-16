package projects

import "env/engine"

func (p *Projects) Resize(newbounds *engine.Boundaries) {
	p.mergeBoundaries(newbounds)
}

func (p *Projects) mergeBoundaries(bounds *engine.Boundaries) {
	bounds.ActualPos = bounds.Pos
	bounds.ActualSize = bounds.Size
	p.Bounds = *bounds
}
