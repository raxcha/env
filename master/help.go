package master

import (
	"env/apiclient"
	"env/cli"
	"env/cli/menu"
	"env/filesystem"
	"env/utilities"
	"strings"
	"time"
)

func (m Master) GetClients() []cli.Cli {
	return m.Clients
}

func (m Master) GetFocus() int {
	return m.Focus
}

func (m Master) GetFocusedClient() cli.Cli {
	if len(m.Clients) == 0 || m.Focus < 0 || m.Focus >= len(m.Clients) {
		return nil
	}
	return m.Clients[m.Focus]
}

func (m Master) GetUtilities() *utilities.Utilities {
	return m.Utilities
}

func (m Master) GetFilesystem() *filesystem.Filesystem {
	return m.Filesystem
}

func (m Master) GetTheme() utilities.Theme {
	return *m.Utilities.Theme
}

func (m Master) GetSyncClient() *apiclient.Client {
	return m.SyncClient
}

func (m Master) GetInteractionMode() string {
	if m.Tabs != nil && m.Tabs.Switch {
		return "TABS"
	}
	if m.Menu != nil && m.Menu.On {
		return "MENU"
	}
	return "NORMAL"
}

func (m *Master) OpenLauncher(title string, items []menu.Item) {
	if m.Menu == nil {
		return
	}
	m.Menu.Open(title, items)
	m.Draw()
}

func (m *Master) RunTemplateAction(templatePath string) {
	if strings.TrimSpace(templatePath) == "" {
		return
	}
	m.AddClients("editor:"+templatePath, "next")
}

func (m *Master) QueueDelayedAction(label string, duration time.Duration, onCommit func(), onCancel func()) {
	if m.Notifications == nil {
		time.AfterFunc(duration, func() {
			if onCommit != nil {
				onCommit()
			}
		})
		return
	}
	m.Notifications.Add(label, duration, onCommit, onCancel)
}

func (m *Master) initGhostCache() {
	roots := []string{"a.log", "b.rec", "c.rand", "d.fami", "e.proj", ".resources"}
	for _, root := range roots {
		req := filesystem.NewLoadRequest()
		req.Mode = "ghost"
		req.Path = root
		req.Depth = -1
		m.Filesystem.Load(req)
	}
}

func (m *Master) SetFocus(idx int) {
	if len(m.Clients) == 0 {
		m.Focus = 0
		m.Draw()
		return
	}

	if idx < 0 {
		idx = 0
	}

	if idx >= len(m.Clients) {
		idx = len(m.Clients) - 1
	}

	m.Focus = idx
	m.setClientsBounds()
	m.Draw()
}

func (m *Master) SetClients(client cli.Cli, idx int) {
	if idx < 0 || idx >= len(m.Clients) || client == nil {
		return
	}

	m.Clients[idx] = client
	m.SetFocus(idx)
	m.setClientsBounds()
	m.Draw()
}

func (m *Master) MoveClient(from int, to int) {
	if from < 0 || from >= len(m.Clients) || to < 0 || to >= len(m.Clients) || from == to {
		return
	}

	client := m.Clients[from]
	if from < to {
		copy(m.Clients[from:to], m.Clients[from+1:to+1])
	} else {
		copy(m.Clients[to+1:from+1], m.Clients[to:from])
	}

	m.Clients[to] = client
	m.Focus = to
	m.setClientsBounds()
	m.Draw()
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
