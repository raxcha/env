package master

import (
	"env/cli/editor"
	"env/cli/menu"
	"env/filesystem"
	"env/routines"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (m *Master) updateInput(newinput *routines.Input) {
	if newinput.Key == "ctrl+t" {
		newinput.Key = "ctrl+enter"
	}

	if m.shouldIgnoreEmptyBackgroundInput(newinput) {
		return
	}

	if newinput.Key == "ctrl+z" && m.Notifications.On {
		m.Notifications.CancelLatest()

	} else if newinput.Key == "esc" {
		m.Tabs.Switch = false
		m.Menu.Cancel()

	} else if newinput.Key == "tab" {
		m.Tabs.Switch = !m.Tabs.Switch

	} else if newinput.Key == "ctrl+space" {
		m.openLauncher()

	} else if newinput.Key == "ctrl+q" {
		m.Routines.RestoreAndExit()

	} else if m.Menu.On {
		m.Menu.Input(newinput)

	} else if m.Tabs.Switch {

		if newinput.Key == "ctrl+backspace" {
			m.CloseFocusedClient()
		} else {
			m.Tabs.Input(newinput)
		}

	} else if len(m.Clients) != 0 {
		m.Clients[m.Focus].Input(newinput)
	}

	m.Draw()
}

func (m *Master) shouldIgnoreEmptyBackgroundInput(input *routines.Input) bool {
	if len(m.Clients) != 0 || m.Menu.On {
		return false
	}
	if input.Key == "ctrl+space" || input.Key == "ctrl+q" {
		return false
	}
	if input.Key == "ctrl+z" && m.Notifications.On {
		return false
	}
	return true
}

func (m *Master) openLauncher() {
	if !m.stageEnabled("menu") {
		return
	}

	todayPath := m.generateLog(0)
	items := []menu.Item{
		{Name: "open today", Command: "today editor:" + todayPath, Run: func(_ string) { m.AddClients("today", "next") }},
		{Name: "do something", Command: "template do something", Run: func(_ string) { m.openTemplateLauncher() }},
		{
			Name:    "insert timestamp",
			Command: "timestamp",
			Run: func(_ string) {
				if e := m.focusedEditor(); e != nil {
					e.InsertTextAtCursor(time.Now().Format("02.01.2006 15:04"))
				}
			},
		},
		{Name: "open picker", Command: "picker", Run: func(_ string) { m.AddClients("picker", "next") }},
		{Name: "open projects", Command: "projects", Run: func(_ string) { m.AddClients("projects", "next") }},
		{Name: "open zettelkasten", Command: "zettelkasten", Run: func(_ string) { m.AddClients("zettelkasten", "next") }},
	}

	if m.Mode == "monocle" {
		items = append(items, menu.Item{Name: "change to fibonacci mode", Command: "fibonacci", Run: func(_ string) { m.setMode("fibonacci") }})
	} else {
		items = append(items, menu.Item{Name: "change to monocle mode", Command: "monocle", Run: func(_ string) { m.setMode("monocle") }})
	}

	items = append(items, menu.Item{Name: "change theme", Command: "theme colors appearance", Run: func(_ string) { m.openThemeLauncher() }})
	items = append(items, menu.Item{Name: "edit .options", Command: "editor:.options", Run: func(_ string) { m.AddClients("editor:.options", "next") }})

	items = append(items, menu.Item{
		Name:       "run command",
		Command:    "__prompt__",
		Pin:        true,
		PromptOnly: true,
		Run: func(prompt string) {
			prompt = strings.TrimSpace(prompt)
			if prompt != "" {
				m.AddClients(prompt, "next")
			}
		},
	})

	m.Menu.Open("main", items)
}

func (m *Master) openThemeLauncher() {
	if m == nil || m.Menu == nil || m.Utilities == nil {
		return
	}

	items := []menu.Item{}
	var previousTheme = m.Utilities.Theme
	current := ""
	if m.Utilities.Theme != nil {
		current = m.Utilities.Theme.Name
	}

	for _, theme := range m.Utilities.Themes {
		if theme == nil || strings.TrimSpace(theme.Name) == "" {
			continue
		}

		name := theme.Name
		themeName := name
		right := ""
		if name == current {
			right = "on"
		}

		items = append(items, menu.Item{
			Name:    name,
			Command: "theme:" + name,
			Right:   right,
			Preview: func(_ string) {
				m.setTheme(themeName)
			},
			Run: func(_ string) {
				m.setTheme(themeName)
				writeRootThemeOption(themeName)
			},
		})
	}

	if len(items) == 0 {
		return
	}

	m.Menu.Open("themes", items)
	m.selectThemeMenuItem(current)
	if current != "" {
		m.setTheme(current)
	}
	m.Menu.OnCancel = func() {
		if previousTheme != nil {
			m.Utilities.Theme = previousTheme
		}
	}
}

func (m *Master) selectThemeMenuItem(name string) {
	if m == nil || m.Menu == nil {
		return
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return
	}

	for i, item := range m.Menu.Items {
		if strings.TrimPrefix(item.Command, "theme:") == name {
			m.Menu.Selected = i
			return
		}
	}
}

func (m *Master) setTheme(name string) {
	if m == nil || m.Utilities == nil {
		return
	}

	name = strings.TrimSpace(name)
	for _, theme := range m.Utilities.Themes {
		if theme == nil || theme.Name != name {
			continue
		}

		m.Utilities.Theme = theme
		m.Draw()
		return
	}
}

func writeRootThemeOption(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}

	path := rootOptionsPath()
	data, _ := os.ReadFile(path)
	text := string(data)
	lines := strings.Split(text, "\n")
	hasTrailingNewline := strings.HasSuffix(text, "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = []string{}
	}

	insertAt := len(lines)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			insertAt = i
			break
		}

		key, _, ok := strings.Cut(trimmed, ":")
		if ok && strings.EqualFold(strings.TrimSpace(key), "theme") {
			lines[i] = "theme: " + name
			writeOptionsLines(path, lines, hasTrailingNewline)
			return
		}
	}

	lines = append(lines, "")
	copy(lines[insertAt+1:], lines[insertAt:])
	lines[insertAt] = "theme: " + name
	writeOptionsLines(path, lines, true)
}

func rootOptionsPath() string {
	if root := strings.TrimSpace(os.Getenv("ENV_ROOT")); root != "" {
		return filepath.Join(filepath.Clean(root), ".options")
	}

	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		home = "/home/asdf"
	}
	return filepath.Join(home, "prsnl.spc", ".options")
}

func writeOptionsLines(path string, lines []string, trailingNewline bool) {
	content := strings.Join(lines, "\n")
	if trailingNewline && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	_ = os.WriteFile(path, []byte(content), 0644)
}

func (m *Master) focusedEditor() *editor.Editor {
	if m == nil || len(m.Clients) == 0 || m.Focus < 0 || m.Focus >= len(m.Clients) {
		return nil
	}

	e, ok := m.Clients[m.Focus].(*editor.Editor)
	if !ok {
		return nil
	}

	return e
}

func (m *Master) hasClient(kind string, spec string) bool {
	for _, client := range m.Clients {
		if client != nil && client.GetKind() == kind && filepathClean(client.GetSpec()) == filepathClean(spec) {
			return true
		}
	}
	return false
}

func (m *Master) focusedClientIs(kind string, spec string) bool {
	if len(m.Clients) == 0 || m.Focus < 0 || m.Focus >= len(m.Clients) {
		return false
	}

	client := m.Clients[m.Focus]
	return client != nil && client.GetKind() == kind && filepathClean(client.GetSpec()) == filepathClean(spec)
}

func (m *Master) setMode(mode string) {
	mode = strings.TrimSpace(mode)
	if mode != "monocle" && mode != "fibonacci" {
		return
	}

	m.Mode = mode
	m.setClientsBounds()
	m.Draw()
}

func (m *Master) openTemplateLauncher() {
	items := m.templateLauncherItems()
	if len(items) == 0 || m.Menu == nil {
		return
	}

	m.Menu.Open("templates", items)
}

func (m *Master) templateLauncherItems() []menu.Item {
	if m == nil || m.Filesystem == nil || m.Filesystem.Cache == nil {
		return nil
	}

	root := findMasterPage(m.Filesystem.Cache, "e.proj")
	if root == nil {
		return nil
	}

	projects := append([]*filesystem.Page{}, root.Children...)
	sort.SliceStable(projects, func(i, j int) bool {
		ap := masterMetadataInt(projects[i], "priority")
		bp := masterMetadataInt(projects[j], "priority")
		if ap != bp {
			return ap > bp
		}
		return strings.ToLower(masterPageDisplayName(projects[i])) < strings.ToLower(masterPageDisplayName(projects[j]))
	})

	items := []menu.Item{}
	for _, project := range projects {
		if project == nil {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(project.Name), ".") || strings.HasPrefix(filepathClean(project.Path), "e.proj/.") {
			continue
		}
		projectName := masterPageDisplayName(project)
		projectPriority := masterMetadataInt(project, "priority")
		templates := append([]*filesystem.Page{}, project.Children...)
		sort.SliceStable(templates, func(i, j int) bool {
			return strings.ToLower(masterPageDisplayName(templates[i])) < strings.ToLower(masterPageDisplayName(templates[j]))
		})

		for _, tmpl := range templates {
			if tmpl == nil {
				continue
			}
			templatePath := tmpl.Path
			name := fmt.Sprintf("%s / %s", projectName, masterPageDisplayName(tmpl))
			items = append(items, menu.Item{
				Name:    name,
				Command: "template:" + templatePath,
				Right:   fmt.Sprintf("%d", projectPriority),
				Run:     func(_ string) { m.RunTemplateAction(templatePath) },
			})
		}
	}

	return items
}

func findMasterPage(page *filesystem.Page, path string) *filesystem.Page {
	if page == nil {
		return nil
	}
	if filepathClean(page.Path) == filepathClean(path) {
		return page
	}
	for _, child := range page.Children {
		if found := findMasterPage(child, path); found != nil {
			return found
		}
	}
	return nil
}

func masterPageDisplayName(page *filesystem.Page) string {
	if page == nil {
		return ""
	}
	if page.Metadata != nil {
		if name, ok := page.Metadata["name"].(string); ok && usableMasterDisplayName(name) {
			return strings.TrimSpace(name)
		}
	}
	if strings.TrimSpace(page.Name) != "" {
		return strings.TrimSpace(page.Name)
	}
	return filepathClean(page.Path)
}

func usableMasterDisplayName(name string) bool {
	name = strings.TrimSpace(name)
	return name != "" && !strings.Contains(name, "{") && !strings.Contains(name, "}")
}

func masterMetadataInt(page *filesystem.Page, key string) int {
	if page == nil || page.Metadata == nil {
		return 0
	}

	switch value := page.Metadata[key].(type) {
	case int:
		return value
	case float64:
		return int(value)
	case string:
		var out int
		fmt.Sscanf(strings.TrimSpace(value), "%d", &out)
		return out
	default:
		return 0
	}
}

func filepathClean(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "."
	}
	path = strings.ReplaceAll(path, "\\", "/")
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return "."
	}
	return path
}
