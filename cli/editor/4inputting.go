package editor

import (
	"env/filesystem"
	"env/routines"
	"path/filepath"
	"strings"
	"time"
)

func (e *Editor) Input(newinput *routines.Input) {
	if e.Page == nil {
		return
	}

	if newinput.Key == "ctrl+p" {
		e.cyclePanelMode()
		return
	}

	if e.panelMode() == "sidebar" && e.Sidebar != nil {
		e.Sidebar.Input(newinput)
		return
	}

	if e.Sidebar != nil && e.Sidebar.On && e.Sidebar.Switch {
		switch newinput.Key {
		case "up", "down", "left", "right", "enter", "ctrl+enter":
			e.Sidebar.Input(newinput)
			return
		}
	}

	switch newinput.Key {

	case "up":
		e.saveState()
		e.moveVertical(-1)

	case "down":
		e.saveState()
		e.moveVertical(1)

	case "left":
		e.saveState()
		e.moveHorizontal(-1)

	case "right":
		e.saveState()
		e.moveHorizontal(1)

	case "ctrl+left":
		if e.cursorIsOnHeader() {
			e.setCursorX(0)
			return
		}
		e.moveWordLeft()

	case "ctrl+right":
		if e.cursorIsOnHeader() {
			e.setCursorX(len([]rune(e.headerEditablePath())))
			return
		}
		e.moveWordRight()

	case "ctrl+up":
		if e.cursorIsOnHeader() {
			return
		}
		e.saveState()
		e.moveLineUp()

	case "ctrl+down":
		if e.cursorIsOnHeader() {
			return
		}
		e.saveState()
		e.moveLineDown()

	case "backspace":
		if e.idProtection() {
			return
		}
		e.saveState()
		e.backspace()

	case "enter":
		if e.idProtection() {
			return
		}
		e.saveState()
		e.enter()

	case "delete":
		if e.idProtection() {
			return
		}
		e.saveState()
		e.delete()

	case "ctrl+backspace":
		if e.idProtection() {
			return
		}
		e.saveState()
		e.ctrlbackspace()

	case "ctrl+enter":
		if e.cursorIsOnHeader() {
			e.moveVertical(1)
			return
		}
		if e.openReferenceUnderCursorInNewTab() {
			return
		}
		e.saveState()
		e.ctrlenter()

	case "ctrl+delete":
		if e.idProtection() {
			return
		}
		e.saveState()
		e.ctrldelete()

	case "ctrl+z":
		e.undoAndRestore()

	case "ctrl+y":
		e.redoo()

	case "ctrl+c":
		e.ctrlc()

	case "ctrl+x":
		e.saveState()
		e.ctrlx()

	case "ctrl+v":
		e.saveState()
		e.ctrlv()

	case "ctrl+d":
		e.saveState()
		e.ctrld()

	case "char":
		if e.idProtection() {
			return
		}
		if newinput.Char == ' ' {
			e.saveState()
		}
		e.char(newinput.Char)

	case "ctrl+s":
		e.savee()

	case "ctrl+o":
		if e.Sidebar != nil && e.Sidebar.On {
			e.Sidebar.Switch = !e.Sidebar.Switch
			if e.Sidebar.Switch {
				e.Sidebar.RequestSelectedPreview()
			}
		}

	}

}

func (e *Editor) cyclePanelMode() {
	switch e.PanelMode {
	case "sidebar":
		e.PanelMode = "editor"
	case "editor":
		e.PanelMode = "whole"
	default:
		e.PanelMode = "sidebar"
	}

	if e.PanelMode == "sidebar" && e.Sidebar != nil {
		e.Sidebar.Switch = true
		e.Sidebar.RequestSelectedPreview()
	}
}

func (e *Editor) panelMode() string {
	switch e.PanelMode {
	case "sidebar", "editor", "whole":
		return e.PanelMode
	default:
		return "whole"
	}
}

func (e *Editor) undoAndRestore() {
	e.undoo()

	if e.Parent == nil {
		return
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	// RestoreLastPatch not available — no-op
}

func (e *Editor) InsertTextAtCursor(text string) {
	if e == nil || text == "" {
		return
	}

	if len(e.Content) == 0 {
		e.Content = []string{""}
	}

	if len(e.Cursor) < 2 {
		e.Cursor = []int{0, 0}
	}

	e.Cursor[1] = clamp(e.Cursor[1], 0, len(e.Content)-1)

	lineLen := len([]rune(e.Content[e.Cursor[1]]))
	e.Cursor[0] = clamp(e.Cursor[0], 0, lineLen)

	if e.idProtection() {
		return
	}

	e.saveState()

	for _, r := range text {
		if r == '\n' {
			e.enter()
			continue
		}

		e.insertCharAt(e.Cursor[0], r)
		e.moveHorizontal(1)
	}

	if e.Page != nil {
		e.Page.Content = e.Content
	}
}

func (e *Editor) idProtection() bool {
	if !e.logBlockAutomationAllowed() {
		return false
	}

	if e.cursorIsOnProtectedID() || e.cursorIsOnProtectedShortID() {
		return true
	}
	return false
}

func (e *Editor) cursorIsOnProtectedShortID() bool {
	if len(e.Content) == 0 {
		return false
	}

	y := e.Cursor[1]
	if y < 0 || y >= len(e.Content) {
		return false
	}

	line := e.Content[y]

	if !strings.HasPrefix(line, "-- ") && !strings.HasPrefix(line, "~~ ") {
		return false
	}

	start, rest := splitIntoTwo(line[3:], "|")

	if rest == "" || !isShortId(start) {
		return false
	}

	return e.Cursor[0] < 14
}

func (e *Editor) cursorIsOnProtectedID() bool {
	if len(e.Content) == 0 {
		return false
	}

	y := e.Cursor[1]

	if y < 0 || y >= len(e.Content) {
		return false
	}

	line := e.Content[y]

	if !isIdLine(line) {
		return false
	}

	if y == 0 {
		return false
	}

	return isBlockHeaderLine(e.Content[y-1])
}

func (e *Editor) savee() {
	if e.Page == nil {
		e.Page = draftPageForPath(e.Path)
	}

	savingHeader := e.cursorIsOnHeader()
	oldPath := e.Page.Path
	if oldPath == "" {
		oldPath = e.Path
	}
	newPath := strings.TrimSpace(e.Path)
	if newPath == "" {
		newPath = oldPath
		e.Path = oldPath
	}

	e.Content = ensureEditableContent(e.Content)
	if e.logBlockAutomationAllowed() {
		ensured := e.ensureId()
		e.applyEnsuredID(ensured)
	}

	e.Page.Content = e.Content

	if e.Page.Stage == "" {
		e.Page.Stage = "draft"
	} else if e.Page.Stage != "draft" {
		e.Page.Stage = "edited"
	}

	if e.Page.Name == "" {
		e.Page.Name = filepath.Base(e.Path)
	}

	if e.Page.Type == "" {
		if shouldNewPageBeDeep(e.Path) {
			e.Page.Type = "deep"
		} else {
			e.Page.Type = "shallow"
		}
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	if fs.Find(oldPath) == nil && e.Page.Stage == "draft" {
		_, cached := fs.NewDraft(oldPath)
		if cached != nil {
			e.Page = cached
		}
	}

	if savingHeader && filepath.Clean(oldPath) != filepath.Clean(newPath) {
		e.Page.Content = copyLines(e.Content)
		fs.EditContent(e.Page, copyLines(e.Content))

		ok, moved := fs.EditPath(e.Page, newPath)
		if !ok || moved == nil {
			e.Path = oldPath
			e.Page.Path = oldPath
			e.Page.Name = filepath.Base(oldPath)
			return
		}

		e.Page = moved
		e.Path = moved.Path
		e.Name = filepath.Base(moved.Path)
		if e.Sidebar != nil {
			e.requestSidebarPage(e.Path)
		}
		e.AppliedContent = copyLines(e.Content)
		return
	}

	e.Page.Path = e.Path

	if sameLines(e.Content, e.AppliedContent) {
		page := e.Page
		commitPage := cloneEditorPageTree(page)
		cacheSnapshot := cloneEditorPageTree(fs.Cache)
		if page.Stage == "draft" || page.Stage == "edited" || page.Stage == "doomed" {
			page.Stage = "local"
			page.Diff = nil
			page.Og = cloneEditorPageTree(page)
		}
		e.Parent.QueueDelayedAction("sync "+relativePersonalPath(page.Path), 5*time.Second, func() {
			req := filesystem.NewSyncRequest()
			req.Branch = commitPage
			req.PageOnly = true
			fs.Sync(req)
		}, func() {
			if cacheSnapshot == nil {
				return
			}
			fs.Cache = cacheSnapshot
			if restored := findPageByPath(cacheSnapshot, page.Path); restored != nil {
				e.Page = restored
			}
		})
		return
	}

	e.Page.Content = copyLines(e.Content)
	fs.EditContent(e.Page, copyLines(e.Content))
	e.AppliedContent = copyLines(e.Content)
}

func (e *Editor) ctrld() {
	if e.cursorIsOnHeader() {
		return
	}

	e.insertString(e.Cursor[1]+1, e.Content[e.Cursor[1]])
}

func (e *Editor) ctrlv() {

	for i, clipped := range sharedClipboard {
		e.insertString(e.Cursor[1]+i+1, clipped)
		e.setCursorX(0)
		e.moveVertical(1)
	}
}

func (e *Editor) ctrlc() {
	e.clip()
}

func (e *Editor) ctrlx() {
	if e.cursorIsOnHeader() {
		return
	}

	if len(e.Content) == 0 {
		return
	}

	e.clip()
	e.Content = removeAt(e.Content, e.Cursor[1])

	if len(e.Content) == 0 {
		e.Content = []string{""}
		e.Cursor = []int{0, 0}
		return
	}

	if e.Cursor[1] >= len(e.Content) {
		e.Cursor[1] = len(e.Content) - 1
	}

	e.setCursorX(0)
}

func (e *Editor) saveState() {

	e.Undo = append([]EditorState{newEditorState(e.Content, e.Cursor, e.WantedX)}, e.Undo...)
	e.Redo = nil
}

func (e *Editor) undoo() bool {
	if len(e.Undo) == 0 {
		return false
	}

	current := newEditorState(e.Content, e.Cursor, e.WantedX)
	e.Redo = append([]EditorState{current}, e.Redo...)

	state := e.Undo[0]

	e.Content = state.Content
	e.Cursor = state.Cursor
	e.WantedX = state.WantedX

	if e.Page != nil {
		e.Page.Content = e.Content
	}

	e.Undo = removeAt(e.Undo, 0)

	return true
}

func (e *Editor) redoo() {

	if len(e.Redo) == 0 {
		return
	}

	current := newEditorState(e.Content, e.Cursor, e.WantedX)
	e.Undo = append([]EditorState{current}, e.Undo...)

	state := e.Redo[0]

	e.Content = state.Content
	e.Cursor = state.Cursor
	e.WantedX = state.WantedX

	e.Redo = removeAt(e.Redo, 0)

}

func (e *Editor) char(char rune) {
	if e.cursorIsOnHeader() {
		e.insertHeaderCharAt(e.Cursor[0], char)
		e.moveHorizontal(1)
		return
	}

	e.insertCharAt(e.Cursor[0], char)
	e.moveHorizontal(1)
}

func (e *Editor) ctrldelete() {
	if e.cursorIsOnHeader() {
		idx := e.findNextHeaderSpace()
		e.removeHeaderRange(e.Cursor[0], idx)
		return
	}

	if e.Cursor[0] == len(e.Content[e.Cursor[1]]) || len(e.Content[e.Cursor[1]]) == 0 {
		e.delete()
		return
	}

	idx := e.findNextSpace()
	e.removeStringRange(e.Cursor[0], idx)
}

func (e *Editor) ctrlenter() {

	e.insertString(e.Cursor[1]+1, "")
	e.setCursorX(0)
	e.moveVertical(1)

}

func (e *Editor) ctrlbackspace() {
	if e.cursorIsOnHeader() {
		if e.Cursor[0] <= 0 {
			return
		}
		idx := e.findPrevHeaderSpace()
		e.removeHeaderRange(idx, e.Cursor[0])
		e.setCursorX(idx)
		return
	}

	if e.Cursor[0] == 0 || len(e.Content[e.Cursor[1]]) == 0 {
		e.backspace()
		return
	}
	idx := e.findPrevSpace()
	e.removeStringRange(idx, e.Cursor[0])
	e.setCursorX(idx)
}

func (e *Editor) delete() {
	if e.cursorIsOnHeader() {
		e.removeHeaderCharAt(e.Cursor[0])
		return
	}

	if e.Cursor[0] == len(e.Content[e.Cursor[1]]) || len(e.Content[e.Cursor[1]]) == 0 {
		e.collapseNextLine(e.Cursor[1])
	} else {
		e.removeCharAt(e.Cursor[0])
	}

}

func (e *Editor) enter() {
	if e.cursorIsOnHeader() {
		e.moveVertical(1)
		return
	}

	if len(e.Content) == 0 {
		e.Content = []string{""}
		e.Cursor = []int{0, 0}
		return
	}

	e.Cursor[1] = clamp(e.Cursor[1], 0, len(e.Content)-1)

	line := []rune(e.Content[e.Cursor[1]])
	x := clamp(e.Cursor[0], 0, len(line))

	if x == len(line) || len(line) == 0 {
		e.insertString(e.Cursor[1]+1, "")
		e.moveVertical(1)
		e.setCursorX(0)
		return
	}

	toadd := string(line[x:])
	e.insertString(e.Cursor[1]+1, toadd)
	e.Content[e.Cursor[1]] = string(line[:x])
	e.moveVertical(1)
	e.setCursorX(0)
}

func (e *Editor) backspace() {
	if e.cursorIsOnHeader() {
		if e.Cursor[0] <= 0 {
			return
		}
		e.removeHeaderCharAt(e.Cursor[0] - 1)
		e.moveHorizontal(-1)
		return
	}

	if e.Cursor[0] == 0 {
		if e.Cursor[1] == 0 {
			return
		}
		size := len([]rune(e.Content[e.Cursor[1]-1]))
		e.collapseLine(e.Cursor[1])
		e.moveVertical(-1)
		e.setCursorX(size)

	} else {
		e.removeCharAt(e.Cursor[0] - 1)
		e.moveHorizontal(-1)
	}
}

func (e *Editor) moveLineUp() {

	e.moveLine(e.Cursor[1], e.Cursor[1]-1)
	e.moveVertical(-1)
}

func (e *Editor) moveLineDown() {

	e.moveLine(e.Cursor[1], e.Cursor[1]+1)
	e.moveVertical(1)
}

func (e *Editor) moveWordLeft() {

	idx := e.findPrevSpace()
	e.setCursorX(idx)
}

func (e *Editor) moveWordRight() {

	idx := e.findNextSpace()
	e.setCursorX(idx)
}

func (e *Editor) moveVertical(delta int) {
	if len(e.Content) == 0 {
		e.Content = []string{""}
	}

	if e.cursorIsOnHeader() {
		if delta > 0 {
			e.Cursor[1] = 0
			e.Cursor[0] = clamp(e.Cursor[0], 0, len([]rune(e.Content[0])))
		}
		return
	}

	if e.Cursor[1] == 0 && delta < 0 {
		e.Cursor[1] = -1
		e.Cursor[0] = clamp(e.Cursor[0], 0, len([]rune(e.headerEditablePath())))
		return
	}

	newy := e.Cursor[1] + delta
	newy = clamp(newy, 0, len(e.Content)-1)
	e.Cursor[1] = newy
	e.Cursor[0] = clamp(e.Cursor[0], 0, len([]rune(e.Content[e.Cursor[1]])))

}

func (e *Editor) moveHorizontal(delta int) {
	if e.cursorIsOnHeader() {
		newx := e.Cursor[0] + delta
		e.Cursor[0] = clamp(newx, 0, len([]rune(e.headerEditablePath())))
		return
	}

	if len(e.Content) == 0 {
		e.Cursor = []int{0, 0}
		return
	}

	e.Cursor[1] = clamp(e.Cursor[1], 0, len(e.Content)-1)

	linelength := len([]rune(e.Content[e.Cursor[1]]))

	if delta < 0 && e.Cursor[0] == 0 && e.Cursor[1] > 0 {
		e.Cursor[1]--
		e.Cursor[0] = len([]rune(e.Content[e.Cursor[1]]))
		return
	}

	if delta > 0 && e.Cursor[0] == linelength && e.Cursor[1] < len(e.Content)-1 {
		e.Cursor[1]++
		e.Cursor[0] = 0
		return
	}

	newx := e.Cursor[0] + delta
	newx = clamp(newx, 0, linelength)

	e.Cursor[0] = newx
}

// ...

func removeAt[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}

	return append(slice[:index], slice[index+1:]...)
}

func (e *Editor) insertCharAt(index int, ch rune) {
	s := e.Content[e.Cursor[1]]
	runes := []rune(s)

	if index < 0 {
		index = 0
	}

	if index > len(runes) {
		index = len(runes)
	}

	runes = append(runes, 0)
	copy(runes[index+1:], runes[index:])
	runes[index] = ch

	e.Content[e.Cursor[1]] = string(runes)
}

func (e *Editor) removeStringRange(start, end int) {

	s := e.Content[e.Cursor[1]]
	runes := []rune(s)

	if start < 0 {
		start = 0
	}

	if end > len(runes) {
		end = len(runes)
	}

	if start >= end {
		e.Content[e.Cursor[1]] = s
	}

	runes = append(runes[:start], runes[end:]...)

	e.Content[e.Cursor[1]] = string(runes)
}

func (e *Editor) insertString(index int, value string) {

	lines := e.Content
	if index < 0 {
		index = 0
	}

	if index > len(lines) {
		index = len(lines)
	}

	lines = append(lines[:index], append([]string{value}, lines[index:]...)...)

	e.Content = lines

	if e.Page != nil {
		e.Page.Content = e.Content
	}
}

func (e *Editor) removeCharAt(index int) {

	s := e.Content[e.Cursor[1]]
	runes := []rune(s)

	if index < 0 || index >= len(runes) {
		e.Content[e.Cursor[1]] = s
	}

	runes = append(runes[:index], runes[index+1:]...)

	e.Content[e.Cursor[1]] = string(runes)
}

func (e *Editor) collapseNextLine(i int) {
	lines := e.Content

	if i < 0 || i >= len(lines)-1 {
		return
	}

	lines[i] += lines[i+1]

	lines = append(lines[:i+1], lines[i+2:]...)

	e.Content = lines

	if e.Page != nil {
		e.Page.Content = e.Content
	}
}

func (e *Editor) collapseLine(i int) {

	if e.Cursor[1] == 0 {
		return
	}

	lines := e.Content
	if i <= 0 || i >= len(lines) {
		e.Content = lines
	}

	lines[i-1] += lines[i]

	lines = append(lines[:i], lines[i+1:]...)

	e.Content = lines

	if e.Page != nil {
		e.Page.Content = e.Content
	}
}

func (e *Editor) moveLine(from, to int) {

	slice := e.Content
	if from < 0 || from >= len(slice) || to < 0 || to >= len(slice) {
		e.Content = slice
		return
	}

	item := slice[from]

	slice = append(slice[:from], slice[from+1:]...)

	slice = append(slice[:to], append([]string{item}, slice[to:]...)...)

	e.Content = slice
}

func (e *Editor) cursorIsOnHeader() bool {
	return e != nil && len(e.Cursor) >= 2 && e.Cursor[1] < 0
}

func (e *Editor) headerEditablePath() string {
	if e == nil {
		return ""
	}

	path := e.Path
	if path == "" && e.Page != nil {
		path = e.Page.Path
	}

	return relativePersonalPath(path)
}

func (e *Editor) setHeaderEditablePath(path string) {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "prsnl.spc/")
	if path == "" {
		path = "untitled"
	}

	e.Path = path
	e.Name = filepath.Base(path)
}

func (e *Editor) insertHeaderCharAt(index int, ch rune) {
	runes := []rune(e.headerEditablePath())
	index = clamp(index, 0, len(runes))

	runes = append(runes, 0)
	copy(runes[index+1:], runes[index:])
	runes[index] = ch
	e.setHeaderEditablePath(string(runes))
}

func (e *Editor) removeHeaderCharAt(index int) {
	runes := []rune(e.headerEditablePath())
	if index < 0 || index >= len(runes) {
		return
	}

	runes = append(runes[:index], runes[index+1:]...)
	e.setHeaderEditablePath(string(runes))
}

func (e *Editor) removeHeaderRange(start int, end int) {
	runes := []rune(e.headerEditablePath())
	start = clamp(start, 0, len(runes))
	end = clamp(end, 0, len(runes))
	if start >= end {
		return
	}

	runes = append(runes[:start], runes[end:]...)
	e.setHeaderEditablePath(string(runes))
}

func (e *Editor) findNextHeaderSpace() int {
	runes := []rune(e.headerEditablePath())
	for i := e.Cursor[0] + 1; i < len(runes); i++ {
		if runes[i] == ' ' || runes[i] == '/' {
			return i
		}
	}
	return len(runes)
}

func (e *Editor) findPrevHeaderSpace() int {
	runes := []rune(e.headerEditablePath())
	for i := e.Cursor[0] - 1; i >= 0; i-- {
		if runes[i] == ' ' || runes[i] == '/' {
			return i + 1
		}
	}
	return 0
}

func (e *Editor) findNextSpace() int {

	runes := []rune(e.Content[e.Cursor[1]])
	for i := e.Cursor[0] + 1; i < len(runes)-1; i++ {
		if runes[i] == ' ' {
			return i
		}
	}
	return len(runes)
}

func (e *Editor) findPrevSpace() int {

	line := e.Content[e.Cursor[1]]
	for i := e.Cursor[0] - 1; i >= 0; i-- {
		if line[i] == ' ' {
			return i
		}
	}
	return 0
}

func (e *Editor) setCursorY(y int) {

	e.Cursor[1] = y
}

func (e *Editor) setCursorX(x int) {
	if e.cursorIsOnHeader() {
		e.Cursor[0] = clamp(x, 0, len([]rune(e.headerEditablePath())))
		return
	}

	e.Cursor[0] = x
}

func clamp(num int, min int, max int) int {

	if num < min {
		return min
	}
	if num > max {
		return max
	}
	return num
}

func (e *Editor) clip() {

	if e.cursorIsOnHeader() {
		sharedClipboard = []string{e.headerEditablePath()}
		e.Clipboard = copyLines(sharedClipboard)
		return
	}

	sharedClipboard = []string{e.Content[e.Cursor[1]]}
	e.Clipboard = copyLines(sharedClipboard)
}

func isBlockHeaderLine(line string) bool {
	line = strings.TrimSpace(line)

	if len(line) == 0 {
		return false
	}

	runes := []rune(line)

	if !strings.ContainsRune("`~>=", runes[0]) {
		return false
	}

	if len(runes) >= 2 && strings.ContainsRune("`~>=", runes[1]) {
		return false
	}

	return true
}

func (e *Editor) openReferenceUnderCursorInNewTab() bool {
	if e.openIDUnderCursorInNewTab() {
		return true
	}

	if e.openDateUnderCursorInNewTab() {
		return true
	}

	if e.handleProjectOrFamilyUnderCursorInNewTab() {
		return true
	}

	return false
}

func (e *Editor) openDateUnderCursorInNewTab() bool {
	if e == nil || e.Parent == nil {
		return false
	}

	ok, date := e.dateUnderCursor()
	if !ok {
		return false
	}

	e.Parent.AddClients("editor:a.log/"+date, "next")
	return true
}

func (e *Editor) dateUnderCursor() (bool, string) {
	if e == nil || len(e.Cursor) < 2 || e.Cursor[1] < 0 || e.Cursor[1] >= len(e.Content) {
		return false, ""
	}

	line := []rune(e.Content[e.Cursor[1]])
	if len(line) == 0 {
		return false, ""
	}

	x := e.Cursor[0]
	if x < 0 {
		x = 0
	}
	if x >= len(line) {
		x = len(line) - 1
	}

	idx := x
	if !isDateRune(line[idx]) {
		idx = x - 1
	}

	if idx < 0 || idx >= len(line) || !isDateRune(line[idx]) {
		return false, ""
	}

	start := idx
	for start > 0 && isDateRune(line[start-1]) {
		start--
	}

	end := idx + 1
	for end < len(line) && isDateRune(line[end]) {
		end++
	}

	date := string(line[start:end])
	if !isEditorDate(date) {
		return false, ""
	}

	return true, date
}

func isDateRune(r rune) bool {
	return (r >= '0' && r <= '9') || r == '.'
}

func (e *Editor) handleProjectOrFamilyUnderCursorInNewTab() bool {
	if e == nil || e.Parent == nil || len(e.Cursor) < 2 {
		return false
	}

	ok, marker, name := e.projectOrFamilyUnderCursor()
	if !ok {
		return false
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return true
	}

	var path string

	switch marker {
	case '@':
		path = "e.proj/" + name

		if fs.Find(path) == nil {
			return true
		}

	case '#':
		e.Parent.AddClients("zettelkasten:"+name, "next")
		return true

	default:
		return false
	}

	e.Parent.AddClients("editor:"+path, "next")
	return true
}

func (e *Editor) projectOrFamilyUnderCursor() (bool, rune, string) {
	if e == nil || len(e.Cursor) < 2 || e.Cursor[1] < 0 || e.Cursor[1] >= len(e.Content) {
		return false, 0, ""
	}

	line := []rune(e.Content[e.Cursor[1]])
	if len(line) == 0 {
		return false, 0, ""
	}

	x := e.Cursor[0]
	if x < 0 {
		x = 0
	}
	if x >= len(line) {
		x = len(line) - 1
	}

	idx := x

	if !isReferenceRune(line[idx]) {
		idx = x - 1
	}

	if idx < 0 || idx >= len(line) || !isReferenceRune(line[idx]) {
		return false, 0, ""
	}

	start := idx
	for start > 0 && isReferenceRune(line[start-1]) {
		start--
	}

	end := idx + 1
	for end < len(line) && isReferenceRune(line[end]) {
		end++
	}

	if start >= len(line) {
		return false, 0, ""
	}

	marker := line[start]
	if marker != '@' && marker != '#' {
		return false, 0, ""
	}

	if end-start <= 1 {
		return false, 0, ""
	}

	name := string(line[start+1 : end])
	name = strings.TrimSpace(name)

	if name == "" {
		return false, 0, ""
	}

	return true, marker, name
}

func isReferenceRune(r rune) bool {
	return r == '@' ||
		r == '#' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '-' ||
		r == '_' ||
		r == '.'
}

func (e *Editor) openIDUnderCursorInNewTab() bool {
	if e == nil || e.Page == nil || e.Parent == nil || len(e.Cursor) < 2 {
		return false
	}

	ok, id := e.idUnderCursor()
	if !ok {
		return false
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return false
	}

	if e.logBlockAutomationAllowed() {
		e.refreshCurrentBlockIDPage(id)
	}

	if path := dedicatedOrReminderPathForID(fs.Cache, id); path != "" {
		if e.focusOpenEditorByID(id, path) {
			return true
		}
		e.Parent.AddClients("editor:"+path, "next")
		return true
	}

	if path := e.currentIDContextPath(id); path != "" {
		if e.focusOpenEditorByID(id, path) {
			return true
		}
		e.Parent.AddClients("editor:"+path, "next")
		return true
	}

	ok, path := resolveIDInCache(fs.Cache, id)
	if !ok || strings.TrimSpace(path) == "" {
		return false
	}

	if e.focusOpenEditorByID(id, path) {
		return true
	}
	e.Parent.AddClients("editor:"+path, "next")
	return true
}

func (e *Editor) focusOpenEditorByID(id string, path string) bool {
	if e == nil || e.Parent == nil || id == "" || strings.TrimSpace(path) == "" {
		return false
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return false
	}

	page := fs.Find(path)
	if page == nil {
		return false
	}

	for i, client := range e.Parent.GetClients() {
		open, ok := client.(*Editor)
		if !ok || open == nil || open == e {
			continue
		}

		if !openEditorIsDerivedBlockPage(open, id) {
			continue
		}

		open.Path = path
		open.Name = filepath.Base(path)
		open.SetPage(page)
		e.Parent.SetFocus(i)
		return true
	}

	return false
}

func openEditorIsDerivedBlockPage(e *Editor, id string) bool {
	if e == nil || id == "" {
		return false
	}

	path := e.Path
	if e.Page != nil && e.Page.Path != "" {
		path = e.Page.Path
	}

	if !strings.HasPrefix(path, "b.rec/") && !strings.HasPrefix(path, "c.rand/") {
		return false
	}

	if strings.HasSuffix(path, "/.reminders") || path == "b.rec/.reminders" || path == "c.rand/.reminders" {
		return false
	}

	if contentHasBlockID(e.Content, id) {
		return true
	}

	return e.Page != nil && contentHasBlockID(e.Page.Content, id)
}

func contentHasBlockID(content []string, id string) bool {
	for _, line := range content {
		if strings.TrimSpace(line) == "id: "+id {
			return true
		}
	}

	return false
}

func (e *Editor) idUnderCursor() (bool, string) {
	if e == nil || len(e.Cursor) < 2 || e.Cursor[1] < 0 || e.Cursor[1] >= len(e.Content) {
		return false, ""
	}

	raw := e.Content[e.Cursor[1]]
	if isIdLine(raw) {
		id := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "id:"))
		return true, id
	}

	if isSpecialLine(raw) {
		id, _ := specialLineIDAndMessage(raw)
		if isShortId(id) {
			return true, id
		}
	}

	line := []rune(e.Content[e.Cursor[1]])
	if len(line) == 0 {
		return false, ""
	}

	x := e.Cursor[0]
	if x < 0 {
		x = 0
	}
	if x > len(line) {
		x = len(line)
	}

	idx := x
	if idx == len(line) || !isIDRune(line[idx]) {
		idx = x - 1
	}

	if idx < 0 || idx >= len(line) || !isIDRune(line[idx]) {
		return false, ""
	}

	start := idx
	for start > 0 && isIDRune(line[start-1]) {
		start--
	}

	end := idx + 1
	for end < len(line) && isIDRune(line[end]) {
		end++
	}

	id := string(line[start:end])
	if len(id) != 8 && len(id) != 12 {
		return false, ""
	}

	return true, id
}

func isIDRune(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

func cloneEditorPageTree(page *filesystem.Page) *filesystem.Page {
	if page == nil {
		return nil
	}

	clone := &filesystem.Page{
		Name:     page.Name,
		Path:     page.Path,
		Type:     page.Type,
		Options:  cloneEditorMap(page.Options),
		Content:  append([]string(nil), page.Content...),
		Metadata: cloneEditorMap(page.Metadata),
		Sorting:  page.Sorting,
		Og:       cloneEditorPageTree(page.Og),
		Stage:    page.Stage,
		Diff:     append([]string(nil), page.Diff...),
	}

	for _, child := range page.Children {
		clone.Children = append(clone.Children, cloneEditorPageTree(child))
	}

	return clone
}

func cloneEditorMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}

	out := map[string]any{}
	for key, value := range in {
		out[key] = value
	}
	return out
}
