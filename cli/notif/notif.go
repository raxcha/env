package notif

import (
	"env/cli"
	"env/filesystem"
	"env/settings"
	"env/types"
	"env/utils"
	"strconv"
	"strings"
	"time"
)

type Notif struct {
	Parent     cli.Parent
	Settings   *settings.Settings
	Filesystem *filesystem.Filesystem
	On         bool
	Sizes      types.Dimensions
	Stack      []*types.Patch
}

func CreateNotif(parent cli.Parent) *Notif {
	return &Notif{Parent: parent, Settings: parent.GetSettings(), Filesystem: parent.GetFilesystem(), On: true}
}

func (n *Notif) AddPatch(patch *types.Patch) {
	if patch != nil {
		n.Stack = append(n.Stack, patch)
	}
	n.Refresh()
}

func (n *Notif) Refresh() {
	pending := n.Stack[:0]
	for _, patch := range n.Stack {
		select {
		case <-patch.Done:
		default:
			pending = append(pending, patch)
		}
	}
	n.Stack = pending
}

func (n *Notif) Resize(newsizes types.Dimensions) { n.Sizes = newsizes }

func (n *Notif) Input(input *types.Input) bool {
	n.Refresh()
	if len(n.Stack) == 0 {
		return false
	}
	patch := n.Stack[len(n.Stack)-1]
	switch input.Description {
	case "ctrl+S", "ctrl+s":
		n.Filesystem.ConfirmPatch(patch)
	case "ctrl+Z", "ctrl+z":
		n.Filesystem.CancelPatch(patch)
	default:
		return false
	}
	n.Refresh()
	return true
}

func (n *Notif) Draw() *types.Queue {
	n.Refresh()
	if !n.On || len(n.Stack) == 0 {
		return n.Settings.GenerateQueue(n.Sizes.Full, nil, false)
	}
	lines := []string{}
	width := 0
	for _, patch := range n.Stack {
		remaining := max(0, time.Until(patch.Deadline))
		seconds := int((remaining + time.Second - 1) / time.Second)
		for _, description := range patch.Description {
			label := strings.Join(description, " ")
			if len(description) == 3 {
				label = strconv.Quote(description[0]) + " | " + description[1] + " > " + description[2]
			}
			text := " " + strconv.Itoa(seconds) + "s " + label + " "
			width = max(width, utils.VisibleLength(text))
			lines = append(lines, "§YX0 "+text)
		}
	}
	sizes := n.Sizes
	sizes.Size[0] = min(width, max(0, n.Sizes.Size[0]))
	sizes.Size[1] = min(len(lines), max(0, n.Sizes.Size[1]))
	sizes.Pos[0] += n.Sizes.Size[0] - sizes.Size[0]
	frame := n.Settings.GenerateFrame(sizes, lines, 0, []int{0, 0, 0, 0})
	return n.Settings.GenerateQueue(sizes.Full, []*types.Frame{frame}, false)
}
