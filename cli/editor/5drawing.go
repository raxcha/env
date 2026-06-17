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
	blockFades := blockFadeLevels(rawLines)
	cursorBlock := blockIndexForLine(rawLines, displayCursor[1])
	state := NewPreviewStyleState()
	colorSpecialLines := isALogEditorPath(displayPath)

	for i := range lines {
		dashedSize := 0
		if i != displayCursor[1] && dashed.MatchString(rawLines[i]) {
			dashedSize = DashedLineVisualSize(rawLines, i, e.Utilities)
		}
		showBlockToken := cursorBlock >= 0 && blockIndexForLine(rawLines, i) == cursorBlock
		lines[i] = StyleContentLine(lines[i], i == displayCursor[1], w, &state, colorSpecialLines, showBlockToken, dashedSize)
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
	showHeader := !previewingSidebar
	showFooter := false
	headerH := 0
	footerH := 0
	if showHeader && editorH > headerH+footerH {
		headerH = 1
	}
	if !previewingSidebar {
		fitsWithFooter := len(displayContent) <= editorH-headerH-1
		showFooter = displayCursor[1] >= len(displayContent)-1 || fitsWithFooter
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
	contentRowFades := contentFadeRows(lines, blockFades, start, w)
	contentRows := len(lines)
	if showHeader {
		lines = append([]string{"§8B0 " + strings.Repeat(" ", w)}, lines...)
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
	applyEditorBlockFades(&finalframe, lineX, y+headerH, w, contentRowFades)

	if footerH > 0 {
		footerY := y + headerH + contentRows
		footerframe := e.Utilities.GenerateFrame(
			engine.Boundaries{Fullsize: e.Bounds.Fullsize, Pos: routines.Bound{x, footerY}, Size: routines.Bound{panelW, 1}},
			[]string{editorFooterLine(panelW)},
			0,
		)
		finalframe = e.Utilities.MergeFrames(*footerframe, finalframe)
	}

	if showHeader {
		headerX := lineX
		headerW := w
		if e.Numbers && !e.Zenmode {
			shift := 2
			if shift > lineX-x {
				shift = lineX - x
			}
			headerX -= shift
			headerW += shift
		}
		headerframe := e.Utilities.GenerateFrame(
			engine.Boundaries{Fullsize: e.Bounds.Fullsize, Pos: routines.Bound{headerX, y}, Size: routines.Bound{headerW, 1}},
			[]string{e.shiftHeaderLine(e.headerLine(headerW+2, e.cursorIsOnHeader(), e.Cursor[0]), headerW)},
			0,
		)
		finalframe = e.Utilities.MergeFrames(*headerframe, finalframe)
	}

	if sidebarActive {
		divider := e.drawSidebarDivider()
		finalframe = e.Utilities.MergeFrames(divider, finalframe)

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

func (e Editor) drawSidebarDivider() engine.Frame {
	if e.Bounds == nil || len(e.Bounds.Fullsize) < 2 || len(e.Bounds.ActualPos) < 2 || len(e.Bounds.ActualSize) < 2 || e.Sidebar == nil || len(e.Sidebar.Bounds.ActualSize) < 1 {
		return engine.Frame{}
	}

	fullW := e.Bounds.Fullsize[0]
	fullH := e.Bounds.Fullsize[1]
	if fullW <= 0 || fullH <= 0 {
		return engine.Frame{}
	}

	theme := e.Utilities.Theme
	if theme == nil {
		bg := engine.RGB{}
		fg := engine.RGB{R: 90, G: 90, B: 90}
		return sidebarDividerFrame(e.Bounds.Fullsize, e.Bounds.ActualPos, e.Bounds.ActualSize, e.Sidebar.Bounds.ActualSize[0], fg, bg)
	}

	bg := theme.Background
	fg := utilities.MixRGB(theme.Background, theme.Foreground, 0.28)
	return sidebarDividerFrame(e.Bounds.Fullsize, e.Bounds.ActualPos, e.Bounds.ActualSize, e.Sidebar.Bounds.ActualSize[0], fg, bg)
}

func sidebarDividerFrame(fullsize routines.Bound, pos routines.Bound, size routines.Bound, sidebarWidth int, fg engine.RGB, bg engine.RGB) engine.Frame {
	fullW := fullsize[0]
	fullH := fullsize[1]
	cells := make([]engine.Cell, fullW*fullH)
	for i := range cells {
		cells[i] = engine.Cell{Visible: false}
	}

	x := pos[0] + sidebarWidth
	y0 := pos[1]
	y1 := pos[1] + size[1]
	if x < 0 || x >= fullW {
		return engine.Frame{Size: fullsize, Cells: cells, Timeout: 0}
	}
	if y0 < 0 {
		y0 = 0
	}
	if y1 > fullH {
		y1 = fullH
	}

	for y := y0; y < y1; y++ {
		cells[y*fullW+x] = engine.Cell{Char: '│', Fg: &fg, Bg: &bg, Visible: true}
	}

	return engine.Frame{Size: fullsize, Cells: cells, Timeout: 0}
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
var blockFadeToken = regexp.MustCompile(`!!([0-9])`)
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

func StyleContentLine(raw string, current bool, width int, state *PreviewStyleState, colorSpecialLines bool, showBlockToken bool, dashedSize ...int) string {
	if state == nil {
		initial := NewPreviewStyleState()
		state = &initial
	}

	line := raw
	if !showBlockToken {
		line = blockFadeToken.ReplaceAllString(line, "")
	}
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

func blockFadeLevels(lines []string) []float64 {
	out := make([]float64, len(lines))
	for i := 0; i < len(lines); {
		if isLogBlockBoundaryLine(lines[i]) {
			i++
			continue
		}

		end := i + 1
		for end < len(lines) && !isLogBlockBoundaryLine(lines[end]) {
			end++
		}

		fade := blockFadeAmount(lines[i:end])
		if fade > 0 {
			for j := i; j < end; j++ {
				out[j] = fade
			}
		}

		i = end
	}

	return out
}

func blockIndexForLine(lines []string, target int) int {
	if target < 0 || target >= len(lines) || isLogBlockBoundaryLine(lines[target]) {
		return -1
	}

	block := 0
	for i := 0; i < len(lines); {
		if isLogBlockBoundaryLine(lines[i]) {
			i++
			continue
		}

		end := i + 1
		for end < len(lines) && !isLogBlockBoundaryLine(lines[end]) {
			end++
		}

		if target >= i && target < end {
			return block
		}

		block++
		i = end
	}

	return -1
}

func blockFadeAmount(lines []string) float64 {
	for _, line := range lines {
		match := blockFadeToken.FindStringSubmatch(line)
		if len(match) < 2 {
			continue
		}

		value := int([]rune(match[1])[0] - '0')
		return (1 - float64(value)/9) * 0.82
	}

	return 0
}

func contentFadeRows(lines []string, fades []float64, start int, width int) []float64 {
	rows := []float64{}
	if width < 1 {
		width = 1
	}

	for i, line := range lines {
		idx := start + i
		fade := 0.0
		if idx >= 0 && idx < len(fades) {
			fade = fades[idx]
		}

		visual := VisibleLength(line)
		count := (visual + width - 1) / width
		if count < 1 {
			count = 1
		}
		for range count {
			rows = append(rows, fade)
		}
	}

	return rows
}

func applyEditorBlockFades(frame *engine.Frame, x int, y int, w int, fades []float64) {
	if frame == nil || len(frame.Size) < 2 || w <= 0 {
		return
	}

	fullW := frame.Size[0]
	fullH := frame.Size[1]
	for row, fade := range fades {
		if fade <= 0 {
			continue
		}

		screenY := y + row
		if screenY < 0 || screenY >= fullH {
			continue
		}

		for screenX := x; screenX < x+w && screenX < fullW; screenX++ {
			if screenX < 0 {
				continue
			}
			idx := screenY*fullW + screenX
			if idx < 0 || idx >= len(frame.Cells) {
				continue
			}

			cell := &frame.Cells[idx]
			if !cell.Visible || cell.Fg == nil || cell.Bg == nil {
				continue
			}

			fg := utilities.MixRGB(*cell.Fg, *cell.Bg, fade)
			cell.Fg = &fg
		}
	}
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
