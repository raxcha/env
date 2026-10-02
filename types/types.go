package types

import (
	
	"time"
	"sync"
)

type Pos [2]int

type Size [2]int

type Dimensions struct {

	Full Size
	
	Pos Pos

	Size Size
}

type Input struct {

	Meta bool

	Description string
	Char rune
}

type Queue struct {

	Size Size
	Frames []*Frame

	Cycle bool
}

type Frame struct {

	Sizes Dimensions
	Cells []*Cell

	Timeout int
}


type Cell struct {

	Bold, Italic, Underline bool
	Invisible bool
	Char rune

	Ansi Ansi
}


type Diff struct {

	Size Size

	Cells []*Cell
	Indexes []int
}


type Style struct {
	
	Bold, Italic, Underline bool

	Uppercase, Invisible bool
	Almost, All bool

	Ansi Ansi
	StdAnsi Ansi
}

type Ansi struct { Bg, Fg string }

type Color struct { R, G, B int }

type Theme struct {

	Color01 Color
	Color02 Color
	Color03 Color
	Color04 Color
	Color05 Color
	Color06 Color
	Color07 Color
	Color08 Color
	Color09 Color
	Color10 Color
	Color11 Color
	Color12 Color
	Color13 Color
	Color14 Color
	Color15 Color
	Color16 Color

	Background Color
	Foreground Color
	Cursor Color
}

type Page struct {

	Path string
	Name string
	Type string

	Content []string
	Metadata map[string]any

	Og *Page
	Stage string

	Children []*Page
	Sorting string

	Token string
}


type Patch struct {

	Description [][]string

	Commands []func()
	OnConfirm func()
	OnCancel func()
	Timer *time.Timer
	Deadline time.Time
	Done chan struct{}
	Once sync.Once
}


type Loading struct {

	CreateDraft bool

	Api bool
	Path string
	Mode string

	Meta bool
	Cont bool
	Depth int

	Sort string
	Filters []string
	Latest int

	Token string
}


type Syncing struct {

	Api bool
	Branch *Page
	Hard bool

	Filters []string
}
