package status

import "env/engine"

func (s *Status) Resize(bounds *engine.Boundaries) {
	if bounds == nil {
		return
	}

	s.Bounds = bounds
}