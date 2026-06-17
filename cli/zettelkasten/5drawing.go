package zettelkasten

import (
	"env/engine"
	"env/routines"
	"env/utilities"
	"fmt"
	"strings"
)

func (z *Zettelkasten) Draw() *engine.Queue {
	if len(z.Bounds.Fullsize) < 2 || len(z.Bounds.Pos) < 2 || len(z.Bounds.Size) < 2 {
		return utilities.NewQueue()
	}

	z.rebuild()
	z.clampSelections()

	if z.panelMode() == "preview" {
		visual := z.drawVisual()
		q := utilities.NewQueue()
		q.Frames = append(q.Frames, *visual)
		q.Size = z.Bounds.Fullsize
		return q
	}

	tags := z.drawTags()
	overlaps := z.drawOverlaps()
	prompt := z.drawPrompt()

	frame := z.Utilities.MergeFrames(*tags, *overlaps)
	frame = z.Utilities.MergeFrames(frame, *prompt)
	if z.panelMode() == "both" {
		visual := z.drawVisual()
		seps := z.drawSeparators()
		frame = z.Utilities.MergeFrames(frame, *visual)
		frame = z.Utilities.MergeFrames(frame, seps)
	}

	q := utilities.NewQueue()
	q.Frames = append(q.Frames, frame)
	q.Size = z.Bounds.Fullsize
	return q
}

func (z *Zettelkasten) layout() (zettelRect, zettelRect, zettelRect, zettelRect) {
	x := z.Bounds.Pos[0] + 1
	y := z.Bounds.Pos[1]
	w := z.Bounds.Size[0] - 1
	h := z.Bounds.Size[1]

	if w < 1 {
		w = 1
	}

	if z.panelMode() == "preview" {
		return zettelRect{}, zettelRect{}, zettelRect{}, zettelRect{X: z.Bounds.Pos[0], Y: y, W: z.Bounds.Size[0], H: h}
	}

	if z.panelMode() == "both" && zettelVerticalPanels(w, h) {
		promptH := 1
		visualH := h / 2
		if visualH < 1 {
			visualH = 1
		}
		if visualH > h-promptH-3 {
			visualH = h - promptH - 3
		}
		if visualH < 1 {
			visualH = 1
		}

		sepH := 1
		listH := h - visualH - sepH - promptH
		if listH < 2 {
			listH = 2
		}

		tagsH := listH / 2
		if tagsH < 1 {
			tagsH = 1
		}

		notesH := listH - tagsH - 1
		if notesH < 1 {
			notesH = 1
		}

		visual := zettelRect{X: x, Y: y, W: w, H: visualH}
		tags := zettelRect{X: x, Y: y + visualH + sepH, W: w, H: tagsH}
		notes := zettelRect{X: x, Y: tags.Y + tagsH + 1, W: w, H: notesH}
		prompt := zettelRect{X: x, Y: y + h - 1, W: w, H: promptH}

		return tags, notes, prompt, visual
	}

	leftW := w / 4
	if z.panelMode() == "list" {
		leftW = w
	}
	if leftW < 24 {
		leftW = 24
	}
	if z.panelMode() == "both" && leftW > w-10 {
		leftW = w / 2
	}
	if leftW < 1 {
		leftW = 1
	}

	promptH := 1
	leftContentH := h - promptH
	if leftContentH < 2 {
		leftContentH = 2
	}
	tagsH := leftContentH / 2
	if tagsH < 1 {
		tagsH = 1
	}

	notesH := h - tagsH - promptH - 1
	if notesH < 1 {
		notesH = 1
	}

	tags := zettelRect{X: x, Y: y, W: leftW, H: tagsH}
	notes := zettelRect{X: x, Y: y + tagsH + 1, W: leftW, H: notesH}
	prompt := zettelRect{X: x, Y: y + h - 1, W: leftW, H: promptH}
	visual := zettelRect{X: x + leftW + 1, Y: y, W: w - leftW - 1, H: h}
	if visual.W < 1 {
		visual.W = 1
	}
	return tags, notes, prompt, visual
}

func zettelVerticalPanels(w int, h int) bool {
	return w < h*2
}

func (z *Zettelkasten) panelMode() string {
	switch z.PanelMode {
	case "list", "preview", "both":
		return z.PanelMode
	default:
		return "both"
	}
}

func (z *Zettelkasten) drawTags() *engine.Frame {
	tagsRect, _, _, _ := z.layout()

	contentW := tagsRect.W - 3
	if contentW < 1 {
		contentW = 1
	}

	lines := z.listLines(tagsRect.W, tagsRect.H, len(z.Tags), z.SelectedTag, func(i int) string {
		item := z.Tags[i]
		if item == nil {
			return ""
		}
		return z.tagLine(item, contentW)
	}, z.Focus == "tags")

	return z.Utilities.GenerateFrame(engine.Boundaries{Fullsize: z.Bounds.Fullsize, Pos: routines.Bound{tagsRect.X, tagsRect.Y}, Size: routines.Bound{tagsRect.W, tagsRect.H}}, lines, 0)
}

func (z *Zettelkasten) tagLine(item *ZettelTagItem, width int) string {
	if item == nil {
		return z.fitLine("", width)
	}

	label := "# " + item.Name
	count := fmt.Sprintf("%d", len(item.Paths))

	labelRunes := []rune(label)
	countRunes := []rune(count)

	if width <= 0 {
		return ""
	}

	rightPad := 1

	if len(countRunes)+rightPad >= width {
		return z.fitLine(count+strings.Repeat(" ", rightPad), width)
	}

	maxLabelW := width - len(countRunes) - rightPad - 1
	if maxLabelW < 1 {
		maxLabelW = 1
	}

	if len(labelRunes) > maxLabelW {
		label = string(labelRunes[:maxLabelW])
		labelRunes = []rune(label)
	}

	gap := width - len(labelRunes) - len(countRunes) - rightPad
	if gap < 1 {
		gap = 1
	}

	return label + strings.Repeat(" ", gap) + count + strings.Repeat(" ", rightPad)
}

func zettelGreekLetter(index int) string {
	letters := []string{
		"α", "β", "γ", "δ", "ε", "ζ", "η", "θ",
		"ι", "κ", "λ", "μ", "ν", "ξ", "ο", "π",
		"ρ", "σ", "τ", "υ", "φ", "χ", "ψ", "ω",
	}

	if index < 0 {
		index = 0
	}

	return letters[index%len(letters)]
}

func (z *Zettelkasten) drawOverlaps() *engine.Frame {
	_, overlapRect, _, _ := z.layout()

	lines := z.overlapLines(overlapRect.W, overlapRect.H, len(z.Overlaps), z.SelectedOverlap, func(i int) string {
		item := z.Overlaps[i]
		if item == nil {
			return ""
		}

		return z.overlapLine(item, overlapRect.W-2)
	}, z.Focus == "overlaps")

	return z.Utilities.GenerateFrame(engine.Boundaries{Fullsize: z.Bounds.Fullsize, Pos: routines.Bound{overlapRect.X, overlapRect.Y}, Size: routines.Bound{overlapRect.W, overlapRect.H}}, lines, 0)
}

func (z *Zettelkasten) listLinesTop(w, h, total, selected int, render func(int) string, active bool) []string {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	contentW := w - 3
	if contentW < 1 {
		contentW = 1
	}

	lines := []string{}

	start := cutLines(h, selected, total)
	end := start + h
	if end > total {
		end = total
	}

	for i := start; i < end; i++ {
		line := render(i)

		prefix := "  "
		if i == selected {
			runes := []rune(line)

			if len(runes) >= 3 {
				leftPad := string(runes[:1])
				mid := string(runes[1 : len(runes)-1])
				rightPad := string(runes[len(runes)-1:])

				if active {
					line = leftPad + "¤KK ‹b " + mid + "›b ¤ " + rightPad
				} else {
					line = leftPad + "¤K0 " + mid + "¤ " + rightPad
				}
			}
		}

		line = prefix + z.fitLine(line, contentW)
		line = z.fitLine(line, w)

		if i == selected {
			if active {
				line = "¤KK ‹b " + line + "›b ¤ "
			} else {
				line = "¤K0 " + line + "¤ "
			}
		}

		lines = append(lines, line)
	}

	for len(lines) < h {
		lines = append(lines, z.fitLine("", w))
	}

	if len(lines) > h {
		lines = lines[:h]
	}

	return lines
}

func (z *Zettelkasten) drawPrompt() *engine.Frame {
	_, _, promptRect, _ := z.layout()
	line := " ‹b " + "λ›b  " + z.Prompt
	lines := []string{z.fitLine(line, promptRect.W)}
	return z.Utilities.GenerateFrame(engine.Boundaries{Fullsize: z.Bounds.Fullsize, Pos: routines.Bound{promptRect.X, promptRect.Y}, Size: routines.Bound{promptRect.W, promptRect.H}}, lines, 0)
}

func (z *Zettelkasten) drawVisual() *engine.Frame {
	_, _, _, visualRect := z.layout()
	lines := z.visualLines(visualRect.W, visualRect.H)
	return z.Utilities.GenerateFrame(engine.Boundaries{Fullsize: z.Bounds.Fullsize, Pos: routines.Bound{visualRect.X, visualRect.Y}, Size: routines.Bound{visualRect.W, visualRect.H}}, lines, 0)
}

func (z *Zettelkasten) listLines(w, h, total, selected int, render func(int) string, active bool) []string {
	return z.zettelPickerStyleLines(w, h, total, selected, render, active, nil, false)
}

func (z *Zettelkasten) familyLines(w, h, total, selected int, render func(int) string, active bool) []string {
	return z.zettelPickerStyleLines(w, h, total, selected, render, active, func(i int, line string) string {
		return "‹b " + zettelGreekLetter(i) + " ›b " + line
	}, true)
}

func (z *Zettelkasten) overlapLines(w, h, total, selected int, render func(int) string, active bool) []string {
	return z.zettelPickerStyleLines(w, h, total, selected, render, active, nil, true)
}

func (z *Zettelkasten) overlapLine(item *ZettelOverlapItem, width int) string {
	if item == nil {
		return z.fitVisibleLine("", width)
	}

	percent := item.Percent
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	barW := 8
	if width < 24 {
		barW = 5
	}
	if barW < 1 {
		barW = 1
	}

	fill := int(float64(barW) * float64(percent) / 100.0)
	if fill < 0 {
		fill = 0
	}
	if fill > barW {
		fill = barW
	}

	bar := strings.Repeat("█", fill) + strings.Repeat("░", barW-fill)
	right := fmt.Sprintf("%3d%% %s", percent, bar)
	labelW := width - z.Utilities.VisibleLength(right) - 1
	if labelW < 1 {
		labelW = 1
	}

	label := "# " + item.Name
	label = z.fitVisibleLine(label, labelW)
	return label + " " + right
}

func (z *Zettelkasten) zettelPickerStyleLines(w, h, total, selected int, render func(int) string, active bool, decorate func(int, string) string, fullHighlight bool) []string {
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

	start := cutLines(h, selected, total)
	end := start + h
	if end > total {
		end = total
	}

	for i := start; i < end; i++ {
		line := render(i)
		if decorate != nil {
			line = decorate(i, line)
		}

		line = strings.Repeat(" ", paddingLeft) + z.fitVisibleLine(line, contentW) + strings.Repeat(" ", paddingRight)

		if i == selected {
			line = z.selectedZettelItemLine(line, w, active, fullHighlight)
		}

		lines = append(lines, z.fitVisibleLine(line, w))
	}

	for len(lines) < h {
		lines = append(lines, z.fitVisibleLine("", w))
	}

	if len(lines) > h {
		lines = lines[:h]
	}

	return lines
}

func (z *Zettelkasten) selectedZettelItemLine(line string, width int, active bool, fullHighlight bool) string {
	if width <= 0 {
		return line
	}

	visible := z.Utilities.VisibleLength(line)

	if visible > width {
		line = z.Utilities.CutVisible(line, width)
		visible = z.Utilities.VisibleLength(line)
	}

	if visible < width {
		line += strings.Repeat(" ", width-visible)
	}

	highlightW := z.Utilities.VisibleLength(strings.TrimRight(line, " ")) + 2

	if fullHighlight {
		highlightW = width - 1
	}

	if highlightW > width {
		highlightW = width
	}
	if highlightW < 0 {
		highlightW = 0
	}

	left := z.Utilities.CutVisible(line, highlightW)
	right := z.Utilities.CutVisibleFrom(line, highlightW)

	if active {
		return "¤KK " + left + "¤ " + right
	}

	return "¤K0 " + left + "¤ " + right
}

func (z *Zettelkasten) fitLine(line string, width int) string {
	return z.fitVisibleLine(line, width)
}

func (z *Zettelkasten) fitVisibleLine(line string, width int) string {
	if width <= 0 {
		return ""
	}

	if z.Utilities == nil {
		r := []rune(line)
		if len(r) > width {
			return string(r[:width])
		}
		return line + strings.Repeat(" ", width-len(r))
	}

	visible := z.Utilities.VisibleLength(line)
	if visible > width {
		line = z.Utilities.CutVisible(line, width)
		visible = z.Utilities.VisibleLength(line)
	}

	if visible < width {
		line += strings.Repeat(" ", width-visible)
	}

	return line
}

func (z *Zettelkasten) drawSeparators() engine.Frame {
	if len(z.Bounds.Fullsize) < 2 || len(z.Bounds.Pos) < 2 || len(z.Bounds.Size) < 2 {
		return engine.Frame{}
	}

	fullW := z.Bounds.Fullsize[0]
	fullH := z.Bounds.Fullsize[1]
	th := z.Parent.GetTheme()
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
			ch = '┼'
		}
		cells[i] = engine.Cell{Char: ch, Fg: &dividerFg, Bg: &th.Background, Visible: true}
	}

	tagsRect, notesRect, _, visualRect := z.layout()
	if zettelVerticalPanels(z.Bounds.Size[0]-1, z.Bounds.Size[1]) {
		sepVisualY := visualRect.Y + visualRect.H
		for x := visualRect.X; x < visualRect.X+visualRect.W; x++ {
			put(x, sepVisualY, '─')
		}

		sepTagsY := notesRect.Y - 1
		for x := tagsRect.X; x < tagsRect.X+tagsRect.W; x++ {
			put(x, sepTagsY, '─')
		}
	} else {
		sepX := visualRect.X - 1
		for y := z.Bounds.Pos[1]; y < z.Bounds.Pos[1]+z.Bounds.Size[1]; y++ {
			put(sepX, y, '│')
		}

		sepTagsY := notesRect.Y - 1
		for x := tagsRect.X; x < tagsRect.X+tagsRect.W; x++ {
			put(x, sepTagsY, '─')
		}

		put(sepX, sepTagsY, '┤')
	}

	return engine.Frame{Size: z.Bounds.Fullsize, Cells: cells, Timeout: 0}
}
