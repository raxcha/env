package routines

type Routines struct {

	Size chan *Bound
	Input chan *Input 
}

type Bound []int

type Input struct {

	Key string
	Char rune
}
