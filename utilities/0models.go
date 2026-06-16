package utilities

import "env/engine"

type Utilities struct {
	
	Theme *Theme
	Themes []*Theme
}

type Style struct {

	Bold, Italic, Underline bool
	Fg, Bg *engine.RGB
	StdFg, StdBg *engine.RGB
	Uppercase bool
	Almost, All bool
	Visible bool
	Wrap int
	Paused bool
	Saved  *Style
	Cursor bool
	SavedFg *engine.RGB
	SavedBg *engine.RGB
}

type Theme struct {
	Name    string
	Author  string
	Variant string
	Raw     rawTheme

	Black   engine.RGB
	Red     engine.RGB
	Green   engine.RGB
	Yellow  engine.RGB
	Blue    engine.RGB
	Magenta engine.RGB
	Cyan    engine.RGB
	White   engine.RGB

	BrightBlack   engine.RGB
	BrightRed     engine.RGB
	BrightGreen   engine.RGB
	BrightYellow  engine.RGB
	BrightBlue    engine.RGB
	BrightMagenta engine.RGB
	BrightCyan    engine.RGB
	BrightWhite   engine.RGB

	Background engine.RGB
	Foreground engine.RGB
	Cursor     engine.RGB
}

type Padding struct {
	Top, Right, Bottom, Left int
}

type BoxOpts struct {
	W, H    int
	Padding Padding
	Title   string
}

type rawTheme struct {
	Name    string `json:"name"`
	Author  string `json:"author"`
	Variant string `json:"variant"`

	Color01 string `json:"color_01"`
	Color02 string `json:"color_02"`
	Color03 string `json:"color_03"`
	Color04 string `json:"color_04"`
	Color05 string `json:"color_05"`
	Color06 string `json:"color_06"`
	Color07 string `json:"color_07"`
	Color08 string `json:"color_08"`
	Color09 string `json:"color_09"`
	Color10 string `json:"color_10"`
	Color11 string `json:"color_11"`
	Color12 string `json:"color_12"`
	Color13 string `json:"color_13"`
	Color14 string `json:"color_14"`
	Color15 string `json:"color_15"`
	Color16 string `json:"color_16"`

	Background string `json:"background"`
	Foreground string `json:"foreground"`
	Cursor     string `json:"cursor"`

	Hash string `json:"hash"`
}
