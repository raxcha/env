package editor

import (
	"env/engine"
	"env/routines"
	"env/utilities"
	"regexp"
	"strconv"
	"strings"
)

func (e Editor) Draw() *engine.Queue {

	if e.Bounds == nil ||
		len(e.Bounds.ActualPos) < 2 ||
		len(e.Bounds.ActualSize) < 2 ||
		len(e.Bounds.Fullsize) < 2 {
		return utilities.NewQueue()
	}

	// DO A LOADING PAGE ...
	if e.Page == nil {
		return utilities.NewQueue()
	}

	if len(e.Content) == 0 {
		e.Content = []string{""}
	}

	x := e.Bounds.ActualPos[0]
	y := e.Bounds.ActualPos[1]
	w := e.Bounds.ActualSize[0]
	h := e.Bounds.ActualSize[1]

	if e.panelMode() == "sidebar" {
		return e.drawSidebarPanelOnly()
	}

	sidebarActive := e.panelMode() == "whole" && e.Sidebar != nil && e.Sidebar.On && e.Sidebar.Switch && !e.Zenmode && e.Bounds.Mode != "fibonacci"

	displayContent := e.Content
	displayCursor := e.Cursor
	displayPath := e.Path
	previewingSidebar := false

	if sidebarActive && e.Sidebar != nil && e.Sidebar.Preview != nil {
		if e.Sidebar.Preview.Path != e.Path {
			displayContent = e.Sidebar.Preview.Content
			displayCursor = []int{0, -1}
			displayPath = e.Sidebar.Preview.Path
			previewingSidebar = true
		}
	}

	sidebarWidth := 0
	if sidebarActive {
		sidebarWidth = e.Sidebar.Bounds.ActualSize[0]

		if sidebarWidth < 1 {
			sidebarWidth = int(float64(e.Bounds.ActualSize[0]) * 0.25)
		}

		if sidebarWidth > w-1 {
			sidebarWidth = w - 1
		}

		x += sidebarWidth
		w -= sidebarWidth
	}

	if e.Zenmode {
		zenW := int(float64(w) * 0.65)
		if zenW < 1 {
			zenW = 1
		}
		x += (w - zenW) / 2
		w = zenW
	}

	panelW := w
	numsize := len(strconv.Itoa(len(displayContent))) + 1
	numbersWidth := 0
	gap := 0

	if e.Numbers && !e.Zenmode {
		numbersWidth = numsize + 1
		gap = 1
		w -= numbersWidth + gap
	}

	if w < 1 {
		w = 1
	}

	numbers := []string{}
	lines := append([]string(nil), displayContent...)
	rawLines := append([]string(nil), displayContent...)
	state := NewPreviewStyleState()
	colorSpecialLines := isALogEditorPath(displayPath)

	for i := range lines {
		dashedSize := 0
		if i != displayCursor[1] && dashed.MatchString(rawLines[i]) {
			dashedSize = DashedLineVisualSize(rawLines, i, e.Utilities)
		}
		lines[i] = StyleContentLine(lines[i], i == displayCursor[1], w, &state, colorSpecialLines, dashedSize)
		if previewingSidebar {
			lines[i] = sidebarPreviewLineStyle(lines[i])
		}
		if i == displayCursor[1] {
			actual := CursorActualIndex(lines[i], displayCursor[0])
			lines[i] = InsertCursorMarker(lines[i], actual)
		}
	}

	selected := displayCursor[1]
	if selected < 0 {
		selected = 0
	}

	editorH := h
	showHeader := !previewingSidebar && displayCursor[1] <= 0
	showFooter := !previewingSidebar && displayCursor[1] >= len(displayContent)-1
	headerH := 0
	footerH := 0
	if showHeader && editorH > headerH+footerH {
		headerH = 1
	}
	if showFooter && editorH > headerH+footerH {
		footerH = 1
	}

	contentH := editorH - headerH - footerH
	if contentH < 1 {
		contentH = 1
	}

	start := CutLines(contentH, selected, len(lines))

	lines = lines[start:min(start+contentH, len(lines))]
	contentRows := len(lines)
	if showHeader {
		lines = append([]string{e.shiftHeaderLine(e.headerLine(w+2, e.cursorIsOnHeader(), e.Cursor[0]), w)}, lines...)
	}
	if footerH > 0 {
		lines = append(lines, "§A80 "+strings.Repeat(" ", w))
	}
	for len(lines) < editorH {
		lines = append(lines, "")
	}

	if e.Numbers {
		if showHeader {
			numbers = append(numbers, "§8B0 "+strings.Repeat(" ", numbersWidth+gap))
		}
		for i, line := range lines {
			if showHeader && i == 0 {
				continue
			}
			if footerH > 0 && i == contentRows+headerH {
				numbers = append(numbers, "§A80 "+strings.Repeat(" ", numbersWidth+gap))
				continue
			}

			contentIndex := i - headerH
			if contentIndex < 0 || contentIndex >= contentRows {
				numbers = append(numbers, "§AB0 "+strings.Repeat(" ", numbersWidth+gap))
				continue
			}

			prefix := numberPrefixFromLine(line)

			if start+contentIndex == displayCursor[1] {
				prefix += "‹b "
			}

			visual := VisibleLength(line)
			rows := 1

			if w > 0 {
				rows = (visual + w - 1) / w
				if rows < 1 {
					rows = 1
				}
			}

			for j := 0; j < rows; j++ {
				if j == 0 {
					numbers = append(numbers, prefix+RightNumber(start+contentIndex+1, numsize+1))
				} else {
					numbers = append(numbers, prefix+strings.Repeat(" ", numsize+1))
				}
			}
		}
	}

	numbersX := x
	lineX := x

	if e.Numbers && !e.Zenmode {
		lineX += numbersWidth + gap
	}

	linesframe := e.Utilities.GenerateFrame(
		engine.Boundaries{Fullsize: e.Bounds.Fullsize, Pos: routines.Bound{lineX, y}, Size: routines.Bound{w, editorH}},
		lines,
		0,
	)

	finalframe := *linesframe

	if e.Numbers && !e.Zenmode {
		numbersframe := e.Utilities.GenerateFrame(
			engine.Boundaries{Fullsize: e.Bounds.Fullsize, Pos: routines.Bound{numbersX, y}, Size: routines.Bound{numbersWidth + gap, editorH}},
			numbers,
			0,
		)

		finalframe = e.Utilities.MergeFrames(*numbersframe, *linesframe)
	}

	if footerH > 0 {
		footerY := y + headerH + contentRows
		footerframe := e.Utilities.GenerateFrame(
			engine.Boundaries{Fullsize: e.Bounds.Fullsize, Pos: routines.Bound{x, footerY}, Size: routines.Bound{panelW, 1}},
			[]string{editorFooterLine(panelW)},
			0,
		)
		finalframe = e.Utilities.MergeFrames(*footerframe, finalframe)
	}

	if sidebarActive {
		sidebarframe := e.Sidebar.Draw()

		if len(sidebarframe.Frames) > 0 {
			finalframe = e.Utilities.MergeFrames(sidebarframe.Frames[0], finalframe)
		}
	}

	q := utilities.NewQueue()
	q.Frames = append(q.Frames, finalframe)
	q.Size = e.Bounds.Fullsize
	return q
}

func (e Editor) drawSidebarPanelOnly() *engine.Queue {
	if e.Sidebar == nil || e.Bounds == nil || len(e.Bounds.ActualSize) < 2 || len(e.Bounds.ActualPos) < 2 || len(e.Bounds.Fullsize) < 2 {
		return utilities.NewQueue()
	}

	original := e.Sidebar.Bounds
	copy := original
	copy.Fullsize = e.Bounds.Fullsize
	copy.Pos = e.Bounds.ActualPos
	copy.Size = e.Bounds.ActualSize
	copy.ActualPos = e.Bounds.ActualPos
	copy.ActualSize = e.Bounds.ActualSize
	e.Sidebar.Bounds = copy
	q := e.Sidebar.Draw()
	e.Sidebar.Bounds = original

	return &q
}

var metadata = `^\s*[A-Za-z0-9_-]+(?:\s+[A-Za-z0-9_-]+)*\s*:`
var hashtag = `#[^\s,.;:!?()\[\]{}"']+`
var tags = `@[^\s,.;:!?()\[\]{}"']+`
var measurements = `!\d+\b`
var timestamps = `\b(?:\d{2}\.\d{2}\.\d{4}|\d{2}:\d{2})\b`
var dashed = regexp.MustCompile(`^\s*-+\s*$`)

type PreviewStyleState struct {
	Meta  bool
	Block rune
}

func NewPreviewStyleState() PreviewStyleState {
	return PreviewStyleState{
		Meta:  true,
		Block: ' ',
	}
}

func isALogEditorPath(path string) bool {
	path = strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
	path = strings.TrimPrefix(path, "/home/asdf/prsnl.spc/")
	path = strings.TrimPrefix(path, "prsnl.spc/")
	path = strings.Trim(path, "/")
	return path == "a.log" || strings.HasPrefix(path, "a.log/")
}

func StyleContentLine(raw string, current bool, width int, state *PreviewStyleState, colorSpecialLines bool, dashedSize ...int) string {
	if state == nil {
		initial := NewPreviewStyleState()
		state = &initial
	}

	line := raw
	isBlockHeader := len(raw) > 0 && strings.ContainsRune(">`~=", rune(raw[0]))
	isDashSpecialLine := strings.HasPrefix(raw, "--")
	isTildeSpecialLine := strings.HasPrefix(raw, "~~")

	if isBlockHeader {
		state.Block = rune(raw[0])
	}
	if isDashedLine(raw) {
		state.Meta = false
	}
	if isSeparatorLine(raw) {
		state.Block = ' '
	}

	if !current && dashed.MatchString(raw) {
		size := 0
		if len(dashedSize) > 0 {
			size = dashedSize[0]
		}
		if size < 1 {
			size = 1
		}
		if width > 1 && size > width-1 {
			size = width - 1
		}
		line = strings.Repeat("─", size)
	}

	if isBlockHeader {
		line, _ = wrapRegexMatches(line, hashtag, "‹b ", "›b ")
		line, _ = wrapRegexMatches(line, tags, "‹b ", "›b ")
	} else {
		line, _ = wrapRegexMatches(line, hashtag, "‹b ", "›b ")
		line, _ = wrapRegexMatches(line, tags, "‹b ", "›b ")
	}

	line, _ = wrapRegexMatches(line, measurements, "‹b ", "›b ")
	line, _ = wrapRegexMatches(line, timestamps, "‹b ", "›b ")

	if current {
		line = replaceLine(line, "*", "¬b *", "*¬b ")
		line = replaceLine(line, "%", "¬i %", "%¬i ")
		line = replaceLine(line, "_", "¬u _", "_¬u ")
		line = replaceLine(line, "^", "¬U ^", "^¬U ")
		line = replaceLine(line, "$", "¬a $", "$¬a ")
		line = replaceLine(line, "&", "¬A &", "&¬A ")
	} else {
		line = strings.ReplaceAll(line, "*", "¬b ")
		line = strings.ReplaceAll(line, "%", "¬i ")
		line = strings.ReplaceAll(line, "_", "¬u ")
		line = strings.ReplaceAll(line, "^", "¬U ")
		line = strings.ReplaceAll(line, "$", "¬a ")
		line = strings.ReplaceAll(line, "&", "¬A ")
	}

	if state.Meta {
		re := regexp.MustCompile(metadata)
		loc := re.FindStringIndex(line)
		if loc != nil {
			key := line[loc[0] : loc[1]-1]
			rest := line[loc[1]:]
			line = "‹b " + key + "›b :" + rest
		}
	}

	if isSpecialLine(raw) {
		line = boldSpecialLinePrefix(line)
	}

	prefix := "§"
	if state.Block == '>' {
		prefix = "§1B"
	}
	if state.Block == '=' {
		prefix = "§5B"
	}
	if state.Block == '`' {
		prefix = "§4B"
	}
	if state.Block == '~' {
		prefix = "§3B"
	}
	if colorSpecialLines && isDashSpecialLine {
		prefix = "§3B"
	}
	if colorSpecialLines && isTildeSpecialLine {
		prefix = "§5B"
	}
	if state.Block == ' ' {
		prefix = "§AB"
	}
	if current {
		prefix = "§BA"
	}
	prefix += "999 "

	if isBlockHeader {
		return prefix + "‹b " + line + "›b "
	}

	return prefix + line
}

func wrapRegexMatches(str string, pattern string, before string, after string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return str, err
	}

	matches := re.FindAllStringIndex(str, -1)
	if len(matches) == 0 {
		return str, nil
	}

	result := []byte{}
	last := 0

	for _, match := range matches {
		start := match[0]
		end := match[1]

		// texto antes do match
		result = append(result, str[last:start]...)

		// before + match + after
		result = append(result, before...)
		result = append(result, str[start:end]...)
		result = append(result, after...)

		last = end
	}

	// resto da string
	result = append(result, str[last:]...)

	return string(result), nil
}

func replaceLine(str, target, before, after string) string {

	parts := strings.Split(str, target)
	if len(parts) == 1 {
		return str
	}
	var b strings.Builder
	b.WriteString(parts[0])
	for i := 1; i < len(parts); i++ {
		if i%2 == 1 {
			b.WriteString(before)
		} else {
			b.WriteString(after)
		}
		b.WriteString(parts[i])
	}
	return b.String()
}

func getLinePrefix(line string) string {
	if !strings.HasPrefix(line, "§") {
		return ""
	}

	runes := []rune(line)
	if len(runes) < 4 {
		return ""
	}

	i := 3
	for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
		i++
	}

	if i >= len(runes) || runes[i] != ' ' {
		return ""
	}

	return string(runes[:i+1])
}

func sidebarPreviewLineStyle(line string) string {
	prefix := getLinePrefix(line)
	if prefix == "" {
		return "¤A8 " + line + "¤ "
	}

	runes := []rune(prefix)
	bg := 'A'
	if len(runes) > 1 {
		bg = runes[1]
	}

	return prefix + "¤" + string(bg) + "8 " + strings.TrimPrefix(line, prefix) + "¤ "
}

func editorFooterLine(width int) string {
	const message = "[ © 2002 nausea . All rights reserved ]"
	if width <= 0 {
		return ""
	}

	msg := []rune(message)
	if len(msg) >= width {
		return string(msg[:width])
	}

	left := (width - len(msg)) / 2
	right := width - len(msg) - left
	return "¤A8 " + strings.Repeat(" ", left) + message + strings.Repeat(" ", right) + "¤ "
}

func (e Editor) shiftHeaderLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	if strings.HasPrefix(line, "§") {
		prefix := getLinePrefix(line)
		if prefix != "" {
			body := strings.TrimPrefix(line, prefix)
			return prefix + e.Utilities.CutVisible(body, width)
		}
	}
	return e.Utilities.CutVisible(line, width)
}

func numberPrefixFromLine(line string) string {
	prefix := getLinePrefix(line)

	if len(prefix) < 3 {
		return "§AB0 "
	}

	runes := []rune(prefix)

	if len(runes) < 3 {
		return "§AB0 "
	}

	bg := runes[1]

	return "§" + string(bg) + "AB "
}
