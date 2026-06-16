package engine

import "env/routines"

type Engine struct {

	Queue chan Queue
	Frame chan Frame
	LastFrame Frame 
	Diff chan Diff
	stop chan struct{}
}

type Queue struct {

	Size routines.Bound
	Frames []Frame
	Cycle bool
}

type Frame struct {

	Size routines.Bound
	Cells []Cell
	Timeout int
}

type Diff struct {

	Size routines.Bound
	Cells []Cell
	Indexing []int
}

type Cell struct {

	Char rune
	Fg *RGB
	Bg *RGB
	Bold, Italic, Underline bool
	Visible, Uppercase bool
}

type Boundaries struct {

	Fullsize routines.Bound
	Pos routines.Bound
	Size routines.Bound

	Mode string
	Index int

	ActualSize routines.Bound
	ActualPos routines.Bound
}

type RGB struct {
	R, G, B int
}