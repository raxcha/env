package editor

import (
	"crypto/rand"
	"env/filesystem"
	"env/utilities"
	"fmt"
	"math/big"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ensuredID struct {
	Kind    string
	ID      string
	Message string
	Prefix  rune
	Header  int
}

type blockDraftData struct {
	ID          string
	Name        string
	Tags        []string
	Projects    []string
	Time        string
	Content     []string
	Hidden      bool
	Measurement string
	Owner       string
	Kind        string
	Modified    bool
	TemplateRef string
	Template    *filesystem.Page
}

func (e *Editor) ensureId() *ensuredID {
	if len(e.Content) == 0 {
		return nil
	}

	y := e.Cursor[1]

	if y < 0 {
		y = 0
	}

	if y >= len(e.Content) {
		y = len(e.Content) - 1
	}

	headerIndex := -1

	if isLogBlockBoundaryLine(e.Content[y]) {
		return nil
	}

	line := e.Content[y]
	if strings.HasPrefix(line, "--") || strings.HasPrefix(line, "~~") {
		if isIdStart(line) {
			id, rest := specialLineIDAndMessage(line)
			return &ensuredID{Kind: "special", ID: id, Message: rest, Prefix: rune(line[0])}
		}

		start, rest := splitIntoTwo(line[2:], "|")

		if len(rest) == 0 {
			rest = start
		}

		id := generateId(8)
		e.Content[y] = strings.Repeat(string(line[0]), 2) + " " + id + " | " + rest
		return &ensuredID{Kind: "special", ID: id, Message: rest, Prefix: rune(line[0])}
	}

	for i := y; i >= 0; i-- {
		if i != y && isLogBlockBoundaryLine(e.Content[i]) {
			break
		}

		if isBlockHeaderLine(e.Content[i]) {
			headerIndex = i
			break
		}

		if i != y && isEditorTimestampLine(e.Content[i]) {
			break
		}
	}

	if headerIndex >= 0 {
		id := ""
		if headerIndex+1 < len(e.Content) && isIdLine(e.Content[headerIndex+1]) {
			id = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(e.Content[headerIndex+1]), "id:"))
		} else {
			id = generateId(12)
			e.insertString(headerIndex+1, "id: "+id)
			if headerIndex+1 <= y {
				e.Cursor[1]++
			}
		}
		return &ensuredID{Kind: "block", ID: id, Prefix: rune(e.Content[headerIndex][0]), Header: headerIndex}
	}

	if y >= len(e.Content) {
		y = len(e.Content) - 1
	}

	if y < 0 {
		return nil
	}
	return nil
}

func specialLineIDAndMessage(line string) (string, string) {
	if len(line) < 2 {
		return "", strings.TrimSpace(line)
	}

	id, rest := splitIntoTwo(line[2:], "|")
	return strings.TrimSpace(id), strings.TrimSpace(rest)
}

func isEditorTimestampLine(line string) bool {
	line = strings.TrimSpace(line)

	if line == "" {
		return false
	}

	if isEditorDateTime(line) {
		return true
	}

	if isEditorDate(line) {
		return true
	}

	if isEditorClock(line) {
		return true
	}

	return false
}

func isEditorDateTime(line string) bool {
	parts := strings.Fields(line)
	if len(parts) != 2 {
		return false
	}

	return isEditorDate(parts[0]) && isEditorClock(parts[1])
}

func isEditorDate(line string) bool {
	parts := strings.Split(line, ".")
	if len(parts) != 3 {
		return false
	}

	return len(parts[0]) == 2 && len(parts[1]) == 2 && len(parts[2]) == 4 &&
		onlyDigits(parts[0]) && onlyDigits(parts[1]) && onlyDigits(parts[2])
}

func isEditorClock(line string) bool {
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return false
	}

	return len(parts[0]) == 2 && len(parts[1]) == 2 &&
		onlyDigits(parts[0]) && onlyDigits(parts[1])
}

func onlyDigits(value string) bool {
	if value == "" {
		return false
	}

	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func generateId(size int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, size)

	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			result[i] = chars[i%len(chars)]
			continue
		}

		result[i] = chars[n.Int64()]
	}

	return string(result)
}

func isIdStart(line string) bool {

	id, rest := splitIntoTwo(line[2:], "|")
	if rest == "" {
		return false
	}
	if isShortId(id) {
		return true
	}
	return false
}

func isShortId(id string) bool {
	id = strings.TrimSpace(id)

	if len(id) != 8 {
		return false
	}

	for _, r := range id {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}

		return false
	}

	return true
}

func isIdLine(line string) bool {
	line = strings.TrimSpace(line)

	const prefix = "id: "

	if !strings.HasPrefix(line, prefix) {
		return false
	}

	id := strings.TrimPrefix(line, prefix)

	if len(id) != 12 {
		return false
	}

	for _, r := range id {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}

		return false
	}

	return true
}

func isDashedLine(line string) bool {
	line = strings.TrimSpace(line)

	if line == "" {
		return false
	}

	for _, char := range line {
		if char != '-' {
			return false
		}
	}

	return true
}

func isSeparatorLine(line string) bool {
	line = strings.TrimSpace(line)

	if line == "" {
		return true
	}

	for _, char := range line {
		if char != '-' {
			return false
		}
	}

	return true
}

func isLogBlockBoundaryLine(line string) bool {
	return isSeparatorLine(line)
}

func isSpecialLine(line string) bool {
	if isDashedLine(line) {
		return false
	}

	return strings.HasPrefix(line, "--") || strings.HasPrefix(line, "~~")
}

func boldSpecialLinePrefix(line string) string {
	idx := strings.Index(line, "|")
	if idx == -1 {
		return line
	}

	return "‹b " + line[:idx] + "›b " + line[idx:]
}

func VisibleLength(str string) int {

	count := 0
	runes := []rune(str)

	for i := 0; i < len(runes); i++ {

		r := runes[i]
		if strings.ContainsRune("§¬‹›¤¶", r) {

			nextspace := findNextSpace(runes[i:])
			if nextspace == -1 {
				break
			}
			i += nextspace
			continue
		}
		if r == '¦' || r == '¶' {
			continue
		}
		count++
	}
	return count
}

func findNextSpace(runes []rune) int {

	for i, r := range runes {
		if r == ' ' {
			return i
		}
	}
	return -1
}

func RightNumber(n int, size int) string {

	return fmt.Sprintf("%*d", size, n)
}

func CutLines(height int, selected int, total int) int {
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

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func InsertCursorMarker(line string, x int) string {
	runes := []rune(line)

	if x < 0 {
		x = 0
	}

	if x > len(runes) {
		x = len(runes)
	}

	if x == len(runes) {
		return string(runes) + "¶  ¶ "
	}

	return string(runes[:x]) + "¶ " + string(runes[x]) + "¶ " + string(runes[x+1:])
}

func CursorActualIndex(line string, visualIndex int) int {
	runes := []rune(line)

	if visualIndex <= 0 {
		for i := 0; i < len(runes); {
			if span := editorMarkupSpan(runes, i); span > 0 {
				i += span
				continue
			}
			return i
		}
		return len(runes)
	}

	visible := 0
	for i := 0; i < len(runes); {
		if span := editorMarkupSpan(runes, i); span > 0 {
			i += span
			continue
		}

		if visible == visualIndex {
			return i
		}

		visible++
		i++
	}

	return len(runes)
}

func editorMarkupSpan(runes []rune, i int) int {
	if i < 0 || i >= len(runes) {
		return 0
	}

	switch runes[i] {
	case '§', '¤':
		for j := i + 1; j < len(runes); j++ {
			if runes[j] == ' ' {
				return j - i + 1
			}
		}
		return len(runes) - i
	case '¬', '‹', '›':
		j := i + 1
		for j < len(runes) {
			if runes[j] == ' ' {
				j++
				break
			}
			if !strings.ContainsRune("biuUaAv", runes[j]) {
				break
			}
			j++
		}
		return j - i
	case '¦', '¶':
		return 1
	default:
		return 0
	}
}

func splitIntoTwo(str string, where string) (string, string) {

	if where == "" {
		return strings.TrimSpace(str), ""
	}

	idxfirst := strings.Index(str, where)
	if idxfirst == -1 {
		return strings.TrimSpace(str), ""
	}

	part1 := strings.TrimSpace(str[:idxfirst])

	idxlast := idxfirst
	for idxlast < len(str) && strings.ContainsRune(where, rune(str[idxlast])) {
		idxlast += len(where)
	}

	part2 := ""
	if idxlast < len(str) {
		part2 = strings.TrimSpace(str[idxlast:])
	}

	return part1, part2
}

func splitIntoMore(str string, where string) []string {

	if where == "" {
		return strings.Fields(str)
	}

	parts := strings.FieldsFunc(str, func(r rune) bool {
		return strings.ContainsRune(where, r)
	})

	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}

	return parts
}

func closestAboveBelowBiggestRawSize(lines []string, index int, utilities *utilities.Utilities) int {
	above := 0
	below := 0

	for j := index - 1; j >= 0; j-- {
		if strings.TrimSpace(lines[j]) == "" {
			continue
		}

		if dashed.MatchString(lines[j]) {
			continue
		}

		above = utilities.VisibleLength(lines[j])
		break
	}

	for j := index + 1; j < len(lines); j++ {
		if strings.TrimSpace(lines[j]) == "" {
			continue
		}

		if dashed.MatchString(lines[j]) {
			continue
		}

		below = utilities.VisibleLength(lines[j])
		break
	}

	if above > below {
		return above
	}

	return below
}

func DashedLineVisualSize(lines []string, index int, utilities *utilities.Utilities) int {
	return closestAboveBelowBiggestRawSize(lines, index, utilities)
}

func sidebarRootPath(path string) string {
	path = strings.Trim(path, "/")

	if path == "" {
		return ""
	}

	parts := strings.Split(path, "/")
	root := parts[0]

	switch root {
	case "a.log", "b.rec", "c.rand":
		return root

	case "d.fami":
		// Se abriu só d.fami, mostra as famílias.
		if len(parts) == 1 {
			return "d.fami"
		}

		// Se abriu d.fami/silva/... mostra apenas a família atual:
		// d.fami/silva
		return strings.Join(parts[:2], "/")

	case "e.proj":
		if len(parts) == 1 {
			return "e.proj"
		}

		return strings.Join(parts[:2], "/")

	default:
		return ""
	}
}

func (e *Editor) requestSidebarPage(currentPath string) {
	if e.Parent == nil || e.Sidebar == nil {
		return
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	sidebarPath := sidebarRootPath(currentPath)
	if sidebarPath == "" || sidebarPath == "." {
		return
	}

	e.Sidebar.Path = sidebarPath
	e.Sidebar.CurrentPath = e.Path
	e.Sidebar.Sort = sidebarSortForPath(sidebarPath, e.Sidebar.Sort)

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	req.Path = sidebarPath
	req.Depth = sidebarDepthForPath(sidebarPath)
	req.Cont = false
	req.Meta = true
	req.Opts = true
	req.Sort = e.Sidebar.Sort

	fs.Load(req)

	if strings.HasPrefix(sidebarPath, "e.proj/") {
		eventsReq := filesystem.NewLoadRequest()
		eventsReq.Mode = "fresh"
		eventsReq.Path = "b.rec"
		eventsReq.Depth = 1
		eventsReq.Cont = false
		eventsReq.Meta = true
		eventsReq.Opts = true
		eventsReq.Sort = "time"

		fs.Load(eventsReq)
	}
}

func draftPageForPath(path string) *filesystem.Page {
	page := &filesystem.Page{
		Name:     filepath.Base(path),
		Path:     path,
		Type:     "shallow",
		Options:  map[string]any{},
		Content:  []string{""},
		Metadata: map[string]any{},
		Children: []*filesystem.Page{},
		Stage:    "draft",
		Sorting:  "basic",
	}

	if shouldNewPageBeDeep(path) {
		page.Type = "deep"
	}

	return page
}

func shouldNewPageBeDeep(path string) bool {
	if path == "d.fami" || path == "e.proj" {
		return true
	}

	// Para famílias/projetos novos, cria como pasta com index.
	if strings.HasPrefix(path, "d.fami/") {
		return true
	}

	if strings.HasPrefix(path, "e.proj/") {
		return true
	}

	return false
}

func ensureEditableContent(content []string) []string {
	if len(content) == 0 {
		return []string{""}
	}

	return content
}

func copyLines(lines []string) []string {
	return append([]string(nil), lines...)
}

func sameLines(a []string, b []string) bool {
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

func pageForSingleSync(page *filesystem.Page) *filesystem.Page {
	if page == nil {
		return &filesystem.Page{}
	}

	copy := *page
	copy.Content = copyLines(page.Content)
	copy.Children = nil
	return &copy
}

func (e *Editor) applyEnsuredID(ensured *ensuredID) {
	if ensured == nil || ensured.ID == "" || e == nil || e.Parent == nil {
		return
	}

	switch ensured.Kind {
	case "block":
		e.ensureDedicatedIDPage(ensured)
	case "special":
		e.ensureReminderEntry(ensured)
	}
}

func (e *Editor) logBlockAutomationAllowed() bool {
	if e == nil {
		return false
	}

	path := strings.TrimSpace(e.Path)
	if e.Page != nil && strings.TrimSpace(e.Page.Path) != "" {
		path = strings.TrimSpace(e.Page.Path)
	}

	path = filepath.ToSlash(filepath.Clean(path))
	path = strings.TrimPrefix(path, "prsnl.spc/")
	path = strings.TrimPrefix(path, "/home/asdf/prsnl.spc/")

	return strings.HasPrefix(path, "a.log/")
}

func (e *Editor) ensureDedicatedIDPage(ensured *ensuredID) {
	base := dedicatedBaseForBlockPrefix(ensured.Prefix)
	if base == "" {
		return
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	existing := findDedicatedPageByID(fs.Cache, ensured.ID)
	title, content, ok := e.dedicatedBlockDraft(ensured, existing)
	if !ok {
		return
	}

	path := base + "/" + title
	e.ensureIDDraftAtPath(ensured.ID, path, content)
}

func (e *Editor) refreshCurrentBlockIDPage(id string) {
	if e == nil || id == "" || len(e.Cursor) < 2 {
		return
	}

	y := e.Cursor[1]
	if y <= 0 || y >= len(e.Content) {
		return
	}

	if !isIdLine(e.Content[y]) || !isBlockHeaderLine(e.Content[y-1]) {
		return
	}

	lineID := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(e.Content[y]), "id:"))
	if lineID != id {
		return
	}

	e.ensureDedicatedIDPage(&ensuredID{
		Kind:   "block",
		ID:     id,
		Prefix: rune(e.Content[y-1][0]),
		Header: y - 1,
	})
}

func (e *Editor) ensureIDDraftAtPath(id string, path string, content []string) *filesystem.Page {
	if e == nil || e.Parent == nil || id == "" || path == "" {
		return nil
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return nil
	}

	page := findDedicatedPageByID(fs.Cache, id)
	if page != nil {
		if filepath.Clean(page.Path) != filepath.Clean(path) {
			_, page = fs.EditPath(page, path)
		}

		if page != nil {
			fs.EditContent(page, copyLines(content))
		}
		return page
	}

	return e.ensureDraftOrSync(path, content)
}

func (e *Editor) dedicatedBlockDraft(ensured *ensuredID, existing *filesystem.Page) (string, []string, bool) {
	if ensured == nil || ensured.Header < 0 || ensured.Header >= len(e.Content) {
		return "", nil, false
	}

	header := e.Content[ensured.Header]
	title := dedicatedBlockTitle(header)
	if title == "" {
		return "", nil, false
	}

	if !e.blockHasTextAfterID(ensured.Header) {
		return "", nil, false
	}

	data := e.blockDraftData(ensured, header, existing)
	if data.Template != nil {
		content := ensureTemplateRefLine(applyBlockTemplate(data.Template.Content, data), data.TemplateRef)
		return title, preserveModifiedBody(content, existing, data.Modified), true
	}

	metadata := []string{
		"id: " + data.ID,
		"name: " + data.Name,
	}
	if data.Kind != "" {
		metadata = append(metadata, "kind: "+data.Kind)
	}
	metadata = append(metadata,
		"time: "+data.Time,
		"owner: "+data.Owner,
		"tags: "+strings.Join(data.Tags, ", "),
		"projects: "+strings.Join(data.Projects, ", "),
		"measurement: "+data.Measurement,
		"hidden: "+yesNo(data.Hidden),
		"modified: "+yesNo(data.Modified),
		"---",
	)

	return title, preserveModifiedBody(append(metadata, data.Content...), existing, data.Modified), true
}

func (e *Editor) blockDraftData(ensured *ensuredID, header string, existing *filesystem.Page) blockDraftData {
	hidden := blockHeaderHidden(header)
	content := []string{}
	if !hidden {
		content = e.blockBodyAfterID(ensured.Header)
	}
	tags, projects := extractInlineTagsAndRefs(append([]string{header}, content...))

	owner := firstOwnerFromHeader(header)
	kind := blockKindForPrefix(ensured.Prefix)
	templateRef := blockTemplateRefFromPage(existing)
	if templateRef == "" {
		templateRef = blockDefaultTemplateRef(kind)
	}
	modified := blockModifiedFromPage(existing)
	measurement := blockBodyMeasurement(content)
	if measurement == "" {
		measurement = blockMeasurement(header)
	}
	return blockDraftData{
		ID:          ensured.ID,
		Name:        blockNameFromHeader(header),
		Tags:        tags,
		Projects:    projects,
		Time:        e.releaseTimeAboveCursor(),
		Content:     content,
		Hidden:      hidden,
		Measurement: measurement,
		Owner:       owner,
		Kind:        kind,
		Modified:    modified,
		TemplateRef: templateRef,
		Template:    e.blockTemplate(templateRef),
	}
}

func dedicatedBlockTitle(header string) string {
	if strings.TrimSpace(header) == "" {
		return ""
	}

	prefix := string([]rune(header)[0])
	name := blockNameFromHeader(header)
	if name == "" {
		return ""
	}

	tags, refs := extractInlineTagsAndRefs([]string{header})
	ref := ""
	if prefix == "`" {
		if len(tags) == 0 {
			return ""
		}
		ref = "#" + firstPathPart(tags[0])
	} else {
		if len(refs) == 0 {
			return ""
		}
		ref = "@" + firstPathPart(refs[0])
	}

	return prefix + name + ref
}

func blockNameFromHeader(header string) string {
	fields := strings.Fields(header)
	if len(fields) < 2 {
		return ""
	}

	start := 1
	if fields[start] == "!" {
		start++
	}

	parts := []string{}
	for _, field := range fields[start:] {
		if strings.HasPrefix(field, "@") || strings.HasPrefix(field, "#") {
			break
		}
		parts = append(parts, slugPart(field))
	}

	return strings.Join(nonEmptyStrings(parts), "-")
}

func blockHeaderHidden(header string) bool {
	fields := strings.Fields(header)
	return len(fields) > 1 && fields[1] == "!"
}

func blockKindForPrefix(prefix rune) string {
	switch prefix {
	case '>':
		return "plan"
	case '=':
		return "session"
	case '~':
		return "task"
	case '`':
		return "idea"
	default:
		return ""
	}
}

func blockDefaultTemplateRef(kind string) string {
	switch strings.TrimSpace(kind) {
	case "plan", "session", "task", "idea":
		return "e.proj/.templates/" + kind
	default:
		return ""
	}
}

func blockTemplateRefFromPage(page *filesystem.Page) string {
	if page == nil {
		return ""
	}

	return metadataValue(page.Content, "template")
}

func blockModifiedFromPage(page *filesystem.Page) bool {
	if page == nil {
		return false
	}

	return strings.EqualFold(metadataValue(page.Content, "modified"), "yes")
}

func metadataValue(content []string, key string) string {
	prefix := key + ":"
	for _, line := range content {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		}
	}
	return ""
}

func (e *Editor) blockTemplate(ref string) *filesystem.Page {
	if e == nil || e.Parent == nil {
		return nil
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return nil
	}

	path := templatePathFromRef(ref)
	if path == "" {
		return nil
	}

	page := fs.Find(path)
	if page != nil && len(page.Content) > 0 {
		return page
	}

	if strings.HasPrefix(path, "e.proj/.templates/") {
		return builtinBlockTemplate(filepath.Base(path))
	}

	return page
}

func builtinBlockTemplate(kind string) *filesystem.Page {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		kind = "session"
	}

	return &filesystem.Page{
		Name: kind,
		Path: "e.proj/.templates/" + kind,
		Type: "shallow",
		Content: []string{
			"id: {id}",
			"name: {name}",
			"kind: {kind}",
			"release-time: {time}",
			"owner: {owner}",
			"tags: {tags}",
			"projects: {projects}",
			"measurement: {measurement}",
			"hidden: {yes/no}",
			"modified: {yes/no}",
			"---",
			"{content}",
		},
		Metadata: map[string]any{"name": kind},
	}
}

func templatePathFromRef(ref string) string {
	ref = strings.TrimSpace(ref)
	ref = strings.Trim(ref, `"'.,;!?()[]{}<>`)
	ref = strings.Trim(ref, "/")
	if ref == "" {
		return ""
	}

	if strings.HasPrefix(ref, "@") {
		parts := strings.Split(strings.Trim(strings.TrimPrefix(ref, "@"), "/"), "/")
		if len(parts) >= 2 && parts[0] != "" && parts[1] != "" {
			return "e.proj/" + parts[0] + "/" + parts[1]
		}
		return ""
	}

	if strings.HasPrefix(ref, "#") {
		parts := strings.Split(strings.Trim(strings.TrimPrefix(ref, "#"), "/"), "/")
		if len(parts) >= 2 && parts[0] != "" && parts[1] != "" {
			return "d.fami/" + parts[0] + "/" + parts[1]
		}
		return ""
	}

	if strings.HasPrefix(ref, "e.proj/") || strings.HasPrefix(ref, "d.fami/") {
		return ref
	}

	return ""
}

func firstOwnerFromHeader(header string) string {
	for _, field := range strings.Fields(header) {
		field = strings.Trim(field, `"'.,;:!?()[]{}<>`)
		if len(field) < 2 {
			continue
		}

		if strings.HasPrefix(field, "@") {
			return "@" + firstPathPart(strings.TrimPrefix(field, "@"))
		}
		if strings.HasPrefix(field, "#") {
			return "#" + firstPathPart(strings.TrimPrefix(field, "#"))
		}
	}

	return ""
}

func blockMeasurement(header string) string {
	for _, field := range strings.Fields(header) {
		field = strings.Trim(field, `"'.,;:!?()[]{}<>`)
		if strings.HasPrefix(field, "&") && len(field) > 1 {
			return strings.TrimPrefix(field, "&")
		}
	}

	return ""
}

func blockBodyMeasurement(content []string) string {
	re := regexp.MustCompile(`(?:^|[^\w])!(\d+)(?:$|[^\w])`)
	for _, line := range content {
		match := re.FindStringSubmatch(line)
		if len(match) >= 2 {
			return match[1]
		}
	}

	return ""
}

func applyBlockTemplate(template []string, data blockDraftData) []string {
	out := []string{}
	for _, line := range template {
		if strings.Contains(line, "{content}") || strings.TrimSpace(line) == "{content" {
			out = append(out, data.Content...)
			continue
		}

		out = append(out, replaceBlockTemplatePlaceholders(line, data))
	}

	return out
}

func ensureTemplateRefLine(content []string, ref string) []string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return content
	}

	line := "template: " + ref
	for i, current := range content {
		trimmed := strings.TrimSpace(current)
		if trimmed == "---" {
			break
		}
		if strings.HasPrefix(trimmed, "template:") {
			content[i] = line
			return content
		}
	}

	for i, current := range content {
		trimmed := strings.TrimSpace(current)
		if trimmed == "---" {
			break
		}
		if strings.HasPrefix(trimmed, "id:") {
			return insertLineCopy(content, i+1, line)
		}
	}

	return append([]string{line}, content...)
}

func preserveModifiedBody(content []string, existing *filesystem.Page, modified bool) []string {
	if !modified || existing == nil {
		return content
	}

	_, existingBody, ok := splitMetadataBody(existing.Content)
	if !ok {
		return content
	}

	header, _, ok := splitMetadataBody(content)
	if !ok {
		return content
	}

	out := copyLines(header)
	out = append(out, existingBody...)
	return out
}

func splitMetadataBody(content []string) ([]string, []string, bool) {
	for i, line := range content {
		if strings.TrimSpace(line) == "---" {
			header := copyLines(content[:i+1])
			body := copyLines(content[i+1:])
			return header, body, true
		}
	}

	return copyLines(content), []string{}, false
}

func insertLineCopy(content []string, index int, line string) []string {
	out := make([]string, 0, len(content)+1)
	out = append(out, content[:index]...)
	out = append(out, line)
	out = append(out, content[index:]...)
	return out
}

func replaceBlockTemplatePlaceholders(line string, data blockDraftData) string {
	yesNoValue := yesNo(data.Hidden)
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "modified:") {
		yesNoValue = yesNo(data.Modified)
	}

	replacements := map[string]string{
		"{id}":          data.ID,
		"{name}":        data.Name,
		"{tags}":        strings.Join(data.Tags, ", "),
		"{projects}":    strings.Join(data.Projects, ", "),
		"{time}":        data.Time,
		"{hidden}":      yesNo(data.Hidden),
		"{measurement}": data.Measurement,
		"{owner}":       data.Owner,
		"{kind}":        data.Kind,
		"{modified}":    yesNo(data.Modified),
		"{yes/no}":      yesNoValue,
	}

	for from, to := range replacements {
		line = strings.ReplaceAll(line, from, to)
	}

	return line
}

func extractInlineTagsAndRefs(lines []string) (tags []string, refs []string) {
	seenTags := map[string]bool{}
	seenRefs := map[string]bool{}

	for _, line := range lines {
		for _, field := range strings.Fields(line) {
			field = cleanInlineMarker(field)
			if len(field) < 2 {
				continue
			}

			if strings.HasPrefix(field, "#") {
				tag := strings.TrimPrefix(field, "#")
				if tag != "" && !seenTags[tag] {
					tags = append(tags, tag)
					seenTags[tag] = true
				}
			}
			if strings.HasPrefix(field, "@") {
				ref := strings.TrimPrefix(field, "@")
				if ref != "" && !seenRefs[ref] {
					refs = append(refs, ref)
					seenRefs[ref] = true
				}
			}
		}
	}
	return tags, refs
}

func cleanInlineMarker(value string) string {
	return strings.Trim(value, `"'.,;:!?()[]{}<>`)
}

func firstPathPart(value string) string {
	value = strings.Trim(value, "/")
	if value == "" {
		return ""
	}
	return strings.Split(value, "/")[0]
}

func slugPart(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.Trim(value, `"'.,;:!?()[]{}<>`)
	value = strings.ReplaceAll(value, "_", "-")
	return value
}

func nonEmptyStrings(values []string) []string {
	out := []string{}
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func (e *Editor) blockBodyAfterID(header int) []string {
	start := header + 2
	if start > len(e.Content) {
		return []string{}
	}

	end := len(e.Content)
	for i := start; i < len(e.Content); i++ {
		if i > start && isLogBlockBoundaryLine(e.Content[i]) {
			end = i
			break
		}
		if i > start && isBlockHeaderLine(e.Content[i]) {
			end = i
			break
		}
	}

	return copyLines(e.Content[start:end])
}

func (e *Editor) blockHasTextAfterID(header int) bool {
	for _, line := range e.blockBodyAfterID(header) {
		if strings.TrimSpace(line) != "" {
			return true
		}
	}

	return false
}

func (e *Editor) releaseTimeAboveCursor() string {
	if e == nil || len(e.Cursor) < 2 {
		return time.Now().Format("02.01.2006 15:04")
	}

	limit := e.Cursor[1]
	if limit < 0 {
		limit = 0
	}
	if limit > len(e.Content) {
		limit = len(e.Content)
	}

	return latestTimestamp(e.Content[:limit])
}

func latestTimestamp(lines []string) string {
	clock := ""
	for i := len(lines) - 1; i >= 0; i-- {
		fields := strings.Fields(lines[i])
		for _, field := range fields {
			field = strings.Trim(field, `"'.,;:!?()[]{}<>`)
			if isEditorDate(field) {
				if clock == "" {
					clock = "00:00"
				}
				return field + " " + clock
			}
			if clock == "" && isEditorClock(field) {
				clock = field
			}
		}
	}

	return time.Now().Format("02.01.2006 15:04")
}

func dedicatedBaseForBlockPrefix(prefix rune) string {
	switch prefix {
	case '~', '>', '=':
		return "b.rec"
	case '`':
		return "c.rand"
	default:
		return ""
	}
}

func (e *Editor) ensureReminderEntry(ensured *ensuredID) {
	base := ""
	switch ensured.Prefix {
	case '~':
		base = "b.rec"
	case '-':
		base = "c.rand"
	default:
		return
	}

	path := base + "/.reminders"
	page := e.ensureDraftOrSync(path, []string{})
	if page == nil {
		return
	}

	page.Content = ensureLeadingBlankLine(page.Content)

	tag := "#" + ensured.ID
	if reminderEntryExists(page.Content, tag) {
		e.syncPageOnly(page)
		return
	}

	message := strings.TrimSpace(ensured.Message)
	if message == "" {
		message = tag
	}

	marker := "--"
	if ensured.Prefix == '~' {
		marker = "~~"
	}

	entry := []string{marker + " " + message, tag, ""}
	page.Content = append([]string{""}, append(entry, page.Content[1:]...)...)
	if page.Stage != "draft" {
		e.Parent.GetFilesystem().EditContent(page, copyLines(page.Content))
	}
}

func ensureLeadingBlankLine(content []string) []string {
	if len(content) > 0 && strings.TrimSpace(content[0]) == "" {
		return content
	}

	return append([]string{""}, content...)
}

func reminderEntryExists(content []string, tag string) bool {
	for _, line := range content {
		if strings.TrimSpace(line) == tag {
			return true
		}
	}

	return false
}

func (e *Editor) ensureDraftOrSync(path string, content []string) *filesystem.Page {
	if e == nil || e.Parent == nil || strings.TrimSpace(path) == "" {
		return nil
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return nil
	}

	page := fs.Find(path)
	if page != nil {
		if page.Stage == "" || page.Stage == "ghost" {
			page.Stage = "draft"
			page.Name = filepath.Base(path)
			if page.Type == "" {
				page.Type = "shallow"
			}
			page.Content = copyLines(content)
			return page
		}

		e.syncPageOnly(page)
		return page
	}

	_, page = fs.NewDraft(path)
	if page == nil {
		return nil
	}

	page.Content = copyLines(content)
	if len(page.Content) == 0 {
		page.Content = []string{}
	}
	return page
}

func (e *Editor) syncPageOnly(page *filesystem.Page) {
	if e == nil || e.Parent == nil || page == nil {
		return
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	req := filesystem.NewSyncRequest()
	req.Branch = page
	req.PageOnly = true
	fs.Sync(req)
}

func (e *Editor) headerLine(width int, current bool, cursorX int) string {
	path := e.Path
	if e.Page != nil && e.Page.Path != "" {
		path = e.Page.Path
	}
	if current {
		path = e.Path
	}

	stage := ""
	if e.Page != nil {
		stage = e.Page.Stage
	}

	displayPath := relativePersonalPath(path)
	if current {
		actual := CursorActualIndex(displayPath, cursorX)
		displayPath = InsertCursorMarker(displayPath, actual)
	}

	label := stageSymbol(stage) + " " + displayPath
	if width > 0 && !current {
		label = fitHeaderLabel(label, width)
	}

	style := "§8B0 "
	if current {
		style = "§BA999 "
	}

	return style + "‹b " + label + "›b "
}

func stageSymbol(stage string) string {
	switch strings.TrimSpace(stage) {
	case "draft":
		return "󰦨"
	case "edited":
		return "󱩽"
	case "local":
		return "󰓦"
	case "api":
		return "󱂛"
	case "doomed":
		return "󱂥"
	case "ghost":
		return "󰊠"
	default:
		return "•"
	}
}

func relativePersonalPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "." {
		return "prsnl.spc"
	}

	return strings.TrimPrefix(path, "prsnl.spc/")
}

func fitHeaderLabel(label string, width int) string {
	if width <= 0 || runeLen(label) <= width {
		return label
	}

	return cutWithMiddle(label, width)
}

func cutWithMiddle(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}

	if width <= 3 {
		return string(runes[:width])
	}

	left := (width - 3) / 2
	right := width - 3 - left
	return string(runes[:left]) + "..." + string(runes[len(runes)-right:])
}

func runeLen(value string) int {
	return len([]rune(value))
}

func sidebarDepthForPath(path string) int {
	path = strings.Trim(path, "/")

	if path == "a.log" || path == "b.rec" || path == "c.rand" {
		return 1
	}

	if strings.HasPrefix(path, "d.fami/") {
		return -1
	}

	if path == "d.fami" {
		return 1
	}

	if strings.HasPrefix(path, "e.proj") {
		return -1
	}

	return 1
}

func addLastEditedTime(content []string) []string {
	now := time.Now().Format("02.01.2006 15:04")
	key := "last-edited-time:"
	for i, line := range content {
		if strings.HasPrefix(strings.TrimSpace(line), key) {
			content[i] = key + " " + now
			return content
		}
		if strings.TrimSpace(line) == "---" {
			break
		}
	}
	return content
}

func resolveIDInCache(page *filesystem.Page, id string) (bool, string) {
	if page == nil || id == "" {
		return false, ""
	}
	for _, line := range page.Content {
		if isIdLine(line) && strings.Contains(line, id) {
			return true, page.Path
		}
	}
	for _, child := range page.Children {
		if ok, path := resolveIDInCache(child, id); ok {
			return true, path
		}
	}
	return false, ""
}

func dedicatedOrReminderPathForID(root *filesystem.Page, id string) string {
	if page := findDedicatedPageByID(root, id); page != nil {
		return page.Path
	}

	for _, path := range []string{"b.rec/" + id, "c.rand/" + id} {
		if findPageByPath(root, path) != nil {
			return path
		}
	}

	for _, path := range []string{"b.rec/.reminders", "c.rand/.reminders"} {
		page := findPageByPath(root, path)
		if page == nil {
			continue
		}

		if reminderEntryExists(page.Content, "#"+id) {
			return path
		}
	}

	return ""
}

func findDedicatedPageByID(root *filesystem.Page, id string) *filesystem.Page {
	if root == nil || id == "" {
		return nil
	}

	if isDedicatedIDPage(root, id) {
		return root
	}

	for _, child := range root.Children {
		if found := findDedicatedPageByID(child, id); found != nil {
			return found
		}
	}

	return nil
}

func isDedicatedIDPage(page *filesystem.Page, id string) bool {
	if page == nil {
		return false
	}

	if !strings.HasPrefix(page.Path, "b.rec/") && !strings.HasPrefix(page.Path, "c.rand/") {
		return false
	}

	if strings.HasSuffix(page.Path, "/.reminders") || page.Path == "b.rec/.reminders" || page.Path == "c.rand/.reminders" {
		return false
	}

	for _, line := range page.Content {
		if strings.TrimSpace(line) == "id: "+id {
			return true
		}
	}

	return false
}

func (e *Editor) currentIDContextPath(id string) string {
	if e == nil || id == "" || len(e.Cursor) < 2 {
		return ""
	}

	y := e.Cursor[1]
	if y < 0 || y >= len(e.Content) {
		return ""
	}

	line := e.Content[y]
	if isSpecialLine(line) {
		prefix := rune(line[0])
		switch prefix {
		case '~':
			return "b.rec/.reminders"
		case '-':
			return "c.rand/.reminders"
		}
	}

	header := e.blockHeaderAbove(y)
	if header < 0 {
		return ""
	}

	title := dedicatedBlockTitle(e.Content[header])
	if title == "" {
		return ""
	}

	base := dedicatedBaseForBlockPrefix(rune(e.Content[header][0]))
	if base == "" {
		return ""
	}

	return base + "/" + title
}

func (e *Editor) blockHeaderAbove(y int) int {
	if e == nil {
		return -1
	}

	if y >= len(e.Content) {
		y = len(e.Content) - 1
	}

	for i := y; i >= 0; i-- {
		if i != y && isLogBlockBoundaryLine(e.Content[i]) {
			return -1
		}

		if isBlockHeaderLine(e.Content[i]) {
			return i
		}

		if i != y && isEditorTimestampLine(e.Content[i]) {
			return -1
		}
	}

	return -1
}

func findPageByPath(page *filesystem.Page, path string) *filesystem.Page {
	if page == nil {
		return nil
	}

	if filepath.Clean(page.Path) == filepath.Clean(path) {
		return page
	}

	for _, child := range page.Children {
		if found := findPageByPath(child, path); found != nil {
			return found
		}
	}

	return nil
}

func sidebarSortForPath(sidebarPath string, fallback string) string {
	root := strings.Split(strings.Trim(sidebarPath, "/"), "/")[0]

	switch root {
	case "d.fami", "e.proj":
		return "basic"

	default:
		if fallback == "" {
			return "basic"
		}
		return fallback
	}
}
