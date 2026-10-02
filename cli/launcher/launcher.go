package launcher

import (
	"env/cli"
	"env/settings"
	"env/types"
	"env/utils"
	"sort"
	"strings"
)

type Launcher struct {
	Parent   cli.Parent
	Settings *settings.Settings

	On    bool
	Sizes types.Dimensions

	Spec     string
	AllItems []item
	Items    []item

	Focus         int
	Prompt        string
	previousTheme types.Theme
	Error         string
}

type item struct {
	Label    string
	Commands []func()
}

func CreateLauncher(parent cli.Parent) *Launcher {

	m := Launcher{

		Parent:   parent,
		Settings: parent.GetSettings(),

		On:    false,
		Sizes: types.Dimensions{},

		Spec:     "default",
		AllItems: []item{},
		Items:    []item{},

		Focus:  0,
		Prompt: "",
	}

	m.Refresh()

	return &m
}

func (m *Launcher) Open(spec string) {
	if m.On {
		m.Close()
	}

	m.On = true
	m.Spec = spec
	m.Prompt = ""
	m.Focus = 0
	m.Error = ""

	all := []item{}
	switch spec {

	case "default":
		all = []item{
			{Label: "󰃶 open today", Commands: []func(){m.openToday, m.openToday}},
			{Label: "󰱒 do something", Commands: []func(){func() { m.Open("templates") }, func() { m.Open("templates") }}},
			{Label: " open picker", Commands: []func(){func() {}, func() {}}},
			{Label: " open zettelkasten", Commands: []func(){func() {}, func() {}}},
			{Label: " open projects", Commands: []func(){func() {}, func() {}}},
			{Label: " edit .options", Commands: []func(){func() {}, func() {}}},
			{Label: " change theme", Commands: []func(){func() { m.Open("themes") }, func() { m.Open("themes") }}},
			{Label: "󰩈 exit", Commands: []func(){func() {}, func() {}}},
		}
	case "templates":
		all = m.templates()
	case "themes":
		m.previousTheme = m.Settings.Theme
		all = m.themes()
	}

	m.AllItems = all
	m.Refresh()
}

func (m *Launcher) Close() {
	if m.On && m.Spec == "themes" {
		m.Settings.Theme = m.previousTheme
	}
	m.On = false
}

func (m *Launcher) themes() []item {
	names := make([]string, 0, len(m.Settings.Themes))
	for name := range m.Settings.Themes {
		names = append(names, name)
	}
	sort.Strings(names)
	items := make([]item, 0, len(names))
	for _, name := range names {
		choose := func() {
			if err := m.Parent.GetFilesystem().SaveTheme(name); err != nil {
				m.Error = "Could not save theme: " + err.Error()
				return
			}
			if m.Settings.Options == nil {
				m.Settings.Options = map[string]any{}
			}
			m.Settings.Options["theme"] = name
			m.Settings.Theme = m.Settings.Themes[name]
			m.On = false
		}
		items = append(items, item{Label: name, Commands: []func(){choose, choose}})
		if name == m.Settings.Options["theme"] {
			m.Focus = len(items) - 1
		}
	}
	return items
}

func (m *Launcher) previewTheme() {
	if m.On && m.Spec == "themes" && m.Focus >= 0 && m.Focus < len(m.Items) {
		m.Settings.Theme = m.Settings.Themes[m.Items[m.Focus].Label]
	}
}

func (m *Launcher) openToday() {
	m.On = false
	m.Parent.AddClients("today", "after", true)
}

func (m *Launcher) templates() []item {
	fs := m.Parent.GetFilesystem()
	req := fs.NewLoading()
	req.Path = "proj"
	req.Depth = 2
	req.Mode = "fresh"
	req.Cont = false
	ok, root := fs.LoadLocal(req)
	if !ok || root == nil {
		return nil
	}
	items := templateItems(root)
	for i := range items {
		reference := items[i].Label
		use := func() {
			if parent, ok := m.Parent.(interface{ PrepareTemplate(string) bool }); ok && parent.PrepareTemplate(reference) {
				m.On = false
			}
		}
		items[i].Commands = []func(){use, use}
	}
	return items
}

func templateItems(root *types.Page) []item {
	projects := append([]*types.Page(nil), root.Children...)
	sort.SliceStable(projects, func(i, j int) bool {
		a, _ := projects[i].Metadata["priority"].(int)
		b, _ := projects[j].Metadata["priority"].(int)
		if a != b {
			return a > b
		}
		return projects[i].Name < projects[j].Name
	})
	items := []item{}
	for _, project := range projects {
		if !strings.HasPrefix(project.Name, "@") {
			continue
		}
		children := append([]*types.Page(nil), project.Children...)
		sort.SliceStable(children, func(i, j int) bool { return children[i].Name < children[j].Name })
		for _, template := range children {
			if strings.HasPrefix(template.Name, ".") {
				items = append(items, item{Label: project.Name + "/" + template.Name})
			}
		}
	}
	return items
}

func (m *Launcher) Refresh() {

	m.Items = nil
	for _, item := range m.AllItems {
		if strings.Contains(strings.ToLower(item.Label), strings.ToLower(m.Prompt)) {
			m.Items = append(m.Items, item)
		}
	}
	m.Focus = max(0, min(m.Focus, len(m.Items)-1))
	m.previewTheme()

}

func (m *Launcher) Resize(newsizes types.Dimensions) {

	m.Sizes = newsizes
}

func (m *Launcher) Input(newinput *types.Input) {

	switch newinput.Description {

	case "char", "number":
		m.char(newinput.Char)

	case "backspace":
		m.backspace()

	case "up":
		m.move(-1)

	case "down":
		m.move(1)

	case "enter":
		m.enter(0)

	case "ctrl+enter":
		m.enter(1)
	}
}

func (m *Launcher) char(r rune) {
	m.Prompt += string(r)
	m.Focus = 0
	m.Refresh()
}

func (m *Launcher) enter(num int) {
	if m.Focus < 0 || m.Focus >= len(m.Items) {
		return
	}
	commands := m.Items[m.Focus].Commands
	if num >= 0 && num < len(commands) && commands[num] != nil {
		commands[num]()
	}
}

func (m *Launcher) move(delta int) {
	final := m.Focus + delta
	if final >= 0 && final <= len(m.Items)-1 {
		m.Focus = final
		m.previewTheme()
	}
}

func (m *Launcher) backspace() {
	runes := []rune(m.Prompt)
	if len(runes) > 0 {
		m.Prompt = string(runes[:len(runes)-1])
	}
	m.Focus = 0
	m.Refresh()
}

func (m *Launcher) Draw() *types.Queue {

	if !m.On {
		return &types.Queue{}
	}

	m.Refresh()

	lines := []string{"> " + m.Prompt}
	if m.Error != "" {
		lines[0] += " — " + m.Error
	}
	visible := max(1, m.Sizes.Size[1]-5)
	start := max(0, m.Focus-visible+1)
	end := min(len(m.Items), start+visible)
	for _, itm := range m.Items[start:end] {
		lines = append(lines, itm.Label+"     ")
	}

	lines = utils.DrawBox(lines, []int{0, 1, 0, 1})

	if len(m.Items) > 0 && m.Focus >= 0 && m.Focus < len(m.Items) {
		focusRow := m.Focus - start + 2
		lines[focusRow] = "§YX0 ‹b " + lines[focusRow] + "›b "
	}

	boxWidth := 0
	for _, line := range lines {
		if w := utils.VisibleLength(line); w > boxWidth {
			boxWidth = w
		}
	}
	boxHeight := len(lines)

	posX := max(0, (m.Sizes.Size[0]-boxWidth)/2)
	posY := max(0, (m.Sizes.Size[1]-boxHeight)/2)

	sizes := m.Sizes
	sizes.Pos[0] += posX
	sizes.Pos[1] += posY
	sizes.Size[0] = boxWidth
	sizes.Size[1] = boxHeight

	frame := m.Settings.GenerateFrame(sizes, lines, 0, []int{0, 0, 0, 0})
	return m.Settings.GenerateQueue(m.Sizes.Full, []*types.Frame{frame}, false)
}
