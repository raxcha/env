package master

import (
	"env/engine"
	"env/routines"
	"math/rand"
	"strings"
	"time"
)

func (m *Master) Draw() {

	queues := []engine.Queue{}

	if len(m.Clients) == 0 {
		if q := m.emptyBackgroundQueue(); len(q.Frames) > 0 {
			queues = append(queues, q)
		}
	}

	switch m.Mode {
	case "fibonacci":
		for _, client := range m.Clients {
			q := client.Draw()
			if q != nil {
				queues = append(queues, *q)
			}
		}
	default:
		if len(m.Clients) > 0 && m.Focus >= 0 && m.Focus < len(m.Clients) {
			q := m.Clients[m.Focus].Draw()
			if q != nil {
				queues = append(queues, *q)
			}
		}
	}

	if m.Mode == "fibonacci" {
		if q := m.fibonacciDividersQueue(); len(q.Frames) > 0 {
			queues = append(queues, q)
		}
	}
	if m.Tabs.On {
		queues = append(queues, m.Tabs.Draw())
	}
	if m.Menu.On {
		queues = append(queues, m.Menu.Draw())
	}
	if m.Status.On {
		q := m.Status.Draw()
		if q != nil {
			queues = append(queues, *q)
		}
	}
	if m.Notifications.On {
		q := m.Notifications.Draw()
		if q != nil {
			queues = append(queues, *q)
		}
	}

	finalqueue := m.Utilities.MergeQueues(queues...)

	select {
	case m.Engine.Queue <- finalqueue:
	default:
		select {
		case <-m.Engine.Queue:
		default:
		}
		select {
		case m.Engine.Queue <- finalqueue:
		default:
		}
	}
}

func (m *Master) emptyBackgroundQueue() engine.Queue {
	if m == nil || m.Utilities == nil || len(m.Size) < 2 || m.Size[0] <= 0 || m.Size[1] <= 0 {
		return engine.Queue{}
	}

	rainQueue := m.emptyMatrixRainQueue()
	skullQueue := m.emptySkullQueue()
	return m.Utilities.MergeQueues(rainQueue, skullQueue)
}

func (m *Master) emptySkullQueue() engine.Queue {
	artFrames := emptyBackgroundArtFrames()
	motionFrames := emptyBackgroundSkullMotionFrames(artFrames, m.EmptySkullDirection)
	artW := 0
	artH := 0
	topAnchor := emptyBackgroundTopAnchor(motionFrames)
	for _, lines := range motionFrames {
		h := len(lines) + 2
		if h > artH {
			artH = h
		}
		leftPad := emptyBackgroundFrameLeftPad(lines, topAnchor)
		for _, line := range lines {
			if w := leftPad + m.Utilities.VisibleLength(line); w > artW {
				artW = w
			}
		}
	}
	if artW <= 0 || artH <= 0 {
		return engine.Queue{}
	}

	x := (m.Size[0] - artW) / 2
	if x < 0 {
		x = 0
	}

	topReserve := 0
	if m.Tabs != nil && m.Tabs.On {
		topReserve = 1
	}
	bottomReserve := 0
	if m.Status != nil && m.Status.On {
		bottomReserve = 1
	}
	availableH := m.Size[1] - topReserve - bottomReserve
	if availableH < 1 {
		availableH = m.Size[1]
		topReserve = 0
	}

	y := topReserve + (availableH-artH)/2
	if y < topReserve {
		y = topReserve
	}
	if y < 0 {
		y = 0
	}

	frames := []engine.Frame{}
	start := emptyBackgroundTimedPhase(emptyBackgroundSkullFrameDurations(len(motionFrames)))
	for i := 0; i < len(motionFrames); i++ {
		idx := (start + i) % len(motionFrames)
		lines := motionFrames[idx]
		normalizedLines := m.emptyBackgroundNormalizeFrame(lines, artW, artH, topAnchor)
		normalizedLines = emptyBackgroundAddFooter(normalizedLines, artW)
		frame := m.Utilities.GenerateFrame(
			engine.Boundaries{Fullsize: m.Size, Pos: routines.Bound{x, y}, Size: routines.Bound{artW, artH}},
			emptyBackgroundStyledLines(normalizedLines),
			emptyBackgroundSkullFrameDuration(idx),
		)
		if frame != nil {
			emptyBackgroundTransparentOutsideSpaces(frame, normalizedLines, x, y, artW)
			frames = append(frames, *frame)
		}
	}

	return engine.Queue{Size: m.Size, Frames: frames, Cycle: len(frames) > 1}
}

func randomEmptySkullDirection() int {
	if rand.New(rand.NewSource(time.Now().UnixNano())).Intn(2) == 0 {
		return 1
	}
	return -1
}

func emptyBackgroundSkullMotionFrames(frames [][]string, direction int) [][]string {
	if len(frames) == 0 {
		return nil
	}
	if direction == 0 {
		direction = 1
	}

	const count = 240
	out := make([][]string, 0, count)
	idx := 0
	for range count {
		out = append(out, frames[idx])
		idx = positiveMod(idx+direction, len(frames))
	}
	return out
}

func emptyBackgroundSkullFrameDurations(count int) []int {
	durations := make([]int, count)
	for i := range durations {
		durations[i] = emptyBackgroundSkullFrameDuration(i)
	}
	return durations
}

func emptyBackgroundSkullFrameDuration(idx int) int {
	weights := []int{5, 3, 3, 4, 3, 3, 4, 3, 3, 4}
	return weights[positiveMod(idx, len(weights))] * 30
}

func emptyBackgroundTimedPhase(durations []int) int {
	total := 0
	for _, duration := range durations {
		if duration > 0 {
			total += duration
		}
	}
	if total <= 0 {
		return 0
	}

	pos := int(time.Now().UnixMilli() % int64(total))
	for i, duration := range durations {
		if duration <= 0 {
			continue
		}
		if pos < duration {
			return i
		}
		pos -= duration
	}
	return 0
}

func emptyBackgroundPhase(frameMS int64, count int) int {
	if frameMS <= 0 || count <= 0 {
		return 0
	}
	return int((time.Now().UnixMilli() / frameMS) % int64(count))
}

func emptyBackgroundTopAnchor(frames [][]string) int {
	anchor := 0
	for _, lines := range frames {
		if len(lines) == 0 {
			continue
		}
		first := emptyBackgroundFirstNonSpace(lines[0])
		if first > anchor {
			anchor = first
		}
	}
	return anchor
}

func emptyBackgroundFrameLeftPad(lines []string, topAnchor int) int {
	if len(lines) == 0 {
		return 0
	}
	first := emptyBackgroundFirstNonSpace(lines[0])
	if first < 0 || first >= topAnchor {
		return 0
	}
	return topAnchor - first
}

func emptyBackgroundAddFooter(lines []string, width int) []string {
	label := "prsnl.spc"
	if width <= 0 {
		return lines
	}

	out := append([]string(nil), lines...)
	if len(out) >= 2 {
		out[len(out)-2] = centerText(label, width)
	}
	return out
}

func centerText(text string, width int) string {
	textW := len([]rune(text))
	if textW >= width {
		return text
	}
	return strings.Repeat(" ", (width-textW)/2) + text
}

func emptyBackgroundFirstNonSpace(line string) int {
	for i, r := range []rune(line) {
		if r != ' ' {
			return i
		}
	}
	return -1
}

func emptyBackgroundTransparentOutsideSpaces(frame *engine.Frame, lines []string, x int, y int, width int) {
	if frame == nil || len(frame.Size) < 2 || width <= 0 {
		return
	}

	fullW := frame.Size[0]
	for row, line := range lines {
		first, last := emptyBackgroundLineBody(line)
		for col := 0; col < width; col++ {
			cellX := x + col
			cellY := y + row
			if cellX < 0 || cellX >= frame.Size[0] || cellY < 0 || cellY >= frame.Size[1] {
				continue
			}

			i := cellY*fullW + cellX
			if i < 0 || i >= len(frame.Cells) || frame.Cells[i].Char != ' ' {
				continue
			}

			if col < first || col > last {
				frame.Cells[i].Char = 0
			}
		}
	}
}

func emptyBackgroundLineBody(line string) (int, int) {
	runes := []rune(line)
	first := len(runes)
	last := -1
	for i, r := range runes {
		if r != ' ' {
			if i < first {
				first = i
			}
			if i > last {
				last = i
			}
		}
	}
	return first, last
}

func (m *Master) emptyMatrixRainQueue() engine.Queue {
	w, h := m.Size[0], m.Size[1]
	if w <= 0 || h <= 0 {
		return engine.Queue{}
	}

	columns := emptyMatrixRainColumns(w, h)
	start := emptyBackgroundPhase(70, 240)
	frames := []engine.Frame{}
	for i := 0; i < 240; i++ {
		step := (start + i) % 240
		frame := m.Utilities.GenerateFrame(
			engine.Boundaries{Fullsize: m.Size, Pos: routines.Bound{0, 0}, Size: routines.Bound{w, h}},
			emptyMatrixRainLines(w, h, step, columns),
			70,
		)
		if frame != nil {
			m.fadeEmptyMatrixRain(frame)
			frames = append(frames, *frame)
		}
	}

	return engine.Queue{Size: m.Size, Frames: frames, Cycle: len(frames) > 1}
}

type emptyMatrixRainColumn struct {
	Active        bool
	Period        int
	Phase         int
	Tail          int
	GlyphOffset   int
	FlickerPeriod int
	FlickerPhase  int
}

func emptyMatrixRainColumns(w int, h int) []emptyMatrixRainColumn {
	seed := time.Now().UnixMilli() / 60000
	rng := rand.New(rand.NewSource(seed + int64(w*1009+h*9176)))
	columns := make([]emptyMatrixRainColumn, w)

	for x := 0; x < w; x++ {
		period := emptyMatrixRainRandomPeriod(rng, h)
		columns[x] = emptyMatrixRainColumn{
			Active:        rng.Intn(9) != 0,
			Period:        period,
			Phase:         rng.Intn(period),
			Tail:          7 + rng.Intn(13),
			GlyphOffset:   rng.Intn(2048),
			FlickerPeriod: []int{8, 10, 12, 15}[rng.Intn(4)],
			FlickerPhase:  rng.Intn(240),
		}
	}

	return columns
}

func emptyMatrixRainRandomPeriod(rng *rand.Rand, h int) int {
	choices := []int{48, 60, 80, 120, 240}
	minPeriod := h + 12
	filtered := []int{}
	for _, period := range choices {
		if period >= minPeriod {
			filtered = append(filtered, period)
		}
	}
	if len(filtered) == 0 {
		return 240
	}
	return filtered[rng.Intn(len(filtered))]
}

func emptyMatrixRainLines(w int, h int, step int, columns []emptyMatrixRainColumn) []string {
	lines := make([]string, h)
	for y := 0; y < h; y++ {
		var b strings.Builder
		b.Grow(w + 6)
		b.WriteString("§AB0 ")
		for x := 0; x < w; x++ {
			b.WriteRune(emptyMatrixRainCell(columns[x], x, y, step))
		}
		lines[y] = b.String()
	}
	return lines
}

func (m *Master) fadeEmptyMatrixRain(frame *engine.Frame) {
	if frame == nil || m == nil || m.Utilities == nil || m.Utilities.Theme == nil {
		return
	}
	bg := m.Utilities.Theme.Background
	for i := range frame.Cells {
		if frame.Cells[i].Char == 0 || frame.Cells[i].Char == ' ' || frame.Cells[i].Fg == nil {
			continue
		}
		fg := fadeRGB(bg, *frame.Cells[i].Fg, 0.38)
		frame.Cells[i].Fg = &fg
	}
}

func fadeRGB(bg engine.RGB, fg engine.RGB, amount float64) engine.RGB {
	return engine.RGB{
		R: int(float64(bg.R)*(1-amount) + float64(fg.R)*amount),
		G: int(float64(bg.G)*(1-amount) + float64(fg.G)*amount),
		B: int(float64(bg.B)*(1-amount) + float64(fg.B)*amount),
	}
}

func emptyMatrixRainCell(column emptyMatrixRainColumn, x int, y int, step int) rune {
	if !column.Active || column.Period <= 0 {
		return ' '
	}

	head := positiveMod(step+column.Phase, column.Period) - column.Tail
	dist := head - y

	if dist < 0 || dist > column.Tail {
		return ' '
	}
	if dist > 1 && positiveMod(step+y+column.FlickerPhase, column.FlickerPeriod) == 0 {
		return ' '
	}

	ch := emptyMatrixRainGlyph(x, y, step+dist+column.GlyphOffset)
	return ch
}

func emptyMatrixRainGlyph(x int, y int, step int) rune {
	const glyphs = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz@#$%&*+=-"
	runes := []rune(glyphs)
	return runes[positiveMod(x*11+y*7+step*13, len(runes))]
}

func positiveMod(v int, mod int) int {
	if mod <= 0 {
		return 0
	}
	v %= mod
	if v < 0 {
		v += mod
	}
	return v
}

func emptyBackgroundStyledLines(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = "§A80 ‹b " + line + "›b "
	}
	return out
}

func (m *Master) emptyBackgroundNormalizeFrame(lines []string, width int, height int, topAnchor int) []string {
	leftPad := emptyBackgroundFrameLeftPad(lines, topAnchor)
	out := make([]string, 0, height)
	for _, line := range lines {
		if leftPad > 0 {
			line = strings.Repeat(" ", leftPad) + line
		}
		out = append(out, line)
	}
	for len(out) < height {
		out = append(out, "")
	}
	return out
}

func emptyBackgroundArtFrames() [][]string {
	return [][]string{
		strings.Split(strings.TrimRight(emptyBackgroundArtFrameOne, "\n"), "\n"),
		strings.Split(strings.TrimRight(emptyBackgroundArtFrameTwo, "\n"), "\n"),
		strings.Split(strings.TrimRight(emptyBackgroundArtFrameThree, "\n"), "\n"),
		strings.Split(strings.TrimRight(emptyBackgroundArtFrameFour, "\n"), "\n"),
		strings.Split(strings.TrimRight(emptyBackgroundArtFrameFive, "\n"), "\n"),
		strings.Split(strings.TrimRight(emptyBackgroundArtFrameSix, "\n"), "\n"),
		strings.Split(strings.TrimRight(emptyBackgroundArtFrameSeven, "\n"), "\n"),
		strings.Split(strings.TrimRight(emptyBackgroundArtFrameEight, "\n"), "\n"),
		strings.Split(strings.TrimRight(emptyBackgroundArtFrameNine, "\n"), "\n"),
		strings.Split(strings.TrimRight(emptyBackgroundArtFrameTen, "\n"), "\n"),
	}
}

const emptyBackgroundArtFrameOne = "" +
	"       _,,._\n" +
	"    ,d$$$$$SIi:.\n" +
	"  ,$$$SSSS$$SSIi:.\n" +
	" j$$$$SSSS$$$SIIi:.\n" +
	".S$$$$SS$$$$$SSIi:.\n" +
	"j?ᵒ`‾`?S$SI7ᵒ\"ᵃ?IL:.\n" +
	"?:     $$S?     `?i'\n" +
	"iL    j$?$k.     I7\n" +
	"$$$b%d$'  `$b,_.dS:  \n" +
	"?SSIiS?    S$?I?$Si \n" +
	" ‾`?IS$L_,d$SIi:`ᵃ'     \n" +
	"    ?$$$SS$SIi'       \n" +
	"    j:?i:i?.:·:       \n" +
	"      \"` `^    "

const emptyBackgroundArtFrameTwo = "" +
	"        _:,,._\n" +
	"     ,d$$$$$SSIi:.\n" +
	"   ,$$$SSSSS$$$S$Ii::\n" +
	"  d$$$$SSSS$$$$SSiiI:.\n" +
	" d$$$$SS$$$$$SSSi:iII:.\n" +
	" jᵒ`^?SSI7ᵒ\"ᵃ?IL:iiISIi:\n" +
	" ?   :$I?     ?$:iIS$I:.\n" +
	"jL _,$?$L     j7b:iISi:\n" +
	"?$d$'  `$b,_,d$$$:iIi?\n" +
	"i$$:    S$?I?$$$S%u.?'\n" +
	" ‾?L_,d$SIi:`ᵃ?S?ᵃ` '\n" +
	"  I$$$S$$SIi'     ,'\n" +
	"  :i?i:i?.i.\n" +
	"      \"` `^"

const emptyBackgroundArtFrameThree = "" +
	"           _.,,._\n" +
	"      _,oS$$$$$SSIi:,_\n" +
	"    ,d$$$$$$$S$$$SSISi:.\n" +
	"   ,$$$$$SSSS$$$$SIISSIi:.\n" +
	"  j$$$$SS$$$$$$SIISS$$SIi:\n" +
	" .\"I7‾`ᵒᵃ$$iI$SiIS$$$$SIi:\n" +
	" `.jI    `?L:iIIS$$$$SII:.\n" +
	" ,ᵃ$$.    j$b:iIS$$S?iIi:.\n" +
	" ? i$L,_.d$d$:ijSSSliSi:.\n" +
	" i.j$S?S$$$$S%u,ᵒ?iISi::.\n" +
	",d$$SIi?ᵃ?$S?ᵃ`   ‾`‾.:'\n" +
	"i$$SI$$k.         ..'\n" +
	":?::.i?·^"

const emptyBackgroundArtFrameFour = "" +
	"            _.,,,._\n" +
	"       .,uS$$$$$$$SSi:,_\n" +
	"     ,d$$S$$$$S$$$SS$SIi:.\n" +
	"    j$S$$S7IS$$S$$$$S$SIi::\n" +
	"   d$$$7iIS$$$$$$SS$$SIIi:.\n" +
	"  $$$7:iIS$S$$$SIS$$$$SIi::\n" +
	"  j ?k:iISIS$S&IS$$$SSIIi::\n" +
	".d? j$b:iIj$$S7IIS$$S7IIi::.\n" +
	"`$$u$$7.:j$$SiiIS$$$SIIIi::.\n" +
	" biS$S$p,._?%u,`?S$SSIIi::.\n" +
	"d$$$Si?ᵃ?$S?ᵃ\"`^\"4ISIii::.\n" +
	"ᵃ?$$k.           .?7I:.\n" +
	"^:?i:\"`          .·'‾"

const emptyBackgroundArtFrameFive = "" +
	"           _.,,,._\n" +
	"      .,oS$$$$$$$SSi:,_\n" +
	"    ,d$$S$$$$S$$$$SSSIi:.\n" +
	"   j$S$$$$$S$$S$$$$$$Sli::\n" +
	"  j$$$$$$SS$$$$$$$S$$SSII:.\n" +
	" j$7iIS$$$$S$$$$$$IS$$SIi::\n" +
	"d$7:jS$SIiIS$$$$$SS$$$Si::.\n" +
	"?7.j$$Siid$$$$$$SS$$$S7i::.\n" +
	"j:.?$liid$$$$SIS$$$SSSIi::.\n" +
	"$k,_`$IS$$$$SIIS$$$SSIii:.\n" +
	"?'  ‾`ᵒᵃ$S?ᵃ::::iISSii::.\n" +
	" k.·:'   i7   ··::::::··\n" +
	" ``      \"'`"

const emptyBackgroundArtFrameSix = "" +
	"        _.,,._\n" +
	"     ,d$$$$$SSi:,\n" +
	" a,d$S$$$$$$$$SSIi:.\n" +
	" dS$$$$$S$$S$$$$$SIk.\n" +
	":S$$$$SS$$$$$$$$$$SIi\n" +
	"IS$$$$$S$$$$$$$$$SIi:\n" +
	"SS$$$$$7S$$$$$S$$SLi·\n" +
	"?$$$$$7jSS$$$$S$SIi:\n" +
	"ji$$S&j$Si?S$$SSIi:·\n" +
	"?$IuiI$$$Ski?S7Ii:·' \n" +
	" ?\\Sbp.`ᵒᵃ?S$7':i:·\n" +
	" ? \"‾ ·:::. .::.."

const emptyBackgroundArtFrameSeven = "" +
	"          _.,,._\n" +
	"      ,uS$$$$$SSIi:,\n" +
	"   ,d$$S$$$$$$$$$$SIi:.\n" +
	"  dIS$$$$$$$$S$$$$$SSSik\n" +
	" jIS$$$$$$$$$$$$$$$$SIiiL\n" +
	"·IIS$$$$$S$$$$$$$$$$SI:?$\n" +
	":iS$$$$$$7S$$$$SS$$SIii:?k\n" +
	":iS$$$$$7jIS$$$$SS$SIS::·?\n" +
	"·iIS$$S7j$SI?S$$$SSii7 · L\n" +
	" :iS$SSi$$$SL`?S$SIi?_·o$$\n" +
	"  ?ISi7 `ᵒᵃ?Sb,`ᵃᵒ'‾`  _`\"\n" +
	"   ᵃ?ᵃ'··:::·`?S$i'  · :\n" +
	"           ··  `?'"

const emptyBackgroundArtFrameEight = "" +
	"          _.,,._\n" +
	"     _,dS$$$$$SSIi:,\n" +
	"   ,dS$S$$$$$$$$SSSIi:.\n" +
	"  dIS$$$$$$$$S$$$$SSSSIk\n" +
	" dIS$$$$$$$$$$$$$$$SSSiiL\n" +
	"iLSS$$$$SS$$$$$$$$$SSIi?$k\n" +
	"SiSS$$S$SSS$$$$S$$SSIIi:S?\n" +
	"iiS$$$$ISSIS$$$$SSSIIi?·j.\n" +
	":iIS$SIIS$SI?S$$SIIi i7'jI$:\n" +
	" :iISSiiiS$SLi?SI ?ᵃ',od$S$\n" +
	"  ·:iSi:?S$SI?'ᵃᵒ'‾`^ᵒᵃ?Sk\n" +
	"    `ᵒᵃ.:?S$Si      ·::iI$$\n" +
	"          `ᵒᵃ'        \".ᵃ:'"

const emptyBackgroundArtFrameNine = "" +
	"        _.,,,._\n" +
	"    _,d$$$$$$S$Sik,\n" +
	"  ,i$$S$$$$$$$$SSSIk:\n" +
	" dISS$$$$$$$SS$$$$SSIk\n" +
	"jIS$$$$$$$$$SIIS$$$SSIk\n" +
	"SSS$$$$SS$$SSIIi$$S7ᵃ??k\n" +
	"?SS$$$$$$SSSIi:d$'   :?\n" +
	":SS$$$$$SSIi::j$S    j$7\n" +
	" ?IS$$SSIi::.,?$$$up%?$'\n" +
	"  ?ISIb,‾`ᵒ\"?S$$ᵃᵒ‾,d$'\n" +
	"   `ᵒ?$?'    `?ᵃk.:iIS$$k.\n" +
	"                `?.:iI$$i\n" +
	"                 \".\"ᵃ ^ '"

const emptyBackgroundArtFrameTen = "" +
	"        _.,,._\n" +
	"    _,d$$$$$$$$b.\n" +
	"  ,d$S$$$$S$$$$SSb.\n" +
	" dSS$SIS$$$$$$$$$Sib\n" +
	"jIS$$SII$$$$$$$$$S$Sk\n" +
	"SI$$IIiid$$S?iI$$$S7ᵃk.\n" +
	":iS$IIi:j7'     ?$S?   i\n" +
	".iSSI:::$$      j$$L   ?\n" +
	":?i:.,d$$k,_.,d$$7‾?p,$\n" +
	" iI:`ᵃ?$$$SS?I?$$7  $$?\n" +
	"  ?:.  ?S$$SII$$L.,J$\"'\n" +
	"   `.   ‾  :IS$S$$$Sk\n" +
	"           .:i?i:?:.?\n" +
	"                \"` `"
