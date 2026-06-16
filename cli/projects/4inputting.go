package projects

import (
	"crypto/rand"
	"env/cli/menu"
	"env/filesystem"
	"env/routines"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (p *Projects) Input(newinput *routines.Input) {
	switch newinput.Key {
	case "up":
		p.moveVertical(1)

	case "down":
		p.moveVertical(-1)

	case "ctrl+p":
		p.cyclePanelMode()

	case "char":
		p.Prompt += string(newinput.Char)

		if p.Context.Kind == "events" {
			p.Selected = 0
			p.updateItems()
		}

	case "backspace":
		p.backspacePrompt()

	case "ctrl+enter":
		p.openSelectedActions()

	case "enter":
		p.enter()

	case "left":
		p.goBack()
	}
}

func (p *Projects) cyclePanelMode() {
	switch p.PanelMode {
	case "list":
		p.PanelMode = "preview"
	case "preview":
		p.PanelMode = "both"
	default:
		p.PanelMode = "list"
	}
}

func (p *Projects) moveVertical(delta int) {
	if len(p.Items) == 0 {
		p.Selected = 0
		return
	}

	p.Selected += delta
	if p.Selected < 0 {
		p.Selected = 0
	}
	if p.Selected >= len(p.Items) {
		p.Selected = len(p.Items) - 1
	}
}

func (p *Projects) backspacePrompt() {
	if len(p.Prompt) == 0 {
		return
	}

	p.Prompt = p.Prompt[:len(p.Prompt)-1]

	if p.Context.Kind == "events" {
		p.Selected = 0
		p.updateItems()
	}
}

func (p *Projects) enter() {
	item := p.selectedItem()
	if item == nil {
		return
	}

	if item.Kind == "back" || item.Name == "..." {
		p.goBack()
		return
	}

	switch item.Kind {
	case "project":
		p.Context = ProjectContext{
			Kind:        "project",
			ProjectPath: item.Path,
			ProjectName: item.Name,
		}
		p.Path = item.Path
		p.Selected = 0
		p.updateItems()

	case "virtual-templates":
		p.Context.Kind = "templates"
		p.Path = item.Path
		p.Selected = 0
		p.updateItems()

	case "virtual-events":
		p.Context.Kind = "events"
		p.Path = item.Path
		p.Selected = 0
		p.updateItems()
	}
}

func (p *Projects) openNewTemplateInEditor(source *filesystem.Page) {
	path := strings.TrimSpace(p.newTemplatePath())
	if path == "" {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs != nil && fs.Find(path) == nil {
		_, draft := fs.NewDraft(path)
		if draft != nil {
			draft.Type = "shallow"
			draft.Stage = "draft"
			draft.Content = p.newTemplateContent(source)
			draft.Metadata = map[string]any{
				"project":  p.Context.ProjectName,
				"owned-by": p.Context.ProjectName,
			}
		}
	}

	p.Prompt = ""
	p.updateItems()
	p.Parent.AddClients("editor:"+path, "next")
}

func (p *Projects) openNewTemplateMenu() {
	if p.Parent == nil {
		return
	}
	if strings.TrimSpace(p.newTemplatePath()) == "" {
		return
	}

	templates := p.projectBaseTemplates()
	items := make([]menu.Item, 0, len(templates))
	for _, tmpl := range templates {
		if tmpl == nil {
			continue
		}
		source := tmpl
		items = append(items, menu.Item{
			Name:    templateName(source),
			Command: "template:" + source.Path,
			Run:     func(_ string) { p.openNewTemplateInEditor(source) },
		})
	}

	if len(items) == 0 {
		p.openNewTemplateInEditor(nil)
		return
	}

	p.Parent.OpenLauncher("NEW TEMPLATE", items)
}

func (p *Projects) createNewProject() {
	if p.Parent == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	name := p.newProjectName()
	if name == "" {
		return
	}

	projectPath := filepath.ToSlash(filepath.Join("e.proj", name))
	if fs.Find(projectPath) != nil {
		p.Parent.AddClients("editor:"+projectPath, "next")
		return
	}

	_, project := fs.NewDraft(projectPath)
	if project != nil {
		project.Type = "deep"
		project.Stage = "draft"
		project.Metadata = map[string]any{"priority": 0}
		project.Content = []string{
			"priority: 0",
			"---",
			"",
		}
		project.Children = []*filesystem.Page{}
	}

	_, templates := fs.NewDraft(filepath.ToSlash(filepath.Join(projectPath, ".templates")))
	if templates != nil {
		templates.Type = "deep"
		templates.Stage = "draft"
		templates.Content = []string{"---", ""}
		templates.Children = []*filesystem.Page{}
	}

	planPath := filepath.ToSlash(filepath.Join("b.rec", ">kickstart@"+name))
	planID := projectGenerateID(12)
	_, plan := fs.NewDraft(planPath)
	if plan != nil {
		plan.Type = "shallow"
		plan.Stage = "draft"
		planTemplate := p.projectTemplateByName("kickstart")
		plan.Metadata = map[string]any{
			"id":       planID,
			"name":     "kickstart",
			"kind":     "plan",
			"owned-by": name,
			"time":     time.Now(),
		}
		plan.Content = p.newPlanContent(name, planID, planTemplate)
	}
	p.appendKickstartBlockToTodayLog(name, planID)

	p.Prompt = ""
	p.Path = projectPath
	p.Context = ProjectContext{Kind: "project", ProjectPath: projectPath, ProjectName: name}
	p.requestProjects()
	p.requestEvents()
	p.updateItems()
	p.Parent.AddClients("editor:"+planPath, "next")
}

func (p *Projects) projectBaseTemplates() []*filesystem.Page {
	if p.Parent != nil && p.Parent.GetFilesystem() != nil {
		if templatesRoot := p.Parent.GetFilesystem().Find("e.proj/.templates"); templatesRoot != nil && len(templatesRoot.Children) > 0 {
			return projectTemplateChildrenOrFallback(templatesRoot)
		}
	}

	if diskTemplates := loadProjectBaseTemplatesFromDisk(); len(diskTemplates) > 0 {
		return diskTemplates
	}

	root := p.projectsRoot()
	if root == nil {
		return []*filesystem.Page{builtinProjectTemplate("kickstart")}
	}
	templatesRoot := findProjectPage(root, "e.proj/.templates")
	if templatesRoot == nil || len(templatesRoot.Children) == 0 {
		return []*filesystem.Page{builtinProjectTemplate("kickstart")}
	}

	return projectTemplateChildrenOrFallback(templatesRoot)
}

func projectTemplateChildrenOrFallback(templatesRoot *filesystem.Page) []*filesystem.Page {
	out := []*filesystem.Page{}
	for _, child := range templatesRoot.Children {
		if child == nil {
			continue
		}
		name := strings.TrimSpace(child.Name)
		if name == "" || name == "index" {
			continue
		}
		out = append(out, child)
	}

	if len(out) == 0 {
		out = append(out, builtinProjectTemplate("kickstart"))
	}

	return out
}

func loadProjectBaseTemplatesFromDisk() []*filesystem.Page {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	dir := filepath.Join(home, "prsnl.spc", "e.proj", ".templates")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	out := []*filesystem.Page{}
	for _, entry := range entries {
		if entry == nil || entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if name == "" || name == "index" {
			continue
		}

		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		content := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
		for len(content) > 0 && content[len(content)-1] == "" {
			content = content[:len(content)-1]
		}
		out = append(out, &filesystem.Page{
			Name:     name,
			Path:     filepath.ToSlash(filepath.Join("e.proj", ".templates", name)),
			Type:     "shallow",
			Content:  content,
			Metadata: filesystem.ParseMetadataFromContent(content, name),
			Stage:    "local",
		})
	}

	return out
}

func (p *Projects) projectTemplateByName(name string) *filesystem.Page {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, tmpl := range p.projectBaseTemplates() {
		if tmpl == nil {
			continue
		}
		if strings.ToLower(templateName(tmpl)) == name || strings.ToLower(filepath.Base(tmpl.Path)) == name {
			return tmpl
		}
	}
	return builtinProjectTemplate(name)
}

func builtinProjectTemplate(name string) *filesystem.Page {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "blank"
	}

	content := []string{
		"name: {name}",
		"kind: session",
		"owned-by: {project-name}",
		"time: " + time.Now().Format("02.01.2006 15:04"),
		"---",
		"",
	}

	if name == "kickstart" {
		content = []string{
			"name: kickstart",
			"kind: plan",
			"owned-by: {project-name}",
			"time: " + time.Now().Format("02.01.2006 15:04"),
			"---",
			"",
			"# kickstart",
			"",
		}
	}

	return &filesystem.Page{
		Name:     name,
		Path:     filepath.ToSlash(filepath.Join("e.proj", ".templates", name)),
		Type:     "shallow",
		Content:  content,
		Metadata: map[string]any{"name": name},
		Stage:    "local",
	}
}

func (p *Projects) appendKickstartBlockToTodayLog(projectName string, id string) {
	if p == nil || p.Parent == nil || strings.TrimSpace(projectName) == "" || strings.TrimSpace(id) == "" {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	path := "a.log/" + time.Now().Format("02.01.2006")
	page := fs.Find(path)
	if page == nil {
		_, page = fs.NewDraft(path)
	}
	if page == nil {
		return
	}

	lines := append([]string(nil), page.Content...)
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > 0 {
		lines = append(lines, "")
	}
	lines = append(lines,
		"> ! kickstart @"+projectName,
		"id: "+id,
		"",
	)

	fs.EditContent(page, lines)
}

func projectGenerateID(size int) string {
	if size < 1 {
		size = 12
	}

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

func (p *Projects) newPlanContent(projectName string, id string, source *filesystem.Page) []string {
	if source != nil && len(source.Content) > 0 {
		values := projectTemplateValues("kickstart", projectName, "plan")
		for k, v := range map[string]string{
			"name":         "kickstart",
			"template":     "kickstart",
			"project":      projectName,
			"project-name": projectName,
			"id":           id,
		} {
			values[k] = v
		}
		return projectApplyTemplate(source.Content, values)
	}

	return []string{
		"id: " + id,
		"name: kickstart",
		"kind: plan",
		"owned-by: " + projectName,
		"time: " + time.Now().Format("02.01.2006 15:04"),
		"---",
		"",
	}
}

func (p *Projects) goBack() {
	switch p.Context.Kind {
	case "templates", "events":
		p.Context.Kind = "project"
		p.Path = p.Context.ProjectPath

	case "project":
		p.Context = ProjectContext{Kind: "root"}
		p.Path = "e.proj"

	default:
		return
	}

	p.Selected = 0
	p.updateItems()
}

func (p *Projects) openSelectedInEditor() {
	item := p.selectedItem()
	if item == nil {
		return
	}

	if item.Kind == "back" || item.Name == "..." {
		return
	}

	if item.Kind == "virtual-templates" || item.Kind == "virtual-events" {
		return
	}

	if item.Kind == "add-template" {
		return
	}

	path := strings.TrimSpace(item.Path)
	if path == "" && item.Page != nil {
		path = strings.TrimSpace(item.Page.Path)
	}
	if path == "" {
		return
	}

	p.Parent.AddClients("editor:"+path, "next")
}

func (p *Projects) openSelectedActions() {
	if p.Parent == nil {
		return
	}

	item := p.selectedItem()
	if item == nil {
		return
	}

	if item.Kind == "back" || item.Name == "..." {
		return
	}

	switch item.Kind {
	case "new-project":
		p.createNewProject()

	case "template":
		p.openTemplateActions(item)

	case "add-template":
		p.openNewTemplateMenu()

	case "event":
		p.openSelectedInEditor()

	case "project":
		p.openSelectedInEditor()

	default:
		return
	}
}

func (p *Projects) openTemplateActions(item *ProjectItem) {
	if item == nil || item.Page == nil {
		return
	}

	page := item.Page
	items := []menu.Item{
		{
			Name:    "use template",
			Command: "use template",
			Run: func(_ string) {
				p.Parent.RunTemplateAction(page.Path)
			},
		},
		{
			Name:    "delete template",
			Command: "delete template",
			Run: func(_ string) {
				p.deleteTemplate(page)
			},
		},
		{
			Name:    "open in editor",
			Command: "open in editor",
			Run: func(_ string) {
				p.Parent.AddClients("editor:"+page.Path, "next")
			},
		},
	}

	p.Parent.OpenLauncher("TEMPLATE ACTIONS", items)
}

func (p *Projects) deleteTemplate(page *filesystem.Page) {
	if p.Parent == nil || page == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	fs.Doom(page, true)
	syncReq := filesystem.NewSyncRequest()
	syncReq.Branch = page
	syncReq.Hard = true
	fs.Sync(syncReq)
}
