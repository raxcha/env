package zettelkasten

import (
	"env/cli"
	"env/engine"
	"env/filesystem"
	"env/utilities"
)

type Zettelkasten struct {
	Parent    cli.Parent
	Bounds    engine.Boundaries
	Utilities *utilities.Utilities

	Name string
	Path string

	RandPage *filesystem.Page
	FamiPage *filesystem.Page

	TagPaths     map[string][]string
	FamilyTitles map[string][]string

	Tags     []*ZettelTagItem
	Overlaps []*ZettelOverlapItem

	SelectedTag     int
	SelectedOverlap int
	Focus           string // tags, overlaps
	Prompt          string
	PanelMode       string

	On    bool
	Float bool
}

type ZettelTagItem struct {
	Name  string
	Paths []string
}

type ZettelOverlapItem struct {
	Name         string
	Paths        []string
	Shared       int
	Percent      int
	Popularity   int
	SelectedBase int
}

type ZettelNoteItem struct {
	Page *filesystem.Page
	Name string
	Path string
}

type zettelRect struct {
	X int
	Y int
	W int
	H int
}
