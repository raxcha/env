package editor

import (
	"env/cli"
	"env/cli/sidebar"
	"env/engine"
	"env/filesystem"
	"path/filepath"
)

func CreateEditor(path string, parent cli.Parent) *Editor {

	if path == "" {
		path = "untitled"
	}

	e := Editor{
		Parent:    parent,
		Bounds:    &engine.Boundaries{},
		Utilities: parent.GetUtilities(),

		Name:           filepath.Base(path),
		Path:           path,
		Page:           draftPageForPath(path),
		Content:        []string{""},
		AppliedContent: []string{""},

		Cursor:    []int{0, 0},
		WantedX:   0,
		Clipboard: []string{},
		Undo:      []EditorState{},
		Redo:      []EditorState{},

		Numbers: true,
		Wrap:    true,
		Zenmode: false,

		Float: false,
		On:    true,

		Sidebar: sidebar.CreateSidebar(path, parent),
	}

	e.requestPage()
	e.requestSidebarPage(e.Path)

	return &e
}

func newEditorState(content []string, cursor []int, wantedx int) EditorState {
	contentCopy := append([]string(nil), content...)
	cursorCopy := append([]int(nil), cursor...)

	return EditorState{
		Content: contentCopy,
		Cursor:  cursorCopy,
		WantedX: wantedx,
	}
}

func (e *Editor) requestPage() {
	if e.Parent == nil {
		return
	}

	fs := e.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	if e.Path == "" {
		return
	}

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	if cached := fs.Find(e.Path); cached != nil && (cached.Stage == "draft" || cached.Stage == "edited") {
		req.Mode = "stale"
	}
	req.Path = e.Path
	req.Depth = 0
	req.Cont = true
	req.Meta = true
	req.Opts = true

	fs.Load(req)
}
