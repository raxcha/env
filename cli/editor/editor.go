package editor

import (
	"env/cli"
	"env/filesystem"
	"env/settings"
	"env/types"
	"env/utils"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)


type Editor struct {

	Parent cli.Parent
	Settings *settings.Settings
	Filesystem *filesystem.Filesystem

	Mode string
	Layout string
	Orientation string

	Sizes *types.Dimensions
	SelectSizes *types.Dimensions
	VisualSizes *types.Dimensions

	Numbers bool
	Wrap bool
	Zen bool

	Undo []state
	Redo []state
	Collapsing map[string]bool

	Spec string
	Request *types.Loading
	Page *types.Page

	Content []string
	Cursor [3]int
	Preview []string
	PathEditor *Editor
	EditingPath bool
	PathError string

	Specc string
	Requests []*types.Loading
	Modules []*entry
	AllEntries []*entry
	Entries []*entry

	Prompt string
	Focus int

}

type state struct {

	Content []string
	Cursor [3]int
}

type entry struct {

	Type string
	Module int

	Page *types.Page
	Children []*entry

	Depth int
	Label string

	Collapsed bool
}

func CreateEditor (parent cli.Parent, spec string) *Editor {

	e := Editor {

		Parent: parent,
		Settings: parent.GetSettings(),
		Filesystem: parent.GetFilesystem(),

		Mode: parent.GetMode(),
		Layout: parent.GetLayout(),
		Orientation: "horizontal",

		Sizes: &types.Dimensions{},
		SelectSizes: &types.Dimensions{},
		VisualSizes: &types.Dimensions{},

		Numbers: true,
		Wrap: true,
		Zen: false,

		Undo: []state{},
		Redo: []state{},
		Collapsing: map[string]bool{},

		Spec: spec,
		Request: &types.Loading{},
		Page: nil,

		Content: []string{""},
		Cursor: [3]int{0, 0, 0},
		Preview: nil,

		Specc: strings.Split(spec, "/")[0],
		Requests: []*types.Loading{},
		Modules: make([]*entry, 12),
		AllEntries: []*entry{},
		Entries: []*entry{},

		Prompt: "",
		Focus: 0,
	}

	e.restartRequests()

	return &e
}

func (e *Editor) newLoading(path string, api bool, mode string, sort string, filter string, depth int, latest int) *types.Loading {

	req := e.Filesystem.NewLoading()
	req.Api = api
	req.Mode = mode
	req.Path = path
	req.Meta = true
	req.Cont = true
	req.Depth = depth
	req.Sort = sort
	req.Filters = []string{filter}
	req.Latest = latest
	req.Token = utils.GenerateId(16)
	return req
}

func (e *Editor) restartRequests() {
	
	e.Request = e.newLoading(e.Spec, false, "fresh", "", "", 0, -1)
	e.Request.CreateDraft = true

	e.Requests = []*types.Loading{}

	root := filepath.Dir(filepath.Clean(e.Spec))
	sorting := "time"
	switch e.Specc {
	case "log", "rec", "rand":
		root = e.Specc
	case "fami", "proj":
		root = e.Specc
		sorting = "priority"
	}
	if filepath.Clean(e.Spec) == root {
		e.Request.Sort = sorting
		e.Request.Depth = -1
		e.Request.Latest = 5
	} else {
		e.Requests = append(e.Requests, e.newLoading(root, false, "fresh", sorting, "", -1, 5))
	}

	e.makeRequests()
}

func (e *Editor) makeRequests() {

	e.loadOnce(e.Request)

	for _, req := range e.Requests {
		e.loadOnce(req)
	}
}

func (e *Editor) loadOnce(req *types.Loading) {
	
	reply := e.Filesystem.Load(req)

	go func() {

		page, ok := <-reply
		if !ok || page == nil {
			return
		}

		for i, req := range append([]*types.Loading{e.Request}, e.Requests...) {

			if page.Token == req.Token {

				e.newPageEntry(page, i)

				if i == 0 {
					if e.Page == nil {
						e.Content = append([]string{}, page.Content...)
					}
					e.Page = page
				}

				if len(e.Content) == 0 { e.Content = []string{""} }
			}
		}
		
	}()
}

func (e *Editor) newPageEntry(page *types.Page, module int) *entry {

	en := &entry{

		Type: "page",
		Module: module,

		Page: page,
		Children: []*entry{},

		Depth: utils.PathDepth(page.Path),
		Label: page.Name,

		Collapsed: false,
	}

	for _, child := range page.Children {
		en.Children = append(en.Children, e.newPageEntry(child, -1))
	}

	if module != -1 {
		e.Modules[module] = en
	}

	return en
}

func (e *Editor) flattenPageEntry(en *entry) []*entry {

	if en == nil {
		return nil
	}

	flat := []*entry{en}

	for _, child := range en.Children {
		flat = append(flat, e.flattenPageEntry(child)...)	
	}

	return flat
}

func (e *Editor) generateAllEntries() {
	e.AllEntries = []*entry{}
	seen := map[string]bool{}
	var appendTree func(*entry)
	appendTree = func(en *entry) {
		if en == nil || en.Page == nil || seen[en.Page.Path] {
			return
		}
		seen[en.Page.Path] = true
		e.AllEntries = append(e.AllEntries, en)
		for _, child := range en.Children {
			appendTree(child)
		}
	}
	if len(e.Modules) == 0 {
		e.generateEntries()
		return
	}
	opened := e.Modules[0]
	// Put the surrounding tree first and include unsaved pages in their parent.
	for _, mod := range e.Modules[1:] {
		if mod == nil || mod.Page == nil || mod.Page.Type == "error" {
			continue
		}
		if opened != nil && opened.Page != nil {
			var parent *entry
			found := false
			for _, en := range e.flattenPageEntry(mod) {
				if en.Page.Path == opened.Page.Path {
					en.Page = opened.Page
					found = true
				}
				if en.Page.Path == filepath.Dir(opened.Page.Path) {
					parent = en
				}
			}
			if !found && parent != nil {
				parent.Children = append(parent.Children, opened)
			}
		}
		appendTree(mod)
	}
	appendTree(opened)
	e.generateEntries()
}

func (e *Editor) generateEntries() {

	entries := []*entry{}

	blackList := []string{}

	var blacklistChildren func([]*entry)

	blacklistChildren = func(children []*entry) {

		for _, child := range children {

			blackList = append(blackList, child.Page.Path)
			blacklistChildren(child.Children)
		}
	}

	for _, en := range e.AllEntries {

		if en.Type == "page" && len(en.Children) > 0 && en.Collapsed {
			blacklistChildren(en.Children)
		}

		if en.Type == "page" && !slices.Contains(blackList, en.Page.Path) {
			entries = append(entries, en)
		}
	}

	depth := 0
	if len(entries) > 0 { depth = entries[0].Depth + 1 }
	entries = append(entries, &entry{Type: "load-more", Label: "load more!", Depth: depth})
	e.Entries = entries
	e.Focus = max(0, min(e.Focus, len(entries)-1))
}

func (e *Editor) Refresh(hard bool) {

	e.Layout = e.Parent.GetLayout()
	e.Mode = e.Parent.GetMode()

	if hard {
		e.restartRequests()
	}
} 

func (e *Editor) Resize(newsizes types.Dimensions) {

	e.Refresh(false)
	e.generateAllEntries()

	e.Sizes = &newsizes

	switch 2*newsizes.Size[1] > newsizes.Size[0] {

	case false:

		selectt := newsizes
		selectt.Pos[0] = selectt.Pos[0]
		selectt.Pos[1] = selectt.Pos[1]
		selectt.Size[0] = e.getSelectWidth()
		selectt.Size[1] = selectt.Size[1]
		e.SelectSizes = &selectt

		visuall := newsizes
		visuall.Pos[0] += e.getSelectWidth()
		visuall.Pos[1] = visuall.Pos[1]
		visuall.Size[0] -= e.getSelectWidth()
		visuall.Size[1] = visuall.Size[1]
		e.VisualSizes = &visuall

	case true:

		visuall := newsizes
		visuall.Pos[0] = visuall.Pos[0]
		visuall.Pos[1] = visuall.Pos[1]
		visuall.Size[1] -= e.getSelectHeight()
		e.VisualSizes = &visuall

		selectt := newsizes
		selectt.Pos[0] = selectt.Pos[0]
		selectt.Pos[1] += visuall.Size[1]
		selectt.Size[0] = selectt.Size[0]
		selectt.Size[1] = e.getSelectHeight()
		e.SelectSizes = &selectt

	}

	if e.Layout != "split" {
		e.SelectSizes = e.Sizes
		e.VisualSizes = e.Sizes
	}

}

func (e *Editor) getSelectHeight() int {

	if len(e.AllEntries) > e.Sizes.Size[1]/2 {
		return e.Sizes.Size[1] / 2
	}
	return len(e.AllEntries)
}

func (e *Editor) getSelectWidth() int {

	maxWidth := 0
	for _, en := range e.Entries {
		maxWidth = max(maxWidth, utils.VisibleLength(e.selectLabel(en)))
	}
	return min(maxWidth+4, max(0, e.Sizes.Size[0]/2))
}

func (e *Editor) Draw() *types.Queue {

	e.Resize(*e.Sizes)

	queues := []*types.Queue{}

	switch e.Layout {
	case "select":
		queues = append(queues, e.drawSelect())
	case "visual":
		queues = append(queues, e.drawVisual())
	case "split":
		queues = append(queues, e.drawSelect())
		queues = append(queues, e.drawVisual())
	}

	return e.Settings.MergeQueues(queues...)
}

func (en *entry) descendantPages() int {
	total := 0
	for _, child := range en.Children {
		if child == nil {
			continue
		}
		if child.Type == "page" && child.Page != nil {
			total++
		}
		total += child.descendantPages()
	}
	return total
}

func (e *Editor) selectLabel(en *entry) string {
	baseDepth := 0
	if len(e.Entries) > 0 {
		baseDepth = e.Entries[0].Depth
	}
	label := " " + strings.Repeat("  ", max(0, en.Depth-baseDepth)) + en.Label
	if en.Type == "page" && en.Page != nil && en.Page.Type == "deep" {
		label += utils.SuperscriptString(strconv.Itoa(en.descendantPages()))
	}
	return label
}

func (e *Editor) drawSelect() *types.Queue {

	lines := []string{}

	for i, en := range e.Entries {

		label := e.selectLabel(en)
		if en.Type == "load-more" {
			label = "‹b " + label + "›b "
		}

		if i == e.Focus {
			label = "§yx0 " + label
		} else {
			label = "§xy0 " + label
		}

		lines = append(lines, label)
	}

	lines = lines[e.selectViewport():]

	sizes := *e.SelectSizes
	if e.Layout != "split" {
		sizes.Pos[1]++
		sizes.Size[1] = max(0, sizes.Size[1]-1)
	}
	frame := e.Settings.GenerateFrame(sizes, lines, 0, []int{0, 0, 0, 0})
	if e.Layout != "split" {
		frame = e.Settings.MergeFrames(frame, e.drawPath(*e.SelectSizes))
	}
	return e.Settings.GenerateQueue(e.SelectSizes.Full, []*types.Frame{frame}, false)
}

func (e *Editor) selectViewport() int {

	total := len(e.Entries)
	if total <= 1 || e.SelectSizes == nil {
		return 0
	}

	height := e.SelectSizes.Size[1]
	if e.Layout != "split" { height = max(0, height-1) }
	if height <= 0 || total <= height {
		return 0
	}

	focus := max(0, min(e.Focus, total-1))

	return focus * (total - height) / (total - 1)
}

func (e *Editor) drawVisual() *types.Queue {

	e.initPathEditor()
	content := e.Content

	preview := false
	if e.Preview != nil {
		content = e.Preview
		preview = true
	}

	if e.Layout == "split" {
		content = nil
		preview = true
		if e.Focus >= 0 && e.Focus < len(e.Entries) {
			en := e.Entries[e.Focus]
			if en != nil && en.Type == "page" && en.Page != nil {
				content = en.Page.Content
			}
		}
	}

	wrapText := "0"
	if e.Wrap {
		wrapText = "999"
	}

	gutterWidth := len(strconv.Itoa(len(content))) + 2
	if e.Zen || !e.Numbers {
		gutterWidth = 0
	}

	gutterWidth = min(gutterWidth, max(0, e.VisualSizes.Size[0]))

	gutterSizes := *e.VisualSizes
	gutterSizes.Size[0] = gutterWidth
	gutterSizes.Pos[1]++
	gutterSizes.Size[1] = max(0, gutterSizes.Size[1]-1)

	contentSizes := *e.VisualSizes
	contentSizes.Pos[1]++
	contentSizes.Size[1] = max(0, contentSizes.Size[1]-1)
	if e.Zen {
		contentSizes.Pos[0] = int(float64(contentSizes.Size[0]) * 0.16)
		contentSizes.Size[0] = int(float64(contentSizes.Size[0]) * 0.68)
	} else {
		contentSizes.Pos[0] += gutterWidth
		contentSizes.Size[0] -= gutterWidth
	}

	contentMargin := []int{0, 1, 0, 0}
	contentWidth := max(0, contentSizes.Size[0]-contentMargin[1]-contentMargin[3])

	lines := []string{}

	for i, line := range content {

		if !preview && !e.EditingPath && i == e.Cursor[1] {

			line = utils.ReplacePaired(line, "*", "*‹b ", "›b *")
			line = utils.ReplacePaired(line, "%", "%‹i ", "›i %")
			line = utils.ReplacePaired(line, "_", "_‹u ", "›u _")
			line = utils.ReplacePaired(line, "^", "^‹U ", "›U ^")
			line = utils.ReplacePaired(line, "$", "$‹a ", "›a $")
			line = utils.ReplacePaired(line, "&", "&‹A ", "›A &")

		} else {

			line = utils.ReplacePaired(line, "*", "‹b ", "›b ")
			line = utils.ReplacePaired(line, "%", "‹i ", "›i ")
			line = utils.ReplacePaired(line, "_", "‹u ", "›u ")
			line = utils.ReplacePaired(line, "^", "‹U ", "›U ")
			line = utils.ReplacePaired(line, "$", "‹a ", "›a ")
			line = utils.ReplacePaired(line, "&", "‹A ", "›A ")

		}

		line = utils.ReplaceRegex(line, dateRegex, "‹b ", "›b ")
		line = utils.ReplaceRegex(line, timeRegex, "‹b ", "›b ")
		line = utils.ReplaceRegex(line, tagRegex, "‹b ", "›b ")
		line = utils.ReplaceRegex(line, projectRegex, "‹b ", "›b ")
		line = utils.ReplaceRegex(line, metadataRegex, "‹b ", "›b ")

		if !preview && !e.EditingPath && i == e.Cursor[1] {

			runes := []rune(line)
			idx := utils.RealIndex(line, e.Cursor[0])

			if idx < len(runes) {
				line = string(runes[:idx]) + "¤bx " + string(runes[idx]) + "¤ " + string(runes[idx+1:])
			} else {
				line += "¤bx  ¤ "
			}

			line = "§yx" + wrapText + " " + line + "¤yx "

		} else {

			line = "§xy" + wrapText + " " + line
		}

		lines = append(lines, line)

	}

	first := 0
	if !preview {
		first = e.visualViewport(lines, contentWidth)
	}
	lines = lines[first:]

	numbers := []string{}
	
	for i := range len(lines) {
		
		if e.Zen || !e.Numbers { break }

		parts := 1
		if e.Wrap && contentWidth > 0 {
			parts = min(1000, max(1, (utils.VisibleLength(lines[i])+contentWidth-1)/contentWidth))
		}

		for j := range parts {

			num := fmt.Sprintf("%"+strconv.Itoa(max(1, gutterWidth-1))+"d", first+i+1)

			if j != 0 {
				num = strings.Repeat(" ", gutterWidth)
			}

			if !preview && !e.EditingPath && first+i == e.Cursor[1] {
				num = "§yx0 " + num
			} else {
				num = "§xy0 " + num
			}

			numbers = append(numbers, num)
		}
	}

	numbersFrame := e.Settings.GenerateFrame(gutterSizes, numbers, 0, []int{0, 0, 0, 0})
	contentFrame := e.Settings.GenerateFrame(contentSizes, lines, 0, contentMargin)

	finalFrame := e.Settings.MergeFrames(numbersFrame, contentFrame, e.drawPath(*e.VisualSizes))
	return e.Settings.GenerateQueue(e.VisualSizes.Full, []*types.Frame{finalFrame}, false)
}

var dateRegex = `\b\d{4}\.\d{2}\.\d{2}\b`
var timeRegex = `\b(?:[01]\d|2[0-3]):[0-5]\d\b`
var tagRegex = `#[A-Za-z0-9_]+(?:\s|$)`
var projectRegex = `@[A-Za-z0-9_]+(?:\s|$)`
var metadataRegex = `^([^:]+):`

func (e *Editor) visualViewport(lines []string, width int) int {

	if len(lines) <= 1 || e.VisualSizes == nil { return 0 }

	height := max(0, e.VisualSizes.Size[1]-1)
	if height <= 0 { return 0 }

	rows := make([]int, len(lines))
	total := 0

	for i, line := range lines {
		rows[i] = 1
		if e.Wrap && width > 0 {
			rows[i] = min(1000, max(1, (utils.VisibleLength(line)+width-1)/width))
		}
		total += rows[i]
	}

	if total <= height { return 0 }

	cursor := max(0, min(e.Cursor[1], len(lines)-1))
	position := 0
	for i := 0; i < cursor; i++ {
		position += rows[i]
	}

	target := position * (total - height) / max(1, total - rows[cursor])
	target = max(target, position + min(rows[cursor], height) - height)

	first := 0
	offset := 0
	for first < cursor && offset < target {
		offset += rows[first]
		first++
	}

	return first
}

func (e *Editor) Input(newinput types.Input) {

	switch e.Layout {
	case "select", "split":
		e.InputSelect(&newinput)
	case "visual":
		e.InputVisual(&newinput)
	}
}


func (e *Editor) InputVisual(newinput *types.Input) {
	if e.inputPath(newinput) { return }


	if len(e.Content) == 0 { e.Content = []string{""} }
	if len(e.Undo) == 0 { e.savestate() }

	before := state{Content: slices.Clone(e.Content), Cursor: e.Cursor}

	switch newinput.Description {
	case "up":
		e.movey(-1)
	case "down":
		e.movey(1)
	case "left":
		e.movex(-1)
	case "right":
		e.movex(1)
	case "char", "number":
		e.typing(newinput.Char)
	case "ctrl+left":
		e.movex(e.nextspace(-1))
	case "ctrl+right":
		e.movex(e.nextspace(1))
	case "ctrl+up":
		e.moveline(-1)
	case "ctrl+down":
		e.moveline(1)
	case "backspace":
		e.backspace()
	case "ctrl+backspace":
		e.ctrlbackspace()
	case "delete":
		e.delete()
	case "ctrl+delete":
		e.ctrldelete()
	case "enter":
		e.enter()
	case "ctrl+enter":
		e.ctrlenter()
	case "ctrl+Z", "ctrl+z":
		e.undo()
		return
	case "ctrl+Y", "ctrl+y":
		e.redo()
		return
	case "ctrl+C", "ctrl+c":
		e.copyline()
	case "ctrl+D", "ctrl+d":
		e.savestate()
		e.duplicateline()
		e.savestate()
	case "ctrl+X", "ctrl+x":
		e.savestate()
		e.cutline()
		e.savestate()
	case "ctrl+V", "ctrl+v":
		e.savestate()
		e.pasteline()
		e.savestate()
	case "ctrl+S", "ctrl+s":
		e.savePage()

	}

	e.clampcursor()

	if !slices.Equal(before.Content, e.Content) {
		e.Redo = nil
	}

	if before.Cursor[1] != e.Cursor[1] || len(before.Content) != len(e.Content) ||
		((newinput.Description == "char" || newinput.Description == "number") && newinput.Char == ' ') {
		e.savestate()
	}
}

func (e *Editor) savePage() {
	if e.Page == nil || e.Filesystem == nil {
		return
	}
	switch e.Page.Stage {
	case "draft", "edit", "local", "*api", "api", "move":
	default:
		return
	}
	if e.PathEditor != nil {
		if err := e.Filesystem.RepathPage(e.Page, e.PathEditor.Content[0]); err != nil {
			e.PathError = err.Error()
			return
		}
		e.PathError = ""
		e.PathEditor.Content[0] = e.Page.Path
		e.Spec = e.Page.Path
		e.Specc = strings.Split(e.Spec, "/")[0]
		if e.Request != nil { e.Request.Path = e.Page.Path }
		for _, en := range e.AllEntries {
			if en.Page != nil {
				en.Label = en.Page.Name
				en.Depth = utils.PathDepth(en.Page.Path)
			}
		}
	}
	if !e.EditingPath { e.ensureLogEvent() }
	e.Filesystem.EditPage(e.Page, e.Content)
	e.Filesystem.Sync(&types.Syncing{Branch: e.Page, Hard: false})
}

func (e *Editor) ensureLogEvent() {
	if e.Page == nil || !strings.HasPrefix(filepath.ToSlash(filepath.Clean(e.Page.Path)), "log/") {
		return
	}
	y := e.Cursor[1]
	if y < 0 || y >= len(e.Content) || e.Content[y] == "" || !strings.ContainsRune("`>=~", rune(e.Content[y][0])) {
		return
	}
	hasTime := false
	if y > 0 && len(e.Content[y-1]) == 5 {
		_, err := time.Parse("15:04", e.Content[y-1])
		hasTime = err == nil
	}
	if !hasTime {
		e.Content = slices.Insert(e.Content, y, time.Now().Format("15:04"))
		y++
		e.Cursor[1] = y
	}
	if y+1 >= len(e.Content) || !strings.HasPrefix(e.Content[y+1], "id") {
		e.Content = slices.Insert(e.Content, y+1, "id: "+utils.GenerateId(12))
	}
	if e.Filesystem == nil {
		return
	}
	id := strings.TrimSpace(strings.TrimPrefix(e.Content[y+1], "id:"))
	if len(id) != 12 || strings.ContainsFunc(id, func(r rune) bool {
		return !(r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z')
	}) {
		return
	}
	root := "rec"
	if e.Content[y][0] == '`' {
		root = "rand"
	}
	title := strings.Join(strings.FieldsFunc(strings.ToLower(e.Content[y][1:]), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}), "-")
	page := e.Filesystem.CacheEventDraft(root, id, title)
	if e.Parent != nil {
		e.Parent.AddClients("editor:"+page.Path, "after", false)
	}
}

func (e *Editor) copyline() {

	if len(e.Content) == 0 || e.Parent == nil { return }

	e.Parent.SetClipboard(e.Content[e.Cursor[1]] + "\n")
}

func (e *Editor) duplicateline() {

	if len(e.Content) == 0 { return }

	e.Content = slices.Insert(e.Content, e.Cursor[1]+1, e.Content[e.Cursor[1]])
	e.movey(1)
	e.clampcursor()
}

func (e *Editor) cutline() {

	if len(e.Content) == 0 || e.Parent == nil { return }

	e.copyline()
	e.Content = slices.Delete(e.Content, e.Cursor[1], e.Cursor[1]+1)
	if len(e.Content) == 0 { e.Content = []string{""} }
	e.Cursor[1] = min(e.Cursor[1], len(e.Content)-1)
	e.clampcursor()
}

func (e *Editor) pasteline() {

	if len(e.Content) == 0 || e.Parent == nil { return }

	str := e.Parent.GetClipboard()
	if str == "" { return }

	lines := strings.Split(strings.TrimSuffix(str, "\n"), "\n")
	e.Content = slices.Insert(e.Content, e.Cursor[1]+1, lines...)
	e.movey(1)
	e.Cursor[0] = 0
}

func (e *Editor) savestate() {

	if len(e.Undo) > 0 && slices.Equal(e.Undo[len(e.Undo)-1].Content, e.Content) {
		e.Undo[len(e.Undo)-1].Cursor = e.Cursor
		return
	}

	e.Undo = append(e.Undo, state{Content: slices.Clone(e.Content), Cursor: e.Cursor})
}

func (e *Editor) undo() {

	if len(e.Undo) == 0 { return }

	last := e.Undo[len(e.Undo)-1]

	if slices.Equal(last.Content, e.Content) {
		if len(e.Undo) == 1 { return }
		e.Undo = e.Undo[:len(e.Undo)-1]
		last = e.Undo[len(e.Undo)-1]
	}

	e.Redo = append(e.Redo, state{Content: slices.Clone(e.Content), Cursor: e.Cursor})
	e.Content = slices.Clone(last.Content)
	e.Cursor = last.Cursor
}

func (e *Editor) redo() {

	if len(e.Redo) == 0 { return }

	last := e.Redo[len(e.Redo)-1]
	e.Redo = e.Redo[:len(e.Redo)-1]
	e.Content = slices.Clone(last.Content)
	e.Cursor = last.Cursor
	e.savestate()
}

func (e *Editor) ctrlenter() {

	if len(e.Content) == 0 { return }

	e.Content = slices.Insert(e.Content, e.Cursor[1]+1, "")
	e.movey(1)
	e.Cursor[0] = 0
}

func (e *Editor) enter() {

	if len(e.Content) == 0 { return }

	e.clampcursor()

	str := []rune(e.Content[e.Cursor[1]])
	e.Content[e.Cursor[1]] = string(str[:e.Cursor[0]])
	e.Content = slices.Insert(e.Content, e.Cursor[1]+1, string(str[e.Cursor[0]:]))
	e.movey(1)
	e.Cursor[0] = 0
}

func (e *Editor) ctrldelete() {

	if len(e.Content) == 0 { return }

	e.clampcursor()

	str := []rune(e.Content[e.Cursor[1]])

	if e.Cursor[0] < len(str) {
		delta := e.nextspace(1)
		e.Content[e.Cursor[1]] = string(str[:e.Cursor[0]]) + string(str[e.Cursor[0]+delta:])
	} else {
		e.delete()
	}
}

func (e *Editor) delete() {

	if len(e.Content) == 0 { return }

	e.clampcursor()

	str := []rune(e.Content[e.Cursor[1]])

	if e.Cursor[0] < len(str) {
		e.Content[e.Cursor[1]] = string(str[:e.Cursor[0]]) + string(str[e.Cursor[0]+1:])
	} else if e.Cursor[1]+1 < len(e.Content) {
		e.Content[e.Cursor[1]] += e.Content[e.Cursor[1]+1]
		e.Content = append(e.Content[:e.Cursor[1]+1], e.Content[e.Cursor[1]+2:]...)
	}
}

func (e *Editor) ctrlbackspace() {

	if len(e.Content) == 0 { return }

	e.clampcursor()

	if e.Cursor[0] > 0 {
		delta := e.nextspace(-1)
		str := []rune(e.Content[e.Cursor[1]])
		e.Content[e.Cursor[1]] = string(str[:e.Cursor[0]+delta]) + string(str[e.Cursor[0]:])
		e.movex(delta)
	} else {
		e.backspace()
	}
}

func (e *Editor) backspace() {

	if len(e.Content) == 0 { return }

	e.clampcursor()

	if e.Cursor[0] > 0 {
		str := []rune(e.Content[e.Cursor[1]])
		e.Content[e.Cursor[1]] = string(str[:e.Cursor[0]-1]) + string(str[e.Cursor[0]:])
		e.movex(-1)
	} else if e.Cursor[1] > 0 {
		e.Cursor[0] = utf8.RuneCountInString(e.Content[e.Cursor[1]-1])
		e.Content[e.Cursor[1]-1] += e.Content[e.Cursor[1]]
		e.Content = append(e.Content[:e.Cursor[1]], e.Content[e.Cursor[1]+1:]...)
		e.movey(-1)
	}
}

func (e *Editor) moveline(delta int) {

	r := e.Cursor[1] + delta
	if e.Cursor[1] >= 0 && e.Cursor[1] < len(e.Content) && r >= 0 && r < len(e.Content) {
		e.Content[e.Cursor[1]], e.Content[r] = e.Content[r], e.Content[e.Cursor[1]]
		e.Cursor[1] = r
	}
}

func (e *Editor) clampcursor() {
	t := utf8.RuneCountInString(e.Content[e.Cursor[1]])
	e.Cursor[0] = max(0, min(e.Cursor[0], t))
}

func (e *Editor) movey(delta int) {
	r := e.Cursor[1] + delta
	if r >= 0 && r < len(e.Content) {
		e.Cursor[1] = r
	}
}

func (e *Editor) movex(delta int) {
	e.Cursor[0] += delta
}

func (e *Editor) nextspace(delta int) int {

	str := []rune(e.Content[e.Cursor[1]])
	pos := max(0, min(e.Cursor[0], len(str)))
	start := pos

	if delta > 0 {
		for pos < len(str) && !unicode.IsSpace(str[pos]) {
			pos++
		}
		for pos < len(str) && unicode.IsSpace(str[pos]) {
			pos++
		}
	} else if delta < 0 {
		for pos > 0 && unicode.IsSpace(str[pos-1]) {
			pos--
		}
		for pos > 0 && !unicode.IsSpace(str[pos-1]) {
			pos--
		}
	}

	return pos - start
}

func (e *Editor) typing(char rune) {
	e.clampcursor()
	e.Content[e.Cursor[1]] = string([]rune(e.Content[e.Cursor[1]])[:e.Cursor[0]]) + string(char) + string([]rune(e.Content[e.Cursor[1]])[e.Cursor[0]:])
	e.movex(1)
}

func (e *Editor) InputSelect(newinput *types.Input) {

	switch newinput.Description {
	case "up":
		e.movefocus(-1)
	case "down":
		e.movefocus(1)
	case "left":
		e.collapse(true)
	case "right":
		e.collapse(false)
	case "enter", "ctrl+enter":
		if e.Focus < 0 || e.Focus >= len(e.Entries) {
			return
		}
		en := e.Entries[e.Focus]
		if en != nil && en.Type == "load-more" {
			amount := 1
			if newinput.Description == "ctrl+enter" { amount = 4 }
			e.loadMore(amount)
			return
		}
		if en == nil || en.Type != "page" || en.Page == nil {
			return
		}
		e.Parent.AddClients("editor:"+en.Page.Path, "after", newinput.Description == "enter")

	}

	e.clampfocus()
}

func (e *Editor) loadMore(amount int) {
	req := e.Request
	module := 0
	if len(e.Requests) > 0 {
		req = e.Requests[0]
		module = 1
	}
	if req == nil || e.Filesystem == nil { return }
	req.Latest = max(0, req.Latest) + amount
	// Refresh only the select tree; preserve the editor's content and cursor.
	page, ok := <-e.Filesystem.Load(req)
	if !ok || page == nil || page.Type == "error" { return }
	collapsed := map[string]bool{}
	for _, en := range e.AllEntries {
		if en.Page != nil { collapsed[en.Page.Path] = en.Collapsed }
	}
	root := e.newPageEntry(page, module)
	for _, en := range e.flattenPageEntry(root) {
		en.Collapsed = collapsed[en.Page.Path]
	}
	e.generateAllEntries()
	e.Focus = len(e.Entries)-1
}

func (e *Editor) collapse(b bool) {
	if e.Focus < 0 || e.Focus >= len(e.Entries) {
		return
	}
	en := e.Entries[e.Focus]
	if en == nil || en.Type != "page" || en.Page == nil || len(en.Children) == 0 {
		return
	}
	en.Collapsed = b
	e.generateEntries()
}

func (e *Editor) movefocus(delta int) {
	e.Focus += delta
}

func (e *Editor) clampfocus() {
	e.Focus = max(0, min(e.Focus, len(e.Entries)-1))
}

func (e *Editor) GetType() string { return "editor" }

func (e *Editor) GetSpec() string { return e.Spec }

func (e *Editor) GetName() string { return filepath.Base(e.Spec) }

func (e *Editor) GetShortLabel() string { return e.GetName() }

func (e *Editor) GetLongLabel() string { return e.Spec }

func (e *Editor) GetTokens() []string {

	tokens := []string{}

	if e.Request != nil {
		tokens = append(tokens, e.Request.Token)
	}

	for _, req := range e.Requests {
		if req != nil {
			tokens = append(tokens, req.Token)
		}
	}

	return tokens
}

func (e *Editor) GetStage() string {
	if e.Page == nil {
		return "ghost"
	}

	if e.Page.Stage != "draft" && e.Page.Og != nil && !utils.CompareContent(e.Page.Og.Content, e.Content) {
		return "edit"
	}

	return e.Page.Stage
}