package master

import (
	"env/engine"
	"env/routines"
)

func (m *Master) updateBoundaries(fullsize *routines.Bound) {

	m.Size = *fullsize

	m.setFixedClientBounds()

	m.setClientsBounds()

	m.Draw()

}

func (m *Master) setFixedClientBounds() {
	m.Menu.Resize(m.calcMenuBounds())
	m.Status.Resize(m.calcStatusBounds())
	m.Notifications.Resize(m.calcNotificationsBounds())
}

func (m *Master) setClientsBounds() {
	switch m.Mode {

	case "monocle":
		m.Tabs.Resize(m.calcMonocleTabsBounds())
	case "fibonacci":
		m.Tabs.Resize(m.calcFibonacciTabsBounds())
	}

	for i, client := range m.Clients {

		switch m.Mode {
		case "monocle":
			client.Resize(m.calcMonocleBounds(i))
		case "fibonacci":
			client.Resize(m.calcFibonacciBounds(i))
		}
	}
}

func (m *Master) calcFibonacciBounds(index int) *engine.Boundaries {
	b := newBoundaries()
	b.Fullsize = m.Size
	b.Index = index
	b.Mode = "fibonacci"

	rects := m.fibonacciClientRects()
	if index < 0 || index >= len(rects) {
		return b
	}

	rect := rects[index]
	b.Pos = routines.Bound{rect[0], rect[1]}
	b.Size = routines.Bound{rect[2], rect[3]}

	if m.Tabs.On && b.Size[1] > 1 {
		b.Pos[1]++
		b.Size[1]--
	}

	return b

}

func (m *Master) fibonacciClientRects() []routines.Bound {
	count := len(m.Clients)
	rects := make([]routines.Bound, count)
	if count == 0 || len(m.Size) < 2 {
		return rects
	}

	insetX, insetW := m.horizontalInset()
	w := maxInt(0, insetW)
	h := maxInt(0, m.Size[1]-1)
	if w == 0 || h == 0 {
		return rects
	}

	m.fillFibonacciRects(rects, 0, insetX, 0, w, h, true)
	return rects
}

func (m *Master) fillFibonacciRects(rects []routines.Bound, index int, x int, y int, w int, h int, vertical bool) {
	if index >= len(rects) {
		return
	}

	if index == len(rects)-1 || w <= 1 || h <= 1 {
		rects[index] = routines.Bound{x, y, w, h}
		return
	}

	if vertical {
		firstW := halfSplit(w)
		rects[index] = routines.Bound{x, y, firstW, h}
		m.fillFibonacciRects(rects, index+1, x+firstW, y, w-firstW, h, false)
		return
	}

	firstH := halfSplit(h)
	rects[index] = routines.Bound{x, y, w, firstH}
	m.fillFibonacciRects(rects, index+1, x, y+firstH, w, h-firstH, true)
}

func halfSplit(size int) int {
	if size <= 2 {
		return 1
	}

	part := size / 2
	if part < 1 {
		return 1
	}
	if part >= size {
		return size - 1
	}
	return part
}

func (m *Master) calcMonocleBounds(index int) *engine.Boundaries {

	b := newBoundaries()

	b.Fullsize = m.Size
	if len(m.Size) < 2 {
		return b
	}

	x, w := m.horizontalInset()
	y := 0
	h := m.Size[1] - 1

	if m.Tabs.On {
		y++
		h--
	}

	if index == m.Focus && m.Focus >= 0 && m.Focus < len(m.Clients) && m.Clients[m.Focus].IsSidebarOn() {
		sb := m.Clients[m.Focus].GetSidebarBounds()
		if sb != nil && len(sb.ActualSize) >= 1 && sb.ActualSize[0] > 0 {
			x += sb.ActualSize[0]
			w -= sb.ActualSize[0]
		}
	}

	b.Pos = routines.Bound{x, y}
	b.Size = routines.Bound{maxInt(0, w), maxInt(0, h)}

	b.Index = -1
	b.Mode = "monocle"
	return b

}

func (m *Master) calcMenuBounds() *engine.Boundaries {

	b := newBoundaries()
	b.Fullsize = m.Size
	b.Index = -1
	b.Mode = m.Mode

	if len(m.Size) < 2 {
		return b
	}

	sw, sh := m.Size[0], m.Size[1]

	w := sw / 2
	if w < 36 {
		w = 36
	}
	if w > 64 {
		w = 64
	}
	if w > sw {
		w = sw
	}

	h := sh * 2 / 3
	if h < 8 {
		h = 8
	}
	if h > 24 {
		h = 24
	}
	if h > sh {
		h = sh
	}

	x := (sw - w) / 2
	y := (sh - h) / 2

	b.Pos = routines.Bound{x, y}
	b.Size = routines.Bound{w, h}

	return b

}

func (m *Master) calcStatusBounds() *engine.Boundaries {

	b := newBoundaries()

	b.Fullsize = m.Size
	if len(m.Size) < 2 {
		return b
	}
	x, w := m.horizontalInset()
	b.Pos = routines.Bound{x, m.Size[1] - 1}
	b.Size = routines.Bound{w, 1}
	b.Index = -1
	b.Mode = m.Mode

	return b

}

func (m *Master) calcNotificationsBounds() *engine.Boundaries {

	b := newBoundaries()
	b.Fullsize = m.Size
	b.Index = -1
	b.Mode = m.Mode

	if len(m.Size) < 2 {
		return b
	}

	y := 0
	if m.Tabs.On {
		y = 1
	}

	x, w := m.horizontalInset()
	b.Pos = routines.Bound{x, y}
	b.Size = routines.Bound{w, maxInt(0, m.Size[1]-y)}

	return b

}

func (m *Master) calcMonocleTabsBounds() *engine.Boundaries {

	b := newBoundaries()

	b.Fullsize = m.Size
	if len(m.Size) < 2 {
		return b
	}
	x, w := m.horizontalInset()
	b.Pos = routines.Bound{x, 0}
	b.Size = routines.Bound{w, 1}
	b.Index = -1
	b.Mode = "monocle"

	return b
}

func (m *Master) calcFibonacciTabsBounds() *engine.Boundaries {

	b := newBoundaries()

	b.Fullsize = m.Size
	if len(m.Size) < 2 {
		return b
	}
	x, w := m.horizontalInset()
	b.Pos = routines.Bound{x, 0}
	b.Size = routines.Bound{w, m.Size[1]}
	b.Index = -1
	b.Mode = "fibonacci"

	return b
}

func newBoundaries() *engine.Boundaries {

	return &engine.Boundaries{
		Fullsize:   routines.Bound{0, 0},
		Pos:        routines.Bound{0, 0},
		Size:       routines.Bound{0, 0},
		Mode:       "",
		Index:      -1,
		ActualSize: nil,
		ActualPos:  nil,
	}
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *Master) horizontalInset() (int, int) {
	if len(m.Size) < 1 {
		return 0, 0
	}

	if m.Size[0] <= 2 {
		return 0, maxInt(0, m.Size[0])
	}

	return 1, m.Size[0] - 2
}
