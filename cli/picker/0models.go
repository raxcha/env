package picker

import (
	"env/cli"
	"env/engine"
	"env/filesystem"
	"env/utilities"
)

type Picker struct {
	Parent    cli.Parent
	Bounds    engine.Boundaries
	Utilities *utilities.Utilities

	Cache         *filesystem.Page
	Flat          []*filesystem.Page
	Items         []*Match
	Selected      int
	PreviewOffset int
	Prompt        string
	Mode          string
	Sorting       string
	Scope         string
	PanelMode     string

	Name      string
	Path      string
	StartPath string

	Switch bool
	On     bool
	Float  bool
}

type Match struct {
	Page  *filesystem.Page
	Score int
	Kind  string
}

const (
	pageStageGhost    = "ghost"
	pageStageLocal    = "local"
	pageStageApi      = "api"
	pageStageDoom     = "doomed"
	pageStageDraft    = "draft"
	pageStageEdit     = "edited"
	pageStageId       = "replacing"
	pageStageConflict = "conflicted"
)
