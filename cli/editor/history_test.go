package editor

import (
	"env/types"
	"slices"
	"testing"
)

func TestHistory(t *testing.T) {

	e := Editor{Content: []string{""}}

	steps := []struct {
		input string
		text string
		content []string
		cursor [3]int
	}{
		{text: "olá ", content: []string{"olá "}, cursor: [3]int{4, 0}},
		{text: "mundo", content: []string{"olá mundo"}, cursor: [3]int{9, 0}},
		{input: "ctrl+z", content: []string{"olá "}, cursor: [3]int{4, 0}},
		{input: "ctrl+Z", content: []string{""}},
		{input: "ctrl+z", content: []string{""}},
		{input: "ctrl+y", content: []string{"olá "}, cursor: [3]int{4, 0}},
		{input: "ctrl+Y", content: []string{"olá mundo"}, cursor: [3]int{9, 0}},
		{input: "enter", content: []string{"olá mundo", ""}, cursor: [3]int{0, 1}},
		{text: "ação", content: []string{"olá mundo", "ação"}, cursor: [3]int{4, 1}},
		{input: "up", content: []string{"olá mundo", "ação"}, cursor: [3]int{4, 0}},
		{input: "ctrl+z", content: []string{"olá mundo", ""}, cursor: [3]int{0, 1}},
		{input: "ctrl+z", content: []string{"olá mundo"}, cursor: [3]int{9, 0}},
		{input: "ctrl+y", content: []string{"olá mundo", ""}, cursor: [3]int{0, 1}},
		{text: "novo", content: []string{"olá mundo", "novo"}, cursor: [3]int{4, 1}},
		{input: "ctrl+y", content: []string{"olá mundo", "novo"}, cursor: [3]int{4, 1}},
		{input: "ctrl+z", content: []string{"olá mundo", ""}, cursor: [3]int{0, 1}},
	}

	for i, step := range steps {
		for _, char := range step.text {
			e.InputVisual(&types.Input{Description: "char", Char: char})
		}
		if step.input != "" {
			e.InputVisual(&types.Input{Description: step.input})
		}
		if !slices.Equal(e.Content, step.content) || e.Cursor != step.cursor {
			t.Fatalf("step %d: got %q %v, want %q %v", i, e.Content, e.Cursor, step.content, step.cursor)
		}
	}
}

func TestHistoryLineChanges(t *testing.T) {

	for _, input := range []string{"enter", "ctrl+enter", "delete", "ctrl+delete", "backspace", "ctrl+backspace", "ctrl+up", "ctrl+down"} {
		t.Run(input, func(t *testing.T) {
			e := Editor{Content: []string{"um", "dois"}, Cursor: [3]int{2, 0}}
			if input == "backspace" || input == "ctrl+backspace" || input == "ctrl+up" {
				e.Cursor = [3]int{0, 1}
			}
			before := state{Content: slices.Clone(e.Content), Cursor: e.Cursor}
			e.InputVisual(&types.Input{Description: input})
			after := state{Content: slices.Clone(e.Content), Cursor: e.Cursor}
			e.InputVisual(&types.Input{Description: "ctrl+z"})
			if !slices.Equal(e.Content, before.Content) || e.Cursor != before.Cursor {
				t.Fatalf("undo: got %q %v, want %q %v", e.Content, e.Cursor, before.Content, before.Cursor)
			}
			e.InputVisual(&types.Input{Description: "ctrl+y"})
			if !slices.Equal(e.Content, after.Content) || e.Cursor != after.Cursor {
				t.Fatalf("redo: got %q %v, want %q %v", e.Content, e.Cursor, after.Content, after.Cursor)
			}
		})
	}
}
