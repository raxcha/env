package status

import (
	"env/cli"
	"env/settings"
	"env/types"
	"env/utils"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Status struct {

	Parent   cli.Parent
	Settings *settings.Settings

	Sizes   types.Dimensions

	Meta    bool
	Entries []entry
}

type entry struct {

	Label string
	Pos   int
}

func CreateStatus(parent cli.Parent) *Status {

	s := Status{

		Parent:   parent,
		Settings: parent.GetSettings(),

		Sizes:    types.Dimensions{},

		Meta:     false,
		Entries:   []entry{},
	}

	s.Refresh()

	return &s
}

func (s *Status) Refresh() {

	inputmode := s.Parent.GetMode() + ":" + s.Parent.GetLayout()

	client := ""

	if s.Parent.GetClientsSize() > 0 {
		
		client = s.Parent.GetFocusedClient().GetType() + ":" + s.Parent.GetFocusedClient().GetLongLabel()
	}

	clock := time.Now().Format("15:04")

	full := s.Sizes.Full

	size := strconv.Itoa(full[0]) + "x" + strconv.Itoa(full[1])

	s.Entries = []entry{

		{Label: inputmode, Pos: -1},
		{Label: size, Pos: 2},
		{Label: client, Pos: 0},
		{Label: clock, Pos: 1},

	}
}

func (s *Status) Resize(newsizes types.Dimensions) {
	s.Sizes = newsizes
}

func (s *Status) Input(newinput *types.Input) {
	s.Meta = newinput.Meta
}

func (s *Status) Draw() *types.Queue {

	s.Refresh()

	left, center, right := []entry{}, []entry{}, []entry{}
	total := 0

	for _, en := range s.Entries {

		total += utils.VisibleLength(en.Label) + 4

		if total >= s.Sizes.Size[0] { break }

		if en.Pos < 0 {

			left = append(left, en)

		} else if en.Pos == 0 {

			center = append(center, en)

		} else if en.Pos > 0 {

			right = append(right, en)
		}
	}

	sort.Slice(left, func(i, j int) bool {
		return left[i].Pos < left[j].Pos
	})

	sort.Slice(center, func(i, j int) bool {
		return center[i].Pos < center[j].Pos
	})

	sort.Slice(right, func(i, j int) bool {
		return right[i].Pos < right[j].Pos
	})

	leftt, centerr, rightt := "", "", ""

	for _, en := range append(append(left, center...), right...) {

		if en.Pos < 0 {

			leftt += " -" + en.Label + "- "
			
		} else if en.Pos == 0 {

			centerr += " -" + en.Label + "- "

		} else if en.Pos > 0 {

			rightt += " -" + en.Label + "- "

		}
	}

	remaining := s.Sizes.Size[0] - utils.VisibleLength(leftt+centerr+rightt)

	half1 := remaining / 2
	half2 := remaining - half1

	line := leftt + strings.Repeat(" ", half1) + centerr + strings.Repeat(" ", half2) + rightt

	if !s.Meta {

		line = "§xy0 ‹b " + line + "›b "

	} else {

		line = "§yx0 ‹b " + line + "›b "
		
	}

	frame := s.Settings.GenerateFrame(s.Sizes, []string{line}, 0, []int{0, 0, 0, 0})

	return s.Settings.GenerateQueue(s.Sizes.Full, []*types.Frame{frame}, false)
}
