package tabs

import (
	"env/engine"
	"strings"
)

func (t *Tabs) Draw() engine.Queue {
	if t.Bounds.Mode == "fibonacci" {
		return t.drawFibonacci()
	}

	frames := []engine.Frame{t.drawMonocleFrame(0)}
	return *t.Utilities.GenerateQueue(t.Bounds.Fullsize, frames, false)
}

func (t *Tabs) drawMonocleFrame(timeout int) engine.Frame {
	fullW := 0
	if len(t.Bounds.Size) > 0 {
		fullW = t.Bounds.Size[0]
	}

	prefix := "λ "
	if t.Switch {
		prefix = "π "
	}

	prefixW := t.Parent.GetUtilities().VisibleLength(prefix)
	baseX := 0
	baseW := fullW
	if len(t.Bounds.Fullsize) > 0 {
		baseW = t.Bounds.Fullsize[0]
	}

	merged := t.Utilities.GenerateFrame(
		engine.Boundaries{Fullsize: t.Bounds.Fullsize, Pos: []int{baseX, t.Bounds.Pos[1]}, Size: []int{baseW, 1}},
		[]string{"§8F0 " + strings.Repeat(" ", baseW)},
		timeout,
	)

	prefixFrame := t.Utilities.GenerateFrame(
		engine.Boundaries{Fullsize: t.Bounds.Fullsize, Pos: t.Bounds.Pos, Size: []int{prefixW, 1}},
		[]string{tabsInlineStyle(-1) + "‹b " + prefix + "›b ¤ "},
		timeout,
	)
	next := t.Utilities.MergeFrames(*prefixFrame, *merged)
	merged = &next

	x := t.Bounds.Pos[0] + prefixW
	for i, client := range t.Parent.GetClients() {
		maxW := 36
		if len(t.Parent.GetClients()) > 0 && fullW > 0 {
			availableW := fullW - prefixW
			if availableW < 0 {
				availableW = fullW
			}
			maxW = availableW / len(t.Parent.GetClients())

			if maxW < 16 {
				maxW = 16
			}

			if maxW > 56 {
				maxW = 56
			}
		}

		selected := i == t.Parent.GetFocus()
		label := NormalClientLabelStyled(client, maxW, selected)

		if t.Switch {
			label = SwitchClientLabel(i, client, maxW)
		}

		toadd := " " + label + " "
		tabW := t.Parent.GetUtilities().VisibleLength(toadd)

		if fullW > 0 && x-t.Bounds.Pos[0] >= fullW {
			break
		}
		if fullW > 0 && x-t.Bounds.Pos[0]+tabW > fullW {
			tabW = fullW - (x - t.Bounds.Pos[0])
			toadd = t.Parent.GetUtilities().CutVisible(toadd, tabW)
		}

		style := tabsInlineStyle(i)
		tab := toadd
		if selected {
			tab = selectedTabText(toadd)
		}

		if tabW > 0 {
			tabFrame := t.Utilities.GenerateFrame(
				engine.Boundaries{Fullsize: t.Bounds.Fullsize, Pos: []int{x, t.Bounds.Pos[1]}, Size: []int{tabW, 1}},
				[]string{style + tab + "¤ "},
				timeout,
			)
			next := t.Utilities.MergeFrames(*tabFrame, *merged)
			merged = &next
		}

		x += tabW
	}

	return *merged
}

func (t *Tabs) drawFibonacci() engine.Queue {
	frames := []engine.Frame{t.drawFibonacciFrame(0)}
	return *t.Utilities.GenerateQueue(t.Bounds.Fullsize, frames, false)
}

func (t *Tabs) drawFibonacciFrame(timeout int) engine.Frame {
	var merged *engine.Frame

	for i, client := range t.Parent.GetClients() {
		bounds := client.GetBounds()
		if bounds == nil || len(bounds.ActualPos) < 2 || len(bounds.ActualSize) < 2 {
			continue
		}

		w := bounds.ActualSize[0]
		if w <= 0 {
			continue
		}

		y := bounds.ActualPos[1] - 1
		if y < 0 {
			y = bounds.ActualPos[1]
		}

		selected := i == t.Parent.GetFocus()
		label := NormalClientLabelStyled(client, w, selected)
		if t.Switch {
			label = SwitchClientLabel(i, client, w)
		}

		style := tabsInlineStyle(i)
		line := style + centerVisible(t.Parent.GetUtilities(), label, w)
		if selected {
			line = style + selectedTabText(centerVisible(t.Parent.GetUtilities(), label, w))
		}
		line += "¤ "

		frame := t.Utilities.GenerateFrame(
			engine.Boundaries{Fullsize: t.Bounds.Fullsize, Pos: []int{bounds.ActualPos[0], y}, Size: []int{w, 1}},
			[]string{line},
			timeout,
		)
		if merged == nil {
			merged = frame
			continue
		}

		next := t.Utilities.MergeFrames(*frame, *merged)
		merged = &next
	}

	if merged == nil {
		return engine.Frame{Size: t.Bounds.Fullsize, Timeout: timeout}
	}

	return *merged
}

func centerVisible(u interface {
	VisibleLength(string) int
	CutVisible(string, int) string
}, value string, width int) string {
	if width <= 0 {
		return ""
	}

	if u.VisibleLength(value) > width {
		return u.CutVisible(value, width)
	}

	remaining := width - u.VisibleLength(value)
	left := remaining / 2
	right := remaining - left
	return strings.Repeat(" ", left) + value + strings.Repeat(" ", right)
}

func tabsInlineStyle(index int) string {
	if index < 0 {
		return "¤8F "
	}
	return "¤8F "
}

func tabsSelectedStyle() string {
	return tabsInlineStyle(0)
}

func selectedTabText(value string) string {
	return " ‹b ¤8f " + strings.TrimSpace(value) + " ¤ ›b "
}
