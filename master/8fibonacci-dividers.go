package master

import (
	"env/engine"
	"env/utilities"
)

const (
	fibonacciDividerVertical = 1 << iota
	fibonacciDividerHorizontal
)

func (m *Master) fibonacciDividersQueue() engine.Queue {
	if len(m.Clients) < 2 || len(m.Size) < 2 || m.Size[0] <= 0 || m.Size[1] <= 0 {
		return engine.Queue{}
	}

	rects := m.fibonacciClientRects()
	if len(rects) < 2 {
		return engine.Queue{}
	}

	marks := map[int]int{}
	for i := 0; i < len(rects); i++ {
		for j := i + 1; j < len(rects); j++ {
			m.markFibonacciDivider(marks, rects[i], rects[j])
		}
	}

	if len(marks) == 0 {
		return engine.Queue{}
	}

	fg, bg := m.fibonacciDividerColors()
	cells := make([]engine.Cell, m.Size[0]*m.Size[1])
	for i := range cells {
		cells[i] = engine.Cell{Visible: false}
	}

	for idx, mark := range marks {
		if idx < 0 || idx >= len(cells) {
			continue
		}

		cells[idx] = engine.Cell{
			Char:    fibonacciDividerRune(mark),
			Fg:      &fg,
			Bg:      &bg,
			Visible: true,
		}
	}

	frame := engine.Frame{
		Size:    m.Size,
		Cells:   cells,
		Timeout: 0,
	}

	return engine.Queue{
		Size:   m.Size,
		Frames: []engine.Frame{frame},
	}
}

func (m *Master) markFibonacciDivider(marks map[int]int, a []int, b []int) {
	if len(a) < 4 || len(b) < 4 {
		return
	}

	ax, ay, aw, ah := a[0], a[1], a[2], a[3]
	bx, by, bw, bh := b[0], b[1], b[2], b[3]
	ar, ab := ax+aw, ay+ah
	br, bb := bx+bw, by+bh

	if ar == bx {
		m.markVerticalDivider(marks, bx, maxInt(ay, by), minInt(ab, bb))
	}
	if br == ax {
		m.markVerticalDivider(marks, ax, maxInt(ay, by), minInt(ab, bb))
	}
	if ab == by {
		m.markHorizontalDivider(marks, by, maxInt(ax, bx), minInt(ar, br))
	}
	if bb == ay {
		m.markHorizontalDivider(marks, ay, maxInt(ax, bx), minInt(ar, br))
	}
}

func (m *Master) markVerticalDivider(marks map[int]int, x int, y0 int, y1 int) {
	if x < 0 || len(m.Size) < 2 || x >= m.Size[0] {
		return
	}

	y0 = maxInt(0, y0)
	y1 = minInt(m.Size[1], y1)
	for y := y0; y < y1; y++ {
		marks[y*m.Size[0]+x] |= fibonacciDividerVertical
	}
}

func (m *Master) markHorizontalDivider(marks map[int]int, y int, x0 int, x1 int) {
	if y < 0 || len(m.Size) < 2 || y >= m.Size[1] {
		return
	}

	x0 = maxInt(0, x0)
	x1 = minInt(m.Size[0], x1)
	for x := x0; x < x1; x++ {
		marks[y*m.Size[0]+x] |= fibonacciDividerHorizontal
	}
}

func (m *Master) fibonacciDividerColors() (engine.RGB, engine.RGB) {
	theme := m.Utilities.Theme
	if theme == nil {
		return engine.RGB{R: 90, G: 90, B: 90}, engine.RGB{}
	}

	fg := utilities.MixRGB(theme.Background, theme.Foreground, 0.28)
	return fg, theme.Background
}

func fibonacciDividerRune(mark int) rune {
	if mark&fibonacciDividerVertical != 0 && mark&fibonacciDividerHorizontal != 0 {
		return '┼'
	}
	if mark&fibonacciDividerHorizontal != 0 {
		return '─'
	}
	return '│'
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
