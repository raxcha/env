package utilities

import "env/engine"

func (u *Utilities) MergeFrames (new, old engine.Frame) engine.Frame {
	if len(new.Cells) != len(old.Cells) {
		return old
	}

	mergedCells := make([]engine.Cell, len(old.Cells))
	copy(mergedCells, old.Cells)

	for i := 0; i < len(new.Cells); i++ {
		// Só sobrepomos se a célula for visível E não for apenas uma margem (Char 0)
		if new.Cells[i].Visible && new.Cells[i].Char != 0 {
			mergedCells[i] = new.Cells[i]
		}
	}

	return engine.Frame{
		Size:    old.Size,
		Cells:   mergedCells,
		Timeout: new.Timeout,
	}
}
