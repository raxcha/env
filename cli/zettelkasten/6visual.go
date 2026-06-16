package zettelkasten

import (
	"fmt"
	"math"
	"strings"
)

const zettelVisualManualDebug = false
const zettelVisualLineStrength = 0

var zettelVisualDebugSelected = "B"
var zettelVisualDebugAX = -1
var zettelVisualDebugAY = -1
var zettelVisualDebugBX = -1
var zettelVisualDebugBY = -1

type ZettelVisualNode struct {
	ID        string
	Label     string
	X         int
	Y         int
	Connected bool
	Strength  int
	Bold      bool
	Ray       zettelVisualRay
	Step      int
	MaxStep   int
}

type zettelVisualCell struct {
	Ch       rune
	Strength int
	Bold     bool
	Label    bool
}

type ZettelVisualConnectionContext struct {
	From  ZettelVisualNode
	To    ZettelVisualNode
	Index int

	X int
	Y int

	Width  int
	Height int

	grid    [][]zettelVisualCell
	stopped bool
}

func (ctx *ZettelVisualConnectionContext) Draw(ch rune) {
	if ctx == nil || ctx.grid == nil || ch == 0 {
		return
	}
	if ctx.Y < 0 || ctx.Y >= ctx.Height || ctx.X < 0 || ctx.X >= ctx.Width {
		return
	}
	cell := &ctx.grid[ctx.Y][ctx.X]
	if cell.Label {
		return
	}
	cell.Ch = zettelVisualMergeRunes(cell.Ch, ch)
	cell.Strength = zettelVisualLineStrength
}

func zettelVisualMergeRunes(existing rune, incoming rune) rune {
	if existing == 0 || existing == ' ' {
		return incoming
	}
	if incoming == 0 || incoming == ' ' || incoming == existing {
		return existing
	}

	if zettelVisualIsNodeRune(existing) || zettelVisualIsNodeRune(incoming) {
		if zettelVisualIsNodeRune(existing) {
			return existing
		}
		return incoming
	}

	existingUp := zettelVisualIsUpDiagonalRune(existing)
	existingDown := zettelVisualIsDownDiagonalRune(existing)
	incomingUp := zettelVisualIsUpDiagonalRune(incoming)
	incomingDown := zettelVisualIsDownDiagonalRune(incoming)
	if (existingUp && incomingDown) || (existingDown && incomingUp) {
		return existing
	}

	if zettelVisualIsCurvedRune(existing) && zettelVisualIsCurvedRune(incoming) {
		return existing
	}

	if zettelVisualIsVerticalRune(existing) && zettelVisualIsHorizontalRune(incoming) {
		return '┼'
	}
	if zettelVisualIsHorizontalRune(existing) && zettelVisualIsVerticalRune(incoming) {
		return '┼'
	}

	return incoming
}

func zettelVisualIsNodeRune(ch rune) bool {
	return ch == '●'
}

func zettelVisualIsUpDiagonalRune(ch rune) bool {
	return ch == '╱' || ch == '⟋'
}

func zettelVisualIsDownDiagonalRune(ch rune) bool {
	return ch == '╲' || ch == '⟍'
}

func zettelVisualIsVerticalRune(ch rune) bool {
	switch ch {
	case '│', '⎹', '|', '⎸':
		return true
	default:
		return false
	}
}

func zettelVisualIsHorizontalRune(ch rune) bool {
	switch ch {
	case '─', '_', '-', '‾', '⎽', '⎼', '⎻', '⎺':
		return true
	default:
		return false
	}
}

func zettelVisualIsCurvedRune(ch rune) bool {
	return zettelVisualIsUpDiagonalRune(ch) ||
		zettelVisualIsDownDiagonalRune(ch) ||
		zettelVisualIsVerticalRune(ch) ||
		zettelVisualIsHorizontalRune(ch)
}

func (ctx *ZettelVisualConnectionContext) Move(dx int, dy int) {
	if ctx == nil {
		return
	}
	ctx.X += dx
	ctx.Y += dy
}

func (ctx *ZettelVisualConnectionContext) Place(x int, y int) {
	if ctx == nil {
		return
	}
	ctx.X = x
	ctx.Y = y
}

func (ctx *ZettelVisualConnectionContext) Stop() {
	if ctx == nil {
		return
	}
	ctx.stopped = true
}

func (ctx *ZettelVisualConnectionContext) Stopped() bool {
	if ctx == nil {
		return true
	}
	return ctx.stopped
}

func (ctx *ZettelVisualConnectionContext) DX() int {
	if ctx == nil {
		return 0
	}
	return ctx.To.X - ctx.From.X
}

func (ctx *ZettelVisualConnectionContext) DY() int {
	if ctx == nil {
		return 0
	}
	return ctx.To.Y - ctx.From.Y
}

func (ctx *ZettelVisualConnectionContext) CursorDXToTarget() int {
	if ctx == nil {
		return 0
	}
	return ctx.To.X - ctx.X
}

func (ctx *ZettelVisualConnectionContext) CursorDYToTarget() int {
	if ctx == nil {
		return 0
	}
	return ctx.To.Y - ctx.Y
}

func (ctx *ZettelVisualConnectionContext) AtTarget() bool {
	if ctx == nil {
		return true
	}
	return ctx.X == ctx.To.X && ctx.Y == ctx.To.Y
}

// Mantive seus sets aqui se quiser reaproveitar manualmente dentro de
// zettelVisualConnection(ctx).

type ZettelVisualConnectionSet struct {
	Pattern string
	DX      int
	DY      int
}

var sets = []ZettelVisualConnectionSet{
	// 0: vertical exato para cima
	{Pattern: "│", DX: 0, DY: -1},

	// 1..7: subindo
	{Pattern: "⎸|", DX: 0, DY: -2},
	{Pattern: "╱", DX: 1, DY: -1},
	{Pattern: "⟋ ", DX: 2, DY: -1},
	{Pattern: "⎽⎼⎻⎺", DX: 4, DY: -1},
	{Pattern: "_-‾", DX: 3, DY: -1},
	{Pattern: "__---‾‾", DX: 7, DY: -1},
	{Pattern: "_⎽⎽⎼⎼⎻⎻⎺⎺‾", DX: 10, DY: -1},

	// 8: horizontal / meio
	{Pattern: "─", DX: 1, DY: 0},

	// 9..15: descendo
	{Pattern: "‾⎺⎺⎻⎻⎼⎼⎽⎽_", DX: 10, DY: 1},
	{Pattern: "‾‾---__", DX: 7, DY: 1},
	{Pattern: "‾-_", DX: 3, DY: 1},
	{Pattern: "⎺⎻⎼⎽", DX: 4, DY: 1},
	{Pattern: "⟍ ", DX: 2, DY: 1},
	{Pattern: "╲", DX: 1, DY: 1},
	{Pattern: "⎸|", DX: 0, DY: 2},

	// 16: vertical exato para baixo
	{Pattern: "│", DX: 0, DY: 1},
}

func chooseSet(dx int, dy int) ZettelVisualConnectionSet {
	// exatamente horizontal: traços descontínuos mais longos
	if dy == 0 {
		return ZettelVisualConnectionSet{
			Pattern: "─── ",
			DX:      4,
			DY:      0,
		}
	}

	// exatamente vertical
	if dx == 0 {
		return ZettelVisualConnectionSet{
			Pattern: "⎹|⎸",
			DX:      0,
			DY:      zettelVisualSign(dy),
		}
	}

	adx := zettelVisualAbs(dx)
	ady := zettelVisualAbs(dy)

	if dy > 0 {
		return chooseSetDown(adx, ady)
	}
	return chooseSetUp(adx, ady)
}

// chooseSetDown selects the pattern for a downward-going connection.
// adx and ady are the absolute horizontal and vertical distances.
func chooseSetDown(adx, ady int) ZettelVisualConnectionSet {
	switch {
	// quase horizontal: padrão muito longo
	case adx >= 25*ady:
		return ZettelVisualConnectionSet{Pattern: "‾⎺⎺⎻⎻⎼⎼⎽⎽_", DX: 10, DY: 1}
	case adx >= 10*ady:
		return ZettelVisualConnectionSet{Pattern: "‾⎺⎺⎻⎻⎼⎼⎽⎽_", DX: 10, DY: 1}
	case adx >= 5*ady:
		return ZettelVisualConnectionSet{Pattern: "‾‾---__", DX: 7, DY: 1}
	// diagonal gentil: ~4 cols por linha
	case adx >= 3*ady:
		return ZettelVisualConnectionSet{Pattern: "⎺⎻⎼⎽", DX: 4, DY: 1}
	// diagonal moderada: ~2-3 cols por linha
	case adx >= 2*ady:
		return ZettelVisualConnectionSet{Pattern: "⟍ ", DX: 2, DY: 1}
	// steep e muito inclinado: Bresenham com char diagonal
	default:
		return ZettelVisualConnectionSet{Pattern: "╲", DX: 1, DY: 1}
	}
}

// chooseSetUp selects the pattern for an upward-going connection.
func chooseSetUp(adx, ady int) ZettelVisualConnectionSet {
	switch {
	case adx >= 25*ady:
		return ZettelVisualConnectionSet{Pattern: "_⎽⎽⎼⎼⎻⎻⎺⎺‾", DX: 10, DY: -1}
	case adx >= 10*ady:
		return ZettelVisualConnectionSet{Pattern: "_⎽⎽⎼⎼⎻⎻⎺⎺‾", DX: 10, DY: -1}
	case adx >= 5*ady:
		return ZettelVisualConnectionSet{Pattern: "__---‾‾", DX: 7, DY: -1}
	case adx >= 3*ady:
		return ZettelVisualConnectionSet{Pattern: "⎽⎼⎻⎺", DX: 4, DY: -1}
	case adx >= 2*ady:
		return ZettelVisualConnectionSet{Pattern: "⟋ ", DX: 2, DY: -1}
	default:
		return ZettelVisualConnectionSet{Pattern: "╱", DX: 1, DY: -1}
	}
}

// ...

func zettelVisualConnection(ctx *ZettelVisualConnectionContext) {
	if ctx == nil {
		return
	}

	// Normalize to always draw left-to-right so pattern chars read correctly.
	// Patterns are designed for left-to-right; drawing right-to-left inverts them.
	if ctx.To.X < ctx.From.X {
		ctx.From, ctx.To = ctx.To, ctx.From
	}

	ctx.Place(ctx.From.X, ctx.From.Y)
	ctx.Index = 0
	ctx.stopped = false

	dx := ctx.To.X - ctx.From.X
	dy := ctx.To.Y - ctx.From.Y

	set := chooseSet(dx, dy)
	pattern := []rune(set.Pattern)

	adx := zettelVisualAbs(dx)
	ady := zettelVisualAbs(dy)

	sx := zettelVisualSign(dx)
	sy := zettelVisualSign(dy)

	if adx == 0 && ady == 0 {
		return
	}

	// Pure vertical.
	if adx == 0 {
		for ctx.Y != ctx.To.Y && !ctx.Stopped() {
			ctx.Draw(zettelVisualPatternChar(pattern, ctx.Index, ady))
			ctx.Move(0, sy)
			ctx.Index++
		}
		return
	}

	// Pure horizontal: cycle the dash pattern, skipping spaces (gaps between dashes).
	if ady == 0 {
		for ctx.X != ctx.To.X && !ctx.Stopped() {
			ch := pattern[ctx.Index%len(pattern)]
			if ch != ' ' {
				ctx.Draw(ch)
			}
			ctx.Move(sx, 0)
			ctx.Index++
		}
		return
	}

	// Short patterns (1-2 chars): use Bresenham along the longer axis for exactly
	// 1 char per step — no gaps and no side-by-side doubles.
	// Always draw the diagonal char (╱ or ╲); ignore any accent char.
	if len(pattern) <= 2 {
		ch := zettelVisualDiagChar(pattern)
		if adx >= ady {
			// Step along x (wider than tall).
			errY := ady - adx/2
			for ctx.X != ctx.To.X && !ctx.Stopped() {
				ctx.Draw(ch)
				ctx.Move(sx, 0)
				errY += ady
				if errY > 0 {
					ctx.Move(0, sy)
					errY -= adx
				}
				ctx.Index++
			}
		} else {
			// Step along y (taller than wide).
			errX := adx - ady/2
			for ctx.Y != ctx.To.Y && !ctx.Stopped() {
				ctx.Draw(ch)
				ctx.Move(0, sy)
				errX += adx
				if errX > 0 {
					ctx.Move(sx, 0)
					errX -= ady
				}
				ctx.Index++
			}
		}
		return
	}

	zettelVisualDrawPatternSegments(ctx, set, sx, sy)
}

func zettelVisualDrawPatternSegments(
	ctx *ZettelVisualConnectionContext,
	set ZettelVisualConnectionSet,
	sx int,
	sy int,
) {
	if ctx == nil || ctx.Stopped() {
		return
	}

	pattern := []rune(set.Pattern)
	if len(pattern) == 0 {
		return
	}

	stepX := zettelVisualAbs(set.DX)
	if stepX == 0 {
		stepX = len(pattern)
	}

	guard := (ctx.Width + ctx.Height + 1) * 4
	for guard > 0 && !ctx.Stopped() && !ctx.AtTarget() {
		guard--
		for i := 0; i < stepX && !ctx.Stopped() && !ctx.AtTarget(); i++ {
			if ctx.X == ctx.To.X {
				break
			}
			ch := pattern[i%len(pattern)]
			if ch != ' ' {
				ctx.Draw(ch)
			}
			ctx.Move(sx, 0)
			ctx.Index++
		}
		if ctx.Y != ctx.To.Y {
			ctx.Move(0, sy)
		}
	}
}

// zettelVisualMirrorPattern reverses a pattern and swaps ╱↔╲.
// Used when drawing right-to-left so arc characters read correctly.
func zettelVisualMirrorPattern(pattern []rune) []rune {
	out := make([]rune, len(pattern))
	for i, ch := range pattern {
		switch ch {
		case '╱':
			ch = '╲'
		case '╲':
			ch = '╱'
		}
		out[len(pattern)-1-i] = ch
	}
	return out
}

// zettelVisualDiagChar returns the diagonal rune (╱ or ╲) from a short pattern,
// falling back to the last rune if neither is present.
func zettelVisualDiagChar(pattern []rune) rune {
	for _, ch := range pattern {
		if ch == '╱' || ch == '╲' {
			return ch
		}
	}
	if len(pattern) > 0 {
		return pattern[len(pattern)-1]
	}
	return '╱'
}

func zettelVisualDrawConnectionCycle(
	ctx *ZettelVisualConnectionContext,
	pattern []rune,
	sx int,
	sy int,
	width int,
	height int,
) {
	if ctx == nil || ctx.Stopped() {
		return
	}

	if width < 0 {
		width = -width
	}
	if height < 0 {
		height = -height
	}

	// Caso muito vertical: não há coluna horizontal suficiente nesse ciclo.
	// Ainda assim desenha algo coerente e preserva o cursor manual.
	if width == 0 {
		if height == 0 {
			return
		}
		for i := 0; i < height && !ctx.Stopped() && !ctx.AtTarget(); i++ {
			ctx.Move(0, sy)
			ctx.Draw(zettelVisualPatternChar(pattern, i, height))
			ctx.Index++
		}
		return
	}

	verticalError := 0
	verticalMoved := 0

	for i := 0; i < width && !ctx.Stopped() && !ctx.AtTarget(); i++ {
		if ctx.X != ctx.To.X {
			ctx.Move(sx, 0)
		}

		verticalError += height
		for verticalError >= width && verticalMoved < height && ctx.Y != ctx.To.Y {
			ctx.Move(0, sy)
			verticalError -= width
			verticalMoved++
		}

		ctx.Draw(zettelVisualPatternChar(pattern, i, width))
		ctx.Index++
	}

	// Se sobrou altura por causa de um ciclo estreito, completa a parte
	// vertical usando o último caractere do pattern esticado.
	for verticalMoved < height && !ctx.Stopped() && !ctx.AtTarget() {
		if ctx.Y != ctx.To.Y {
			ctx.Move(0, sy)
		}
		ctx.Draw(zettelVisualPatternChar(pattern, width-1, width))
		ctx.Index++
		verticalMoved++
	}
}

func zettelVisualDistributedAmount(part int, parts int, total int) int {
	if parts <= 0 {
		return total
	}
	if total <= 0 {
		return 0
	}

	// Diferença entre dois pontos proporcionais inteiros.
	// Ex.: total=37, parts=4 => 9, 9, 10, 9.
	a := part * total / parts
	b := (part + 1) * total / parts
	return b - a
}

func zettelVisualPatternChar(pattern []rune, index int, width int) rune {
	if len(pattern) == 0 {
		return ' '
	}

	if width <= 0 {
		width = len(pattern)
	}

	i := index * len(pattern) / width

	if i < 0 {
		i = 0
	}
	if i >= len(pattern) {
		i = len(pattern) - 1
	}

	return pattern[i]
}

func zettelVisualConnectionChar(ctx *ZettelVisualConnectionContext) rune {
	if ctx == nil {
		return ' '
	}

	// Exemplo principal: escolhe um set de caracteres pela inclinação
	// restante entre o cursor atual e o destino.
	set := chooseSet(ctx.CursorDXToTarget(), ctx.CursorDYToTarget())
	pattern := []rune(set.Pattern)
	if len(pattern) == 0 {
		return '─'
	}

	// Para controlar literalmente os caracteres, edite só isto.
	// Exemplo:
	//   return 'a'
	//   return []rune("abc")[ctx.Index%3]
	return pattern[ctx.Index%len(pattern)]
}

func zettelVisualConnectionMove(ctx *ZettelVisualConnectionContext, steps int) {
	if ctx == nil || steps <= 0 {
		return
	}

	// Movimento padrão: interpola da origem ao destino.
	// Mesmo assim, o cursor só muda através de ctx.Move(dx, dy).
	// Para assumir controle total, substitua esta função por regras próprias.
	nextIndex := ctx.Index + 1
	if nextIndex > steps {
		ctx.Stop()
		return
	}

	t := float64(nextIndex) / float64(steps)
	targetX := ctx.From.X + int(math.Round(float64(ctx.DX())*t))
	targetY := ctx.From.Y + int(math.Round(float64(ctx.DY())*t))

	ctx.Move(targetX-ctx.X, targetY-ctx.Y)
}

// ============================================================
// VISUAL LINES
// ============================================================

func (z *Zettelkasten) visualLines(w int, h int) []string {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	grid := zettelVisualBlankGrid(w, h)
	nodes := z.zettelVisualNodes(w, h)

	for _, node := range nodes {
		zettelVisualDrawNode(grid, w, h, node)
	}

	if len(nodes) >= 2 {
		from := nodes[0]
		for i := 1; i < len(nodes); i++ {
			if !nodes[i].Connected {
				continue
			}
			ctx := &ZettelVisualConnectionContext{
				From:   from,
				To:     nodes[i],
				X:      from.X,
				Y:      from.Y,
				Width:  w,
				Height: h,
				grid:   grid,
			}
			zettelVisualConnection(ctx)
		}
	}

	return zettelVisualGridLines(grid)
}

func (z *Zettelkasten) zettelVisualNodes(w int, h int) []ZettelVisualNode {
	if zettelVisualManualDebug {
		return z.zettelVisualDebugNodes(w, h)
	}

	nodes := []ZettelVisualNode{}

	centerX := w / 2
	centerY := h / 2
	if centerX < 2 {
		centerX = 2
	}
	if centerY < 1 {
		centerY = 1
	}

	tagName := z.selectedTagName()
	if tagName == "" {
		tagName = "tag"
	}

	nodes = append(nodes, ZettelVisualNode{
		ID:       "tag",
		Label:    strings.ToUpper(tagName),
		X:        centerX,
		Y:        centerY,
		Strength: 3,
		Bold:     true,
	})

	overlapCount := len(z.Overlaps)
	outerTags := z.zettelVisualOuterTags(8)
	if overlapCount+len(outerTags) == 0 {
		nodes = append(nodes, ZettelVisualNode{
			ID:       "empty",
			Label:    "empty",
			X:        zettelVisualClamp(w-8, 0, w-1),
			Y:        centerY,
			Strength: 0,
		})
		return nodes
	}

	for i := 0; i < overlapCount; i++ {
		overlap := z.Overlaps[i]
		if overlap == nil {
			continue
		}

		closeness := float64(overlap.Percent) / 100.0
		if closeness < 0 {
			closeness = 0
		}
		if closeness > 1 {
			closeness = 1
		}
		x, y, ray, step, maxStep := zettelVisualAlignedNodePosition(centerX, centerY, w, h, i, overlapCount, closeness, true)

		label := fmt.Sprintf("T%d(%d)", i+1, overlap.Percent)
		if overlap.Name != "" {
			label = fmt.Sprintf("%s(%d)", overlap.Name, overlap.Percent)
		}
		label = zettelVisualMoonPhase(overlap.Percent) + " " + label

		nodes = append(nodes, ZettelVisualNode{
			ID:        fmt.Sprintf("tag-%d", i),
			Label:     label,
			X:         zettelVisualClamp(x, 1, w-2),
			Y:         zettelVisualClamp(y, 0, h-1),
			Connected: true,
			Strength:  zettelVisualStrengthFromCloseness(closeness),
			Bold:      z.Focus == "overlaps" && i == z.SelectedOverlap,
			Ray:       ray,
			Step:      step,
			MaxStep:   maxStep,
		})
	}

	for i, tag := range outerTags {
		if tag == nil {
			continue
		}
		x, y, ray, step, maxStep := zettelVisualAlignedNodePosition(centerX, centerY, w, h, i, len(outerTags), 0, false)
		nodes = append(nodes, ZettelVisualNode{
			ID:       fmt.Sprintf("outer-tag-%d", i),
			Label:    zettelVisualMoonPhase(0) + " " + tag.Name,
			X:        zettelVisualClamp(x, 1, w-2),
			Y:        zettelVisualClamp(y, 0, h-1),
			Strength: 0,
			Ray:      ray,
			Step:     step,
			MaxStep:  maxStep,
		})
	}

	zettelVisualResolveLabelOverlaps(nodes, centerX, centerY, w, h)
	return nodes
}

func (z *Zettelkasten) zettelVisualOuterTags(limit int) []*ZettelTagItem {
	if z == nil || limit <= 0 {
		return nil
	}

	selected := z.selectedTagName()
	overlapNames := map[string]bool{}
	for _, overlap := range z.Overlaps {
		if overlap != nil {
			overlapNames[overlap.Name] = true
		}
	}

	out := []*ZettelTagItem{}
	for _, tag := range z.Tags {
		if tag == nil || tag.Name == "" || tag.Name == selected || overlapNames[tag.Name] {
			continue
		}
		out = append(out, tag)
		if len(out) >= limit {
			break
		}
	}

	return out
}

func zettelVisualMoonPhase(percent int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	phases := []string{"", "", "", "", "", "", "", "", "", "", "", "", "", "󰽢"}
	fullness := math.Sqrt(float64(percent) / 100)
	index := int(math.Round(fullness * float64(len(phases)-1)))
	return phases[index]
}

func zettelVisualStrengthFromCloseness(closeness float64) int {
	switch {
	case closeness >= 0.67:
		return 3
	case closeness >= 0.34:
		return 2
	default:
		return 1
	}
}

type zettelVisualRay struct {
	X int
	Y int
}

func zettelVisualAlignedNodePosition(
	centerX int,
	centerY int,
	w int,
	h int,
	index int,
	count int,
	closeness float64,
	connected bool,
) (int, int, zettelVisualRay, int, int) {
	rays := []zettelVisualRay{
		{X: 0, Y: -1},
		{X: 1, Y: -1},
		{X: 2, Y: -1},
		{X: 4, Y: -1},
		{X: 7, Y: -1},
		{X: 10, Y: -1},
		{X: 1, Y: 0},
		{X: 10, Y: 1},
		{X: 7, Y: 1},
		{X: 4, Y: 1},
		{X: 2, Y: 1},
		{X: 1, Y: 1},
		{X: 0, Y: 1},
		{X: -1, Y: 1},
		{X: -2, Y: 1},
		{X: -4, Y: 1},
		{X: -7, Y: 1},
		{X: -10, Y: 1},
		{X: -1, Y: 0},
		{X: -10, Y: -1},
		{X: -7, Y: -1},
		{X: -4, Y: -1},
		{X: -2, Y: -1},
		{X: -1, Y: -1},
	}

	if len(rays) == 0 {
		return centerX, centerY, zettelVisualRay{}, 0, 0
	}

	seed := zettelVisualScatterSeed(index, count, connected)
	rayIndex := zettelVisualRayIndex(index, count, connected, len(rays), seed)
	ray := rays[rayIndex]
	maxSteps := zettelVisualMaxRaySteps(centerX, centerY, w, h, ray)
	if maxSteps < 1 {
		return centerX, centerY, ray, 0, maxSteps
	}

	distanceScale := 1.0
	if connected {
		distanceScale = 0.12 + math.Pow(1-closeness, 0.78)*0.36
	} else {
		distanceScale = 0.76
	}
	step := int(math.Round(float64(maxSteps) * distanceScale))
	if !connected && count > 1 {
		step += index%3 - 1
	}
	step += zettelVisualStepJitter(seed, maxSteps, connected)
	minStep := 2
	if maxSteps < minStep {
		minStep = maxSteps
	}
	if step < minStep {
		step = minStep
	}
	if step > maxSteps {
		step = maxSteps
	}

	return centerX + ray.X*step, centerY + ray.Y*step, ray, step, maxSteps
}

func zettelVisualRayIndex(index int, count int, connected bool, totalRays int, seed int) int {
	if totalRays <= 0 {
		return 0
	}
	if index < 0 {
		index = 0
	}

	if !connected {
		rayIndex := index % totalRays
		if count > 0 && count <= totalRays {
			rayIndex = index * totalRays / count
		}
		rayIndex += zettelVisualSignedJitter(seed, 2)
		return zettelVisualClamp(rayIndex, 0, totalRays-1)
	}

	lanes := []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22}
	if count > 0 && count <= len(lanes) {
		laneIndex := index * len(lanes) / count
		return zettelVisualClamp(lanes[laneIndex]+zettelVisualSignedJitter(seed, 1), 0, totalRays-1)
	}

	laneIndex := index % len(lanes)
	return zettelVisualClamp(lanes[laneIndex]+zettelVisualSignedJitter(seed, 1), 0, totalRays-1)
}

func zettelVisualScatterSeed(index int, count int, connected bool) int {
	seed := (index + 1) * 1103515245
	seed += (count + 17) * 12345
	if connected {
		seed += 7919
	} else {
		seed += 1543
	}
	if seed < 0 {
		seed = -seed
	}
	return seed
}

func zettelVisualSignedJitter(seed int, radius int) int {
	if radius <= 0 {
		return 0
	}
	return seed%(radius*2+1) - radius
}

func zettelVisualStepJitter(seed int, maxSteps int, connected bool) int {
	if maxSteps <= 3 {
		return 0
	}
	radius := maxSteps / 12
	if connected {
		radius = maxSteps / 10
	}
	if radius < 1 {
		radius = 1
	}
	return zettelVisualSignedJitter(seed/97, radius)
}

func zettelVisualResolveLabelOverlaps(nodes []ZettelVisualNode, centerX int, centerY int, w int, h int) {
	occupied := []zettelVisualLabelRect{}
	for i := range nodes {
		if i == 0 {
			occupied = append(occupied, zettelVisualNodeRect(nodes[i], w, h))
			continue
		}
		if nodes[i].MaxStep < 1 {
			occupied = append(occupied, zettelVisualNodeRect(nodes[i], w, h))
			continue
		}

		best := nodes[i]
		found := false
		for _, step := range zettelVisualCandidateSteps(nodes[i].Step, nodes[i].MaxStep) {
			candidate := nodes[i]
			candidate.Step = step
			candidate.X = zettelVisualClamp(centerX+candidate.Ray.X*step, 1, w-2)
			candidate.Y = zettelVisualClamp(centerY+candidate.Ray.Y*step, 0, h-1)
			rect := zettelVisualNodeRect(candidate, w, h)
			if !zettelVisualRectOverlapsAny(rect, occupied) {
				best = candidate
				found = true
				break
			}
		}
		nodes[i] = best
		occupied = append(occupied, zettelVisualNodeRect(nodes[i], w, h))
		if !found {
			continue
		}
	}
}

func zettelVisualCandidateSteps(start int, maxStep int) []int {
	steps := []int{}
	seen := map[int]bool{}
	add := func(step int) {
		if step < 1 || step > maxStep || seen[step] {
			return
		}
		seen[step] = true
		steps = append(steps, step)
	}
	add(start)
	for delta := 1; delta <= maxStep; delta++ {
		add(start + delta)
		add(start - delta)
	}
	return steps
}

type zettelVisualLabelRect struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
}

func zettelVisualNodeRect(node ZettelVisualNode, w int, h int) zettelVisualLabelRect {
	labelWidth := len([]rune(node.Label))
	x1 := zettelVisualClamp(node.X, 0, w-1)
	x2 := zettelVisualClamp(node.X+labelWidth-1, 0, w-1)
	y := zettelVisualClamp(node.Y, 0, h-1)
	return zettelVisualLabelRect{X1: x1, Y1: y, X2: x2, Y2: y}
}

func zettelVisualRectOverlapsAny(rect zettelVisualLabelRect, others []zettelVisualLabelRect) bool {
	for _, other := range others {
		if rect.Y1 != other.Y1 {
			continue
		}
		if rect.X1 <= other.X2 && rect.X2 >= other.X1 {
			return true
		}
	}
	return false
}

func zettelVisualMaxRaySteps(centerX int, centerY int, w int, h int, ray zettelVisualRay) int {
	maxSteps := 1 << 20
	if ray.X > 0 {
		maxSteps = zettelVisualMin(maxSteps, (w-2-centerX)/ray.X)
	} else if ray.X < 0 {
		maxSteps = zettelVisualMin(maxSteps, (centerX-1)/-ray.X)
	}
	if ray.Y > 0 {
		maxSteps = zettelVisualMin(maxSteps, (h-1-centerY)/ray.Y)
	} else if ray.Y < 0 {
		maxSteps = zettelVisualMin(maxSteps, centerY/-ray.Y)
	}
	if maxSteps == 1<<20 {
		return 0
	}
	return maxSteps
}

func zettelVisualMin(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func (z *Zettelkasten) zettelVisualDebugNodes(w int, h int) []ZettelVisualNode {
	if zettelVisualDebugAX < 0 {
		zettelVisualDebugAX = w / 4
	}
	if zettelVisualDebugAY < 0 {
		zettelVisualDebugAY = h / 2
	}
	if zettelVisualDebugBX < 0 {
		zettelVisualDebugBX = (w * 3) / 4
	}
	if zettelVisualDebugBY < 0 {
		zettelVisualDebugBY = h / 2
	}

	zettelVisualDebugAX = zettelVisualClamp(zettelVisualDebugAX, 0, w-1)
	zettelVisualDebugAY = zettelVisualClamp(zettelVisualDebugAY, 0, h-1)
	zettelVisualDebugBX = zettelVisualClamp(zettelVisualDebugBX, 0, w-1)
	zettelVisualDebugBY = zettelVisualClamp(zettelVisualDebugBY, 0, h-1)

	return []ZettelVisualNode{
		{ID: "A", Label: "A", X: zettelVisualDebugAX, Y: zettelVisualDebugAY, Strength: 3},
		{ID: "B", Label: "B", X: zettelVisualDebugBX, Y: zettelVisualDebugBY, Connected: true, Strength: 2},
	}
}

func zettelVisualDebugToggleSelected() {
	if zettelVisualDebugSelected == "A" {
		zettelVisualDebugSelected = "B"
		return
	}
	zettelVisualDebugSelected = "A"
}

func zettelVisualDebugMoveSelected(dx int, dy int) {
	if zettelVisualDebugSelected == "A" {
		zettelVisualDebugAX += dx
		zettelVisualDebugAY += dy
		return
	}
	zettelVisualDebugBX += dx
	zettelVisualDebugBY += dy
}

// ============================================================
// GRID HELPERS
// ============================================================

func zettelVisualBlankGrid(w int, h int) [][]zettelVisualCell {
	grid := make([][]zettelVisualCell, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]zettelVisualCell, w)
		for x := 0; x < w; x++ {
			grid[y][x] = zettelVisualCell{Ch: ' '}
		}
	}
	return grid
}

func zettelVisualGridLines(grid [][]zettelVisualCell) []string {
	lines := make([]string, len(grid))
	for y := range grid {
		var b strings.Builder
		lastStyle := ""
		for _, cell := range grid[y] {
			ch := cell.Ch
			if ch == 0 {
				ch = ' '
			}
			style := zettelVisualStyle(cell.Strength, cell.Bold, ch)
			if style != lastStyle {
				if lastStyle != "" {
					if strings.Contains(lastStyle, "‹b") {
						b.WriteString("›b ")
					}
					b.WriteString("¤ ")
				}
				b.WriteString(style)
				lastStyle = style
			}
			b.WriteRune(ch)
		}
		if lastStyle != "" {
			if strings.Contains(lastStyle, "‹b") {
				b.WriteString("›b ")
			}
			b.WriteString("¤ ")
		}
		lines[y] = b.String()
	}
	return lines
}

func zettelVisualStyle(strength int, bold bool, ch rune) string {
	if ch == ' ' || ch == 0 {
		return ""
	}
	boldOn := ""
	if bold {
		boldOn = "‹b "
	}
	switch strength {
	case 3:
		return "¤AB0 " + boldOn
	case 2:
		return "¤AB0 " + boldOn
	case 1:
		return "¤AB0 " + boldOn
	default:
		return "¤A80 " + boldOn
	}
}

func zettelVisualDrawNode(grid [][]zettelVisualCell, w int, h int, node ZettelVisualNode) {
	label := node.Label
	labelRunes := []rune(label)
	for i, ch := range labelRunes {
		x := node.X + i
		if x >= w {
			break
		}
		zettelVisualPut(grid, w, h, x, node.Y, ch, node.Strength, node.Bold)
	}
}

func zettelVisualPut(grid [][]zettelVisualCell, w int, h int, x int, y int, ch rune, strength int, bold bool) {
	if y < 0 || y >= h || x < 0 || x >= w || ch == 0 {
		return
	}
	grid[y][x].Ch = ch
	if strength > grid[y][x].Strength {
		grid[y][x].Strength = strength
	}
	grid[y][x].Bold = grid[y][x].Bold || bold
	grid[y][x].Label = true
}

func zettelVisualSign(n int) int {
	if n < 0 {
		return -1
	}
	if n > 0 {
		return 1
	}
	return 0
}

func zettelVisualAbs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func zettelVisualClamp(v int, min int, max int) int {
	if max < min {
		return min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func zettelVisualHeightChar(dy int, frac float64) rune {
	up := []rune("_⎽⎼⎻⎺‾")
	down := []rune("‾⎺⎻⎼⎽_")

	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}

	i := int(frac * float64(len(up)-1))

	if i < 0 {
		i = 0
	}
	if i >= len(up) {
		i = len(up) - 1
	}

	if dy < 0 {
		return up[i]
	}

	return down[i]
}

func zettelVisualHeightCharFromRealY(dy int, realY float64) rune {
	frac := realY - math.Floor(realY)

	if frac < 0 {
		frac += 1
	}

	up := []rune("_⎽⎼⎻⎺‾")
	down := []rune("‾⎺⎻⎼⎽_")

	i := int(frac * float64(len(up)))

	if i < 0 {
		i = 0
	}
	if i >= len(up) {
		i = len(up) - 1
	}

	if dy < 0 {
		return up[i]
	}

	return down[i]
}

func zettelVisualHeightCharFromFrac(frac float64) rune {
	chars := []rune("‾⎺⎻⎼⎽_")

	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}

	i := int(frac * float64(len(chars)-1))

	if i < 0 {
		i = 0
	}
	if i >= len(chars) {
		i = len(chars) - 1
	}

	return chars[i]
}

func zettelVisualMonotonicUpChar(dy int, y int, prevY int) rune {
	if dy < 0 {
		// linha subindo: não reinicia pattern baixo->alto
		if y != prevY {
			return '╱'
		}
		return '‾'
	}

	if dy > 0 {
		// linha descendo
		if y != prevY {
			return '╲'
		}
		return '_'
	}

	return '─'
}

func zettelVisualSmallStepLevel(t float64) int {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	// 6 níveis visuais:
	// 0 1 2 3 4 5
	return int(t * 5.999)
}

func zettelVisualSmallStepChar(dy int, t float64, prevLevel int) rune {
	up := []rune("_⎽⎼⎻⎺‾")
	down := []rune("‾⎺⎻⎼⎽_")

	level := zettelVisualSmallStepLevel(t)

	if level < 0 {
		level = 0
	}
	if level > 5 {
		level = 5
	}

	if dy < 0 {
		return up[level]
	}

	if dy > 0 {
		return down[level]
	}

	return '─'
}

func zettelVisualSmallStepCharByRealY(dy int, fromY int, realY float64) rune {
	up := []rune("_⎽⎼⎻⎺‾")
	down := []rune("‾⎺⎻⎼⎽_")

	progress := math.Abs(realY - float64(fromY))

	frac := progress - math.Floor(progress)

	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}

	i := int(frac * float64(len(up)-1))

	if i < 0 {
		i = 0
	}
	if i >= len(up) {
		i = len(up) - 1
	}

	if dy < 0 {
		return up[i]
	}

	if dy > 0 {
		return down[i]
	}

	return '─'
}

func zettelVisualSmallStepFromError(dy int, err int, maxErr int) rune {
	if maxErr <= 0 {
		return '─'
	}

	up := []rune("⎽⎼⎻⎺")
	down := []rune("⎺⎻⎼⎽")

	t := float64(err)
	if t < 0 {
		t = -t
	}

	frac := t / float64(maxErr)
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}

	i := int(frac * float64(len(up)-1))

	if i < 0 {
		i = 0
	}
	if i >= len(up) {
		i = len(up) - 1
	}

	if dy < 0 {
		return up[i]
	}

	if dy > 0 {
		return down[i]
	}

	return '─'
}
