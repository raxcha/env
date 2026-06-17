package sidebar

import (
	"env/engine"
	"strings"
)

func (s *Sidebar) Resize(newbounds *engine.Boundaries) {

	s.mergeBoundaries(newbounds)
}

func (s *Sidebar) mergeBoundaries(bounds *engine.Boundaries) {

	copy := *bounds
	copy.ActualPos = bounds.Pos
	copy.ActualSize = []int{s.idealWidth(bounds.Size[0]), bounds.Size[1]}
	s.Bounds = copy
	s.updateViewport()
}

func (s *Sidebar) refreshWidth() {
	if len(s.Bounds.Size) < 2 {
		return
	}

	s.Bounds.ActualSize = []int{s.idealWidth(s.Bounds.Size[0]), s.Bounds.Size[1]}
	s.updateViewport()
}

func (s *Sidebar) idealWidth(total int) int {
	if total <= 0 {
		return 0
	}

	maxLabel := 12
	for _, node := range s.Flat {
		label := s.measureNodeLabel(node)
		if label > maxLabel {
			maxLabel = label
		}
	}

	width := maxLabel + 5
	minW := 16
	maxW := total / 2
	if maxW < minW {
		maxW = total
	}

	if width < minW {
		width = minW
	}
	if width > maxW {
		width = maxW
	}
	if width > total-1 {
		width = total - 1
	}
	if width < 1 {
		width = 1
	}

	return width
}

func (s *Sidebar) measureNodeLabel(node *Node) int {
	if node == nil || node.Page == nil || s.Utilities == nil {
		return 0
	}

	depth := node.Depth - 1
	if depth < 0 {
		depth = 0
	}

	return s.Utilities.VisibleLength(strings.Repeat("  ", depth) + sidebarGreekIcon(node.Page) + " " + node.Page.Name)
}
