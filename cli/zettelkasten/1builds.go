package zettelkasten

import (
	"env/cli"
	"env/engine"
	"env/filesystem"
	"path/filepath"
	"strings"
)

func CreateZettelkasten(path string, parent cli.Parent) *Zettelkasten {
	path = normalizeZettelkastenPath(path)

	z := Zettelkasten{
		Parent:    parent,
		Bounds:    engine.Boundaries{},
		Utilities: parent.GetUtilities(),

		Name: "zettelkasten",
		Path: path,

		TagPaths:        map[string][]string{},
		FamilyTitles:    map[string][]string{},
		Tags:            []*ZettelTagItem{},
		Overlaps:        []*ZettelOverlapItem{},
		SelectedTag:     0,
		SelectedOverlap: 0,
		Focus:           "tags",
		Prompt:          "",

		On:    true,
		Float: false,
	}

	z.requestRand()
	z.requestFami()
	z.rebuild()
	return &z
}

func normalizeZettelkastenPath(path string) string {
	path = strings.TrimSpace(strings.Trim(path, "/"))
	path = filepath.ToSlash(filepath.Clean(path))
	if path == "" || path == "." || path == "..." {
		return "zettelkasten"
	}
	return path
}

func (z *Zettelkasten) requestRand() {
	if z == nil || z.Parent == nil || z.Parent.GetFilesystem() == nil {
		return
	}

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	req.Path = "c.rand"
	req.Depth = -1
	req.Meta = true
	req.Cont = true
	req.Opts = true
	req.Sort = "time"
	z.Parent.GetFilesystem().Load(req)
}

func (z *Zettelkasten) requestFami() {
	if z == nil || z.Parent == nil || z.Parent.GetFilesystem() == nil {
		return
	}

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	req.Path = "d.fami"
	req.Depth = -1
	req.Meta = true
	req.Cont = false
	req.Opts = true
	req.Sort = "basic"
	z.Parent.GetFilesystem().Load(req)
}
