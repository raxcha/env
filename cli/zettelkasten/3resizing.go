package zettelkasten

import "env/engine"

func (z *Zettelkasten) Resize(newbounds *engine.Boundaries) {
	if newbounds == nil {
		return
	}
	newbounds.ActualPos = newbounds.Pos
	newbounds.ActualSize = newbounds.Size
	z.Bounds = *newbounds
}
