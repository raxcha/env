package utilities

import "strings"

func (u *Utilities) Box(lines []string, opts BoxOpts) []string {
	w := opts.W
	h := opts.H
	title := opts.Title

	paddingTop := max(0, opts.Padding.Top)
	paddingRight := max(0, opts.Padding.Right)
	paddingBottom := max(0, opts.Padding.Bottom)
	paddingLeft := max(0, opts.Padding.Left)

	if w < 2 || h < 2 {
		return lines
	}

	char1 := "─"
	char2 := "┐"
	char3 := "│"
	char4 := "┘"
	char5 := "─"
	char6 := "└"
	char7 := "│"
	char8 := "┌"

	innerW := w - 2
	innerH := h - 2

	contentW := innerW - paddingLeft - paddingRight
	contentH := innerH - paddingTop - paddingBottom

	if contentW < 0 {
		contentW = 0
	}
	if contentH < 0 {
		contentH = 0
	}

	if len(lines) > contentH {
		lines = lines[:contentH]
	}

	for len(lines) < contentH {
		lines = append(lines, "")
	}

	top := char8 + strings.Repeat(char1, innerW) + char2

	if title != "" && innerW > 0 {
		label := " " + title + " "

		if u.VisibleLength(label) > innerW {
			label = string([]rune(label)[:innerW])
		}

		top = char8 + label + strings.Repeat(char1, innerW-u.VisibleLength(label)) + char2
	}

	boxed := []string{top}

	for i := 0; i < paddingTop; i++ {
		boxed = append(boxed, char7+strings.Repeat(" ", innerW)+char3)
	}

	for _, line := range lines {
		if u.VisibleLength(line) > contentW {
			line = u.CutVisible(line, contentW)
		}

		contentPadding := contentW - u.VisibleLength(line)

		boxed = append(
			boxed,
			char7+
				strings.Repeat(" ", paddingLeft)+
				line+
				strings.Repeat(" ", contentPadding)+
				strings.Repeat(" ", paddingRight)+
				char3,
		)
	}

	for i := 0; i < paddingBottom; i++ {
		boxed = append(boxed, char7+strings.Repeat(" ", innerW)+char3)
	}

	boxed = append(boxed, char6+strings.Repeat(char5, innerW)+char4)

	return boxed
}

func (u *Utilities) CutVisible(str string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(str)
	out := make([]rune, 0, max+16)
	visible := 0
	for i := 0; i < len(runes); {
		span := markupSpan(runes, i)
		if span > 0 {
			out = append(out, runes[i:i+span]...)
			i += span
		} else {
			if visible >= max {
				break
			}
			out = append(out, runes[i])
			visible++
			i++
		}
	}
	return string(out)
}