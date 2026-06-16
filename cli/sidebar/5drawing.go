package sidebar

import (
	"env/engine"
	"env/filesystem"
	"strings"
)

func (s *Sidebar) Draw() engine.Queue {

	if len(s.Bounds.Fullsize) < 2 || len(s.Bounds.ActualSize) < 2 || len(s.Bounds.ActualPos) < 2 {
		return engine.Queue{}
	}

	lines := s.sidebarLines()

	frame := s.Utilities.GenerateFrame(
		engine.Boundaries{Fullsize: s.Bounds.Fullsize, Pos: s.Bounds.ActualPos, Size: s.Bounds.ActualSize},
		lines,
		0,
	)

	frames := []engine.Frame{*frame}
	return *s.Utilities.GenerateQueue(s.Bounds.Fullsize, frames, false)
}

func (s *Sidebar) sidebarLines() []string {
	lines := []string{}

	start := s.ViewStart
	height := 0

	if len(s.Bounds.ActualSize) >= 2 {
		height = s.Bounds.ActualSize[1]
	}

	if height <= 0 {
		height = len(s.Flat)
	}

	end := start + height

	if start < 0 {
		start = 0
	}

	if start > len(s.Flat) {
		start = len(s.Flat)
	}

	if end > len(s.Flat) {
		end = len(s.Flat)
	}

	for i := start; i < end; i++ {
		node := s.Flat[i]

		if node == nil || node.Page == nil {
			continue
		}

		depth := node.Depth - 1
		if depth < 0 {
			depth = 0
		}

		line := strings.Repeat("  ", depth) + sidebarGreekIcon(node.Page) + " " + node.Page.Name

		if len(node.Children) > 0 && !node.Expanded {
			line += " ..."
		}

		if node.Page.Path == s.CurrentPath {
			line = "‹b " + line + "›b "
		}

		if i == s.Cursor {
			line = "¤KK ‹b " + s.fitSidebarLine(line) + "›b ¤ "
		} else {
			line = s.fitSidebarLine(line)
		}

		lines = append(lines, line)
	}

	return lines
}

func (s *Sidebar) fitSidebarLine(line string) string {
	width := 0
	if len(s.Bounds.ActualSize) >= 1 {
		width = s.Bounds.ActualSize[0]
	}
	if width <= 0 || s.Utilities == nil {
		return line
	}

	if s.Utilities.VisibleLength(line) > width {
		line = s.Utilities.CutVisible(line, width)
	}

	if s.Utilities.VisibleLength(line) < width {
		line += strings.Repeat(" ", width-s.Utilities.VisibleLength(line))
	}

	return line
}

func sidebarGreekIcon(page *filesystem.Page) string {
	if page == nil {
		return " "
	}

	if page.Type == "deep" {
		return sidebarGreekUpper(page.Path)
	}

	return sidebarGreekLower(page.Path)
}

func sidebarGreekUpper(seed string) string {
	letters := []string{
		"Α", "Β", "Γ", "Δ", "Ε", "Ζ", "Η", "Θ",
		"Ι", "Κ", "Λ", "Μ", "Ν", "Ξ", "Ο", "Π",
		"Ρ", "Σ", "Τ", "Υ", "Φ", "Χ", "Ψ", "Ω",
	}

	return letters[sidebarHashIndex(seed, len(letters))]
}

func sidebarGreekLower(seed string) string {
	letters := []string{
		"α", "β", "γ", "δ", "ε", "ζ", "η", "θ",
		"ι", "κ", "λ", "μ", "ν", "ξ", "ο", "π",
		"ρ", "σ", "τ", "υ", "φ", "χ", "ψ", "ω",
	}

	return letters[sidebarHashIndex(seed, len(letters))]
}

func sidebarHashIndex(seed string, size int) int {
	if size <= 0 {
		return 0
	}

	hash := 0
	for _, r := range seed {
		hash = int(r) + ((hash << 5) - hash)
	}

	if hash < 0 {
		hash = -hash
	}

	return hash % size
}
