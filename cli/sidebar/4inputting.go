package sidebar

import (
	"env/routines"
	"strings"
)

func (s *Sidebar) Input(newinput *routines.Input) {
	if len(s.Flat) == 0 {
		return
	}

	switch newinput.Key {
	case "up":
		s.moveCursor(-1)

	case "down":
		s.moveCursor(1)

	case "left":
		s.collapseCurrent()

	case "right":
		s.expandCurrent()

	case "enter":
		s.openSelectedInNewTab(true)

	case "ctrl+enter":
		s.openSelectedInNewTab(false)
	}
}
func (s *Sidebar) moveCursor(delta int) {
	s.Cursor += delta

	if s.Cursor < 0 {
		s.Cursor = 0
	}

	if s.Cursor >= len(s.Flat) {
		s.Cursor = len(s.Flat) - 1
	}

	s.updateViewport()
	if s.Switch {
		s.RequestSelectedPreview()
	}
}

func (s *Sidebar) collapseCurrent() {
	if s.Cursor < 0 || s.Cursor >= len(s.Flat) {
		return
	}

	node := s.Flat[s.Cursor]
	if node == nil {
		return
	}

	if len(node.Children) == 0 {
		return
	}

	if !node.Expanded {
		return
	}

	node.Expanded = false
	s.rebuildFlatKeepingRootMode()
	s.updateViewport()
	if s.Switch {
		s.RequestSelectedPreview()
	}
}

func (s *Sidebar) expandCurrent() {
	if s.Cursor < 0 || s.Cursor >= len(s.Flat) {
		return
	}

	node := s.Flat[s.Cursor]
	if node == nil {
		return
	}

	if len(node.Children) == 0 {
		return
	}

	if node.Expanded {
		return
	}

	node.Expanded = true
	s.rebuildFlatKeepingRootMode()
	s.updateViewport()
	if s.Switch {
		s.RequestSelectedPreview()
	}
}

func (s *Sidebar) rebuildFlatKeepingRootMode() {
	if s.Page == nil {
		return
	}

	root := wrapPageTree(s.Page, nil, 0)
	if root == nil {
		s.Flat = []*Node{}
		s.Cursor = 0
		return
	}

	if strings.HasPrefix(s.Page.Path, "e.proj/") {
		s.Flat = s.buildProjectSidebarFlat(root)
	} else if sidebarShouldHideContainer(s.Page.Path) {
		s.Flat = flattenVisibleRootList(root.Children)
	} else {
		s.Flat = flattenVisible(root)
	}

	if s.Cursor >= len(s.Flat) {
		s.Cursor = len(s.Flat) - 1
	}

	if s.Cursor < 0 {
		s.Cursor = 0
	}
	s.refreshWidth()
}

func (s *Sidebar) openSelectedInNewTab(follow bool) {
	if s.Cursor < 0 || s.Cursor >= len(s.Flat) {
		return
	}

	node := s.Flat[s.Cursor]
	if node == nil || node.Page == nil || node.Virtual {
		return
	}

	if node.Page.Path == "" {
		return
	}

	path := node.Page.Path

	mode := "next-stay"
	if follow {
		mode = "next"
		s.Switch = false
		s.PreviewPath = ""
		s.Preview = nil
	}

	s.Parent.AddClients("editor:"+path, mode)
}
