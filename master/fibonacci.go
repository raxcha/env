package master

import "env/types"

// fibonacciLayout peels off half of the remaining area clockwise:
// left, top, right, bottom. Each split reserves one cell for its divider.
func (m *Master) fibonacciLayout() ([]types.Dimensions, map[types.Pos]bool) {
	remaining := m.Sizes
	remaining.Size[0] = max(0, remaining.Size[0])
	remaining.Size[1] = max(0, remaining.Size[1]-1)
	remaining = m.sidebarArea(remaining)
	panes := make([]types.Dimensions, 0, len(m.Clients))
	lines := map[types.Pos]bool{}

	for i := range m.Clients {
		pane := remaining
		if i < len(m.Clients)-1 {
			axis := i % 2
			gap := 0
			if remaining.Size[axis] >= 3 && remaining.Size[1-axis] > 0 {
				gap = 1
			}
			half := (remaining.Size[axis] - gap) / 2
			pane.Size[axis] = half
			remaining.Size[axis] -= half + gap
			divider := pane.Pos
			if i%4 < 2 {
				divider[axis] += half
				remaining.Pos[axis] += half + gap
			} else {
				divider[axis] += remaining.Size[axis]
				pane.Pos[axis] += remaining.Size[axis] + gap
			}
			if gap > 0 {
				for j := 0; j < pane.Size[1-axis]; j++ {
					pos := divider
					pos[1-axis] += j
					lines[pos] = axis == 0
				}
			}
		}
		panes = append(panes, pane)
	}
	return panes, lines
}

func (m *Master) resizeFibonacci() {
	panes, _ := m.fibonacciLayout()
	for i, client := range m.Clients {
		content := panes[i]
		headerHeight := min(1, content.Size[1])
		content.Pos[1] += headerHeight
		content.Size[1] -= headerHeight
		client.Resize(content)
	}
}

// Draw the background before the clients so divider cells never hide content.
func (m *Master) drawFibonacciDividers() *types.Queue {
	_, lines := m.fibonacciLayout()
	frame := m.Settings.GenerateFrame(m.Sizes, nil, 0, []int{0, 0, 0, 0})
	for pos, vertical := range lines {
		mask := 0
		if vertical {
			mask = 1 | 4
		} else {
			mask = 2 | 8
		}
		for i, delta := range []types.Pos{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
			if _, ok := lines[types.Pos{pos[0] + delta[0], pos[1] + delta[1]}]; ok {
				mask |= 1 << i
			}
		}
		char := map[int]rune{5: '│', 10: '─', 7: '├', 13: '┤', 11: '┴', 14: '┬', 15: '┼'}[mask]
		frame.Cells[pos[1]*m.Sizes.Full[0]+pos[0]].Char = char
	}
	return m.Settings.GenerateQueue(m.Sizes.Full, []*types.Frame{frame}, false)
}
