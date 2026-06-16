package engine

import (
	"fmt"
	"unicode"
)

func (e *Engine) startPrinting () {
	go func() {
		for diff := range e.Diff {
			e.print(diff)
		}
	}()
}

func (e *Engine) print (diff Diff) {

	for i, cell := range diff.Cells {
		x := diff.Indexing[i] % diff.Size[0]
		y := diff.Indexing[i] / diff.Size[0]

		moveCursor(x+1, y+1)
		printCell(cell)
	}
}

func moveCursor (x, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

func printCell(c Cell) {

	style := ""

	if c.Bold {
		style += "\033[1m"
	}

	if c.Italic {
		style += "\033[3m"
	}

	if c.Underline {
		style += "\033[4m"
	}

	fg := c.Fg
	bg := c.Bg

	if fg != nil && c.Visible {
		style += fmt.Sprintf("\033[38;2;%d;%d;%dm", fg.R, fg.G, fg.B)
	}

	if bg != nil && c.Visible {
		style += fmt.Sprintf("\033[48;2;%d;%d;%dm", bg.R, bg.G, bg.B)
	}

	fmt.Print(style)

	if c.Char == transparentRune {
		fmt.Print(" ")
		fmt.Print("\033[0m")
		return
	}

	if c.Char == 0 {
		fmt.Print(" ")
	} else if c.Uppercase {
		fmt.Printf("%c", unicode.ToUpper(c.Char))
	} else {
		fmt.Printf("%c", c.Char)
	}

	fmt.Print("\033[0m")
}

const transparentRune = '·'