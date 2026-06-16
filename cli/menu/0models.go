package menu

import (
	"sort"
	"strings"

	"env/engine"
	"env/routines"
	"env/utilities"
)

// menu item ...
type Item struct {
	Name       string
	Command    string
	Pin        bool
	PromptOnly bool
	Right      string
	Score      float64
	Preview    func(prompt string)
	Run        func(prompt string)
}

type Menu struct {
	Utilities *utilities.Utilities

	Title  string
	Mode   string
	Bounds engine.Boundaries

	AllItems []Item
	Items    []Item
	Prompt   string
	Selected int

	OnCancel func()
	On       bool
}

func CreateMenu(u *utilities.Utilities) *Menu {
	return &Menu{
		Bounds:    engine.Boundaries{},
		Utilities: u,
		Title:     "menu",
		Mode:      "main",
		Items:     []Item{},
		AllItems:  []Item{},
		Prompt:    "",
		Selected:  0,
		On:        false,
	}
}

func (m *Menu) Open(title string, items []Item) {

	m.Title = title
	m.Mode = strings.ToLower(strings.TrimSpace(title))
	m.AllItems = append([]Item{}, items...)
	m.Prompt = ""
	m.Selected = 0
	m.OnCancel = nil
	m.On = len(items) > 0
	m.filterItems()
}

func (m *Menu) Close() {
	m.On = false
	m.Items = []Item{}
	m.AllItems = []Item{}
	m.Prompt = ""
	m.Selected = 0
	m.OnCancel = nil
}

func (m *Menu) Cancel() {
	onCancel := m.OnCancel
	m.Close()
	if onCancel != nil {
		onCancel()
	}
}

func (m *Menu) Resize(b *engine.Boundaries) {
	if b == nil {
		return
	}

	b.ActualPos = b.Pos
	b.ActualSize = b.Size
	m.Bounds = *b
}

func (m *Menu) Input(input *routines.Input) {
	if input == nil || !m.On {
		return
	}

	switch input.Key {
	case "esc":
		m.Cancel()

	case "up":
		m.move(-1)

	case "down":
		m.move(1)

	case "char":
		m.Prompt += string(input.Char)
		m.filterItems()

	case "backspace":
		m.backspacePrompt()

	case "enter", "ctrl+enter":
		m.runSelected()
	}
}

func (m *Menu) move(delta int) {
	if len(m.Items) == 0 {
		m.Selected = 0
		return
	}

	m.Selected += delta

	if m.Selected < 0 {
		m.Selected = 0
	}

	if m.Selected >= len(m.Items) {
		m.Selected = len(m.Items) - 1
	}

	m.previewSelected()
}

func (m *Menu) runSelected() {
	if len(m.Items) == 0 || m.Selected < 0 || m.Selected >= len(m.Items) {
		return
	}

	item := m.Items[m.Selected]
	prompt := strings.TrimSpace(m.Prompt)

	m.Close()

	if item.Run != nil {
		item.Run(prompt)
	}
}

func (m *Menu) backspacePrompt() {
	if m.Prompt == "" {
		return
	}

	runes := []rune(m.Prompt)
	if len(runes) == 0 {
		return
	}

	m.Prompt = string(runes[:len(runes)-1])
	m.filterItems()
}

func (m *Menu) filterItems() {
	query := strings.ToLower(strings.TrimSpace(m.Prompt))

	if query == "" {
		m.Items = []Item{}
		for _, item := range m.AllItems {
			if item.PromptOnly {
				continue
			}
			m.Items = append(m.Items, item)
		}
	} else {
		m.Items = []Item{}
		fallbacks := []Item{}

		for _, item := range m.AllItems {
			item.Score = m.scoreItem(query, item)
			if item.PromptOnly {
				fallbacks = append(fallbacks, item)
				continue
			}
			if item.Pin {
				m.Items = append(m.Items, item)
				continue
			}

			if item.Score > -0.75 {
				m.Items = append(m.Items, item)
			}
		}

		if len(m.Items) == 0 {
			m.Items = fallbacks
		}

		sort.SliceStable(m.Items, func(i, j int) bool {
			if m.Items[i].Pin != m.Items[j].Pin {
				return m.Items[i].Pin
			}
			if m.Items[i].Score != m.Items[j].Score {
				return m.Items[i].Score > m.Items[j].Score
			}
			return strings.ToLower(m.Items[i].Name) < strings.ToLower(m.Items[j].Name)
		})
	}

	m.Selected = 0
	if len(m.Items) == 0 {
		m.Selected = -1
	}
	m.previewSelected()
}

func (m *Menu) previewSelected() {
	if len(m.Items) == 0 || m.Selected < 0 || m.Selected >= len(m.Items) {
		return
	}

	item := m.Items[m.Selected]
	if item.Preview == nil {
		return
	}

	item.Preview(strings.TrimSpace(m.Prompt))
}

func (m *Menu) scoreItem(query string, item Item) float64 {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return 0
	}

	name := strings.ToLower(item.Name)
	command := strings.ToLower(strings.TrimSpace(item.Command))

	if command == "" {
		command = name
	}

	text := strings.TrimSpace(name + " " + command)
	if text == "" {
		return -1
	}

	switch {
	case command == "__prompt__":
		return 1
	case menuCommandMatchesQuery(command, query):
		return 0.95
	case name == query || command == query:
		return 1
	case strings.HasPrefix(name, query) || strings.HasPrefix(command, query):
		return 0.75
	case strings.Contains(name, query) || strings.Contains(command, query):
		return 0.45
	case menuSubsequence(query, text):
		return 0
	default:
		return -1
	}
}

func menuCommandMatchesQuery(command string, query string) bool {
	command = strings.ToLower(strings.TrimSpace(command))
	query = strings.ToLower(strings.TrimSpace(query))
	if command == "" || query == "" {
		return false
	}

	for _, key := range strings.Fields(command) {
		key = strings.TrimSuffix(strings.TrimSpace(key), ":")
		if key == "" {
			continue
		}
		if query == key || strings.HasPrefix(query, key+":") {
			return true
		}
	}

	return false
}

func menuSubsequence(query string, text string) bool {
	if query == "" {
		return true
	}

	q := []rune(query)
	t := []rune(text)
	j := 0

	for i := 0; i < len(t) && j < len(q); i++ {
		if t[i] == q[j] {
			j++
		}
	}

	return j == len(q)
}
