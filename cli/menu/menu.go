package menu

import (
	"env/cli"
	"env/settings"
	"env/types"
	"env/utils"
)

type Menu struct {

	Parent   cli.Parent
	Settings *settings.Settings

	On       bool
	Sizes    types.Dimensions

	Spec     string
	AllItems []item
	Items    []item

	Focus    int
	Prompt   string
}

type item struct {
	Label    string
	Commands []func()
}

func CreateMenu(parent cli.Parent) *Menu {

	m := Menu{

		Parent:   parent,
		Settings: parent.GetSettings(),

		On:       false,
		Sizes:    types.Dimensions{},
		
		Spec:     "default",
		AllItems: []item{},
		Items:    []item{},

		Focus:    0,
		Prompt:   "",
	}

	m.Refresh()

	return &m
}

func (m *Menu) Open(spec string) {

	m.On = true
	m.Spec = spec
	m.Prompt = ""

	all := []item{}
	switch spec {

	case "default":
		all = []item{
			{Label: "󰃶 open today", Commands: []func(){func() {}, func() {}}},
			{Label: "󰱒 do something", Commands: []func(){func() {}, func() {}}},
			{Label: " open picker", Commands: []func(){func() {}, func() {}}},
			{Label: " open zettelkasten", Commands: []func(){func() {}, func() {}}},
			{Label: " open projects", Commands: []func(){func() {}, func() {}}},
			{Label: " edit .options", Commands: []func(){func() {}, func() {}}},
			{Label: " change theme", Commands: []func(){func() {}, func() {}}},
			{Label: "󰩈 exit", Commands: []func(){func() {}, func() {}}},
		}
	}

	m.AllItems = all
}

func (m *Menu) Refresh() {

	m.Items = m.AllItems

}

func (m *Menu) Resize(newsizes types.Dimensions) {
	
	m.Sizes = newsizes
}

func (m *Menu) Input(newinput *types.Input) {

	switch newinput.Description {

	case "char":
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

func (m *Menu) char(r rune) {
	m.Prompt += string(r)
}

func (m *Menu) enter(num int) {
	m.Items[m.Focus].Commands[num]()
}

func (m *Menu) move(delta int) {
	final := m.Focus + delta
	if final >= 0 && final <= len(m.Items)-1 {
		m.Focus = final
	}
}

func (m *Menu) backspace() {
	final := len(m.Prompt) - 2
	if final < 0 {
		final = 0
	}
	m.Prompt = m.Prompt[:final]
}

func (m *Menu) Draw() *types.Queue {

	if !m.On {
		return &types.Queue{}
	}

	m.Refresh()

	lines := []string{"> " + m.Prompt}
	for _, itm := range m.Items {
		lines = append(lines, itm.Label+"     ")
	}

	lines = utils.DrawBox(lines, []int{0, 1, 0, 1})

	if len(m.Items) > 0 && m.Focus >= 0 && m.Focus < len(m.Items) {
		focusRow := m.Focus + 2
		lines[focusRow] = "§yx0 ‹b " + lines[focusRow] + "›b "
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
