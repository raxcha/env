package projects

import (
	"env/filesystem"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (p *Projects) prependBackItem() {
	if p.Context.Kind == "root" {
		return
	}

	backPath := "e.proj"

	switch p.Context.Kind {
	case "templates", "events":
		backPath = p.Context.ProjectPath
	case "project":
		backPath = "e.proj"
	}

	p.Items = append([]*ProjectItem{
		{
			Kind: "back",
			Name: "...",
			Path: backPath,
		},
	}, p.Items...)
}

func (p *Projects) updateItems() {
	p.Items = []*ProjectItem{}

	switch p.Context.Kind {
	case "project":
		p.startProjectItems()
	case "templates":
		p.startTemplateItems()
	case "events":
		p.startEventItems()
	default:
		p.Context = ProjectContext{Kind: "root"}
		p.startRootItems()
	}

	p.prependBackItem()

	if p.Selected < 0 {
		p.Selected = 0
	}

	if p.Selected >= len(p.Items) {
		p.Selected = len(p.Items) - 1
	}

	if p.Selected < 0 {
		p.Selected = 0
	}
}

func (p *Projects) startRootItems() {
	p.refreshProjectPriorities()

	p.Items = append(p.Items, &ProjectItem{
		Kind: "new-project",
		Name: "new project",
		Path: p.newProjectPath(),
	})

	root := p.projectsRoot()
	if root == nil {
		return
	}

	projects := append([]*filesystem.Page{}, root.Children...)
	sort.SliceStable(projects, func(i, j int) bool {
		ap := metadataInt(projects[i], "priority")
		bp := metadataInt(projects[j], "priority")

		if ap != bp {
			return ap > bp
		}

		return strings.ToLower(projectDisplayName(projects[i])) < strings.ToLower(projectDisplayName(projects[j]))
	})

	for _, project := range projects {
		if project == nil {
			continue
		}
		if strings.HasPrefix(project.Name, ".") {
			continue
		}

		p.Items = append(p.Items, &ProjectItem{
			Page:     project,
			Kind:     "project",
			Name:     projectDisplayName(project),
			Path:     project.Path,
			Priority: metadataInt(project, "priority"),
			Depth:    0,
		})
	}
}

func (p *Projects) startProjectItems() {
	p.Items = append(p.Items,
		&ProjectItem{
			Kind: "virtual-templates",
			Name: "templates",
			Path: filepath.Join(p.Context.ProjectPath, ".templates"),
		},
		&ProjectItem{
			Kind: "virtual-events",
			Name: "events",
			Path: filepath.Join(p.Context.ProjectPath, ".events"),
		},
	)
}

func (p *Projects) startTemplateItems() {
	project := p.currentProject()
	if project == nil {
		return
	}

	p.Items = append(p.Items, &ProjectItem{
		Kind: "add-template",
		Name: "new template",
		Path: p.newTemplatePath(),
	})

	templatesRoot := p.templatesRoot(project)
	templates := []*filesystem.Page{}
	if templatesRoot != nil {
		templates = append(templates, templatesRoot.Children...)
	} else {
		templates = append(templates, legacyProjectTemplates(project)...)
	}

	sort.SliceStable(templates, func(i, j int) bool {
		return strings.ToLower(templates[i].Name) < strings.ToLower(templates[j].Name)
	})

	for _, tmpl := range templates {
		if tmpl == nil {
			continue
		}
		if strings.TrimSpace(tmpl.Name) == "index" || strings.TrimSpace(tmpl.Name) == ".templates" {
			continue
		}

		p.Items = append(p.Items, &ProjectItem{
			Page:  tmpl,
			Kind:  "template",
			Name:  templateName(tmpl),
			Path:  tmpl.Path,
			Depth: 0,
		})
	}
}
func (p *Projects) startEventItems() {
	project := p.currentProject()
	if project == nil {
		return
	}

	events := p.filteredProjectEvents()

	for _, event := range events {
		if event == nil {
			continue
		}

		p.Items = append(p.Items, &ProjectItem{
			Page:  event,
			Kind:  "event",
			Name:  eventDisplayName(event),
			Path:  event.Path,
			Depth: 0,
		})
	}
}

func (p *Projects) projectsRoot() *filesystem.Page {
	if p.ProjectsPage != nil {
		return p.ProjectsPage
	}

	if p.Parent == nil || p.Parent.GetFilesystem() == nil || p.Parent.GetFilesystem().Cache == nil {
		return nil
	}

	return findProjectPage(p.Parent.GetFilesystem().Cache, "e.proj")
}

func (p *Projects) currentProject() *filesystem.Page {
	root := p.projectsRoot()
	if root == nil || p.Context.ProjectPath == "" {
		return nil
	}

	return findProjectPage(root, p.Context.ProjectPath)
}

func findProjectPage(page *filesystem.Page, path string) *filesystem.Page {
	if page == nil {
		return nil
	}

	if filepath.Clean(page.Path) == filepath.Clean(path) {
		return page
	}

	for _, child := range page.Children {
		found := findProjectPage(child, path)
		if found != nil {
			return found
		}
	}

	return nil
}

func (p *Projects) eventTypeFilter() string {
	filter := strings.TrimSpace(p.Prompt)

	switch filter {
	case "=", ">", "~":
		return filter
	default:
		return ""
	}
}

func eventMatchesType(page *filesystem.Page, filter string) bool {
	if filter == "" {
		return true
	}

	if page == nil {
		return false
	}

	name := strings.TrimSpace(page.Name)
	if strings.HasPrefix(name, filter) {
		return true
	}

	if len(page.Content) > 0 {
		firstLine := strings.TrimSpace(page.Content[0])
		if strings.HasPrefix(firstLine, filter) {
			return true
		}
	}

	return false
}

func (p *Projects) filteredProjectEvents() []*filesystem.Page {
	if p.EventsPage == nil || strings.TrimSpace(p.Context.ProjectName) == "" {
		return []*filesystem.Page{}
	}

	events := []*filesystem.Page{}
	projectName := strings.TrimSpace(p.Context.ProjectName)
	typeFilter := p.eventTypeFilter()

	for _, child := range p.EventsPage.Children {
		if child == nil || child.Metadata == nil {
			continue
		}

		if !projectOwnsEvent(child, projectName) {
			continue
		}

		if !eventMatchesType(child, typeFilter) {
			continue
		}

		events = append(events, child)
	}

	sort.SliceStable(events, func(i, j int) bool {
		at := eventSortTime(events[i])
		bt := eventSortTime(events[j])

		aHasTime := !at.IsZero()
		bHasTime := !bt.IsZero()

		if aHasTime && bHasTime && !at.Equal(bt) {
			return at.After(bt)
		}

		if aHasTime != bHasTime {
			return aHasTime
		}

		return strings.ToLower(events[i].Name) < strings.ToLower(events[j].Name)
	})

	return events
}

func projectDisplayName(page *filesystem.Page) string {
	if page == nil {
		return ""
	}

	if page.Metadata != nil {
		if name, ok := page.Metadata["name"].(string); ok && strings.TrimSpace(name) != "" {
			return strings.TrimSpace(name)
		}
	}

	if strings.TrimSpace(page.Name) != "" {
		return page.Name
	}

	return filepath.Base(page.Path)
}

func templateName(page *filesystem.Page) string {
	if page == nil {
		return ""
	}

	if page.Metadata != nil {
		if name, ok := page.Metadata["name"].(string); ok && usableProjectDisplayName(name) {
			return strings.TrimSpace(name)
		}
	}

	return page.Name
}

func usableProjectDisplayName(name string) bool {
	name = strings.TrimSpace(name)
	return name != "" && !strings.Contains(name, "{") && !strings.Contains(name, "}")
}

func eventDisplayName(page *filesystem.Page) string {
	if page == nil {
		return ""
	}

	if strings.TrimSpace(page.Name) != "" {
		return page.Name
	}

	return filepath.Base(page.Path)
}

func (p *Projects) templatesRoot(project *filesystem.Page) *filesystem.Page {
	if project == nil {
		return nil
	}

	for _, child := range project.Children {
		if child == nil {
			continue
		}
		if strings.TrimSpace(child.Name) == ".templates" || filepath.Base(child.Path) == ".templates" {
			return child
		}
	}

	return nil
}

func legacyProjectTemplates(project *filesystem.Page) []*filesystem.Page {
	if project == nil {
		return []*filesystem.Page{}
	}

	out := []*filesystem.Page{}
	for _, child := range project.Children {
		if child == nil {
			continue
		}
		name := strings.TrimSpace(child.Name)
		if name == "" || name == "index" || strings.HasPrefix(name, ".") {
			continue
		}
		out = append(out, child)
	}

	return out
}

func projectOwnsEvent(page *filesystem.Page, projectName string) bool {
	if page == nil || page.Metadata == nil {
		return false
	}

	projectName = projectDashedName(projectName)
	if projectName == "" {
		return false
	}

	keys := []string{"owned-by", "owner", "project", "projects"}
	for _, key := range keys {
		raw, ok := page.Metadata[key]
		if !ok {
			continue
		}

		for _, candidate := range strings.FieldsFunc(fmt.Sprint(raw), func(r rune) bool {
			return r == ',' || r == ';' || r == ' '
		}) {
			if projectDashedName(candidate) == projectName {
				return true
			}
		}
	}

	return false
}

func projectDashedName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, " ", "-")
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	return strings.Trim(name, "-")
}

func metadataInt(page *filesystem.Page, key string) int {
	if page == nil || page.Metadata == nil {
		return 0
	}

	value := page.Metadata[key]

	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(v))
		return n
	default:
		return 0
	}
}

func (p *Projects) selectedItem() *ProjectItem {
	if len(p.Items) == 0 || p.Selected < 0 || p.Selected >= len(p.Items) {
		return nil
	}

	return p.Items[p.Selected]
}

func (p *Projects) selectedPath() string {
	item := p.selectedItem()
	if item == nil {
		return ""
	}

	return item.Path
}

func (p *Projects) selectPath(path string) bool {
	if path == "" {
		return false
	}

	for i, item := range p.Items {
		if item != nil && filepath.Clean(item.Path) == filepath.Clean(path) {
			p.Selected = i
			return true
		}
	}

	return false
}

func (p *Projects) contextLabel() string {
	switch p.Context.Kind {
	case "project":
		return "projects:" + p.Context.ProjectName
	case "templates":
		return "projects:" + p.Context.ProjectName + "/.templates"
	case "events":
		return "projects:" + p.Context.ProjectName + "/.events"
	default:
		return "projects"
	}
}

func (p *Projects) lineForItem(item *ProjectItem, width int) string {
	if item == nil {
		return ""
	}

	if item.Kind == "back" || item.Name == "..." {
		return " ..."
	}

	left := p.projectItemLabel(item)
	right := p.projectItemRightText(item)

	leftVisible := p.Utilities.VisibleLength(left)
	rightVisible := p.Utilities.VisibleLength(right)

	trailingSpace := 0
	if rightVisible > 0 && width > 0 {
		trailingSpace = 1
	}

	gap := width - leftVisible - rightVisible - trailingSpace

	if gap < 1 && rightVisible > 0 {
		maxLeft := width - rightVisible - trailingSpace - 1
		if maxLeft < 0 {
			maxLeft = 0
		}

		left = p.truncateStyledLabel(left, maxLeft)
		leftVisible = p.Utilities.VisibleLength(left)
		gap = width - leftVisible - rightVisible - trailingSpace
	}

	if gap < 0 {
		gap = 0
	}

	if rightVisible == 0 {
		if p.Utilities.VisibleLength(left) > width {
			left = p.truncateStyledLabel(left, width)
		}

		return left
	}

	return left + strings.Repeat(" ", gap) + right + strings.Repeat(" ", trailingSpace)
}

func eventTimePrefix(page *filesystem.Page) string {
	if page == nil || page.Metadata == nil {
		return ""
	}

	if t, ok := page.Metadata["time"].(time.Time); ok && !t.IsZero() {
		return t.Format("02.01 15:04") + "  "
	}

	return ""
}

func (p *Projects) truncate(text string, width int) string {
	if width <= 0 {
		return ""
	}

	if p.Utilities.VisibleLength(text) <= width {
		return text
	}

	runes := []rune(text)
	out := ""

	for _, r := range runes {
		candidate := out + string(r)
		if p.Utilities.VisibleLength(candidate+"…") > width {
			break
		}
		out = candidate
	}

	if width > 1 {
		return out + "…"
	}

	return out
}

func (p *Projects) fitLine(line string, width int) string {
	if width <= 0 {
		return ""
	}

	if p.Utilities.VisibleLength(line) > width {
		line = p.truncate(line, width)
	}

	for p.Utilities.VisibleLength(line) < width {
		line += " "
	}

	return line
}

func (p *Projects) selectedLine(line string, width int) string {
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

func (p *Projects) selectedFullLine(line string, width int) string {
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

func (p *Projects) previewLines() []string {
	item := p.selectedItem()
	if item == nil {
		return []string{" ..."}
	}

	if item.Kind == "new-project" {
		name := p.newProjectName()
		if name == "" {
			return []string{
				" new project",
				"",
				" escreva o nome no prompt",
				" ctrl+enter: cria projeto e primeiro plano",
			}
		}

		return []string{
			" new project",
			"",
			" name: " + name,
			" path: " + p.newProjectPath(),
			" first plan: " + p.newProjectPlanPath(),
			"",
			" ctrl+enter: criar e abrir primeiro plano",
		}
	}

	if item.Kind == "back" {
		return []string{" ...", "", " enter: voltar"}
	}

	if item.Kind == "virtual-templates" {
		return []string{" templates", "", " enter: listar templates do projeto"}
	}

	if item.Kind == "virtual-events" {
		return []string{" events", "", " enter: listar eventos linkados por owned-by"}
	}

	if item.Kind == "add-template" {
		name := p.newTemplateName()
		if name == "" {
			return []string{
				" new template",
				"",
				" escreva o nome no prompt",
				" ctrl+enter: escolher base e criar template",
			}
		}

		return []string{
			" new template",
			"",
			" name: " + name,
			" path: " + p.newTemplatePath(),
			"",
			" ctrl+enter: escolher base e abrir no editor",
		}
	}

	if item.Page == nil {
		return []string{" ..."}
	}

	if item.Kind == "project" {
		return p.projectLastPlanPreview(item)
	}

	lines := []string{}
	if item.Page.Metadata != nil {
		lines = append(lines, " name: "+projectDisplayName(item.Page))
		if item.Kind == "project" {
			lines = append(lines, fmt.Sprintf(" priority: %d", metadataInt(item.Page, "priority")))
		}
		lines = append(lines, " path: "+item.Page.Path)
		lines = append(lines, "")
	}

	for _, line := range item.Page.Content {
		lines = append(lines, " "+line)
	}

	if len(lines) == 0 {
		lines = []string{" ..."}
	}

	return lines
}

func (p *Projects) projectLastPlanPreview(item *ProjectItem) []string {
	if item == nil || item.Page == nil {
		return []string{" last plan", "", " nenhum projeto selecionado"}
	}

	projectName := projectDisplayName(item.Page)
	plan := p.lastPlanForProject(projectName)
	if plan == nil {
		return []string{
			" last plan",
			"",
			" project: " + projectName,
			" nenhum plano encontrado para este projeto",
		}
	}

	lines := []string{
		" last plan",
		"",
		" project: " + projectName,
		" path: " + plan.Path,
		"",
	}
	lines = append(lines, plan.Content...)
	return lines
}

func (p *Projects) layout() (projectRect, projectRect, projectRect) {
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
	sepH := 0

	if p.panelMode() == "preview" {
		return projectRect{}, projectRect{}, projectRect{X: x, Y: y, W: w, H: h}
	}

	if p.panelMode() == "list" {
		itemsH := h - promptH - sepH
		if itemsH < 1 {
			itemsH = 1
		}
		items := projectRect{X: x, Y: y, W: w, H: itemsH}
		prompt := projectRect{X: x, Y: y + itemsH + sepH, W: w, H: promptH}
		return items, prompt, projectRect{}
	}

	if projectsVerticalPanels(w, h) {
		previewH := h / 2
		if previewH < 1 {
			previewH = 1
		}
		if previewH > h-promptH-2 {
			previewH = h - promptH - 2
		}
		if previewH < 1 {
			previewH = 1
		}

		sepH := 1
		itemsH := h - previewH - sepH - promptH
		if itemsH < 1 {
			itemsH = 1
		}

		items := projectRect{X: x, Y: y + previewH + sepH, W: w, H: itemsH}
		prompt := projectRect{X: x, Y: y + previewH + sepH + itemsH, W: w, H: promptH}
		preview := projectRect{X: x, Y: y, W: w, H: previewH}

		return items, prompt, preview
	}

	leftW := w / 3
	if leftW < 24 {
		leftW = 24
	}
	if leftW > 42 {
		leftW = 42
	}
	if leftW > w-2 {
		leftW = w - 2
	}

	previewW := w - leftW - 1
	if previewW < 1 {
		previewW = 1
	}

	itemsH := h - promptH - sepH
	if itemsH < 1 {
		itemsH = 1
	}

	items := projectRect{X: x, Y: y, W: leftW, H: itemsH}
	prompt := projectRect{X: x, Y: y + itemsH + sepH, W: leftW, H: promptH}
	preview := projectRect{X: x + leftW + 1, Y: y, W: previewW, H: h}

	return items, prompt, preview
}

func projectsVerticalPanels(w int, h int) bool {
	return w < h*2
}

func (p *Projects) panelMode() string {
	switch p.PanelMode {
	case "list", "preview", "both":
		return p.PanelMode
	default:
		return "both"
	}
}

func cutLines(height int, selected int, total int) int {
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

func (p *Projects) projectItemLabel(item *ProjectItem) string {
	if item == nil {
		return ""
	}

	name := item.Name
	if strings.TrimSpace(name) == "" {
		name = item.Path
	}

	switch item.Kind {
	case "new-project":
		return " ‹b + ›b " + name

	case "project":
		return " " + projectPageGreekIcon(item) + name

	case "virtual-templates":
		return " " + projectGreekIcon(item.Path, true) + name

	case "virtual-events":
		return " " + projectGreekIcon(item.Path, true) + name

	case "add-template":
		return " ‹b + ›b " + name

	case "template":
		return " " + projectPageGreekIcon(item) + name

	case "event":
		symbol, cleanName := eventSymbolAndName(item.Page, name)
		return " " + projectPageGreekIcon(item) + "‹b " + symbol + " ›b " + cleanName

	default:
		return " " + projectPageGreekIcon(item) + name
	}
}

func (p *Projects) projectItemRightText(item *ProjectItem) string {
	if item == nil {
		return ""
	}

	switch item.Kind {
	case "project":
		return fmt.Sprintf("%d", item.Priority)

	default:
		return ""
	}
}

func projectGreekIcon(seed string, upper bool) string {
	if upper {
		return "‹b " + projectGreekUpper(seed) + " ›b "
	}

	return "‹b " + projectGreekLower(seed) + " ›b "
}

func projectPageGreekIcon(item *ProjectItem) string {
	if item == nil {
		return projectGreekIcon("", false)
	}

	seed := item.Path
	if item.Page != nil && item.Page.Path != "" {
		seed = item.Page.Path
	}

	upper := item.Page != nil && item.Page.Type == "deep"
	return projectGreekIcon(seed, upper)
}

func projectGreekUpper(seed string) string {
	letters := []string{
		"Α", "Β", "Γ", "Δ", "Ε", "Ζ", "Η", "Θ",
		"Ι", "Κ", "Λ", "Μ", "Ν", "Ξ", "Ο", "Π",
		"Ρ", "Σ", "Τ", "Υ", "Φ", "Χ", "Ψ", "Ω",
	}

	return letters[projectHashIndex(seed, len(letters))]
}

func projectGreekLower(seed string) string {
	letters := []string{
		"α", "β", "γ", "δ", "ε", "ζ", "η", "θ",
		"ι", "κ", "λ", "μ", "ν", "ξ", "ο", "π",
		"ρ", "σ", "τ", "υ", "φ", "χ", "ψ", "ω",
	}

	return letters[projectHashIndex(seed, len(letters))]
}

func projectHashIndex(seed string, size int) int {
	if size <= 0 {
		return 0
	}

	hash := 0

	for _, r := range seed {
		hash = (hash*31 + int(r)) % size
	}

	if hash < 0 {
		hash = -hash
	}

	return hash
}

func (p *Projects) truncateStyledLabel(label string, width int) string {
	if width <= 0 {
		return ""
	}

	if p.Utilities.VisibleLength(label) <= width {
		return label
	}

	return p.Utilities.CutVisible(label, width)
}

func (p *Projects) newTemplateName() string {
	name := strings.TrimSpace(p.Prompt)
	if name == "" {
		return ""
	}

	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.Trim(name, "/")
	name = strings.TrimSpace(name)

	return name
}

func (p *Projects) newTemplatePath() string {
	name := p.newTemplateName()
	if name == "" || p.Context.ProjectPath == "" {
		return ""
	}

	return filepath.ToSlash(filepath.Join(p.Context.ProjectPath, name))
}

func (p *Projects) newTemplateContent(source *filesystem.Page) []string {
	name := p.newTemplateName()
	if name == "" {
		name = "untitled"
	}

	if source != nil && len(source.Content) > 0 {
		values := projectTemplateValues(name, p.Context.ProjectName, "session")
		for k, v := range map[string]string{
			"name":         name,
			"template":     name,
			"project":      p.Context.ProjectName,
			"project-name": p.Context.ProjectName,
		} {
			values[k] = v
		}
		return projectApplyTemplate(source.Content, values)
	}

	return []string{
		"name: " + name,
		"kind: session",
		"owned-by: " + p.Context.ProjectName,
		"---",
		"",
	}
}

func (p *Projects) newProjectName() string {
	name := strings.TrimSpace(p.Prompt)
	if name == "" {
		return ""
	}

	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.Trim(name, "/")
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	return name
}

func (p *Projects) newProjectPath() string {
	name := p.newProjectName()
	if name == "" {
		return "e.proj/"
	}
	return filepath.ToSlash(filepath.Join("e.proj", name))
}

func (p *Projects) newProjectPlanPath() string {
	name := p.newProjectName()
	if name == "" {
		return "b.rec/"
	}
	return filepath.ToSlash(filepath.Join("b.rec", ">kickstart@"+name))
}

func (p *Projects) refreshProjectPriorities() {
	root := p.projectsRoot()
	if root == nil {
		return
	}

	for _, project := range root.Children {
		if project == nil {
			continue
		}
		metadata := filesystem.ParseMetadataFromContent(project.Content, project.Name)
		if project.Metadata == nil {
			project.Metadata = map[string]any{}
		}
		if priority, ok := metadata["priority"]; ok {
			project.Metadata["priority"] = priority
		}
	}
}

func projectApplyTemplate(lines []string, values map[string]string) []string {
	out := append([]string(nil), lines...)
	for i, line := range out {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "modified:") {
			if value, ok := values["modified"]; ok {
				line = strings.ReplaceAll(line, "{yes/no}", value)
			}
		}
		for key, value := range values {
			line = strings.ReplaceAll(line, "{"+key+"}", value)
			line = strings.ReplaceAll(line, "{{"+key+"}}", value)
		}
		out[i] = line
	}
	return out
}

func projectTemplateValues(name string, projectName string, kind string) map[string]string {
	now := time.Now().Format("02.01.2006 15:04")
	return map[string]string{
		"id":           "",
		"name":         name,
		"template":     name,
		"kind":         kind,
		"time":         now,
		"release-time": now,
		"owner":        projectName,
		"project":      projectName,
		"project-name": projectName,
		"projects":     projectName,
		"tags":         "",
		"measurement":  "",
		"hidden":       "no",
		"modified":     "no",
		"yes/no":       "no",
	}
}

func eventSortTime(page *filesystem.Page) time.Time {
	if page == nil || page.Metadata == nil {
		return time.Time{}
	}

	raw, ok := page.Metadata["time"]
	if !ok || raw == nil {
		return time.Time{}
	}

	switch value := raw.(type) {
	case time.Time:
		return value

	case string:
		value = strings.TrimSpace(value)
		if value == "" {
			return time.Time{}
		}

		layouts := []string{
			"02.01.2006 15:04",
			"02.01.2006",
			time.RFC3339,
			"2006-01-02 15:04",
			"2006-01-02",
		}

		for _, layout := range layouts {
			if parsed, err := time.Parse(layout, value); err == nil {
				return parsed
			}
		}

		return time.Time{}

	default:
		return time.Time{}
	}
}

func (p *Projects) lastPlanForCurrentProject() *filesystem.Page {
	return p.lastPlanForProject(p.Context.ProjectName)
}

func (p *Projects) lastPlanForProject(projectName string) *filesystem.Page {
	if p.EventsPage == nil || strings.TrimSpace(projectName) == "" {
		return nil
	}

	projectName = strings.TrimSpace(projectName)
	events := []*filesystem.Page{}
	for _, event := range p.EventsPage.Children {
		if event == nil || !projectOwnsEvent(event, projectName) {
			continue
		}
		events = append(events, event)
	}

	sort.SliceStable(events, func(i, j int) bool {
		at := eventSortTime(events[i])
		bt := eventSortTime(events[j])

		if !at.IsZero() && !bt.IsZero() && !at.Equal(bt) {
			return at.After(bt)
		}
		if !at.IsZero() != !bt.IsZero() {
			return !at.IsZero()
		}
		return strings.ToLower(events[i].Name) < strings.ToLower(events[j].Name)
	})

	for _, event := range events {
		if event == nil {
			continue
		}

		if isPlanEvent(event) {
			return event
		}
	}

	return nil
}

func isPlanEvent(page *filesystem.Page) bool {
	if page == nil {
		return false
	}

	name := strings.TrimSpace(page.Name)
	if strings.HasPrefix(name, ">") {
		return true
	}

	if len(page.Content) > 0 {
		firstLine := strings.TrimSpace(page.Content[0])
		if strings.HasPrefix(firstLine, ">") {
			return true
		}
	}

	return false
}

func pagePath(page *filesystem.Page) string {
	if page == nil {
		return ""
	}

	return page.Path
}

func eventSymbolAndName(page *filesystem.Page, fallbackName string) (string, string) {
	name := strings.TrimSpace(fallbackName)

	if name == "" && page != nil {
		name = strings.TrimSpace(page.Name)
	}

	if name == "" && page != nil && len(page.Content) > 0 {
		name = strings.TrimSpace(page.Content[0])
	}

	if name == "" {
		return "•", ""
	}

	runes := []rune(name)
	first := string(runes[0])

	switch first {
	case "=", ">", "~", "`":
		clean := strings.TrimSpace(string(runes[1:]))
		return first, clean
	default:
		return "•", name
	}
}
