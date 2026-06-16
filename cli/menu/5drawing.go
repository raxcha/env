package menu

import (
	"env/engine"
	"env/routines"
	"env/utilities"
	"strings"
)

func (m *Menu) Draw() engine.Queue {
	q := *utilities.NewQueue()

	if !m.On || m.Utilities == nil {
		return q
	}

	if len(m.Bounds.Fullsize) < 2 {
		return q
	}

	sw, sh := m.Bounds.Fullsize[0], m.Bounds.Fullsize[1]

	// Width and height from full item list so size stays fixed while filtering
	iconOverhead := m.Utilities.VisibleLength("‹b α ›b ")
	w := 20
	for _, item := range m.AllItems {
		need := iconOverhead + m.Utilities.VisibleLength(item.Name) + 4
		if need > w {
			w = need
		}
	}
	w += 4
	if w > sw-2 {
		w = sw - 2
	}
	if w < 20 {
		w = 20
	}

	// Height: 1 prompt + N items + 2 box borders
	itemCount := len(m.AllItems)
	if itemCount == 0 {
		itemCount = 1
	}
	h := itemCount + 3
	if h > sh-2 {
		h = sh - 2
	}
	if h < 4 {
		h = 4
	}

	// Center on screen
	x := (sw - w) / 2
	y := (sh - h) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	contentW := w - 4
	itemsH := h - 3
	if itemsH < 0 {
		itemsH = 0
	}

	lines := []string{}

	prompt := "‹b λ ›b" + "› " + m.Prompt
	if m.Utilities.VisibleLength(prompt) > contentW {
		prompt = m.Utilities.CutVisible(prompt, contentW)
	}
	if m.Utilities.VisibleLength(prompt) < contentW {
		prompt += strings.Repeat(" ", contentW-m.Utilities.VisibleLength(prompt))
	}

	lines = append(lines, prompt)

	if len(m.Items) == 0 {
		line := "no matches"
		if m.Utilities.VisibleLength(line) < contentW {
			line += strings.Repeat(" ", contentW-m.Utilities.VisibleLength(line))
		}
		lines = append(lines, line)
	} else {
		start := m.visibleStart(itemsH)
		end := start + itemsH
		if end > len(m.Items) {
			end = len(m.Items)
		}

		for idx := start; idx < end; idx++ {
			displayName := m.itemDisplayName(m.Items[idx])
			name := menuGreekIcon(idx) + " " + displayName
			right := strings.TrimSpace(m.Items[idx].Right)

			rowW := contentW
			if right != "" || idx == m.Selected {
				rowW = contentW - 1
			}
			if rowW < 0 {
				rowW = 0
			}

			name = m.alignMenuRow(name, right, rowW)

			if idx == m.Selected {
				name = "¤KK ‹b " + name + " ›b ¤ "
			}

			lines = append(lines, name)
		}
	}

	lines = m.Utilities.Box(lines, utilities.BoxOpts{
		W:       w,
		H:       h,
		Padding: utilities.Padding{Top: 0, Right: 1, Bottom: 0, Left: 1},
		Title:   m.Title,
	})

	for i, line := range lines {
		lines[i] = "§AB0 " + line
	}

	b := m.Bounds
	b.Pos = routines.Bound{x, y}
	b.Size = routines.Bound{w, h}
	b.ActualPos = routines.Bound{x, y}
	b.ActualSize = routines.Bound{w, h}
	frame := m.Utilities.GenerateFrame(b, lines, 0)

	q.Frames = append(q.Frames, *frame)
	q.Size = m.Bounds.Fullsize

	return q
}

func (m *Menu) alignMenuRow(left string, right string, width int) string {
	if width < 0 {
		width = 0
	}

	if right == "" {
		if m.Utilities.VisibleLength(left) > width {
			left = m.Utilities.CutVisible(left, width)
		}
		if m.Utilities.VisibleLength(left) < width {
			left += strings.Repeat(" ", width-m.Utilities.VisibleLength(left))
		}
		return left
	}

	rightW := m.Utilities.VisibleLength(right)
	if rightW > width-1 {
		right = m.Utilities.CutVisible(right, width-1)
		rightW = m.Utilities.VisibleLength(right)
	}

	leftW := width - rightW - 1
	if leftW < 0 {
		leftW = 0
	}

	if m.Utilities.VisibleLength(left) > leftW {
		left = m.Utilities.CutVisible(left, leftW)
	}

	gap := width - m.Utilities.VisibleLength(left) - rightW
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}

func (m *Menu) visibleStart(height int) int {
	if height <= 0 || len(m.Items) <= height || m.Selected < 0 {
		return 0
	}

	start := m.Selected - height/2
	if start < 0 {
		start = 0
	}

	maxStart := len(m.Items) - height
	if start > maxStart {
		start = maxStart
	}

	return start
}

func menuSplitVisible(u *utilities.Utilities, s string, leftSize int) (string, string) {
	if u == nil {
		return s, ""
	}

	if leftSize < 0 {
		leftSize = 0
	}

	total := u.VisibleLength(s)

	if leftSize >= total {
		return s, ""
	}

	left := u.CutVisible(s, leftSize)

	rightSize := total - leftSize
	right := u.CutVisible(strings.TrimPrefix(s, left), rightSize)

	return left, right
}

func menuGreekIcon(index int) string {
	icons := []string{
		"‹b α ›b",
		"‹b β ›b",
		"‹b γ ›b",
		"‹b δ ›b",
		"‹b ε ›b",
		"‹b ζ ›b",
		"‹b η ›b",
		"‹b θ ›b",
		"‹b ι ›b",
		"‹b κ ›b",
		"‹b λ ›b",
		"‹b μ ›b",
		"‹b ν ›b",
		"‹b ξ ›b",
		"‹b ο ›b",
		"‹b π ›b",
		"‹b ρ ›b",
		"‹b σ ›b",
		"‹b τ ›b",
		"‹b υ ›b",
		"‹b φ ›b",
		"‹b χ ›b",
		"‹b ψ ›b",
		"‹b ω ›b",
	}

	if len(icons) == 0 {
		return ""
	}

	return icons[index%len(icons)]
}

func (m *Menu) itemDisplayName(item Item) string {
	if item.Command == "__prompt__" {
		prompt := strings.TrimSpace(m.Prompt)
		if prompt != "" {
			return "run: " + prompt
		}
	}

	return item.Name
}

func (m *Menu) GetSpec() string {
	return "asd"
}
