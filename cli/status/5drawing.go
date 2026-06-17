package status

import (
	"env/cli"
	"env/cli/editor"
	"env/cli/picker"
	"env/engine"
	"env/filesystem"
	"env/routines"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type statusSegment struct {
	Text     string
	Style    string
	Priority int
}

func (s *Status) Draw() *engine.Queue {
	if s == nil || !s.On || s.Bounds == nil || len(s.Bounds.Fullsize) < 2 || len(s.Bounds.Size) < 2 {
		return nil
	}

	w := s.Bounds.Size[0]
	if w <= 0 || s.Bounds.Size[1] <= 0 {
		return nil
	}

	line := s.line(w)

	contentFrame := s.Utilities.GenerateFrame(
		engine.Boundaries{Fullsize: s.Bounds.Fullsize, Pos: routines.Bound(s.Bounds.Pos), Size: routines.Bound(s.Bounds.Size)},
		[]string{line},
		0,
	)

	fullW := s.Bounds.Fullsize[0]
	baseFrame := s.Utilities.GenerateFrame(
		engine.Boundaries{Fullsize: s.Bounds.Fullsize, Pos: routines.Bound{0, s.Bounds.Pos[1]}, Size: routines.Bound{fullW, 1}},
		[]string{"§8F0 " + strings.Repeat(" ", fullW)},
		0,
	)
	frame := s.Utilities.MergeFrames(*contentFrame, *baseFrame)

	return &engine.Queue{
		Size:   s.Bounds.Fullsize,
		Frames: []engine.Frame{frame},
		Cycle:  false,
	}
}

func (s *Status) line(width int) string {
	left := []statusSegment{{Text: statusInteractionMode(s.Parent), Style: "‹b ", Priority: 0}}
	left = append(left, statusTimeSegments(width)...)

	client := focusedClient(s.Parent)
	if client == nil {
		return s.fitStatusZones(left, statusSegment{Text: "no-client", Style: "¤88 ", Priority: 0}, []statusSegment{s.screenSizeSegment()}, width)
	}

	center := statusSegment{
		Text:     fmt.Sprintf("%s:%s", client.GetKind(), statusSpec(client)),
		Priority: 0,
	}

	right := s.clientInfoSegments(client)
	right = append(right, s.screenSizeSegment())

	return s.fitStatusZones(left, center, right, width)
}

func statusTimeSegments(width int) []statusSegment {
	now := time.Now()
	if width < 50 {
		return nil
	}
	if width < 78 {
		return []statusSegment{{Text: now.Format("15:04"), Priority: 0}}
	}
	return []statusSegment{
		{Text: statusWeekday(now.Weekday()), Priority: 0},
		{Text: now.Format("15:04"), Priority: 0},
		{Text: now.Format("02.01.2006"), Priority: 0},
	}
}

func statusInteractionMode(parent cli.Parent) string {
	if parent == nil {
		return "NORMAL"
	}
	mode := strings.ToUpper(strings.TrimSpace(parent.GetInteractionMode()))
	switch mode {
	case "MENU":
		return mode
	case "TABS":
		return "TAB"
	default:
		return "NORMAL"
	}
}

func statusWeekday(day time.Weekday) string {
	switch day {
	case time.Monday:
		return "Monday"
	case time.Tuesday:
		return "Tuesday"
	case time.Wednesday:
		return "Wednesday"
	case time.Thursday:
		return "Thursday"
	case time.Friday:
		return "Friday"
	case time.Saturday:
		return "Saturday"
	case time.Sunday:
		return "Sunday"
	default:
		return "-"
	}
}

func (s *Status) screenSizeSegment() statusSegment {
	if s == nil || s.Bounds == nil || len(s.Bounds.Fullsize) < 2 {
		return statusSegment{Text: "-", Priority: 0}
	}
	return statusSegment{Text: fmt.Sprintf("%dx%d", s.Bounds.Fullsize[0], s.Bounds.Fullsize[1]), Priority: 0}
}

func (s *Status) clientInfoSegments(client cli.Cli) []statusSegment {
	if e, ok := client.(*editor.Editor); ok {
		return s.editorInfo(e)
	}
	if p, ok := client.(*picker.Picker); ok {
		return s.pickerInfo(p)
	}
	return nil
}

func (s *Status) editorInfo(e *editor.Editor) []statusSegment {
	cursorX := 0
	cursorY := 0

	if len(e.Cursor) >= 2 {
		cursorX = e.Cursor[0] + 1
		cursorY = e.Cursor[1] + 1
	}

	size := pageSizeBytes(e.GetPath(), e.Content)

	return []statusSegment{
		{Text: fmt.Sprintf("%d:%d", cursorY, cursorX), Priority: 1},
		{Text: statusByteSize(size), Priority: 2},
	}
}

func statusByteSize(size int) string {
	if size < 0 {
		size = 0
	}
	units := []string{"kb", "mb", "gb"}
	value := float64(size) / 1024
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	if value >= 10 {
		return fmt.Sprintf("%.0f%s", value, units[unit])
	}
	return fmt.Sprintf("%.1f%s", value, units[unit])
}

func focusedClient(parent cli.Parent) cli.Cli {
	if parent == nil {
		return nil
	}

	clients := parent.GetClients()
	focus := parent.GetFocus()

	if focus < 0 || focus >= len(clients) {
		return nil
	}

	return clients[focus]
}

func pageSizeBytes(path string, content []string) int {
	if path != "" {
		home, _ := os.UserHomeDir()
		full := filepath.Join(home, "prsnl.spc", path)

		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			return int(info.Size())
		}
	}

	return len([]byte(strings.Join(content, "\n")))
}

func cacheSyncText(parent cli.Parent, e *editor.Editor) string {
	if parent == nil || e == nil || parent.GetFilesystem() == nil {
		return "unknown"
	}

	cached := findCachedPage(parent.GetFilesystem().Cache, e.GetPath())
	if cached == nil {
		return "missing"
	}

	if cached.Stage == "" {
		if sameContent(cached.Content, e.Content) {
			return "synced"
		}

		return "edited"
	}

	return cached.Stage
}

func findCachedPage(page *filesystem.Page, path string) *filesystem.Page {
	if page == nil {
		return nil
	}

	if page.Path == path {
		return page
	}

	for _, child := range page.Children {
		found := findCachedPage(child, path)
		if found != nil {
			return found
		}
	}

	return nil
}

func sameContent(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func (s *Status) pickerInfo(p *picker.Picker) []statusSegment {
	return []statusSegment{
		{Text: "sort:" + nonEmpty(p.Sorting, "-"), Priority: 1},
		{Text: "scope:" + nonEmpty(p.Scope, "-"), Priority: 1},
		{Text: "score:" + nonEmpty(p.Mode, "-"), Priority: 1},
	}
}

func (s *Status) fitSegments(segments []statusSegment, width int) string {
	active := append([]statusSegment(nil), segments...)
	line := s.renderSegments(active)

	for s.visibleLength(line) > width {
		drop := leastImportantSegment(active)
		if drop < 0 || active[drop].Priority <= 0 {
			break
		}

		active = append(active[:drop], active[drop+1:]...)
		line = s.renderSegments(active)
	}

	if s.visibleLength(line) > width && s.Utilities != nil {
		line = s.Utilities.CutVisible(line, width)
	}

	if pad := width - s.visibleLength(line); pad > 0 {
		line += strings.Repeat(" ", pad)
	}

	return line
}

func (s *Status) fitStatusZones(left []statusSegment, center statusSegment, right []statusSegment, width int) string {
	if width <= 0 {
		return ""
	}

	activeRight := append([]statusSegment(nil), right...)
	leftText := s.renderSegments(left)
	centerText := renderStatusSegment(center)
	rightText := s.renderSegmentsInline(activeRight)

	for len(activeRight) > 1 && s.visibleLength(leftText)+s.visibleLength(centerText)+s.visibleLength(rightText)+4 > width {
		drop := leastImportantSegment(activeRight[:len(activeRight)-1])
		if drop < 0 || activeRight[drop].Priority <= 0 {
			break
		}
		activeRight = append(activeRight[:drop], activeRight[drop+1:]...)
		rightText = s.renderSegmentsInline(activeRight)
	}

	leftW := s.visibleLength(leftText)
	rightW := s.visibleLength(rightText)
	centerW := s.visibleLength(centerText)

	rightStart := width - rightW
	if rightStart < 0 {
		rightStart = 0
	}

	centerSlotStart := leftW + 2
	centerSlotEnd := rightStart - 2
	if centerSlotEnd < centerSlotStart {
		centerSlotEnd = centerSlotStart
	}
	centerSlotW := centerSlotEnd - centerSlotStart

	centerStart := centerSlotStart + (centerSlotW-centerW)/2
	if centerStart < centerSlotStart {
		centerStart = centerSlotStart
	}

	centerMax := centerSlotEnd - centerStart
	if centerMax < 0 {
		centerMax = 0
	}
	if centerW > centerMax && s.Utilities != nil {
		centerText = s.Utilities.CutVisible(centerText, centerMax)
		centerW = s.visibleLength(centerText)
	}

	line := leftText
	if s.visibleLength(line) > width && s.Utilities != nil {
		line = s.Utilities.CutVisible(line, width)
	}

	if centerW > 0 {
		pad := centerStart - s.visibleLength(line)
		if pad < 1 {
			pad = 1
		}
		line += strings.Repeat(" ", pad) + centerText
	}

	if rightW > 0 {
		pad := rightStart - s.visibleLength(line)
		if pad < 1 {
			pad = 1
		}
		line += strings.Repeat(" ", pad) + rightText
	}

	if s.visibleLength(line) > width && s.Utilities != nil {
		line = s.Utilities.CutVisible(line, width)
	}
	if pad := width - s.visibleLength(line); pad > 0 {
		line += strings.Repeat(" ", pad)
	}

	return line
}

func leastImportantSegment(segments []statusSegment) int {
	drop := -1
	dropPriority := -1

	for i := len(segments) - 1; i >= 0; i-- {
		if segments[i].Priority > dropPriority {
			drop = i
			dropPriority = segments[i].Priority
		}
	}

	return drop
}

func (s *Status) renderSegments(segments []statusSegment) string {
	if len(segments) == 0 {
		return "§8F0 "
	}

	var b strings.Builder
	b.WriteString("§8F0 ")

	for i, segment := range segments {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(renderStatusSegment(segment))
	}

	return b.String()
}

func (s *Status) renderSegmentsInline(segments []statusSegment) string {
	if len(segments) == 0 {
		return ""
	}

	var b strings.Builder
	for i, segment := range segments {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(renderStatusSegment(segment))
	}

	return b.String()
}

func renderStatusSegment(segment statusSegment) string {
	text := strings.TrimSpace(segment.Text)
	if text == "" {
		text = "-"
	}
	text = statusDashToken(text)

	switch {
	case strings.HasPrefix(segment.Style, "¤"):
		return segment.Style + text + "¤ "
	case strings.HasPrefix(segment.Style, "‹b"):
		return segment.Style + text + "›b "
	default:
		return text
	}
}

func statusDashToken(text string) string {
	text = strings.TrimSpace(text)
	if text == "" || text == "-" {
		return "-"
	}
	return "-" + text + "-"
}

func (s *Status) visibleLength(line string) int {
	if s.Utilities == nil {
		return len([]rune(line))
	}
	return s.Utilities.VisibleLength(line)
}

func statusSpec(client cli.Cli) string {
	if client == nil {
		return "-"
	}

	spec := client.GetSpec()
	if spec == "" {
		spec = client.GetPath()
	}
	if spec == "" {
		spec = client.GetTitle()
	}

	return shortenStatusPath(spec, 46)
}

func shortenStatusPath(path string, limit int) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "-"
	}

	runes := []rune(path)
	if limit <= 0 || len(runes) <= limit {
		return path
	}
	if limit <= 5 {
		return string(runes[:limit])
	}

	parts := strings.Split(path, "/")
	if len(parts) >= 3 {
		candidate := parts[0] + "/.../" + parts[len(parts)-1]
		if len([]rune(candidate)) <= limit {
			return candidate
		}
	}

	head := (limit - 3) / 2
	tail := limit - 3 - head
	return string(runes[:head]) + "..." + string(runes[len(runes)-tail:])
}

func nonEmpty(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
