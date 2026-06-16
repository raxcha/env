package engine


func CreateEngine() *Engine {

	e := Engine {
		Queue: make(chan Queue, 4),
		Frame: make(chan Frame, 4),
		LastFrame: Frame{},
		Diff: make(chan Diff, 4),
		stop: make(chan struct{}),
	}

	e.startOrchestra()
	e.startDiffing()
	e.startPrinting()

	return &e
}