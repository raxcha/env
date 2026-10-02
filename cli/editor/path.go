package editor

import (
	"env/types"
	"env/utils"
	"strings"
)

func (e *Editor) initPathEditor() {
	if e.PathEditor == nil && e.Page != nil {
		e.PathEditor = &Editor{Content: []string{e.Page.Path}}
	}
}

func (e *Editor) inputPath(input *types.Input) bool {
	e.initPathEditor()
	if e.PathEditor == nil {
		return false
	}
	if !e.EditingPath {
		if input.Description != "up" || e.Cursor[1] != 0 {
			return false
		}
		e.EditingPath = true
		e.PathEditor.Cursor[0] = e.Cursor[0]
		e.PathEditor.clampcursor()
		return true
	}
	switch input.Description {
	case "down", "enter", "ctrl+enter", "escape":
		e.EditingPath = false
	case "ctrl+s", "ctrl+S":
		e.savePage()
	case "char", "number", "left", "right", "ctrl+left", "ctrl+right", "backspace", "ctrl+backspace", "delete", "ctrl+delete", "ctrl+z", "ctrl+Z", "ctrl+y", "ctrl+Y":
		e.PathEditor.InputVisual(input)
		e.PathError = ""
	case "ctrl+c", "ctrl+C":
		if e.Parent != nil {
			e.Parent.SetClipboard(e.PathEditor.Content[0])
		}
	case "ctrl+v", "ctrl+V":
		if e.Parent != nil {
			value := strings.NewReplacer("\r", "", "\n", "").Replace(e.Parent.GetClipboard())
			for _, r := range value {
				e.PathEditor.InputVisual(&types.Input{Description: "char", Char: r})
			}
		}
	}
	return true
}

func (e *Editor) drawPath(sizes types.Dimensions) *types.Frame {
	prefix := "  " + utils.StageIcons[e.GetStage()]
	if prefix != "" {
		prefix += " "
	}
	return e.drawPathWithPrefix(sizes, prefix, false)
}

func (e *Editor) drawAlignedPath(sizes types.Dimensions, left int) *types.Frame {
	frame := e.drawPathWithPrefix(sizes, "", true)
	style := "§xy0 "
	if e.EditingPath || e.PathError != "" {
		style = "§YX0 "
	}
	headerSizes := *e.VisualSizes
	headerSizes.Size[1] = min(1, headerSizes.Size[1])
	background := e.Settings.GenerateFrame(headerSizes, []string{style}, 0, []int{0, 0, 0, 0})
	frame = e.Settings.MergeFrames(background, frame)
	icon := utils.StageIcons[e.GetStage()]
	width := utils.VisibleLength(icon) + 1
	if icon == "" || width > sizes.Pos[0]-left {
		return frame
	}
	// Keep the path anchored to the content column; draw the stage in the gutter.
	sizes.Pos[0] -= width
	sizes.Size[0] = width
	sizes.Size[1] = min(1, sizes.Size[1])
	stage := e.Settings.GenerateFrame(sizes, []string{style + icon + " "}, 0, []int{0, 0, 0, 0})
	return e.Settings.MergeFrames(frame, stage)
}

func (e *Editor) drawPathWithPrefix(sizes types.Dimensions, prefix string, interactive bool) *types.Frame {
	e.initPathEditor()
	sizes.Size[1] = min(1, sizes.Size[1])
	path := e.Spec
	if e.PathEditor != nil {
		path = e.PathEditor.Content[0]
	}
	if interactive && e.EditingPath && e.PathEditor != nil {
		runes := []rune(path)
		cursor := min(e.PathEditor.Cursor[0], len(runes))
		start := max(0, cursor-max(1, sizes.Size[0]-utils.VisibleLength(prefix))+2)
		start = min(start, cursor)
		line := string(runes[start:cursor]) + "¤BX "
		if cursor < len(runes) {
			line += string(runes[cursor]) + "¤ " + string(runes[cursor+1:])
		} else {
			line += " ¤ "
		}
		path = "§YX0 " + prefix + line
	} else {
		path = "§xy0 " + prefix + "‹b " + path + "›b "
	}
	if interactive && e.PathError != "" {
		path = "§YX0 " + prefix + e.PathError
	}
	return e.Settings.GenerateFrame(sizes, []string{path}, 0, []int{0, 0, 0, 0})
}
