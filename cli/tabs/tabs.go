package tabs

import (
	"env/cli"
	"env/settings"
	"env/types"
	"env/utils"
	"strings"
)

type Tabs struct {

	Parent   cli.Parent
	Settings *settings.Settings

	On bool
	Sizes   types.Dimensions

	Focus   int
	Mode    string
	Meta    bool

	Entries []entry
}

type entry struct {

	Short string
	Long string

	Focused bool

	Sizes   types.Dimensions
}

func CreateTabs(parent cli.Parent) *Tabs {

	t := Tabs{

		Parent:   parent,
		Settings: parent.GetSettings(),

		On:       false,
		Sizes:    types.Dimensions{},

		Focus:    0,
		Mode:     "",
		Meta:     true,

		Entries:  []entry{},
	}

	t.refresh()

	return &t
}

func (t *Tabs) refresh() {

	t.Focus = t.Parent.GetFocus()
	t.Mode = t.Parent.GetMode()
	t.Sizes = *t.Parent.GetSizes()

	oldSizes := []types.Dimensions{}
	for _, en := range t.Entries {
		oldSizes = append(oldSizes, en.Sizes)
	}

	entries := []entry{}
	for i, client := range t.Parent.GetClients() {

		newLabel := entry{Short: "", Long: "", Focused: false, Sizes: types.Dimensions{}}

		if i <= len(oldSizes)-1 {
			newLabel.Sizes = oldSizes[i]
		}

		if i == t.Focus {
			newLabel.Focused = true
		}

		newLabel.Short = client.GetShortLabel()
		newLabel.Long = client.GetLongLabel() + "[" + string("ASDFGHJKLQWERTYUIOPZXCVBNM1234567890"[i]) + "]"
		
		entries = append(entries, newLabel)
	}

	t.Entries = entries
}

func (t *Tabs) Resize(newsizes types.Dimensions) {

	t.Sizes = newsizes
}

func (t *Tabs) Draw() *types.Queue {

	t.refresh()

	if len(t.Entries) == 0 {
		return &types.Queue{}
	}

	switch t.Mode {

	case "monocle":
		return t.drawMonocle()

	// case "fibonacci":
		// return t.drawFibonacci()

	default:
		return &types.Queue{}
	}
}

func (t *Tabs) drawMonocle() *types.Queue {

	start, end := 0, 0
	top, bot := " λ", "  "
	if t.On {
		top = " Δ"
	}

	for _, en := range t.Entries {

		enel := " " + en.Short + " "

		if en.Focused {

			enel = "‹b " + enel + "›b "
			bot += " " + strings.Repeat("━", utils.VisibleLength(enel)-2) + " "
			start = utils.VisibleLength(top)
			end = start + utils.VisibleLength(enel)

		}

		top += enel

		bot += strings.Repeat(" ", utils.VisibleLength(enel))
	}

	total := utils.VisibleLength(top)

	if total > t.Sizes.Size[0] {

		max := total - t.Sizes.Size[0]
		place := start * max / total

		if end-place > t.Sizes.Size[0] {
			place = end - t.Sizes.Size[0]
		}
		if start < place {
			place = start
		}

		top = utils.VisibleSlice(top, place, place+t.Sizes.Size[0])
		bot = utils.VisibleSlice(bot, place, place+t.Sizes.Size[0])
	}

	if t.Meta {
		top = "§yx0 " + top
		bot = "§xy0 " + bot
	} else {
		top = "§xy0 " + top
		bot = "§xy0 " + bot
	}

	frame := t.Settings.GenerateFrame(t.Sizes, []string{top, bot}, 0, []int{0, 0, 0, 0})

	return t.Settings.GenerateQueue(t.Sizes.Full, []*types.Frame{frame}, false)
}

func (t *Tabs) Input(newinput *types.Input, real bool) {

	if !real {
		t.Meta = newinput.Meta
		return
	}

	switch newinput.Description {

	case "char":
		t.quickSwitch(newinput.Char)

	case "left":
		t.moveFocus(-1)

	case "right":
		t.moveFocus(1)

	case "ctrl+left":
		t.moveClient(-1)

	case "ctrl+right":
		t.moveClient(1)

	case "tab":
		t.Parent.SetFocus(t.Focus)
		t.On = false

	case "ctrl+Q":
		t.closeCurrent()
	}
}

func (t *Tabs) closeCurrent() {

	t.Parent.CloseClient(t.Parent.GetFocus())
	t.refresh()
}

func (t *Tabs) quickSwitch(char rune) {

	newIndex := strings.IndexRune("asdfghjklqwertyuiopzxcvbnm1234567890"[:t.Parent.GetClientsSize()], char)
	if newIndex != -1 {
		t.Parent.SetFocus(newIndex)
		t.On = false
	}
}

func (t *Tabs) moveFocus(delta int) {

	final := t.Parent.GetFocus() + delta
	if final >= 0 && final <= t.Parent.GetClientsSize()-1 {
		t.Parent.SetFocus(final)
	}
}

func (t *Tabs) moveClient(delta int) {

	from := t.Parent.GetFocus()
	to := from + delta
	if to < 0 || to > t.Parent.GetClientsSize()-1 {
		return
	}

	t.Parent.MoveClient(from, to)
	t.Parent.SetFocus(to)
	t.refresh()
}