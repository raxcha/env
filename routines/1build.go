package routines

func CreateRoutines () *Routines {

	r := Routines {
		Size: make(chan *Bound),
		Input: make(chan *Input),
	}

	r.prepareTerminal()
	r.startResizing()
	r.startInputting()
	

	return &r
}