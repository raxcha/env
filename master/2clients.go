package master

import (
	"env/cli"
	"env/cli/editor"
	"env/cli/picker"
	"env/cli/projects"
	"env/cli/zettelkasten"
	"strings"
	"time"
)

func (m *Master) AddClients(arg string, mode string) {

	var toinsert = []cli.Cli{}
	changed := false
	forceNew := strings.Contains(mode, "new")
	stay := strings.Contains(mode, "stay")

	args := splitIntoMore(arg, " ")
	for _, arg := range args {

		name, spec := splitIntoTwo(arg, ":")

		if len(spec) == 0 {
			if inferredName, inferredSpec, ok := inferClientFromPath(name); ok {
				name = inferredName
				spec = inferredSpec
			} else {
				spec = "."
			}
		}
		if spec == "." && name == "today" {
			name = "editor"
			spec = m.generateLog(0)
		}
		if spec == "." && name == "tomorrow" {
			name = "editor"
			spec = m.generateLog(1)
		}
		if spec == "." && name == "editor" {
			name = "editor"
			spec = "untitled"
		}
		if !m.stageEnabled(clientStage(name)) {
			continue
		}

		found := false
		if !forceNew {
			for i, client := range m.Clients {
				if name == client.GetKind() && spec == client.GetSpec() {
					if m.Filesystem != nil {
						if page := m.Filesystem.Find(spec); page != nil {
							client.SetPage(page)
						}
					}
					if !stay {
						m.Focus = i
					}
					changed = true
					found = true
					break
				}
			}
		}
		if found {
			continue
		}

		if !forceNew && name == "editor" && m.Filesystem != nil {
			if page := m.Filesystem.Find(spec); page != nil {
				for i, client := range m.Clients {
					if client.GetKind() != "editor" {
						continue
					}

					client.SetPage(page)
					if client.GetSpec() == spec {
						if !stay {
							m.Focus = i
						}
						changed = true
						found = true
						break
					}
				}
			}
		}
		if found {
			continue
		}

		switch name {
		case "editor":
			e := editor.CreateEditor(spec, m)
			if e.Sidebar != nil && !m.stageEnabled("sidebar") {
				e.Sidebar.On = false
			}
			toinsert = append(toinsert, e)

		case "picker":
			toinsert = append(toinsert, picker.CreatePicker(spec, m))

		case "projects":
			toinsert = append(toinsert, projects.CreateProjects(spec, m))

		case "zettelkasten":
			toinsert = append(toinsert, zettelkasten.CreateZettelkasten(spec, m))

		case "menu":
			continue
		}
	}

	if len(toinsert) == 0 {
		if changed {
			m.setClientsBounds()
			m.Draw()
		}
		return
	}

	n := len(m.Clients)
	focus := m.Focus
	if focus < 0 {
		focus = 0
	}
	if focus > n {
		focus = n
	}

	newFocus := focus
	switch mode {
	case "start":
		m.Clients = append(toinsert, m.Clients...)
		newFocus = 0
	case "end":
		m.Clients = append(m.Clients, toinsert...)
		newFocus = n
	case "prev":
		m.Clients = append(m.Clients[:focus:focus], append(toinsert, m.Clients[focus:]...)...)
		newFocus = focus
	case "next":
		fallthrough
	case "next-stay":
		fallthrough
	case "next-new":
		fallthrough
	case "next-new-stay":
		after := focus + 1
		if after > n {
			after = n
		}
		m.Clients = append(m.Clients[:after:after], append(toinsert, m.Clients[after:]...)...)
		newFocus = after
	default:
		return
	}

	if stay {
		newFocus = focus
	}

	m.Focus = newFocus
	m.setClientsBounds()
	m.Draw()

}

func (m *Master) generateLog(offset int) string {

	t := time.Now().AddDate(0, 0, offset)
	return "a.log/" + t.Format("02.01.2006")
}

func inferClientFromPath(path string) (string, string, bool) {
	root := strings.TrimSpace(path)
	if root == "" {
		return "", "", false
	}

	if i := strings.Index(root, "/"); i >= 0 {
		root = root[:i]
	}

	switch root {
	case "a.log":
		return "editor", path, true
	case "b.rec":
		return "picker", path, true
	case "c.rand", "d.fami":
		return "zettelkasten", path, true
	case "e.proj":
		return "projects", path, true
	default:
		return "", "", false
	}
}

func (m *Master) CloseFocusedClient() {
	if len(m.Clients) == 0 {
		m.Focus = 0
		return
	}

	idx := m.Focus
	if idx < 0 || idx >= len(m.Clients) {
		return
	}

	m.Clients = append(m.Clients[:idx], m.Clients[idx+1:]...)

	if m.Focus >= len(m.Clients) {
		m.Focus = len(m.Clients) - 1
	}

	if m.Focus < 0 && len(m.Clients) > 0 {
		m.Focus = 0
	}

	m.setClientsBounds()
	m.Draw()
}
