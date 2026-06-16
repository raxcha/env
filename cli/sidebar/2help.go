package sidebar

import (
	"env/filesystem"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func wrapPageTree(page *filesystem.Page, parent *Node, depth int) *Node {
	if page == nil {
		return nil
	}

	node := &Node{
		Page:     page,
		Parent:   parent,
		Depth:    depth,
		Expanded: true,
	}

	for _, child := range page.Children {
		childNode := wrapPageTree(child, node, depth+1)
		if childNode != nil {
			node.Children = append(node.Children, childNode)
		}
	}

	return node
}

func flattenVisible(root *Node) []*Node {
	var result []*Node

	var walk func(node *Node)
	walk = func(node *Node) {
		if node == nil {
			return
		}

		result = append(result, node)

		if !node.Expanded {
			return
		}

		for _, child := range node.Children {
			walk(child)
		}
	}

	walk(root)
	return result
}

func (s *Sidebar) SetPage(page *filesystem.Page) {
	if page == nil {
		return
	}

	if s.Path != "" && page.Path != s.Path {
		return
	}

	s.Page = page
	s.Path = page.Path
	s.sortSidebarTree(page)

	root := wrapPageTree(page, nil, 0)

	if root == nil {
		s.Flat = []*Node{}
		s.Cursor = 0
		return
	}

	if strings.HasPrefix(page.Path, "e.proj/") {
		s.Flat = s.buildProjectSidebarFlat(root)
	} else if sidebarShouldHideContainer(page.Path) {
		s.Flat = flattenVisibleRootList(root.Children)
	} else {
		s.Flat = flattenVisible(root)
	}

	if s.Cursor < 0 {
		s.Cursor = 0
	}

	if s.Cursor >= len(s.Flat) {
		s.Cursor = len(s.Flat) - 1
	}

	if len(s.Flat) == 0 {
		s.Cursor = 0
	}

	s.selectCurrentPath()
	s.refreshWidth()
	s.updateViewport()
}

func (s *Sidebar) rootNodes() []*Node {
	if len(s.Flat) == 0 {
		return nil
	}

	roots := []*Node{}

	for _, node := range s.Flat {
		if node == nil {
			continue
		}

		if node.Parent == nil {
			roots = append(roots, node)
			continue
		}

		if node.Depth == 1 && node.Parent.Parent == nil {
			roots = append(roots, node)
		}
	}

	return roots
}

func flattenVisibleRootList(roots []*Node) []*Node {
	result := []*Node{}

	var walk func(*Node)
	walk = func(node *Node) {
		if node == nil {
			return
		}

		result = append(result, node)

		if !node.Expanded {
			return
		}

		for _, child := range node.Children {
			walk(child)
		}
	}

	for _, root := range roots {
		walk(root)
	}

	return result
}

func (s *Sidebar) selectedPage() *filesystem.Page {
	if s.Cursor < 0 || s.Cursor >= len(s.Flat) {
		return nil
	}

	node := s.Flat[s.Cursor]
	if node == nil || node.Virtual {
		return nil
	}

	return node.Page
}

func (s *Sidebar) RequestSelectedPreview() {
	page := s.selectedPage()
	if page == nil {
		return
	}

	s.PreviewPath = page.Path

	if page.Path == s.CurrentPath {
		s.Preview = page
		return
	}

	fs := s.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	req.Path = page.Path
	req.Depth = 0
	req.Cont = true
	req.Meta = true
	req.Opts = true
	req.Sort = s.Sort
	fs.Load(req)
}

func (s *Sidebar) SetPreview(page *filesystem.Page) {
	if page == nil {
		return
	}

	if page.Path != s.PreviewPath {
		return
	}

	s.Preview = page
}

func CutSidebarItems(height int, selected int, total int) int {
	if height <= 0 || total <= height {
		return 0
	}

	if selected < 0 {
		selected = 0
	}

	if selected >= total {
		selected = total - 1
	}

	edge := height / 4
	if edge < 1 {
		edge = 1
	}

	if selected <= edge {
		return 0
	}

	if selected >= total-1-edge {
		return total - height
	}

	progress := float64(selected-edge) / float64((total-1)-(edge*2))
	targetY := edge + int(progress*float64((height-1)-(edge*2)))

	start := selected - targetY
	maxStart := total - height

	if start < 0 {
		return 0
	}

	if start > maxStart {
		return maxStart
	}

	return start
}

func (s *Sidebar) updateViewport() {
	if len(s.Bounds.ActualSize) < 2 {
		s.ViewStart = 0
		return
	}

	height := s.Bounds.ActualSize[1]
	total := len(s.Flat)

	s.ViewStart = CutSidebarItems(height, s.Cursor, total)
}

func (s *Sidebar) sortSidebarTree(page *filesystem.Page) {
	if page == nil {
		return
	}

	for _, child := range page.Children {
		s.sortSidebarTree(child)
	}

	sort.SliceStable(page.Children, func(i, j int) bool {
		return sidebarLessPage(page.Path, page.Children[i], page.Children[j])
	})
}

func sidebarLessPage(context string, a *filesystem.Page, b *filesystem.Page) bool {
	root := sidebarRoot(context)

	switch root {
	case "a.log":
		at, aok := sidebarTitleTime(a)
		bt, bok := sidebarTitleTime(b)
		if aok != bok {
			return aok
		}
		if aok && !at.Equal(bt) {
			return at.After(bt)
		}

	case "b.rec", "c.rand":
		at, aok := sidebarMetadataTime(a)
		bt, bok := sidebarMetadataTime(b)
		if aok != bok {
			return aok
		}
		if aok && !at.Equal(bt) {
			return at.After(bt)
		}
	}

	if a == nil || b == nil {
		return a != nil
	}
	return strings.ToLower(a.Name) < strings.ToLower(b.Name)
}

func sidebarTitleTime(page *filesystem.Page) (time.Time, bool) {
	if page == nil {
		return time.Time{}, false
	}

	name := page.Name
	if name == "" {
		name = filepath.Base(page.Path)
	}

	return sidebarParseTime(name)
}

func sidebarMetadataTime(page *filesystem.Page) (time.Time, bool) {
	if page == nil || page.Metadata == nil {
		return time.Time{}, false
	}

	for _, key := range []string{"time", "release-time", "last-edited-time"} {
		switch value := page.Metadata[key].(type) {
		case time.Time:
			return value, true
		case string:
			if t, ok := sidebarParseTime(value); ok {
				return t, true
			}
		default:
			if t, ok := sidebarParseTime(fmt.Sprint(value)); ok {
				return t, true
			}
		}
	}

	return time.Time{}, false
}

func sidebarParseTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339,
		"02.01.2006 15:04",
		"02.01.2006",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, raw, time.Local)
		if err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

func (s *Sidebar) selectCurrentPath() {
	if s.CurrentPath == "" {
		return
	}

	for i, node := range s.Flat {
		if node != nil && node.Page != nil && node.Page.Path == s.CurrentPath {
			s.Cursor = i
			return
		}
	}
}

func sidebarShowsRootChildren(path string) bool {
	return path == "a.log" ||
		path == "b.rec" ||
		path == "c.rand" ||
		path == "d.fami"
}

func sidebarRoot(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}

	return strings.Split(path, "/")[0]
}

func sidebarShouldHideContainer(path string) bool {
	switch sidebarRoot(path) {
	case "a.log", "b.rec", "c.rand", "d.fami", "e.proj":
		return true
	default:
		return false
	}
}

func (s *Sidebar) SetEvents(page *filesystem.Page) {
	if page == nil {
		return
	}

	if page.Path != "b.rec" {
		return
	}

	s.EventsPage = page

	if s.Page != nil && strings.HasPrefix(s.Page.Path, "e.proj/") {
		root := wrapPageTree(s.Page, nil, 0)
		s.Flat = s.buildProjectSidebarFlat(root)

		if s.Cursor >= len(s.Flat) {
			s.Cursor = len(s.Flat) - 1
		}

		if s.Cursor < 0 {
			s.Cursor = 0
		}

		s.refreshWidth()
		s.updateViewport()
	}
}

func (s *Sidebar) buildProjectSidebarFlat(projectRoot *Node) []*Node {
	if projectRoot == nil || projectRoot.Page == nil {
		return []*Node{}
	}

	projectName := projectMetadataName(projectRoot.Page)
	events := s.filteredProjectEvents(projectName)
	out := []*Node{}

	for _, eventPage := range events {
		eventNode := wrapPageTree(eventPage, nil, 0)
		if eventNode != nil {
			out = append(out, eventNode)
		}
	}

	return flattenVisibleRootList(out)
}

func projectMetadataName(page *filesystem.Page) string {
	if page == nil {
		return ""
	}

	if page.Metadata != nil {
		if name, ok := page.Metadata["name"].(string); ok && strings.TrimSpace(name) != "" {
			return strings.TrimSpace(name)
		}
	}

	return page.Name
}

func (s *Sidebar) filteredProjectEvents(projectName string) []*filesystem.Page {
	if s.EventsPage == nil || projectName == "" {
		return []*filesystem.Page{}
	}

	events := []*filesystem.Page{}

	for _, child := range s.EventsPage.Children {
		if child == nil || child.Metadata == nil {
			continue
		}

		if !sidebarOwnedByProject(child, projectName) {
			continue
		}

		events = append(events, child)
	}

	sort.SliceStable(events, func(i, j int) bool {
		return sidebarLessPage("b.rec", events[i], events[j])
	})

	return events
}

func sidebarOwnedByProject(page *filesystem.Page, projectName string) bool {
	if page == nil || page.Metadata == nil || projectName == "" {
		return false
	}

	projectKey := sidebarDashedName(projectName)
	for _, key := range []string{"owned-by", "owner", "project", "projects"} {
		raw := strings.TrimSpace(fmt.Sprint(page.Metadata[key]))
		if raw == "" {
			continue
		}

		for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ';' || r == ' '
		}) {
			if sidebarDashedName(part) == projectKey {
				return true
			}
		}
	}

	return false
}

func sidebarDashedName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	for strings.Contains(value, "--") {
		value = strings.ReplaceAll(value, "--", "-")
	}
	return strings.Trim(value, "-")
}
