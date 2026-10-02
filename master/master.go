package master

import (
	"time"

	"env/types"
	"env/utils"

	"env/engine"
	"env/filesystem"
	"env/processing"
	"env/settings"

	"env/cli"
	"env/cli/editor"
	"env/cli/empty"
	"env/cli/launcher"
	"env/cli/notif"
	"env/cli/status"
	"env/cli/tabs"
)

type Master struct {
	Processing *processing.Processing
	Engine     *engine.Engine
	Filesystem *filesystem.Filesystem
	Settings   *settings.Settings

	Tabs     *tabs.Tabs
	Status   *status.Status
	Launcher *launcher.Launcher
	Notif    *notif.Notif
	Empty    *empty.Empty

	Sizes types.Dimensions

	Clients []cli.Client
	Focus   int

	Mode   string
	Layout string

	Sidebar   editor.Sidebar
	Clipboard string
}

func CreateMaster(arg string) *Master {

	m := Master{

		Processing: &processing.Processing{},
		Engine:     &engine.Engine{},
		Filesystem: &filesystem.Filesystem{},
		Settings:   &settings.Settings{},

		Tabs:     &tabs.Tabs{},
		Status:   &status.Status{},
		Launcher: &launcher.Launcher{},
		Notif:    &notif.Notif{},
		Empty:    &empty.Empty{},

		Sizes: types.Dimensions{},

		Clients: []cli.Client{},
		Focus:   0,

		Mode:   "monocle",
		Layout: "visual",
	}

	m.Processing = processing.CreateProcessing()
	m.Engine = engine.CreateEngine()
	m.Filesystem = filesystem.CreateFilesystem(false)
	m.Settings = settings.CreateSettings(m.Filesystem.Options)

	m.Tabs = tabs.CreateTabs(&m)
	m.Status = status.CreateStatus(&m)
	m.Launcher = launcher.CreateLauncher(&m)
	m.Notif = notif.CreateNotif(&m)
	m.Empty = empty.CreateEmpty(&m)

	m.AddClients(arg, "after", false)

	m.startListening()

	return &m
}

func (m *Master) AddClients(arg string, mode string, follow bool) {

	args := utils.SplitMore(arg, " ")

	newOnes := []cli.Client{}

	for _, arg := range args {

		key, spec := utils.SplitTwo(arg, ":")

		if spec == "" {
			spec = "."
		}

		var newOne cli.Client

		switch key {

		case "today":
			newOne = editor.CreateEditor(m, "log/"+time.Now().Format("2006.01.02"))

		case "editor":
			newOne = editor.CreateEditor(m, spec)

		default:
			newOne = editor.CreateEditor(m, key)

		}

		existingIdx := -1

		for i, existing := range m.Clients {

			if existing.GetType() == newOne.GetType() && existing.GetSpec() == newOne.GetSpec() {

				existingIdx = i
				break
			}
		}

		if existingIdx != -1 {

			m.Clients[existingIdx] = newOne

			if follow {

				m.Focus = existingIdx
			}

			continue
		}

		newOnes = append(newOnes, newOne)
	}

	idx := -1

	switch mode {

	case "after":

		insertAt := min(m.Focus+1, len(m.Clients))
		m.Clients = append(append(m.Clients[:insertAt:insertAt], newOnes...), m.Clients[insertAt:]...)
		idx = insertAt

	case "end":

		m.Clients = append(m.Clients, newOnes...)
		idx = len(m.Clients) - 1
	}

	if follow && len(newOnes) == 1 {

		m.Focus = idx
	}

	m.applySizes()
}

func (m *Master) startListening() {

	go func() {

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {

			case newSizes := <-m.Processing.Sizes:
				m.Resizing(newSizes)

			case newInput := <-m.Processing.Input:
				m.Inputting(newInput)

			case patch := <-m.Filesystem.Patch:
				m.Notif.AddPatch(patch)
				m.Drawing()

			case <-ticker.C:
				if len(m.Notif.Stack) > 0 {
					m.Notif.Refresh()
					m.Drawing()
				}

			}
		}
	}()
}

func (m *Master) Inputting(newinput *types.Input) {

	m.Tabs.Input(newinput, false)
	m.Status.Input(newinput)

	if m.Notif.Input(newinput) {
		m.Drawing()
		return
	}

	if m.Launcher.On {

		if newinput.Description == "esc" {
			m.Launcher.Close()
		} else {
			m.Launcher.Input(newinput)
		}

	} else if newinput.Meta {

		switch newinput.Description {

		case "enter":
			m.Launcher.Open("default")

		case "backspace":
			m.cycleMode()
			m.Resizing(&m.Sizes)

		case "char":
			if newinput.Char == ' ' {
				m.cycleLayout()
			} else {
				m.Tabs.Input(newinput, true)
			}

		default:
			m.Tabs.Input(newinput, true)
		}

	} else if len(m.Clients) > 0 {
		if m.Layout == "select" || m.Layout == "split" {
			m.Sidebar.Input(newinput)
		} else {
			m.Clients[m.Focus].Input(*newinput)
		}
	}

	m.Drawing()

}

func (m *Master) cycleLayout() {
	switch m.Layout {
	case "visual":
		m.Layout = "select"
	case "select":
		m.Layout = "split"
	case "split":
		m.Layout = "visual"
	}
	m.Resizing(&m.Sizes)
}

func (m *Master) cycleMode() {
	switch m.Mode {
	case "monocle":
		m.Mode = "fibonacci"
	case "fibonacci":
		m.Mode = "monocle"
	}
	m.Resizing(&m.Sizes)
}

func (m *Master) Resizing(newsizes *types.Dimensions) {

	m.Sizes = *newsizes
	m.applySizes()

	m.Drawing()
}

func (m *Master) applySizes() {

	m.resizeStatus()
	m.resizeLauncher()
	m.resizeNotif()
	m.resizeEmpty()

	switch m.Mode {
	case "monocle":
		m.resizeTabs()
		m.resizeMonocle()
	case "fibonacci":
		m.resizeFibonacci()
	}
}

func (m *Master) resizeTabs() {

	sizes := m.Sizes
	sizes.Size[1] = 2
	m.Tabs.Resize(sizes)
}

func (m *Master) resizeStatus() {

	sizes := m.Sizes
	sizes.Size[1] = 1
	sizes.Pos[1] = sizes.Full[1] - 1
	m.Status.Resize(sizes)
}

func (m *Master) resizeLauncher() {

	sizes := m.Sizes
	sizes.Size[1] -= 3
	sizes.Pos[1] += 2
	m.Launcher.Resize(sizes)
}

func (m *Master) resizeNotif() {

	sizes := m.Sizes
	sizes.Size[1] -= 1
	m.Notif.Resize(sizes)
}

func (m *Master) resizeMonocle() {

	sizes := m.Sizes
	sizes.Size[1] -= 3
	sizes.Pos[1] += 2
	sizes = m.sidebarArea(sizes)

	for _, client := range m.Clients {
		client.Resize(sizes)
	}
}

func (m *Master) resizeEmpty() {

	sizes := m.Sizes
	sizes.Size[1] -= 3
	sizes.Pos[1] += 2
	m.Empty.Resize(sizes)
}
func (m *Master) Drawing() {

	if m.Sizes.Full[0] == 0 || m.Sizes.Full[1] == 0 {
		return
	}

	m.applySizes()
	queues := []*types.Queue{}
	if m.Mode == "fibonacci" {
		queues = append(queues, m.drawFibonacciDividers())
		panes, _ := m.fibonacciLayout()
		queues = append(queues, m.Tabs.DrawFibonacci(panes)...)
	}

	queues = append(queues, m.Tabs.Draw())
	queues = append(queues, m.Status.Draw())

	if len(m.Clients) == 0 {
		queues = append(queues, m.Empty.Draw())
	}

	switch m.Mode {
	case "monocle":
		if len(m.Clients) > 0 {
			queues = append(queues, m.Clients[m.Focus].Draw())
		}
	case "fibonacci":
		for _, client := range m.Clients {
			queues = append(queues, client.Draw())
		}
	}

	if m.Layout == "select" || m.Layout == "split" {
		queues = append(queues, m.Sidebar.Draw(m.Settings))
	}

	queues = append(queues, m.Notif.Draw())
	queues = append(queues, m.Launcher.Draw())

	m.Engine.CallEngine(*m.Settings.MergeQueues(queues...))
}

func (m *Master) GetClients() []cli.Client {
	return m.Clients
}

func (m *Master) GetFocus() int {
	return m.Focus
}

func (m *Master) GetMode() string {
	return m.Mode
}

func (m *Master) GetLayout() string {
	return m.Layout
}

func (m *Master) GetFocusedClient() cli.Client {
	if m.Focus < 0 || m.Focus >= len(m.Clients) {
		return nil
	}
	return m.Clients[m.Focus]
}

func (m *Master) SetFocus(focus int) {
	m.Focus = focus
}

func (m *Master) GetClientsSize() int {
	return len(m.Clients)
}

func (m *Master) MoveClient(from int, to int) {

	if from == to || from < 0 || from >= len(m.Clients) || to < 0 || to >= len(m.Clients) {
		return
	}

	client := m.Clients[from]

	m.Clients = append(m.Clients[:from], m.Clients[from+1:]...)
	m.Clients = append(m.Clients[:to], append([]cli.Client{client}, m.Clients[to:]...)...)
	if m.Mode == "fibonacci" {
		m.resizeFibonacci()
	}
}

func (m *Master) CloseClient(idx int) {

	if idx < 0 || idx >= len(m.Clients) {
		return
	}

	m.Clients = append(m.Clients[:idx], m.Clients[idx+1:]...)

	if m.Focus >= len(m.Clients) {
		m.Focus = len(m.Clients) - 1
	}
	if m.Focus < 0 {
		m.Focus = 0
	}

	m.applySizes()
}

func (m *Master) GetSettings() *settings.Settings {
	return m.Settings
}

func (m *Master) GetFilesystem() *filesystem.Filesystem {
	return m.Filesystem
}
func (m *Master) GetEngine() *engine.Engine {
	return m.Engine
}

func (m *Master) GetSizes() *types.Dimensions {
	return &m.Sizes
}

func (m *Master) GetClipboard() string {
	return m.Clipboard
}

func (m *Master) SetClipboard(content string) {
	m.Clipboard = content
}

// sidebarArea reserves space once, before laying out individual editors.
func (m *Master) sidebarArea(area types.Dimensions) types.Dimensions {
	editors := []*editor.Editor{}
	for i, client := range m.Clients {
		if e, ok := client.(*editor.Editor); ok {
			e.SharedSidebar = true
			if m.Mode == "fibonacci" || i == m.Focus {
				editors = append(editors, e)
			}
		}
	}
	if m.Layout != "select" && m.Layout != "split" {
		return area
	}
	content := m.Sidebar.Prepare(editors, area)
	if m.Layout == "select" && len(editors) > 0 {
		m.Sidebar.Sizes = m.Sizes
		m.Sidebar.Sizes.Size[0] = max(0, m.Sizes.Size[0])
		m.Sidebar.Sizes.Size[1] = max(0, m.Sizes.Size[1]-1)
		if m.Mode == "monocle" {
			tabHeight := min(2, m.Sidebar.Sizes.Size[1])
			m.Sidebar.Sizes.Pos[1] += tabHeight
			m.Sidebar.Sizes.Size[1] -= tabHeight
		}
		content.Size = types.Size{}
	}
	return content
}
