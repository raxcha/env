package engine

import (

	"fmt"
	"strings"
	"time"

	"env/types"
	"env/utils"
)

type Engine struct {

	Queue chan types.Queue

	index int
	timer *time.Timer

	queue *types.Queue
	oldframe *types.Frame
	newframe *types.Frame
	diff *types.Diff
}

func CreateEngine() *Engine {

	e := Engine {

		Queue: make(chan types.Queue),

		index: 0,
		timer: time.NewTimer(0),
		
		queue: nil,
		oldframe: nil,
		newframe: nil,
		diff: nil,
	}

	e.timer.Stop()
	e.startEngine()

	return &e
}

func (e *Engine) CallEngine(queue types.Queue) {
	e.Queue <- queue
}

func (e *Engine) startEngine() {

	go func() {
		for {
			select {
			case q := <- e.Queue:
				e.queue = &q
				e.index = 0
				e.timer.Stop()
				e.manageQueue(&q)

			case <-e.timer.C:
				e.manageQueue(e.queue)
			}
		}
	}()
}

func (e *Engine) manageQueue(q *types.Queue) {

	if q == nil { return }

	if e.index >= len(q.Frames) {
		if q.Cycle {
			e.index = 0
			e.manageQueue(q)
		}
		return
	}

	e.oldframe = e.newframe
	e.newframe = q.Frames[e.index]
	e.timer.Reset(time.Duration(q.Frames[e.index].Timeout) * time.Millisecond)
	e.index++
	e.manageFrame()
}

func (e *Engine) manageFrame() {

	if e.oldframe == nil || !utils.CompareSize(e.oldframe.Sizes.Full, e.newframe.Sizes.Full) || len(e.oldframe.Cells) != len(e.newframe.Cells) {

		indexes := []int{}
		for i := range e.newframe.Cells {
			indexes = append(indexes, i)
		}
		e.diff = &types.Diff{Size: e.newframe.Sizes.Full, Cells: e.newframe.Cells, Indexes: indexes}
		e.manageDiff()
		return
	}

	diffcells := []*types.Cell{}
	indexes := []int{}
	for i := range len(e.newframe.Cells) {
		if !utils.CompareCells(e.oldframe.Cells[i], e.newframe.Cells[i]) {
			diffcells = append(diffcells, e.newframe.Cells[i])
			indexes = append(indexes, i)
		}
	}

	e.diff = &types.Diff{Size: e.newframe.Sizes.Full, Cells: diffcells, Indexes: indexes}
	e.manageDiff()
}

func (e *Engine) manageDiff() {

	var output strings.Builder

	if e.oldframe == nil || !utils.CompareSize(e.oldframe.Sizes.Full, e.newframe.Sizes.Full) {
		output.WriteString("\033[0m\033[2J")
	}

	for i, cell := range e.diff.Cells{
		
		if cell.Invisible { continue }

		x := e.diff.Indexes[i] % e.diff.Size[0]
		y := e.diff.Indexes[i] / e.diff.Size[0]

		moveCursor(&output, x+1, y+1)
		printCell(&output, cell)
	}

	fmt.Print(output.String())
}

func moveCursor(output *strings.Builder, x, y int) {
	fmt.Fprintf(output, "\033[%d;%dH", y, x)
}

func printCell(output *strings.Builder, c *types.Cell) {

	style := ""

	if c.Bold { style += "\033[1m" }
	if c.Italic { style += "\033[3m" }
	if c.Underline && c.Char != ' ' { style += "\033[4m" }
	style += c.Ansi.Bg
	style += c.Ansi.Fg
	output.WriteString(style)
	
	if c.Char == ' ' {
		output.WriteByte(' ')
	} else {
		output.WriteRune(c.Char)
	}

	output.WriteString("\033[0m")
}
