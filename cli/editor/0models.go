package editor

import (
	"env/cli"
	"env/cli/sidebar"
	"env/engine"
	"env/filesystem"
	"env/utilities"
)

type Editor struct {
	Parent    cli.Parent
	Bounds    *engine.Boundaries
	Utilities *utilities.Utilities

	Name string
	Path string
	Page *filesystem.Page

	Content        []string
	AppliedContent []string
	Cursor         []int
	WantedX        int
	Undo           []EditorState
	Redo           []EditorState
	Clipboard      []string

	Numbers   bool
	Wrap      bool
	Zenmode   bool
	PanelMode string

	On    bool
	Float bool

	Sidebar *sidebar.Sidebar
}

type EditorState struct {
	Content []string
	Cursor  []int
	WantedX int
}

var sharedClipboard []string
