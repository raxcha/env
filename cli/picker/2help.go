package picker

import (
	"env/cli/editor"
	"env/engine"
	"env/filesystem"
	"env/routines"
	"env/utilities"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	pickerStateGhostSymbol    = "󰊠"
	pickerStateLocalSymbol    = "󰓦"
	pickerStateApiSymbol      = "󱂛"
	pickerStateDoomSymbol     = "󱂥"
	pickerStateDraftSymbol    = "󰦨"
	pickerStateEditSymbol     = "󰤀"
	pickerStateIdSymbol       = ""
	pickerStateConflictSymbol = ""
	pickerStateUnknownSymbol  = "?"
)

func (p *Picker) updateItems() {

	p.RefreshFlat()
	p.StartItems()

	p.ensureReminderItem()

	sort.SliceStable(p.Items, func(i, j int) bool {
		return p.lessItem(p.Items[i], p.Items[j])
	})

	p.pinBottomSpecialItems()

	if p.Selected < 0 {
		p.Selected = 0
	}

	if p.Selected >= len(p.Items) {
		p.Selected = len(p.Items) - 1
	}

	if p.Selected < 0 {
		p.Selected = 0
	}

	if p.StartPath != "" && p.StartPath != "." {
		if p.selectPath(p.StartPath) {
			p.StartPath = ""
		}
	}
}

func (p *Picker) drawItems() *engine.Frame {
	itemsRect, _, _ := p.pickerLayout()

	w := itemsRect.W
	h := itemsRect.H

	if w < 1 {
		w = 1
	}

	if h < 1 {
		h = 1
	}

	paddingLeft := 1
	paddingRight := 1

	contentW := w - paddingLeft - paddingRight
	if contentW < 1 {
		contentW = 1
	}

	items := p.Items
	lines := []string{}

	innerH := h

	start := pickerCutLines(innerH, p.Selected, len(items))
	end := start + innerH

	if end > len(items) {
		end = len(items)
	}

	visibleItems := end - start
	emptyLines := innerH - visibleItems

	for i := 0; i < emptyLines; i++ {
		lines = append(lines, p.fitPickerLine("", w))
	}

	for i := end - 1; i >= start; i-- {
		line := p.itemLine(items[i], contentW)

		if i == p.Selected {
			if items[i] != nil && items[i].Kind == "parent" {
				line = p.selectedFullItemLine(line, contentW)
			} else {
				line = p.selectedItemLine(line, contentW)
			}
		}

		line =
			strings.Repeat(" ", paddingLeft) +
				p.fitPickerLine(line, contentW) +
				strings.Repeat(" ", paddingRight)

		lines = append(lines, p.fitPickerLine(line, w))
	}

	for len(lines) < h {
		lines = append(lines, p.fitPickerLine("", w))
	}

	if len(lines) > h {
		lines = lines[:h]
	}

	return p.Utilities.GenerateFrame(
		engine.Boundaries{Fullsize: p.Bounds.Fullsize, Pos: routines.Bound{itemsRect.X, itemsRect.Y}, Size: routines.Bound{w, h}},
		lines,
		0,
	)
}

func (p *Picker) drawPrompt() *engine.Frame {
	_, promptRect, _ := p.pickerLayout()

	w := promptRect.W
	h := promptRect.H

	if w < 1 {
		w = 1
	}

	if h < 1 {
		h = 1
	}

	context := p.Path
	if context == "" || context == "." {
		context = "prsnl.spc"
	}

	promptLine := " ‹b " + context + " λ›b  " + p.Prompt

	lines := []string{
		p.fitPickerLine(promptLine, w),
	}

	for len(lines) < h {
		lines = append(lines, p.fitPickerLine("", w))
	}

	if len(lines) > h {
		lines = lines[:h]
	}

	return p.Utilities.GenerateFrame(
		engine.Boundaries{Fullsize: p.Bounds.Fullsize, Pos: routines.Bound{promptRect.X, promptRect.Y}, Size: routines.Bound{w, h}},
		lines,
		0,
	)
}

func (p *Picker) modeInfoLine() string {
	return p.Mode + " | " + p.sortingLabel() + " | " + p.searchScopeLabel()
}

func (p *Picker) sortingLabel() string {
	if p.Sorting == "" || p.Sorting == "auto" {
		return "auto:" + p.effectiveSorting()
	}

	return p.Sorting
}

func (p *Picker) drawPreview() *engine.Frame {
	_, _, previewRect := p.pickerLayout()

	w := previewRect.W
	h := previewRect.H

	if w < 1 {
		w = 1
	}

	if h < 1 {
		h = 1
	}

	content := []string{
		" ...",
	}

	if len(p.Items) > 0 && p.Selected >= 0 && p.Selected < len(p.Items) {
		item := p.Items[p.Selected]

		if item != nil && item.Kind == "parent" {
			content = []string{
				" ...",
			}
		} else if item != nil && item.Page != nil {
			page := item.Page
			content = p.pagePreviewLines(page)
		}
	}

	if p.PreviewOffset < 0 {
		p.PreviewOffset = 0
	}

	maxOffset := len(content) - h
	if maxOffset < 0 {
		maxOffset = 0
	}
	if p.PreviewOffset > maxOffset {
		p.PreviewOffset = maxOffset
	}

	firstLine := p.PreviewOffset + 1
	if p.PreviewOffset > 0 {
		content = content[p.PreviewOffset:]
	}

	lines := []string{}
	numberW := len(fmt.Sprintf("%d", firstLine+len(content)-1))
	if numberW < 1 {
		numberW = 1
	}

	numberAreaW := numberW
	contentW := w - numberAreaW - 1
	if contentW < 1 {
		contentW = 1
	}

	styleState := editor.NewPreviewStyleState()
	for i, line := range content {
		prefix := pickerPreviewNumber(firstLine+i, numberW)
		dashedSize := 0
		isDashed := pickerPreviewDashedLine(line)
		if strings.TrimSpace(line) != "" {
			dashedSize = editor.DashedLineVisualSize(content, i, p.Utilities)
		}
		styled := editor.StyleContentLine(line, false, contentW, &styleState, false, dashedSize)
		if isDashed {
			prefix += " "
		}
		styled = pickerInsertAfterStylePrefix(styled, prefix)
		lines = append(lines, p.fitPickerLine(styled, contentW+numberAreaW)+" ")
	}

	for len(lines) < h {
		lines = append(lines, strings.Repeat(" ", numberAreaW)+p.fitPickerLine("", contentW)+" ")
	}

	if len(lines) > h {
		lines = lines[:h]
	}

	return p.Utilities.GenerateFrame(
		engine.Boundaries{Fullsize: p.Bounds.Fullsize, Pos: routines.Bound{previewRect.X, previewRect.Y}, Size: routines.Bound{w, h}},
		lines,
		0,
	)
}

func pickerPreviewLine(markup string, line string) string {
	return markup + " " + strings.TrimLeft(line, " \t")
}

func pickerPreviewNumber(n int, width int) string {
	if width < 1 {
		width = 1
	}

	return fmt.Sprintf("%*d", width, n)
}

func pickerPreviewDashedLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	for _, r := range line {
		if r != '-' {
			return false
		}
	}

	return true
}

func pickerInsertAfterStylePrefix(line string, insert string) string {
	if !strings.HasPrefix(line, "§") {
		return insert + line
	}

	runes := []rune(line)
	if len(runes) < 4 {
		return insert + line
	}

	i := 3
	for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
		i++
	}

	if i >= len(runes) || runes[i] != ' ' {
		return insert + line
	}

	return string(runes[:i+1]) + insert + string(runes[i+1:])
}

func (p *Picker) pagePreviewLines(page *filesystem.Page) []string {
	if page == nil {
		return []string{" ..."}
	}

	lines := []string{}

	diff := pickerPreviewDiff(page)
	if len(diff) > 0 {
		for _, line := range diff {
			switch {
			case strings.HasPrefix(line, "+ "):
				lines = append(lines, pickerPreviewLine("¤22 ", line))
			case strings.HasPrefix(line, "- "):
				lines = append(lines, pickerPreviewLine("¤11 ", line))
			default:
				lines = append(lines, pickerPreviewLine("", line))
			}
		}
		return lines
	}

	if hasVisibleContent(page.Content, p.Utilities) {
		for _, line := range page.Content {
			lines = append(lines, pickerPreviewLine("", p.highlightPickerMatches(line)))
		}
	}

	return lines
}

func pickerPreviewDiff(page *filesystem.Page) []string {
	if page == nil {
		return nil
	}

	if page.Stage == pageStageEdit {
		return pickerBuildDiff(page)
	}

	if len(page.Diff) > 0 {
		return page.Diff
	}

	return pickerBuildDiff(page)
}

func pickerBuildDiff(page *filesystem.Page) []string {
	if page == nil || !pickerHasOriginal(page.Og) {
		return nil
	}

	og := page.Og
	var out []string

	if og.Name != page.Name {
		out = append(out, "name:  "+og.Name+" → "+page.Name)
	}
	if og.Path != page.Path {
		out = append(out, "path:  "+og.Path+" → "+page.Path)
	}
	if og.Type != page.Type {
		out = append(out, "type:  "+og.Type+" → "+page.Type)
	}

	if pickerHasOriginalContent(og) || len(page.Content) == 0 {
		cd := filesystem.FullDiffLines(og.Content, page.Content)
		if len(cd) > 0 {
			if len(out) > 0 {
				out = append(out, "---")
			}
			out = append(out, cd...)
		}
	}

	return out
}

func pickerHasOriginal(page *filesystem.Page) bool {
	if page == nil {
		return false
	}

	return page.Path != "" ||
		page.Name != "" ||
		page.Type != "" ||
		len(page.Content) > 0
}

func pickerHasOriginalContent(page *filesystem.Page) bool {
	return page != nil && len(page.Content) > 0
}

// ...

func hasVisibleContent(lines []string, u *utilities.Utilities) bool {
	for _, line := range lines {
		if u.VisibleLength(line) > 0 {
			return true
		}
	}

	return false
}

func (p *Picker) StartItems() {
	p.Items = []*Match{}

	prompt := strings.ToLower(strings.TrimSpace(p.Prompt))

	if !p.isRootContext() && prompt == "" {
		p.Items = append(p.Items, &Match{
			Page:  nil,
			Score: 0,
			Kind:  "parent",
		})
	}

	for _, page := range p.Flat {
		if page == nil {
			continue
		}

		score := p.scorePage(prompt, page)

		match := &Match{
			Page:  page,
			Score: score,
			Kind:  "page",
		}

		p.Items = append(p.Items, match)
	}
}

func (p *Picker) isRootContext() bool {
	return p.Path == "" || p.Path == "."
}

func (p *Picker) scorePage(prompt string, page *filesystem.Page) int {
	if page == nil {
		return 0
	}

	if prompt == "" {
		return 1
	}

	score := 0

	switch p.Mode {
	case "literal":
		score = p.scorePageLiteral(prompt, page)

	case "fuzzy":
		score = p.scorePageFuzzy(prompt, page)

	case "metadata":
		score = p.scorePageMetadata(prompt, page)

	case "recent":
		score = p.scorePageRecent(prompt, page)

	default:
		score = p.scorePageLiteral(prompt, page)
	}

	if score > 0 {
		return score
	}

	if p.pickerWeakMatch(prompt, page) {
		return 1
	}

	return 0
}

func (p *Picker) pickerWeakMatch(prompt string, page *filesystem.Page) bool {
	if page == nil {
		return false
	}

	prompt = strings.ToLower(strings.TrimSpace(prompt))
	if prompt == "" {
		return true
	}

	name := strings.ToLower(page.Name)
	path := strings.ToLower(page.Path)
	meta := pickerMetadataText(page)
	options := pickerOptionsText(page)
	content := strings.ToLower(strings.Join(page.Content, "\n"))

	text := strings.TrimSpace(strings.Join([]string{
		name,
		path,
		meta,
		options,
		content,
	}, " "))

	if strings.Contains(text, prompt) {
		return true
	}

	for _, word := range strings.Fields(prompt) {
		if word == "" {
			continue
		}

		if !pickerSubsequenceMatch(word, text) {
			return false
		}
	}

	return true
}

func pickerSubsequenceMatch(prompt string, text string) bool {
	if prompt == "" {
		return true
	}

	if text == "" {
		return false
	}

	p := []rune(prompt)
	t := []rune(text)

	j := 0

	for i := 0; i < len(t) && j < len(p); i++ {
		if t[i] == p[j] {
			j++
		}
	}

	return j == len(p)
}

func (p *Picker) scorePageLiteral(prompt string, page *filesystem.Page) int {
	if page == nil {
		return 0
	}

	if prompt == "" {
		return 1
	}

	score := 0
	name := strings.ToLower(page.Name)
	path := strings.ToLower(page.Path)
	content := strings.ToLower(strings.Join(page.Content, "\n"))
	meta := pickerMetadataText(page)

	if name == prompt {
		score += 2200
	}

	if path == prompt {
		score += 1800
	}

	if strings.HasPrefix(name, prompt) {
		score += 1400
	}

	if strings.HasPrefix(path, prompt) {
		score += 900
	}

	if strings.Contains(name, prompt) {
		score += 900
	}

	if strings.Contains(path, prompt) {
		score += 650
	}

	if strings.Contains(meta, prompt) {
		score += 450
	}

	if strings.Contains(content, prompt) {
		score += 250
	}

	for _, word := range strings.Fields(prompt) {
		if strings.Contains(name, word) {
			score += 160
		}

		if strings.Contains(path, word) {
			score += 110
		}

		if strings.Contains(meta, word) {
			score += 70
		}

		if strings.Contains(content, word) {
			score += 30
		}
	}

	return score
}

func (p *Picker) scorePageFuzzy(prompt string, page *filesystem.Page) int {
	if page == nil {
		return 0
	}

	if prompt == "" {
		return 1
	}

	score := 0
	name := strings.ToLower(page.Name)
	path := strings.ToLower(page.Path)

	if name == prompt {
		score += 1600
	}

	if strings.HasPrefix(name, prompt) {
		score += 900
	}

	if strings.Contains(name, prompt) {
		score += 550
	}

	if strings.HasPrefix(path, prompt) {
		score += 420
	}

	if strings.Contains(path, prompt) {
		score += 260
	}

	score += scoreFuzzySequence(prompt, name) * 12
	score += scoreFuzzySequence(prompt, path) * 5

	return score
}

func (p *Picker) scorePageMetadata(prompt string, page *filesystem.Page) int {
	if page == nil {
		return 0
	}

	if prompt == "" {
		return 1
	}

	score := 0
	name := strings.ToLower(page.Name)
	path := strings.ToLower(page.Path)
	meta := pickerMetadataText(page)
	options := pickerOptionsText(page)
	beforeDash := pickerBeforeFirstDashText(page)
	structured := strings.TrimSpace(strings.Join([]string{name, path, meta, options, beforeDash}, " "))

	if name == prompt {
		score += 1200
	}

	if strings.Contains(name, prompt) {
		score += 500
	}

	if strings.Contains(path, prompt) {
		score += 420
	}

	if strings.Contains(meta, prompt) {
		score += 1400
	}

	if strings.Contains(options, prompt) {
		score += 850
	}

	if strings.Contains(beforeDash, prompt) {
		score += 900
	}

	for key, value := range page.Metadata {
		k := strings.ToLower(key)
		v := strings.ToLower(pickerAnyToString(value))

		if k == prompt {
			score += 1600
		}

		if v == prompt {
			score += 1800
		}

		if strings.Contains(k, prompt) {
			score += 700
		}

		if strings.Contains(v, prompt) {
			score += 1000
		}
	}

	for _, word := range strings.Fields(prompt) {
		if strings.Contains(meta, word) {
			score += 180
		}

		if strings.Contains(options, word) {
			score += 130
		}

		if strings.Contains(beforeDash, word) {
			score += 120
		}
	}

	score += scoreFuzzySequence(prompt, structured) * 2

	return score
}

func (p *Picker) scorePageRecent(prompt string, page *filesystem.Page) int {
	if page == nil {
		return 0
	}

	base := p.scorePageLiteral(prompt, page)/2 + p.scorePageFuzzy(prompt, page)/3 + p.scorePageMetadata(prompt, page)/4

	if prompt == "" {
		base = 100
	}

	return base * pickerRecencyMultiplier(page)
}

func (p *Picker) RefreshFlat() {
	p.Flat = []*filesystem.Page{}

	if p.Cache == nil {
		return
	}

	if p.isRootContext() {
		if p.recursiveSearchOn() {
			for _, child := range rootChildren(p.Cache) {
				p.flattenPage(child)
			}
			return
		}

		p.Flat = rootChildren(p.Cache)
		return
	}

	context := findPageByPath(p.Cache, p.Path)
	if context == nil {
		return
	}

	if p.recursiveSearchOn() {
		for _, child := range context.Children {
			p.flattenPage(child)
		}
		return
	}

	p.Flat = context.Children
}

func (p *Picker) recursiveSearchOn() bool {
	return strings.TrimSpace(p.Prompt) != "" && p.Scope == "tree"
}

func (p *Picker) searchScopeLabel() string {
	if p.Scope == "tree" {
		return "tree"
	}

	return "context"
}

func (p *Picker) flattenPage(page *filesystem.Page) {
	if page == nil {
		return
	}

	// não inclui a raiz "." na lista
	if page.Path != "." && page.Path != "" {
		p.Flat = append(p.Flat, page)
	}

	for _, child := range page.Children {
		p.flattenPage(child)
	}
}

func pickerCutLines(height int, selected int, total int) int {
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

	if start < 0 {
		start = 0
	}

	if start > total-height {
		start = total - height
	}

	return start
}

func (p *Picker) itemLine(item *Match, width int) string {
	if item == nil {
		return ""
	}

	if item.Kind == "parent" {
		return " ..."
	}

	if item.Page == nil {
		return ""
	}

	page := item.Page
	name := page.Name

	if name == "" {
		name = page.Path
	}

	childStatus := pickerStateGapText(page)
	ownStatus := pickerOwnStatusIcon(page)
	score := p.pickerScoreText(item)

	leftPad := " "

	label := p.pickerPageLabel(page, name)
	label = p.patchVisualLabel(page, label)
	label = leftPad + label

	labelVisible := p.Utilities.VisibleLength(label)
	scoreVisible := p.Utilities.VisibleLength(score)
	childStatusVisible := p.Utilities.VisibleLength(childStatus)
	ownStatusVisible := p.Utilities.VisibleLength(ownStatus)

	rightVisible := scoreVisible

	if childStatus != "" {
		rightVisible += childStatusVisible + 1
	}

	if ownStatus != "" {
		rightVisible += ownStatusVisible + 1
	}

	trailingSpace := 0
	if rightVisible > 0 && width > 0 {
		trailingSpace = 1
	}

	gap := width - labelVisible - rightVisible - trailingSpace

	if gap < 1 && rightVisible > 0 {
		leftPadVisible := p.Utilities.VisibleLength(leftPad)

		maxLabel := width - rightVisible - trailingSpace - 1
		if maxLabel < 0 {
			maxLabel = 0
		}

		maxInnerLabel := maxLabel - leftPadVisible
		if maxInnerLabel < 0 {
			maxInnerLabel = 0
		}

		label = leftPad + p.truncatePickerLabel(page, name, maxInnerLabel)
		labelVisible = p.Utilities.VisibleLength(label)
		gap = width - labelVisible - rightVisible - trailingSpace
	}

	if gap < 0 {
		gap = 0
	}

	if rightVisible == 0 {
		if p.Utilities.VisibleLength(label) > width {
			leftPadVisible := p.Utilities.VisibleLength(leftPad)
			maxInnerLabel := width - leftPadVisible
			if maxInnerLabel < 0 {
				maxInnerLabel = 0
			}

			label = leftPad + p.truncatePickerLabel(page, name, maxInnerLabel)
		}

		return label
	}

	rightParts := []string{}

	if childStatus != "" {
		rightParts = append(rightParts, childStatus)
	}

	rightParts = append(rightParts, score)

	if ownStatus != "" {
		rightParts = append(rightParts, ownStatus)
	}

	right := strings.Join(rightParts, " ")

	return label + strings.Repeat(" ", gap) + right + strings.Repeat(" ", trailingSpace)
}

func (p *Picker) pickerScoreText(item *Match) string {
	if item == nil {
		return ""
	}

	if strings.TrimSpace(p.Prompt) == "" {
		return ""
	}

	if item.Kind == "parent" {
		return ""
	}

	score := p.pickerDisplayScore(item.Score)

	return fmt.Sprintf("%+0.2f", score)
}

func (p *Picker) selectPath(path string) bool {
	if path == "" || path == "." {
		return false
	}

	for i, item := range p.Items {
		if item == nil || item.Page == nil {
			continue
		}

		if item.Page.Path == path {
			p.Selected = i
			return true
		}
	}

	return false
}

func (p *Picker) findPickerRoot() *filesystem.Page {
	if p.Path == "" || p.Path == "." {
		return p.Cache
	}

	// tenta achar exatamente o path pedido
	root := p.findPageByPath(p.Cache, p.Path)
	if root != nil {
		return root
	}

	// fallback:
	// se picker:b.rec/asdasd ainda não existe,
	// tenta abrir pelo pai b.rec
	parentPath := parentPathOf(p.Path)

	for parentPath != "" && parentPath != "." {
		root = p.findPageByPath(p.Cache, parentPath)
		if root != nil {
			return root
		}

		parentPath = parentPathOf(parentPath)
	}

	return p.Cache
}

func (p *Picker) findPageByPath(page *filesystem.Page, path string) *filesystem.Page {
	if page == nil {
		return nil
	}

	if page.Path == path {
		return page
	}

	for _, child := range page.Children {
		found := p.findPageByPath(child, path)
		if found != nil {
			return found
		}
	}

	return nil
}

func (p *Picker) findParentPage(path string) *filesystem.Page {
	parentPath := parentPathOf(path)

	if parentPath == "" {
		parentPath = "."
	}

	return p.findPageByPath(p.Cache, parentPath)
}

func parentPathOf(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")

	if path == "" || path == "." || path == "prsnl.spc" {
		return "."
	}

	if path == "a.log" ||
		path == "b.rec" ||
		path == "c.rand" ||
		path == "d.fami" ||
		path == "e.proj" ||
		path == ".resources" {
		return "."
	}

	parent := filepath.Dir(path)

	if parent == "." || parent == "/" || parent == "" {
		return "."
	}

	return parent
}

func rootChildren(root *filesystem.Page) []*filesystem.Page {
	if root == nil {
		return nil
	}

	names := []string{
		"a.log",
		"b.rec",
		"c.rand",
		"d.fami",
		"e.proj",
		".resources",
	}

	out := []*filesystem.Page{}
	seen := map[string]bool{}

	for _, name := range names {
		child := findPageByPath(root, name)
		if child != nil {
			out = append(out, child)
			seen[child.Path] = true
		}
	}

	extra := []*filesystem.Page{}
	for _, child := range root.Children {
		if child == nil || seen[child.Path] {
			continue
		}
		extra = append(extra, child)
	}
	sort.SliceStable(extra, func(i, j int) bool {
		return compareBasicPages(extra[i], extra[j]) < 0
	})

	out = append(out, extra...)
	return out
}

func findPageByPath(page *filesystem.Page, path string) *filesystem.Page {
	if page == nil {
		return nil
	}

	if page.Path == path {
		return page
	}

	for _, child := range page.Children {
		found := findPageByPath(child, path)
		if found != nil {
			return found
		}
	}

	return nil
}

func pickerTypeIcon(page *filesystem.Page) string {
	if page == nil {
		return " "
	}

	if page.Type == "deep" {
		return " "
	}

	if page.Type == "shallow" {
		return " "
	}

	return " "
}

func pickerStateGapText(page *filesystem.Page) string {
	if page == nil {
		return ""
	}

	counts := map[string]int{}

	pickerCollectStateCounts(counts, page, false)

	parts := []string{}

	if counts[pageStageId] > 0 {
		parts = append(parts, pickerTinyCount(counts[pageStageId])+pickerStateIdSymbol)
	}

	if counts[pageStageEdit] > 0 {
		parts = append(parts, pickerTinyCount(counts[pageStageEdit])+pickerStateEditSymbol)
	}

	if counts[pageStageDraft] > 0 {
		parts = append(parts, pickerTinyCount(counts[pageStageDraft])+pickerStateDraftSymbol)
	}

	if counts[pageStageDoom] > 0 {
		parts = append(parts, pickerTinyCount(counts[pageStageDoom])+pickerStateDoomSymbol)
	}

	if counts[pageStageConflict] > 0 {
		parts = append(parts, pickerTinyCount(counts[pageStageConflict])+pickerStateConflictSymbol)
	}

	return strings.Join(parts, " ")
}

func pickerCollectStateCounts(counts map[string]int, page *filesystem.Page, includeSelf bool) {
	if page == nil {
		return
	}

	if includeSelf {
		pickerCountState(counts, page.Stage)
	}

	if page.Type != "deep" {
		return
	}

	for _, child := range page.Children {
		pickerCollectStateCounts(counts, child, true)
	}
}

func pickerCountState(counts map[string]int, state string) {
	switch state {
	case pageStageId,
		pageStageEdit,
		pageStageDraft,
		pageStageDoom,
		pageStageConflict:

		counts[state]++
	}
}

func pickerCollectOwnState(states map[string]bool, page *filesystem.Page) {
	if page == nil {
		return
	}

	switch page.Stage {
	case pageStageLocal,
		pageStageApi,
		pageStageGhost,
		pageStageDoom,
		pageStageDraft,
		pageStageEdit,
		pageStageId,
		pageStageConflict:

		states[page.Stage] = true
	}
}

func pickerCollectChildStates(states map[string]bool, page *filesystem.Page) {
	if page == nil {
		return
	}

	for _, child := range page.Children {
		if child == nil {
			continue
		}

		switch child.Stage {
		case pageStageDoom,
			pageStageDraft,
			pageStageEdit,
			pageStageId,
			pageStageConflict:

			states[child.Stage] = true
		}

		pickerCollectChildStates(states, child)
	}
}

func pickerOwnSymbol(name string) (rune, string, bool) {
	runes := []rune(name)

	if len(runes) == 0 {
		return 0, "", false
	}

	first := runes[0]

	if first == '>' ||
		first == '=' ||
		first == '~' ||
		first == '`' {
		rest := strings.TrimLeft(string(runes[1:]), " ")
		return first, rest, true
	}

	return 0, name, false
}

func (p *Picker) pickerPageLabel(page *filesystem.Page, name string) string {
	symbol, rest, ok := pickerOwnSymbol(name)

	if ok {
		return "‹b " + string(symbol) + " ›b " + p.boldPickerMatches(rest)
	}

	if isRemindersPage(name) {
		return "‹b 󰔟 ›b " + p.boldPickerMatches(name)
	}

	if p.isRootContext() {
		greek := pickerGreekIcon(page)

		if greek != "" {
			return "‹b " + greek + " ›b " + p.boldPickerMatches(name)
		}

		return pickerRootGreekIcon(page.Path) + p.boldPickerMatches(name)
	}
	greek := pickerGreekIcon(page)

	if greek != "" {
		return "‹b " + greek + " ›b " + p.boldPickerMatches(name)
	}

	icon := pickerTypeIcon(page)
	return icon + p.boldPickerMatches(name)
}

func (p *Picker) boldPickerMatches(str string) string {
	prompt := strings.ToLower(strings.TrimSpace(p.Prompt))
	if prompt == "" {
		return str
	}

	runes := []rune(str)
	lower := []rune(strings.ToLower(str))
	marked := make([]bool, len(runes))

	words := []string{prompt}

	for _, word := range strings.Fields(prompt) {
		if word != "" && word != prompt {
			words = append(words, word)
		}
	}

	for _, word := range words {
		target := []rune(word)
		if len(target) == 0 || len(target) > len(lower) {
			continue
		}

		for i := 0; i <= len(lower)-len(target); i++ {
			ok := true

			for j := 0; j < len(target); j++ {
				if lower[i+j] != target[j] {
					ok = false
					break
				}
			}

			if ok {
				for j := 0; j < len(target); j++ {
					marked[i+j] = true
				}
			}
		}
	}

	hasLiteral := false
	for _, ok := range marked {
		if ok {
			hasLiteral = true
			break
		}
	}

	if !hasLiteral && p.Mode == "fuzzy" {
		target := []rune(prompt)
		j := 0

		for i := 0; i < len(lower) && j < len(target); i++ {
			if lower[i] == target[j] {
				marked[i] = true
				j++
			}
		}
	}

	out := ""
	open := false

	for i, r := range runes {
		if marked[i] && !open {
			out += "‹a "
			open = true
		}

		if !marked[i] && open {
			out += "›a "
			open = false
		}

		out += string(r)
	}

	if open {
		out += "›a "
	}

	return out
}

func isRemindersPage(name string) bool {
	return strings.ToLower(strings.TrimSpace(name)) == "reminders"
}

func pickerRootGreekIcon(path string) string {
	switch path {
	case "a.log":
		return "‹b Φ ›b "
	case "b.rec":
		return "‹b Σ ›b "
	case "c.rand":
		return "‹b γ ›b "
	case "d.fami":
		return "‹b δ ›b "
	case "e.proj":
		return "‹b ε ›b "
	}

	return "‹b λ ›b "
}

func pickerRandomGreekIcon(name string) string {
	// icons := []rune{'ζ', 'η', '¤', 'ι', 'κ', 'λ', 'μ', 'ν', 'ξ', 'ο', 'π', 'ρ', 'σ', 'τ', 'υ', 'φ', 'χ', '§', '¬'}
	icons := []rune{'ζ', 'η', 'ι', 'κ', 'λ', 'μ', 'ν', 'ξ', 'ο', 'π', 'ρ', 'σ', 'τ', 'υ', 'φ', 'χ', '¬'}

	if len(name) == 0 {
		return "‹b λ ›b "
	}

	sum := 0
	for _, r := range name {
		sum += int(r)
	}

	icon := icons[sum%len(icons)]

	return "‹b " + string(icon) + " ›b "
}

func pickerCalendarIcon() string {
	return "‹b 󰃭 ›b "
}

func (p *Picker) truncatePickerLabel(page *filesystem.Page, name string, width int) string {
	if width <= 0 {
		return ""
	}

	cutRunes := func(str string, max int) string {
		if max <= 0 {
			return ""
		}

		runes := []rune(str)

		if len(runes) <= max {
			return str
		}

		return string(runes[:max])
	}

	withIcon := func(icon string, label string) string {
		iconW := p.Utilities.VisibleLength(icon)

		maxLabelW := width - iconW
		if maxLabelW < 0 {
			maxLabelW = 0
		}

		label = cutRunes(label, maxLabelW)
		label = p.boldPickerMatches(label)

		return icon + label
	}

	symbol, rest, ok := pickerOwnSymbol(name)

	if ok {
		icon := "‹b " + string(symbol) + " ›b "
		return withIcon(icon, rest)
	}

	if isRemindersPage(name) {
		icon := "‹b 󰔟 ›b "
		return withIcon(icon, name)
	}

	if p.isRootContext() {
		greek := pickerGreekIcon(page)

		if greek != "" {
			icon := "‹b " + greek + " ›b "
			return withIcon(icon, name)
		}

		icon := pickerRootGreekIcon(page.Path)
		return withIcon(icon, name)
	}

	greek := pickerGreekIcon(page)

	if greek != "" {
		icon := "‹b " + greek + " ›b "
		return withIcon(icon, name)
	}

	if greek != "" {
		icon := "‹b " + greek + " ›b "
		return withIcon(icon, name)
	}

	icon := pickerTypeIcon(page)
	return withIcon(icon, name)
}

func (p *Picker) selectedItemLine(line string, width int) string {
	if width <= 0 {
		return line
	}

	visible := p.Utilities.VisibleLength(line)

	if visible > width {
		line = p.Utilities.CutVisible(line, width)
		visible = p.Utilities.VisibleLength(line)
	}

	if visible < width {
		line += strings.Repeat(" ", width-visible)
	}

	// O highlight vai até o último caractere real da linha,
	// ou seja: inclui o símbolo da direita, mas não pinta
	// os espaços vazios até a borda da box.
	highlightW := p.Utilities.VisibleLength(strings.TrimRight(line, " ")) + 1

	if highlightW > width {
		highlightW = width
	}
	if highlightW < 0 {
		highlightW = 0
	}

	left := p.Utilities.CutVisible(line, highlightW)
	right := p.Utilities.CutVisibleFrom(line, highlightW)

	return "¤KK " + left + "¤ " + right
}

func pickerMetadataText(page *filesystem.Page) string {
	if page == nil || page.Metadata == nil {
		return ""
	}

	parts := []string{}

	for k, v := range page.Metadata {
		parts = append(parts, strings.ToLower(k))
		v2, ok := v.(string)
		if ok {
			parts = append(parts, strings.ToLower(v2))
		}

	}

	return strings.Join(parts, " ")
}

func scoreFuzzySequence(prompt string, text string) int {
	if prompt == "" || text == "" {
		return 0
	}

	p := []rune(prompt)
	t := []rune(text)

	j := 0
	first := -1
	last := -1
	prev := -1
	score := 0

	for i := 0; i < len(t) && j < len(p); i++ {
		if t[i] != p[j] {
			continue
		}

		if first == -1 {
			first = i
		}

		last = i
		score += 10

		if prev >= 0 && i == prev+1 {
			score += 15
		}

		if i == 0 {
			score += 20
		}

		if i > 0 && isPickerSeparator(t[i-1]) {
			score += 12
		}

		prev = i
		j++
	}

	if j != len(p) {
		return 0
	}

	span := last - first + 1
	gaps := span - len(p)

	score -= gaps * 2
	score -= first / 2

	if score < 1 {
		return 1
	}

	return score
}

func isPickerSeparator(r rune) bool {
	return r == ' ' ||
		r == '-' ||
		r == '_' ||
		r == '/' ||
		r == '.' ||
		r == ':'
}

func pickerBeforeFirstDashText(page *filesystem.Page) string {
	if page == nil {
		return ""
	}

	parts := []string{}

	for _, line := range page.Content {
		if strings.TrimSpace(line) == "---" {
			break
		}

		parts = append(parts, strings.ToLower(line))
	}

	return strings.Join(parts, "\n")
}

func pickerOptionsText(page *filesystem.Page) string {
	if page == nil || page.Options == nil {
		return ""
	}

	parts := []string{}

	for key, value := range page.Options {
		parts = append(parts, strings.ToLower(key))
		parts = append(parts, strings.ToLower(pickerAnyToString(value)))
	}

	return strings.Join(parts, " ")
}

func pickerAnyToString(value any) string {
	if value == nil {
		return ""
	}

	s, ok := value.(string)
	if ok {
		return strings.TrimSpace(s)
	}

	return strings.TrimSpace(fmt.Sprint(value))
}

func pickerRecencyMultiplier(page *filesystem.Page) int {
	t, ok := pickerLastEditedTime(page)
	if !ok {
		return 1
	}

	days := time.Since(t).Hours() / 24

	switch {
	case days < 0:
		return 4
	case days <= 1:
		return 10
	case days <= 3:
		return 8
	case days <= 7:
		return 6
	case days <= 14:
		return 4
	case days <= 30:
		return 3
	case days <= 90:
		return 2
	default:
		return 1
	}
}

func pickerLastEditedTime(page *filesystem.Page) (time.Time, bool) {
	if page == nil || page.Metadata == nil {
		return time.Time{}, false
	}

	raw := pickerAnyToString(page.Metadata["last-edited-time"])
	if raw == "" {
		return time.Time{}, false
	}

	return pickerParseTime(raw)
}

func pickerParseTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"02.01.2006 15:04:05",
		"02.01.2006 15:04",
		"02.01.2006",
	}

	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, raw, time.Local)
		if err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

func pickerMatchPath(match *Match) string {
	if match == nil || match.Page == nil {
		return ""
	}

	return match.Page.Path
}

func (p *Picker) pickerDisplayScore(score int) float64 {
	if score <= 0 {
		return -1.0
	}

	scale := 800.0

	switch p.Mode {
	case "literal":
		scale = 900.0

	case "fuzzy":
		scale = 500.0

	case "metadata":
		scale = 1200.0

	case "recent":
		scale = 2500.0
	}

	s := float64(score)

	return 2*(s/(s+scale)) - 1
}

func pickerGreekIcon(page *filesystem.Page) string {
	if page == nil {
		return ""
	}

	if page.Type == "deep" {
		return pickerGreekUpper(page.Path)
	}

	if page.Type == "shallow" {
		return pickerGreekLower(page.Path)
	}

	return ""
}

func pickerGreekUpper(seed string) string {
	letters := []string{
		"Α", "Β", "Γ", "Δ", "Ε", "Ζ", "Η", "Θ",
		"Ι", "Κ", "Λ", "Μ", "Ν", "Ξ", "Ο", "Π",
		"Ρ", "Σ", "Τ", "Υ", "Φ", "Χ", "Ψ", "Ω",
	}

	return letters[pickerHashIndex(seed, len(letters))]
}

func pickerGreekLower(seed string) string {
	letters := []string{
		"α", "β", "γ", "δ", "ε", "ζ", "η", "θ",
		"ι", "κ", "λ", "μ", "ν", "ξ", "ο", "π",
		"ρ", "σ", "τ", "υ", "φ", "χ", "ψ", "ω",
	}

	return letters[pickerHashIndex(seed, len(letters))]
}

func pickerHashIndex(seed string, size int) int {
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

func (p *Picker) itemRank(item *Match) int {
	if item == nil {
		return 99
	}

	if item.Kind == "parent" {
		return 0
	}

	emptyPrompt := strings.TrimSpace(p.Prompt) == ""

	if p.Path == "b.rec" || p.Path == "c.rand" {
		if emptyPrompt && item.Page != nil {
			switch item.Page.Stage {
			case pageStageConflict:
				return 1
			case pageStageDoom:
				return 2
			case pageStageEdit:
				return 3
			case pageStageDraft:
				return 4
			}
		}
		if p.isPinnedReminder(item) {
			return 5
		}
		return 6
	}

	if emptyPrompt && item.Page != nil {
		switch item.Page.Stage {
		case pageStageConflict:
			return 1
		case pageStageDoom:
			return 2
		case pageStageEdit:
			return 3
		case pageStageDraft:
			return 4
		}
	}

	return 5
}

func (p *Picker) lessItem(a *Match, b *Match) bool {
	ra := p.itemRank(a)
	rb := p.itemRank(b)
	if ra != rb {
		return ra < rb
	}

	if strings.TrimSpace(p.Prompt) != "" && a != nil && b != nil && a.Score != b.Score {
		return a.Score > b.Score
	}

	if cmp, ok := p.compareItemsBySorting(a, b); ok {
		return cmp < 0
	}

	return pickerMatchPath(a) < pickerMatchPath(b)
}

func (p *Picker) compareItemsBySorting(a *Match, b *Match) (int, bool) {
	if a == nil || b == nil || a.Page == nil || b.Page == nil {
		return 0, false
	}

	switch p.effectiveSorting() {
	case "priority":
		ap, aok := pickerPagePriority(a.Page)
		bp, bok := pickerPagePriority(b.Page)
		if cmp, ok := compareIntsDesc(ap, aok, bp, bok); ok {
			return cmp, true
		}
		return compareBasicPages(a.Page, b.Page), true

	case "metadata-time":
		at, aok := pickerPageMetadataTime(a.Page)
		bt, bok := pickerPageMetadataTime(b.Page)
		if cmp, ok := compareTimesDesc(at, aok, bt, bok); ok {
			return cmp, true
		}
		return compareBasicPages(a.Page, b.Page), true

	case "log-time":
		at, aok := pickerPageLogTitleTime(a.Page)
		bt, bok := pickerPageLogTitleTime(b.Page)
		if cmp, ok := compareTimesDesc(at, aok, bt, bok); ok {
			return cmp, true
		}
		return compareBasicPages(a.Page, b.Page), true

	case "time":
		at, aok := pickerBestPageTime(a.Page)
		bt, bok := pickerBestPageTime(b.Page)
		if cmp, ok := compareTimesDesc(at, aok, bt, bok); ok {
			return cmp, true
		}
		return compareBasicPages(a.Page, b.Page), true

	case "basic":
		return compareBasicPages(a.Page, b.Page), true
	}

	return 0, false
}

func (p *Picker) effectiveSorting() string {
	if p.Sorting != "" && p.Sorting != "auto" {
		return p.Sorting
	}

	path := strings.Trim(strings.TrimSpace(p.Path), "/")
	switch {
	case path == "a.log" || strings.HasPrefix(path, "a.log/"):
		return "log-time"
	case path == "b.rec" || strings.HasPrefix(path, "b.rec/"):
		return "metadata-time"
	case path == "c.rand" || strings.HasPrefix(path, "c.rand/"):
		return "metadata-time"
	case path == "e.proj":
		return "priority"
	case strings.HasPrefix(path, "e.proj/"):
		return "basic"
	case path == "d.fami" || strings.HasPrefix(path, "d.fami/"):
		return "basic"
	default:
		return "basic"
	}
}

func compareBasicPages(a *filesystem.Page, b *filesystem.Page) int {
	if a == nil || b == nil {
		return 0
	}

	if a.Type != b.Type {
		if a.Type == "deep" {
			return -1
		}
		if b.Type == "deep" {
			return 1
		}
	}

	an := strings.ToLower(a.Name)
	bn := strings.ToLower(b.Name)
	if an == "" {
		an = strings.ToLower(a.Path)
	}
	if bn == "" {
		bn = strings.ToLower(b.Path)
	}
	if an < bn {
		return -1
	}
	if an > bn {
		return 1
	}
	if a.Path < b.Path {
		return -1
	}
	if a.Path > b.Path {
		return 1
	}
	return 0
}

func compareIntsDesc(a int, aok bool, b int, bok bool) (int, bool) {
	if !aok && !bok {
		return 0, false
	}
	if aok && !bok {
		return -1, true
	}
	if !aok && bok {
		return 1, true
	}
	if a > b {
		return -1, true
	}
	if a < b {
		return 1, true
	}
	return 0, false
}

func compareTimesDesc(a time.Time, aok bool, b time.Time, bok bool) (int, bool) {
	if !aok && !bok {
		return 0, false
	}
	if aok && !bok {
		return -1, true
	}
	if !aok && bok {
		return 1, true
	}
	if a.After(b) {
		return -1, true
	}
	if a.Before(b) {
		return 1, true
	}
	return 0, false
}

func pickerPagePriority(page *filesystem.Page) (int, bool) {
	if page == nil || page.Metadata == nil {
		return 0, false
	}

	switch value := page.Metadata["priority"].(type) {
	case int:
		return value, true
	case float64:
		return int(value), true
	case string:
		value = strings.TrimSpace(value)
		if value == "" {
			return 0, false
		}
		var out int
		_, err := fmt.Sscanf(value, "%d", &out)
		return out, err == nil
	default:
		raw := strings.TrimSpace(fmt.Sprint(value))
		if raw == "" {
			return 0, false
		}
		var out int
		_, err := fmt.Sscanf(raw, "%d", &out)
		return out, err == nil
	}
}

func pickerPageMetadataTime(page *filesystem.Page) (time.Time, bool) {
	if page == nil || page.Metadata == nil {
		return time.Time{}, false
	}

	if value, ok := page.Metadata["time"].(time.Time); ok {
		return value, true
	}

	for _, key := range []string{"time", "release-time", "last-edited-time"} {
		raw := pickerAnyToString(page.Metadata[key])
		if t, ok := pickerParseTime(raw); ok {
			return t, true
		}
	}

	return time.Time{}, false
}

func pickerPageLogTitleTime(page *filesystem.Page) (time.Time, bool) {
	if page == nil {
		return time.Time{}, false
	}

	name := page.Name
	if name == "" {
		name = filepath.Base(page.Path)
	}

	return pickerParseTime(name)
}

func pickerBestPageTime(page *filesystem.Page) (time.Time, bool) {
	if t, ok := pickerPageMetadataTime(page); ok {
		return t, true
	}
	return pickerPageLogTitleTime(page)
}

func (p *Picker) isPinnedReminder(item *Match) bool {
	if item == nil || item.Page == nil {
		return false
	}

	if p.Path != "b.rec" && p.Path != "c.rand" {
		return false
	}

	return isRemindersPage(item.Page.Name)
}

func (p *Picker) highlightPickerMatches(str string) string {
	prompt := strings.ToLower(strings.TrimSpace(p.Prompt))
	if prompt == "" {
		return str
	}

	words := []string{prompt}

	for _, word := range strings.Fields(prompt) {
		if word != "" && word != prompt {
			words = append(words, word)
		}
	}

	runes := []rune(str)
	lowerRunes := []rune(strings.ToLower(str))

	marked := make([]bool, len(runes))

	for _, word := range words {
		wordRunes := []rune(word)
		if len(wordRunes) == 0 {
			continue
		}

		for i := 0; i <= len(lowerRunes)-len(wordRunes); i++ {
			ok := true

			for j := 0; j < len(wordRunes); j++ {
				if lowerRunes[i+j] != wordRunes[j] {
					ok = false
					break
				}
			}

			if ok {
				for j := 0; j < len(wordRunes); j++ {
					marked[i+j] = true
				}
			}
		}
	}

	// fallback fuzzy: se não achou substring literal, destaca a sequência fuzzy
	hasLiteral := false
	for _, ok := range marked {
		if ok {
			hasLiteral = true
			break
		}
	}

	if !hasLiteral && p.Mode == "fuzzy" {
		promptRunes := []rune(prompt)
		j := 0

		for i := 0; i < len(lowerRunes) && j < len(promptRunes); i++ {
			if lowerRunes[i] == promptRunes[j] {
				marked[i] = true
				j++
			}
		}
	}

	out := ""
	open := false

	for i, r := range runes {
		if marked[i] && !open {
			out += "¤bA ‹b "
			open = true
		}

		if !marked[i] && open {
			out += "›b ¤ "
			open = false
		}

		out += string(r)
	}

	if open {
		out += "›b ¤ "
	}

	return out
}

func (p *Picker) pinBottomSpecialItems() {
	if p.Path != "b.rec" && p.Path != "c.rand" {
		return
	}

	parent := (*Match)(nil)
	reminder := (*Match)(nil)
	rest := []*Match{}

	for _, item := range p.Items {
		if item == nil {
			continue
		}

		if item.Kind == "parent" {
			parent = item
			continue
		}

		if item.Page != nil && isRemindersPage(item.Page.Name) {
			reminder = item
			continue
		}

		rest = append(rest, item)
	}

	out := []*Match{}

	if parent != nil {
		out = append(out, parent)
	}

	if reminder != nil {
		out = append(out, reminder)
	}

	out = append(out, rest...)

	p.Items = out
}

func (p *Picker) ensureReminderItem() {
	if p.Path != "b.rec" && p.Path != "c.rand" {
		return
	}

	for _, item := range p.Items {
		if item == nil || item.Page == nil {
			continue
		}

		if isRemindersPage(item.Page.Name) {
			return
		}
	}

	for _, page := range p.Flat {
		if page == nil {
			continue
		}

		if !isRemindersPage(page.Name) {
			continue
		}

		if page.Path != p.Path+"/reminders" {
			continue
		}

		p.Items = append(p.Items, &Match{
			Page:  page,
			Kind:  "page",
			Score: -999999,
		})

		return
	}
}

func (p *Picker) patchVisualLabel(page *filesystem.Page, label string) string {
	if page == nil {
		return label
	}

	switch page.Stage {
	case pageStageDoom:
		return label

	case pageStageDraft:
		// return "‹b " + label + " ›b"

	case pageStageEdit:
		// return "‹u " + label + " ›u"

	case pageStageGhost:
		// return "‹i " + label + " ›i"

	case pageStageId:
		// return "‹b " + label + " ›b"

	default:
		return label
	}
	return label
}

func (p *Picker) pickerLeftX() int {
	return p.Bounds.Pos[0]
}

func (p *Picker) pickerItemsW() int {
	w := p.Bounds.Size[0]

	if p.pickerUsePreview() {
		w = w / 2
	}

	if w < 4 {
		w = 4
	}

	return w
}

func (p *Picker) selectedFullItemLine(line string, width int) string {
	if width <= 0 {
		return line
	}

	visible := p.Utilities.VisibleLength(line)

	if visible > width {
		line = p.Utilities.CutVisible(line, width)
		visible = p.Utilities.VisibleLength(line)
	}

	if visible < width {
		line += strings.Repeat(" ", width-visible)
	}

	return "¤KK " + line + "¤ "
}

func pickerOwnStatusIcon(page *filesystem.Page) string {
	if page == nil {
		return ""
	}

	switch page.Stage {
	case pageStageLocal, "":
		return pickerStateLocalSymbol
	case pageStageApi:
		return pickerStateApiSymbol
	case pageStageGhost:
		return pickerStateGhostSymbol
	case pageStageDoom:
		return pickerStateDoomSymbol
	case pageStageDraft:
		return pickerStateDraftSymbol
	case pageStageEdit:
		return pickerStateEditSymbol
	case pageStageId:
		return pickerStateIdSymbol
	case pageStageConflict:
		return pickerStateConflictSymbol
	default:
		return pickerStateUnknownSymbol
	}
}

func pickerTinyCount(n int) string {

	if n < 1 {
		return ""
	}

	if n > 9 {
		return "⁎"
	}

	digits := map[rune]rune{
		'0': '⁰',
		'1': '¹',
		'2': '²',
		'3': '³',
		'4': '⁴',
		'5': '⁵',
		'6': '⁶',
		'7': '⁷',
		'8': '⁸',
		'9': '⁹',
	}

	out := ""

	for _, r := range fmt.Sprintf("%d", n) {
		if tiny, ok := digits[r]; ok {
			out += string(tiny)
		}
	}

	return out
}

type pickerRect struct {
	X int
	Y int
	W int
	H int
}

func (p *Picker) pickerUsePreview() bool {
	if p.panelMode() == "list" {
		return false
	}
	return true
}

func (p *Picker) panelMode() string {
	switch p.PanelMode {
	case "list", "preview", "both":
		return p.PanelMode
	default:
		return "both"
	}
}

func (p *Picker) pickerLayout() (pickerRect, pickerRect, pickerRect) {
	x := p.Bounds.Pos[0]
	y := p.Bounds.Pos[1]
	w := p.Bounds.Size[0]
	h := p.Bounds.Size[1]

	if w < 1 {
		w = 1
	}

	if h < 1 {
		h = 1
	}

	promptH := 1
	if h < 8 {
		promptH = 2
	}

	if promptH > h {
		promptH = h
	}

	sepH := 0

	if p.panelMode() == "preview" {
		return pickerRect{}, pickerRect{}, pickerRect{X: x, Y: y, W: w, H: h}
	}

	if p.pickerUsePreview() {
		leftW := w / 2
		if leftW < 24 {
			leftW = 24
		}

		if leftW > w-2 {
			leftW = w - 2
		}
		if leftW < 1 {
			leftW = 1
		}

		sepV := 1
		if w < 3 {
			sepV = 0
		}
		previewW := w - leftW - sepV

		if previewW < 1 {
			previewW = 1
		}

		itemsH := h - promptH - sepH
		if itemsH < 1 {
			itemsH = 1
		}

		items := pickerRect{
			X: x,
			Y: y + 1,
			W: leftW,
			H: max(1, itemsH-1),
		}

		prompt := pickerRect{
			X: x,
			Y: y + itemsH + sepH,
			W: leftW,
			H: promptH,
		}

		preview := pickerRect{
			X: x + leftW + sepV,
			Y: y,
			W: previewW,
			H: h,
		}

		return items, prompt, preview
	}

	itemsH := h - promptH - sepH
	if itemsH < 1 {
		itemsH = 1
	}

	items := pickerRect{
		X: x,
		Y: y + 1,
		W: w,
		H: max(1, itemsH-1),
	}

	prompt := pickerRect{
		X: x,
		Y: y + itemsH + sepH,
		W: w,
		H: promptH,
	}

	preview := pickerRect{}

	return items, prompt, preview
}

func (p *Picker) fitPickerLine(line string, width int) string {
	if width <= 0 {
		return ""
	}

	if p.Utilities.VisibleLength(line) > width {
		line = p.Utilities.CutVisible(line, width)
	}

	for p.Utilities.VisibleLength(line) < width {
		line += " "
	}

	return line
}

func mergeBoxChar(a rune, b rune) rune {
	if a == b {
		return a
	}

	if a == 0 {
		return b
	}

	if b == 0 {
		return a
	}

	if (a == '─' && b == '│') || (a == '│' && b == '─') {
		return '┼'
	}

	if a == '─' && b == '┤' {
		return '┤'
	}

	if a == '│' && b == '┤' {
		return '┤'
	}

	if a == '┤' || b == '┤' {
		return '┤'
	}

	return '┼'
}

func (p *Picker) drawPickerSeparators() engine.Frame {
	if len(p.Bounds.Fullsize) < 2 || len(p.Bounds.Pos) < 2 || len(p.Bounds.Size) < 2 {
		return engine.Frame{}
	}

	fullW := p.Bounds.Fullsize[0]
	fullH := p.Bounds.Fullsize[1]

	th := p.Parent.GetTheme()
	dividerFg := utilities.MixRGB(th.Background, th.Foreground, 0.28)

	cells := make([]engine.Cell, fullW*fullH)

	for i := range cells {
		cells[i] = engine.Cell{
			Char:    0,
			Fg:      &th.Foreground,
			Bg:      &th.Background,
			Visible: true,
		}
	}

	put := func(x int, y int, ch rune) {
		if x < 0 || x >= fullW || y < 0 || y >= fullH {
			return
		}

		i := y*fullW + x
		current := cells[i].Char

		if current != 0 && current != ch {
			ch = mergeBoxChar(current, ch)
		}

		cells[i] = engine.Cell{
			Char:    ch,
			Fg:      &dividerFg,
			Bg:      &th.Background,
			Visible: true,
		}
	}

	_, _, previewRect := p.pickerLayout()

	// Linha vertical entre coluna esquerda e preview.
	if p.pickerUsePreview() {
		sepX := previewRect.X - 1

		for y := p.Bounds.Pos[1]; y < p.Bounds.Pos[1]+p.Bounds.Size[1]; y++ {
			put(sepX, y, '│')
		}
	}

	return engine.Frame{
		Size:    p.Bounds.Fullsize,
		Cells:   cells,
		Timeout: 0,
	}
}
