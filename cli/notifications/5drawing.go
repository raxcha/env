package notifications

import (
	"env/engine"
	"env/routines"
	"env/utilities"
	"fmt"
	"strings"
	"time"
)

func (n *Notifications) Draw() *engine.Queue {
	if n == nil || !n.On || n.Utilities == nil || n.Bounds == nil {
		return nil
	}

	fs := n.Bounds.Fullsize
	if len(fs) < 2 {
		return nil
	}

	n.mu.Lock()
	items := make([]*Notification, len(n.Items))
	copy(items, n.Items)
	n.mu.Unlock()

	if len(items) == 0 {
		return nil
	}

	var merged *engine.Frame
	y := n.Bounds.Pos[1]

	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		boxFrame := n.drawNotificationBox(item, y)
		if boxFrame == nil {
			continue
		}

		if merged == nil {
			merged = boxFrame
		} else {
			next := n.Utilities.MergeFrames(*boxFrame, *merged)
			merged = &next
		}

		y += notificationFrameHeight()
		if y >= fs[1] {
			break
		}
	}

	if merged == nil {
		return nil
	}

	return &engine.Queue{
		Size:   fs,
		Frames: []engine.Frame{*merged},
		Cycle:  false,
	}
}

func (n *Notifications) drawNotificationBox(item *Notification, y int) *engine.Frame {
	fs := n.Bounds.Fullsize
	secs := int(time.Until(item.Deadline).Seconds())
	if secs < 0 {
		secs = 0
	}

	line := fmt.Sprintf("%s  ctrl+z %ds", item.Message, secs)
	w := n.Utilities.VisibleLength(line) + 4
	if w < 30 {
		w = 30
	}
	if w > 54 {
		w = 54
	}

	contentW := w - 4
	if n.Utilities.VisibleLength(line) > contentW {
		line = n.Utilities.CutVisible(line, contentW)
	}
	if pad := contentW - n.Utilities.VisibleLength(line); pad > 0 {
		line += strings.Repeat(" ", pad)
	}

	boxed := n.Utilities.Box([]string{line}, utilities.BoxOpts{
		W:       w,
		H:       notificationFrameHeight(),
		Padding: utilities.Padding{Top: 0, Right: 1, Bottom: 0, Left: 1},
		Title:   "pending",
	})

	for i, line := range boxed {
		boxed[i] = "§AB0 " + line
	}

	x := fs[0] - w - 1
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	return n.Utilities.GenerateFrame(
		engine.Boundaries{
			Fullsize:   fs,
			Pos:        routines.Bound{x, y},
			Size:       routines.Bound{w, notificationFrameHeight()},
			ActualPos:  routines.Bound{x, y},
			ActualSize: routines.Bound{w, notificationFrameHeight()},
		},
		boxed,
		0,
	)
}

func notificationFrameHeight() int {
	return 3
}
