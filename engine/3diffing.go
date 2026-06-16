package engine

import "env/routines"

func (e *Engine) startDiffing () {

	go func() {
		for frame := range e.Frame {

			diff := e.diff(e.LastFrame, frame)

			e.LastFrame = frame
			e.Diff <- diff
		}
	}()
}

func (e *Engine) diff(old Frame, new Frame) Diff {
	
	diff := Diff{
		Size: routines.Bound{},
		Cells: []Cell{},
		Indexing: []int{},
	}
	
	if invalidForDiff(old, new) {

		diff.Size = new.Size
		for i, cell := range new.Cells {
			diff.Cells = append(diff.Cells, cell)
			diff.Indexing = append(diff.Indexing, i)
		}
		return diff
	}
		
	diff.Size = old.Size

	limit := old.Size[0] * old.Size[1]
	if len(old.Cells) < limit || len(new.Cells) < limit {
		return diff
	}

	for i := 0 ; i < limit ; i++ {

			n := new.Cells[i] ; o := old.Cells[i]
		if !sameCell(n, o) {
			diff.Cells = append(diff.Cells, n)
			diff.Indexing = append(diff.Indexing, i)
		}
	}
	return diff
}

func invalidForDiff(old Frame, new Frame) bool {

	if len(old.Size) < 2 || len(new.Size) < 2 {
		return true
	}

	return old.Size[0] != new.Size[0] ||
		old.Size[1] != new.Size[1]
}

func sameRGB (a, b *RGB) bool {

	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func sameCell (a, b Cell) bool {

	return a.Char == b.Char &&
		sameRGB(a.Fg, b.Fg) &&
		sameRGB(a.Bg, b.Bg) &&
		a.Bold == b.Bold &&
		a.Italic == b.Italic &&
		a.Underline == b.Underline &&
		a.Visible == b.Visible &&
		a.Uppercase == b.Uppercase
}
