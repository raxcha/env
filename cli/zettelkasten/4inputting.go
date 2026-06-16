package zettelkasten

import "env/routines"

func (z *Zettelkasten) Input(input *routines.Input) {
	if z.visualManualDebugInput(input) {
		return
	}

	switch input.Key {
	case "up":
		z.Focus = "tags"
		z.moveTag(-1)

	case "down":
		z.Focus = "tags"
		z.moveTag(1)

	case "left":
		z.Focus = "overlaps"
		z.moveOverlap(-1)

	case "right":
		z.Focus = "overlaps"
		z.moveOverlap(1)

	case "tab":
		z.toggleFocus()

	case "enter", "ctrl+enter":
		if z.Focus == "overlaps" {
			z.selectOverlapAsTag()
		}

	case "backspace":
		z.backspacePrompt()

	case "char":
		z.Prompt += string(input.Char)
		z.SelectedTag = 0
		z.SelectedOverlap = 0
		z.rebuild()
	}
}

func (z *Zettelkasten) toggleFocus() {
	if z.Focus == "tags" {
		z.Focus = "overlaps"
		return
	}
	z.Focus = "tags"
}

func (z *Zettelkasten) move(delta int) {
	if z.Focus == "overlaps" {
		z.SelectedOverlap += delta
		z.clampSelections()
		return
	}

	old := z.SelectedTag
	z.SelectedTag += delta
	z.clampSelections()
	if old != z.SelectedTag {
		z.SelectedOverlap = 0
		z.rebuildOverlaps()
	}
}

func (z *Zettelkasten) backspacePrompt() {
	if z.Prompt == "" {
		return
	}

	runes := []rune(z.Prompt)
	if len(runes) == 0 {
		return
	}

	z.Prompt = string(runes[:len(runes)-1])
	z.SelectedTag = 0
	z.SelectedOverlap = 0
	z.rebuild()
}

func (z *Zettelkasten) selectOverlapAsTag() {
	item := z.selectedOverlap()
	if item == nil {
		return
	}

	z.Prompt = item.Name
	z.Focus = "tags"
	z.SelectedTag = 0
	z.SelectedOverlap = 0
	z.rebuild()
}

func (z *Zettelkasten) moveTag(delta int) {
	if len(z.Tags) == 0 {
		z.SelectedTag = 0
		return
	}

	z.SelectedTag += delta

	if z.SelectedTag < 0 {
		z.SelectedTag = 0
	}

	if z.SelectedTag >= len(z.Tags) {
		z.SelectedTag = len(z.Tags) - 1
	}
}

func (z *Zettelkasten) moveOverlap(delta int) {
	if len(z.Overlaps) == 0 {
		z.SelectedOverlap = 0
		return
	}

	z.SelectedOverlap += delta

	if z.SelectedOverlap < 0 {
		z.SelectedOverlap = 0
	}

	if z.SelectedOverlap >= len(z.Overlaps) {
		z.SelectedOverlap = len(z.Overlaps) - 1
	}
}

func (z *Zettelkasten) visualManualDebugInput(input *routines.Input) bool {
	if !zettelVisualManualDebug || input == nil {
		return false
	}

	switch input.Key {
	case "tab":
		zettelVisualDebugToggleSelected()
		return true

	case "left":
		zettelVisualDebugMoveSelected(-1, 0)
		return true

	case "right":
		zettelVisualDebugMoveSelected(1, 0)
		return true

	case "up":
		zettelVisualDebugMoveSelected(0, -1)
		return true

	case "down":
		zettelVisualDebugMoveSelected(0, 1)
		return true

	case "char":
		switch input.Char {
		case 'a', 'A':
			zettelVisualDebugSelected = "A"
			return true

		case 'b', 'B':
			zettelVisualDebugSelected = "B"
			return true

		case 'h':
			zettelVisualDebugMoveSelected(-1, 0)
			return true

		case 'l':
			zettelVisualDebugMoveSelected(1, 0)
			return true

		case 'k':
			zettelVisualDebugMoveSelected(0, -1)
			return true

		case 'j':
			zettelVisualDebugMoveSelected(0, 1)
			return true

		case 'r', 'R':
			zettelVisualDebugAX = -1
			zettelVisualDebugAY = -1
			zettelVisualDebugBX = -1
			zettelVisualDebugBY = -1
			return true
		}
	}

	return false
}
