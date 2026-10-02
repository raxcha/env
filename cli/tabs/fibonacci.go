package tabs

import (
	"env/types"
	"env/utils"
)

// Keep each header in its own queue so its bounds exclude other panes.
func (t *Tabs) DrawFibonacci(panes []types.Dimensions) []*types.Queue {
	t.refresh()
	queues := make([]*types.Queue, 0, len(panes))
	for i, pane := range panes {
		if i >= len(t.Entries) || pane.Size[0] <= 0 || pane.Size[1] <= 0 {
			continue
		}
		pane.Size[1] = 1
		en := t.Entries[i]
		label := en.Short
		if t.Meta {
			label = en.Long
		}
		label = utils.VisibleSlice(" "+label+" ", 0, pane.Size[0])
		style := "§xy0 "
		if en.Focused {
			style = "§YX0 ‹b "
		}
		frame := t.Settings.GenerateFrame(pane, []string{style + label}, 0, []int{0, 0, 0, 0})
		queues = append(queues, t.Settings.GenerateQueue(pane.Full, []*types.Frame{frame}, false))
	}
	return queues
}
