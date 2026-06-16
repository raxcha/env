package utilities

import "env/engine"

func (u *Utilities) GenerateQueue (size []int, frames []engine.Frame, cycle bool) *engine.Queue {

	q := engine.Queue{}
	q.Size = size
    	q.Frames = frames
    	q.Cycle = cycle
	return &q
}

// animations ...