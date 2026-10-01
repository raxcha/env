package notif

import (
	"env/filesystem"
	"env/settings"
	"env/types"
	"testing"
	"time"
)

func TestPendingPatchDisplay(t *testing.T) {
	fs := &filesystem.Filesystem{}
	n := &Notif{Settings: &settings.Settings{}, Filesystem: fs, On: true}
	n.Resize(types.Dimensions{Full: types.Size{80, 24}, Size: types.Size{80, 23}})
	patch := &types.Patch{Description: [][]string{{"proj/page", "edit", "local"}}, Deadline: time.Now().Add(5 * time.Second), Done: make(chan struct{})}
	n.AddPatch(patch)
	q := n.Draw()
	if len(q.Frames) != 1 {
		t.Fatal("expected one notification frame")
	}
	frame := q.Frames[0]
	if frame.Sizes.Pos[1] != 0 || frame.Sizes.Pos[0]+frame.Sizes.Size[0] != 80 {
		t.Fatalf("notification not at top right: %+v", frame.Sizes)
	}
	text := ""
	for x := frame.Sizes.Pos[0]; x < 80; x++ {
		text += string(frame.Cells[x].Char)
	}
	if text != ` 5s "proj/page" | edit > local ` {
		t.Fatalf("notification = %q", text)
	}
	deadline := patch.Deadline
	n.Draw()
	if patch.Deadline != deadline {
		t.Fatal("drawing reset the deadline")
	}
	if !n.Input(&types.Input{Description: "ctrl+s"}) {
		t.Fatal("pending patch did not handle apply shortcut")
	}
	if len(n.Draw().Frames) != 0 || len(n.Stack) != 0 {
		t.Fatal("completed notification still visible")
	}
	if n.Input(&types.Input{Description: "ctrl+s"}) {
		t.Fatal("completed patch swallowed save shortcut")
	}
}
