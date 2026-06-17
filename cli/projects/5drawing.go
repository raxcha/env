package projects

import (
	"env/engine"
	"env/routines"
	"env/utilities"
	"strings"
)

func (p *Projects) Draw() *engine.Queue {
	if len(p.Bounds.Fullsize) < 2 || len(p.Bounds.Pos) < 2 || len(p.Bounds.Size) < 2 {
		return utilities.NewQueue()
	}

	p.updateItems()

	if p.panelMode() == "preview" {
		preview := p.drawPreview()
		q := utilities.NewQueue()
		q.Frames = append(q.Frames, *preview)
		q.Size = p.Bounds.Fullsize
		return q
	}

	items := p.drawItems()
	prompt := p.drawPrompt()

	frame := p.Utilities.MergeFrames(*items, *prompt)
	if p.panelMode() == "both" {
		preview := p.drawPreview()
		seps := p.drawSeparators()
		frame = p.Utilities.MergeFrames(frame, *preview)
		frame = p.Utilities.MergeFrames(frame, seps)
	}

	q := utilities.NewQueue()
	q.Frames = append(q.Frames, frame)
	q.Size = p.Bounds.Fullsize
	return q
}

func (p *Projects) drawItems() *engine.Frame {
	itemsRect, _, _ := p.layout()
	w, h := itemsRect.W, itemsRect.H
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	paddingLeft := 1
	paddingRight := 1
	contentW := w - paddingLeft - paddingRight
	if contentW < 1 {
		contentW = 1
	}

	lines := []string{}
	start := cutLines(h, p.Selected, len(p.Items))
	end := start + h
	if end > len(p.Items) {
		end = len(p.Items)
	}

	for i := 0; i < h-(end-start); i++ {
		lines = append(lines, p.fitLine("", w))
	}

	for i := end - 1; i >= start; i-- {
		item := p.Items[i]
		line := p.lineForItem(item, contentW)

		if i == p.Selected {
			line = p.selectedLine(line, contentW)
		}

		line =
			strings.Repeat(" ", paddingLeft) +
				p.fitLine(line, contentW) +
				strings.Repeat(" ", paddingRight)

		lines = append(lines, p.fitLine(line, w))
	}

	for len(lines) < h {
		lines = append(lines, p.fitLine("", w))
	}
	if len(lines) > h {
		lines = lines[:h]
	}

	return p.Utilities.GenerateFrame(engine.Boundaries{Fullsize: p.Bounds.Fullsize, Pos: routines.Bound{itemsRect.X, itemsRect.Y}, Size: routines.Bound{w, h}}, lines, 0)
}

func (p *Projects) drawPrompt() *engine.Frame {
	_, promptRect, _ := p.layout()
	w, h := promptRect.W, promptRect.H
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	line := " ‹b " + "λ›b  " + p.Prompt
	lines := []string{p.fitLine(line, w)}
	for len(lines) < h {
		lines = append(lines, p.fitLine("", w))
	}
	if len(lines) > h {
		lines = lines[:h]
	}

	return p.Utilities.GenerateFrame(engine.Boundaries{Fullsize: p.Bounds.Fullsize, Pos: routines.Bound{promptRect.X, promptRect.Y}, Size: routines.Bound{w, h}}, lines, 0)
}

func (p *Projects) drawPreview() *engine.Frame {
	_, _, previewRect := p.layout()
	w, h := previewRect.W, previewRect.H
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	innerW := w - 2
	if innerW < 1 {
		innerW = 1
	}
	innerH := h - 2
	if innerH < 1 {
		innerH = 1
	}

	content := []string{}
	for _, line := range p.previewLines() {
		content = append(content, p.fitLine(line, innerW))
	}
	for len(content) < innerH {
		content = append(content, p.fitLine("", innerW))
	}
	if len(content) > innerH {
		content = content[:innerH]
	}

	lines := []string{p.fitLine("", w)}
	for _, line := range content {
		lines = append(lines, p.fitLine(" "+line+" ", w))
	}
	for len(lines) < h {
		lines = append(lines, p.fitLine("", w))
	}
	if len(lines) > h {
		lines = lines[:h]
	}

	return p.Utilities.GenerateFrame(engine.Boundaries{Fullsize: p.Bounds.Fullsize, Pos: routines.Bound{previewRect.X, previewRect.Y}, Size: routines.Bound{w, h}}, lines, 0)
}

func (p *Projects) drawSeparators() engine.Frame {
	if len(p.Bounds.Fullsize) < 2 || len(p.Bounds.Pos) < 2 || len(p.Bounds.Size) < 2 {
		return engine.Frame{}
	}

	fullW := p.Bounds.Fullsize[0]
	fullH := p.Bounds.Fullsize[1]
	th := p.Parent.GetTheme()
	dividerFg := utilities.MixRGB(th.Background, th.Foreground, 0.28)
	cells := make([]engine.Cell, fullW*fullH)

	for i := range cells {
		cells[i] = engine.Cell{Char: 0, Fg: &th.Foreground, Bg: &th.Background, Visible: true}
	}

	put := func(x, y int, ch rune) {
		if x < 0 || x >= fullW || y < 0 || y >= fullH {
			return
		}

		i := y*fullW + x
		current := cells[i].Char
		if current != 0 && current != ch {
			ch = projectMergeBoxChar(current, ch)
		}

		cells[i] = engine.Cell{Char: ch, Fg: &dividerFg, Bg: &th.Background, Visible: true}
	}

	_, _, previewRect := p.layout()

	if projectsVerticalPanels(p.Bounds.Size[0], p.Bounds.Size[1]) {
		sepY := previewRect.Y + previewRect.H
		for x := p.Bounds.Pos[0]; x < p.Bounds.Pos[0]+p.Bounds.Size[0]; x++ {
			put(x, sepY, '─')
		}
	} else {
		sepX := previewRect.X - 1
		for y := p.Bounds.Pos[1]; y < p.Bounds.Pos[1]+p.Bounds.Size[1]; y++ {
			put(sepX, y, '│')
		}
	}

	return engine.Frame{Size: p.Bounds.Fullsize, Cells: cells, Timeout: 0}
}

func projectMergeBoxChar(a, b rune) rune {
	if a == b {
		return a
	}

	if a == '┤' || b == '┤' {
		return '┤'
	}

	return '┼'
}
