package picker

import "env/engine"

func (p *Picker) Resize (newbounds *engine.Boundaries) {
	
	p.mergeBoundaries(newbounds)
}

func (p *Picker) mergeBoundaries (bounds *engine.Boundaries) {
	
	bounds.ActualPos = bounds.Pos
	bounds.ActualSize = bounds.Size
	p.Bounds = *bounds
} 