package editor

import (
	"env/settings"
	"env/types"
	"env/utils"
	"path/filepath"
)

const sidebarWidth = 32

// Sidebar is the single navigation surface shared by the open editors.
type Sidebar struct {
	Sizes     types.Dimensions
	rows      []sidebarRow
	focus     int
	collapsed map[string]bool
}
type sidebarRow struct {
	owner   *Editor
	index   int
	label   string
	context string
}

func (s *Sidebar) Prepare(editors []*Editor, area types.Dimensions) types.Dimensions {
	var selected sidebarRow
	hasSelection := s.focus >= 0 && s.focus < len(s.rows)
	if hasSelection {
		selected = s.rows[s.focus]
	}
	s.rows = nil
	seen := map[string]bool{}
	for _, e := range editors {
		e.generateAllEntries()
		context := filepath.Clean(e.Spec)
		if len(e.Requests) > 0 {
			context = e.Requests[0].Path
		} else if e.Request != nil {
			context = e.Request.Path
		}
		if seen[context] {
			continue
		}
		seen[context] = true
		s.rows = append(s.rows, sidebarRow{owner: e, index: -1, label: context, context: context})
		if s.collapsed[context] {
			continue
		}
		for i, en := range e.Entries {
			s.rows = append(s.rows, sidebarRow{owner: e, index: i, label: e.selectLabel(en), context: context})
		}
	}
	s.focus = max(0, min(s.focus, len(s.rows)-1))
	if hasSelection {
		for i, row := range s.rows {
			if row.context == selected.context && row.index == selected.index {
				s.focus = i
				break
			}
		}
	}
	s.Sizes = area
	if len(s.rows) == 0 {
		s.Sizes.Size = types.Size{}
		return area
	}
	if 2*area.Size[1] > area.Size[0] {
		s.Sizes.Size[1] = min(len(s.rows), max(0, area.Size[1]/2))
		area.Size[1] -= s.Sizes.Size[1]
		s.Sizes.Pos[1] += area.Size[1]
	} else {
		s.Sizes.Size[0] = min(sidebarWidth, max(0, area.Size[0]/2))
		area.Pos[0] += s.Sizes.Size[0]
		area.Size[0] -= s.Sizes.Size[0]
	}
	return area
}
func (s *Sidebar) move(delta int) {
	s.focus = max(0, min(s.focus+delta, len(s.rows)-1))
}
func (s *Sidebar) Input(input *types.Input) {
	if len(s.rows) == 0 {
		return
	}
	switch input.Description {
	case "up":
		s.move(-1)
	case "down":
		s.move(1)
	default:
		row := s.rows[s.focus]
		if row.index < 0 {
			if s.collapsed == nil {
				s.collapsed = map[string]bool{}
			}
			switch input.Description {
			case "left":
				s.collapsed[row.context] = true
			case "right":
				s.collapsed[row.context] = false
			case "enter":
				s.collapsed[row.context] = !s.collapsed[row.context]
			}
		} else {
			row.owner.Focus = row.index
			row.owner.InputSelect(input)
		}
	}
	if s.focus < len(s.rows) {
		row := s.rows[s.focus]
		if row.index >= 0 {
			row.owner.Focus = row.index
		}
	}
}
func (s *Sidebar) Draw(settings *settings.Settings) *types.Queue {
	lines := make([]string, 0, len(s.rows))
	for i, row := range s.rows {
		style := "§xy0 "
		if i == s.focus {
			style = "§YX0 "
		}
		label := " " + row.label
		width := max(0, s.Sizes.Size[0])
		if utils.VisibleLength(label) > width {
			ellipsis := "..."[:min(3, width)]
			label = utils.VisibleSlice(label, 0, width-len(ellipsis)) + ellipsis
		}
		if row.index >= 0 && row.owner.Entries[row.index].Type == "load-more" {
			label = "‹b " + label + "›b "
		}
		lines = append(lines, style+label)
	}
	start := 0
	height := max(0, s.Sizes.Size[1])
	if len(lines) > height && len(lines) > 1 {
		start = s.focus * (len(lines) - height) / (len(lines) - 1)
	}
	frame := settings.GenerateFrame(s.Sizes, lines[start:], 0, []int{0, 0, 0, 0})
	return settings.GenerateQueue(s.Sizes.Full, []*types.Frame{frame}, false)
}
