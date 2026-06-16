package utilities

import (
	"env/engine"
	"strconv"
)

func (u *Utilities) GenerateFrame(
	bounds engine.Boundaries,
	lines []string,
	timeout int,
) *engine.Frame {

	theme := u.Theme

	/*
	   - start "§fc1 " > bg, fg and wrap(num) ...
	   - toggle "¬biuUaAv "
	   - on "‹biuUaAv "
	   - off "›biuUaAv "
	   - bg and fg ...
	   	"¤f4 " and "¤ "
	   - "§ " to cancel everything ...
	   - "¦ " ...
	   - "¶ " ...

	*/

	if len(bounds.Fullsize) < 2 || len(bounds.Pos) < 2 || len(bounds.Size) < 2 {
		return newFrame()
	}

	left := max(0, bounds.Pos[0])
	top := max(0, bounds.Pos[1])

	sizex := bounds.Size[0]
	if left+sizex > bounds.Fullsize[0] {
		sizex = max(0, bounds.Fullsize[0]-left)
	}

	sizey := bounds.Size[1]
	if top+sizey > bounds.Fullsize[1] {
		sizey = max(0, bounds.Fullsize[1]-top)
	}

	right := max(0, bounds.Fullsize[0]-left-sizex)
	bot := max(0, bounds.Fullsize[1]-top-sizey)

	cells := make([]engine.Cell, 0, bounds.Fullsize[0]*bounds.Fullsize[1])

	// Célula que representa o fundo padrão do tema (margem)
	bgOnlyCell := engine.Cell{
		Char:    0, // Usamos 0 para identificar que é apenas preenchimento de fundo
		Fg:      &theme.Foreground,
		Bg:      &theme.Background,
		Visible: true,
	}

	for range top * bounds.Fullsize[0] {
		cells = append(cells, bgOnlyCell)
	}

	contentI := 0
	contentJ := 0

	basestl := newStyle(theme)
	stl := *basestl
	mode := "normal"

	for screenY := 0; screenY < sizey; screenY++ {
		visible := 0

		for range left {
			cells = append(cells, bgOnlyCell)
		}

		if contentI >= len(lines) {
			paddingCell := *newCell(basestl)

			for k := 0; k < sizex; k++ {
				cells = append(cells, paddingCell)
			}

			for range right {
				cells = append(cells, bgOnlyCell)
			}

			continue
		}

		runes := []rune(lines[contentI])

		for contentJ < len(runes) {
			r := runes[contentJ]

			nextdelim := findNextSpace(runes[contentJ:])
			if nextdelim == -1 {
				nextdelim = len(runes) - contentJ
			}

			if contentJ == 0 && r == '§' && nextdelim >= 4 {
				basestl.StdBg = theme.Color(runes[contentJ+1], Calm(0.25), SoftOn(theme.Background, 0.18))
				basestl.Bg = basestl.StdBg
				basestl.StdFg = theme.Color(runes[contentJ+2], EnsureContrast(*basestl.StdBg, 4.5))
				basestl.Fg = basestl.StdFg
				basestl.Wrap, _ = strconv.Atoi(string(runes[contentJ+3 : contentJ+nextdelim]))
				stl = *basestl

				contentJ += nextdelim
				if contentJ < len(runes) && runes[contentJ] == ' ' {
					contentJ++
				}

				continue
			}

			if r == '¬' {
				mode = "toggle"
				contentJ++
				continue
			}

			if r == '‹' {
				mode = "on"
				contentJ++
				continue
			}

			if r == '›' {
				mode = "off"
				contentJ++
				continue
			}

			if r == '§' {
				stl = *basestl
				contentJ++

				if contentJ < len(runes) && runes[contentJ] == ' ' {
					contentJ++
				}

				continue
			}

			if r == '¦' {
				if stl.Paused {
					if stl.Saved != nil {
						stl = *stl.Saved
					}

					stl.Paused = false
					stl.Saved = nil
				} else {
					saved := stl

					stl = *basestl
					stl.Paused = true
					stl.Saved = &saved
				}

				contentJ++

				if contentJ < len(runes) && runes[contentJ] == ' ' {
					contentJ++
				}

				continue
			}

			if r == '¶' {
				if stl.Cursor {
					stl.Fg = stl.SavedFg
					stl.Bg = stl.SavedBg

					stl.Cursor = false
					stl.SavedFg = nil
					stl.SavedBg = nil
				} else {
					stl.SavedFg = stl.Fg
					stl.SavedBg = stl.Bg

					stl.Fg = &theme.Background
					stl.Bg = &theme.Cursor

					stl.Cursor = true
				}

				contentJ++

				if contentJ < len(runes) && runes[contentJ] == ' ' {
					contentJ++
				}

				continue
			}

			if r == '¤' {
				if nextdelim == 1 {
					stl.Bg = basestl.StdBg
					stl.Fg = basestl.StdFg
				} else if nextdelim >= 3 {
					stl.Bg = theme.Color(runes[contentJ+1], Calm(0.25), SoftOn(theme.Background, 0.22))
					if nextdelim >= 4 {
						stl.Fg = theme.Color(runes[contentJ+2], EnsureContrast(*stl.Bg, 4.5))
					} else {
						stl.Fg = theme.Color(runes[contentJ+1], EnsureContrast(*stl.Bg, 4.5))
					}
				}

				contentJ += nextdelim

				if contentJ < len(runes) && runes[contentJ] == ' ' {
					contentJ++
				}

				continue
			}

			if mode != "normal" && r == ' ' {
				mode = "normal"
				contentJ++
				continue
			}

			if mode == "toggle" {
				if r == 'b' {
					stl.Bold = !stl.Bold
				}
				if r == 'i' {
					stl.Italic = !stl.Italic
				}
				if r == 'u' {
					stl.Underline = !stl.Underline
				}
				if r == 'U' {
					stl.Uppercase = !stl.Uppercase
				}
				if r == 'a' {
					stl.Almost = !stl.Almost
				}
				if r == 'A' {
					stl.All = !stl.All
				}
				if r == 'v' {
					stl.Visible = !stl.Visible
				}

				contentJ++
				continue
			}

			if mode == "on" {
				if r == 'b' {
					stl.Bold = true
				}
				if r == 'i' {
					stl.Italic = true
				}
				if r == 'u' {
					stl.Underline = true
				}
				if r == 'U' {
					stl.Uppercase = true
				}
				if r == 'a' {
					stl.Almost = true
				}
				if r == 'A' {
					stl.All = true
				}
				if r == 'v' {
					stl.Visible = true
				}

				contentJ++
				continue
			}

			if mode == "off" {
				if r == 'b' {
					stl.Bold = false
				}
				if r == 'i' {
					stl.Italic = false
				}
				if r == 'u' {
					stl.Underline = false
				}
				if r == 'U' {
					stl.Uppercase = false
				}
				if r == 'a' {
					stl.Almost = false
				}
				if r == 'A' {
					stl.All = false
				}
				if r == 'v' {
					stl.Visible = false
				}

				contentJ++
				continue
			}

			if visible >= sizex {
				if stl.Wrap > 0 {
					stl.Wrap--
					break
				}

				contentJ = len(runes)
				break
			}

			visible++

			c := newCell(&stl)
			c.Char = r

			if r == ' ' {
				c.Underline = false
			}

			cells = append(cells, *c)
			contentJ++
		}

		paddingCell := *newCell(basestl)

		for k := visible; k < sizex; k++ {
			cells = append(cells, paddingCell)
		}

		for range right {
			cells = append(cells, bgOnlyCell)
		}

		if contentJ >= len(runes) {
			contentI++
			contentJ = 0

			basestl = newStyle(theme)
			stl = *basestl
			mode = "normal"
		}
	}

	for range bot * bounds.Fullsize[0] {
		cells = append(cells, bgOnlyCell)
	}

	expected := bounds.Fullsize[0] * bounds.Fullsize[1]

	if len(cells) < expected {
		for i := len(cells); i < expected; i++ {
			cells = append(cells, bgOnlyCell)
		}
	}

	if len(cells) > expected {
		cells = cells[:expected]
	}

	newframe := newFrame()
	newframe.Cells = cells
	newframe.Timeout = timeout
	newframe.Size = bounds.Fullsize
	return newframe
}
