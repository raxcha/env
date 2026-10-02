package notif

import (
	"env/filesystem"
	"env/settings"
	"env/types"
	"os"
	"path/filepath"
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

func TestDoomNotificationConfirmAndCancel(t *testing.T) {
	for _, stage := range []string{"draft", "local", "move"} {
		for _, key := range []string{"ctrl+s", "ctrl+z"} {
			t.Run(stage+"/"+key, func(t *testing.T) {
				root := t.TempDir()
				fs, err := filesystem.Open(root)
				if err != nil {
					t.Fatal(err)
				}
				page := fs.NewDraft("note", nil)
				if stage != "draft" {
					if err := os.WriteFile(filepath.Join(root, "note"), []byte("original"), 0600); err != nil {
						t.Fatal(err)
					}
					page.Og = &types.Page{Path: "note"}
				}
				if stage == "move" {
					page.Path = "renamed"
				}
				page.Stage = stage
				fs.Cache[page.Path] = page
				fs.DoomPage(page)
				var patch *types.Patch
				select {
				case patch = <-fs.Patch:
				case <-time.After(time.Second):
					t.Fatal("missing notification")
				}
				t.Cleanup(func() { fs.CancelPatch(patch) })
				n := &Notif{Filesystem: fs}
				n.AddPatch(patch)
				if !n.Input(&types.Input{Description: key}) {
					t.Fatal("shortcut not handled")
				}
				_, err = os.Stat(filepath.Join(root, "note"))
				if key == "ctrl+s" {
					if !os.IsNotExist(err) || fs.Cache[page.Path] != nil {
						t.Fatal("confirmation did not delete page")
					}
					fs.CancelPatch(patch)
					if page.Stage != "doom" {
						t.Fatal("completed deletion was undone")
					}
				} else {
					if page.Stage != stage || fs.Cache[page.Path] != page {
						t.Fatal("cancellation did not restore stage")
					}
					if stage != "draft" && err != nil {
						t.Fatal("cancellation deleted file")
					}
					fs.ConfirmPatch(patch)
					if fs.Cache[page.Path] != page {
						t.Fatal("cancelled deletion was applied")
					}
				}
				if len(n.Stack) != 0 {
					t.Fatal("notification remains after action")
				}
			})
		}
	}
}
