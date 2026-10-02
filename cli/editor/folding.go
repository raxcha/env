package editor

import (
	"env/utils"
	"strings"
	"time"
)

func blockBoundary(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" || utils.DashedLine(line) {
		return true
	}
	for _, layout := range []string{"15:04", "2006.01.02", "2006.01.02 15:04", "15:04 2006.01.02"} {
		if _, err := time.Parse(layout, line); err == nil {
			return true
		}
	}
	return false
}

func (e *Editor) toggleBlockFold() bool {
	start := e.Cursor[1]
	if start < 0 || start >= len(e.Content) || blockBoundary(e.Content[start]) {
		return false
	}
	if start > 0 && !blockBoundary(e.Content[start-1]) {
		return false
	}
	end := start
	for end+1 < len(e.Content) && !blockBoundary(e.Content[end+1]) {
		end++
	}
	if end == start {
		return false
	}
	if e.FoldedBlocks == nil {
		e.FoldedBlocks = make(map[int]int)
	}
	if _, folded := e.FoldedBlocks[start]; folded {
		delete(e.FoldedBlocks, start)
	} else {
		e.FoldedBlocks[start] = end
	}
	return true
}

func (e *Editor) moveVisibleLine(delta int) {
	next := e.Cursor[1] + delta
	for start, end := range e.FoldedBlocks {
		if next > start && next <= end {
			if delta > 0 {
				next = end + 1
			} else {
				next = start
			}
			break
		}
	}
	if next >= 0 && next < len(e.Content) {
		e.Cursor[1] = next
	}
}
