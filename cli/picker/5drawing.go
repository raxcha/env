package picker

import (
	"env/engine"
	"env/utilities"
)

func (p *Picker) Draw() *engine.Queue {
	if len(p.Bounds.Fullsize) < 2 || len(p.Bounds.Pos) < 2 || len(p.Bounds.Size) < 2 {
		q := utilities.NewQueue()
		return q
	}

	p.updateItems()

	itemsframe := p.drawItems()
	promptframe := p.drawPrompt()

	finalframe := p.Utilities.MergeFrames(*promptframe, *itemsframe)

	if p.pickerUsePreview() {
		previewframe := p.drawPreview()
		finalframe = p.Utilities.MergeFrames(*previewframe, finalframe)
	}

	separators := p.drawPickerSeparators()
	finalframe = p.Utilities.MergeFrames(separators, finalframe)

	q := utilities.NewQueue()
	q.Frames = append(q.Frames, finalframe)
	q.Size = p.Bounds.Fullsize

	return q
}