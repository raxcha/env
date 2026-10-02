package settings

import (

	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	
	"env/types"
	"env/utils"
)

type Settings struct {

	Options map[string]any

	Themes  map[string]types.Theme
	Theme   types.Theme
}

func CreateSettings(options map[string]any) *Settings {

	s := Settings{

		Options: options,
		Themes:  map[string]types.Theme{},
	}

	s.loadThemes()
	s.Reload(options)

	return &s
}

func (s *Settings) Reload(options map[string]any) {

	s.Options = options

	themeName, ok := s.Options["theme"].(string)
	if !ok {
		themeName = "Tokyo Night"
	}
	
	wanted, ok1 := s.Themes[themeName]
	if ok1 {
		s.Theme = wanted
	}

	fallback, ok2 := s.Themes["Tokyo Night"]
	if !ok1 && ok2 {
		s.Theme = fallback
	}
}

func (s *Settings) loadThemes() {

	data, err := os.ReadFile("themes.json")
	if err != nil {
		return
	}

	var rawThemes []map[string]any
	if err := json.Unmarshal(data, &rawThemes); err != nil {
		return
	}

	themes := map[string]types.Theme{}

	for _, raw := range rawThemes {

		name, ok := raw["name"].(string)
		if !ok {
			continue
		}

		hex := func(key string) types.Color {
			v, _ := raw[key].(string)
			return hexToRGB(v)
		}

		themes[name] = types.Theme{
			Color01:    hex("color_01"),
			Color02:    hex("color_02"),
			Color03:    hex("color_03"),
			Color04:    hex("color_04"),
			Color05:    hex("color_05"),
			Color06:    hex("color_06"),
			Color07:    hex("color_07"),
			Color08:    hex("color_08"),
			Color09:    hex("color_09"),
			Color10:    hex("color_10"),
			Color11:    hex("color_11"),
			Color12:    hex("color_12"),
			Color13:    hex("color_13"),
			Color14:    hex("color_14"),
			Color15:    hex("color_15"),
			Color16:    hex("color_16"),
			Background: hex("background"),
			Foreground: hex("foreground"),
			Cursor:     hex("cursor"),
		}
	}

	s.Themes = themes
}

func hexToRGB(hex string) types.Color {

	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return types.Color{}
	}

	val, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return types.Color{}
	}

	return types.Color{
		R: int(val >> 16 & 0xFF),
		G: int(val >> 8 & 0xFF),
		B: int(val & 0xFF),
	}
}

func (s *Settings) ChooseColor(r rune, kind string, blendRatio float64) string {

	var color types.Color

	switch r {
	case 'a', 'A':
		color = s.Theme.Color01
	case 'b', 'B':
		color = s.Theme.Color02
	case 'c', 'C':
		color = s.Theme.Color03
	case 'd', 'D':
		color = s.Theme.Color04
	case 'e', 'E':
		color = s.Theme.Color05
	case 'f', 'F':
		color = s.Theme.Color06
	case 'g', 'G':
		color = s.Theme.Color07
	case 'h', 'H':
		color = s.Theme.Color08
	case 'i', 'I':
		color = s.Theme.Color09
	case 'j', 'J':
		color = s.Theme.Color10
	case 'k', 'K':
		color = s.Theme.Color11
	case 'l', 'L':
		color = s.Theme.Color12
	case 'm', 'M':
		color = s.Theme.Color13
	case 'n', 'N':
		color = s.Theme.Color14
	case 'o', 'O':
		color = s.Theme.Color15
	case 'p', 'P':
		color = s.Theme.Color16
	case 'x', 'X':
		color = s.Theme.Background
	case 'y', 'Y':
		color = s.Theme.Foreground
	case 'z', 'Z':
		color = s.Theme.Cursor
	default:
		return ""
	}

	if strings.ToLower(string(r)) == string(r) { blendRatio = 0 }

	color = blendRGB(color, s.Theme.Background, blendRatio)

	layer := 38
	if kind == "bg" {
		layer = 48
	}

	return fmt.Sprintf("\x1b[%d;2;%d;%d;%dm", layer, color.R, color.G, color.B)
}

func blendRGB(a, b types.Color, ratio float64) types.Color {

	if ratio <= 0 {
		return a
	}
	if ratio >= 1 {
		return b
	}

	mix := func(x, y int) int {
		return int(float64(x)*(1-ratio) + float64(y)*ratio + 0.5)
	}

	return types.Color{
		R: mix(a.R, b.R),
		G: mix(a.G, b.G),
		B: mix(a.B, b.B),
	}
}

func (s *Settings) MergeFrames(frames ...*types.Frame) *types.Frame {

	if len(frames) == 0 {
		return &types.Frame{}
	}

	timeout := 0
	cells := append([]*types.Cell{}, frames[0].Cells...)

	minX, minY := frames[0].Sizes.Pos[0], frames[0].Sizes.Pos[1]
	maxX, maxY := minX+frames[0].Sizes.Size[0], minY+frames[0].Sizes.Size[1]

	for _, frame := range frames {
		if frame.Timeout > 0 && (timeout == 0 || frame.Timeout < timeout) {
			timeout = frame.Timeout
		}
	}

	for _, frame := range frames[1:] {

		if len(frame.Cells) != len(cells) {

			return &types.Frame{}
		}

		for y := 0; y < frame.Sizes.Size[1]; y++ {
			for x := 0; x < frame.Sizes.Size[0]; x++ {

				idx := (frame.Sizes.Pos[1]+y)*frame.Sizes.Full[0] + (frame.Sizes.Pos[0] + x)
				if idx < 0 || idx >= len(cells) {
					continue
				}

				cells[idx] = frame.Cells[idx]
			}
		}

		minX = min(minX, frame.Sizes.Pos[0])
		minY = min(minY, frame.Sizes.Pos[1])
		maxX = max(maxX, frame.Sizes.Pos[0]+frame.Sizes.Size[0])
		maxY = max(maxY, frame.Sizes.Pos[1]+frame.Sizes.Size[1])
	}

	sizes := types.Dimensions{
		Full: frames[0].Sizes.Full,
		Pos:  types.Pos{minX, minY},
		Size: types.Size{maxX - minX, maxY - minY},
	}

	return &types.Frame{Sizes: sizes, Cells: cells, Timeout: timeout}
}

func (s *Settings) MergeQueues(queues ...*types.Queue) *types.Queue {

	if len(queues) == 0 {
		return &types.Queue{}
	}

	length := 0
	cycle := false
	for _, q := range queues {
		if len(q.Frames) > length {
			length = len(q.Frames)
		}
		if q.Cycle {
			cycle = true
		}
	}

	frames := []*types.Frame{}

	for i := 0; i < length; i++ {

		current := []*types.Frame{}

		for _, q := range queues {

			if len(q.Frames) == 0 {
				continue
			}

			current = append(current, q.Frames[i%len(q.Frames)])
		}

		frames = append(frames, s.MergeFrames(current...))

	}

	return &types.Queue{Size: queues[0].Size, Frames: frames, Cycle: cycle}
}

func (s *Settings) GenerateQueue(size types.Size, frames []*types.Frame, cycle bool) *types.Queue {

	return &types.Queue{Size: size, Frames: frames, Cycle: cycle}
}

func (s *Settings) GenerateFrame(
	sizes types.Dimensions,
	text []string,
	timeout int,
	margin []int,
) *types.Frame {

	/*
		start: "§fc1 "
		colors: "¤f4 " , "¤ "
		toggle: "¬biuUaAv "
		on: "‹biuUaAv "
		off: "›biuUaAv "
		pause-toggle: "∅ "
	*/

	frameSizes := sizes

	sizes.Pos[0] += margin[3]
	sizes.Pos[1] += margin[0]
	sizes.Size[0] -= margin[1] + margin[3]
	sizes.Size[1] -= margin[0] + margin[2]

	if sizes.Size[0] < 0 {
		sizes.Size[0] = 0
	}
	if sizes.Size[1] < 0 {
		sizes.Size[1] = 0
	}

	cells := []types.Cell{}
	backgrounds := []string{}

	for _, line := range text {

		start := len(cells)
		size := 0
		wrap := 0
		style := s.NewStyle()

		runes := []rune(line)

		for j := 0; j < len(runes); j++ {
			r := runes[j]

			switch r {
			case '§', '¤', '¬', '‹', '›', '∅':

				k := utils.NextSpace(runes[j:])

				var unitWrap int
				style, unitWrap = s.InterpretRunes(style, runes[j:j+k])
				if r == '§' {
					wrap = unitWrap
				}

				j += k

				continue

			}

			if size == sizes.Size[0] {

				if wrap <= 0 || sizes.Size[0] == 0 {
					break
				}

				wrap--
				size = 0

			}

			cell := newCell(style)
			cell.Char = r
			cells = append(cells, cell)

			size++
		}

		for ; size < sizes.Size[0]; size++ {
			blank := newCell(style)
			blank.Char = ' '
			cells = append(cells, blank)
		}

		for range max(1, (len(cells)-start)/max(1, sizes.Size[0])) {
			backgrounds = append(backgrounds, style.StdAnsi.Bg)
		}
	}

	if sizes.Size[0] > 0 {
		rowsProduced := len(cells) / sizes.Size[0]
		for range sizes.Size[1] - rowsProduced {
			for range sizes.Size[0] {
				blank := newCell(s.NewStyle())
				blank.Char = ' '
				cells = append(cells, blank)
			}
		}
	}

	full := make([]*types.Cell, sizes.Full[0]*sizes.Full[1])
	for i := range full {
		blank := newCell(s.NewStyle())
		full[i] = &blank
	}

	for y := 0; y < min(sizes.Size[1], len(backgrounds)); y++ {
		for x := 0; x < frameSizes.Size[0]; x++ {

			posX := frameSizes.Pos[0] + x
			posY := sizes.Pos[1] + y
			if posX < 0 || posX >= sizes.Full[0] || posY < 0 || posY >= sizes.Full[1] { continue }

			full[posY*sizes.Full[0]+posX].Ansi.Bg = backgrounds[y]
		}
	}

	for y := 0; y < sizes.Size[1]; y++ {
		for x := 0; x < sizes.Size[0]; x++ {

			srcIdx := y*sizes.Size[0] + x
			dstIdx := (sizes.Pos[1]+y)*sizes.Full[0] + (sizes.Pos[0] + x)

			full[dstIdx] = &cells[srcIdx]
		}
	}

	return &types.Frame{Sizes: frameSizes, Cells: full, Timeout: timeout}
}

func newCell(style *types.Style) types.Cell {

	cell := types.Cell{
		Char:      ' ',
		Ansi:      style.Ansi,
		Bold:      style.Bold,
		Italic:    style.Italic,
		Underline: style.Underline,
		Invisible: false,
	}

	if style.All {
		cell.Bold = true
		cell.Italic = true
		cell.Underline = true

	} else if style.Almost {
		cell.Bold = true
		cell.Italic = false
		cell.Underline = true
	}

	return cell
}

func (s *Settings) NewStyle() *types.Style {

	return &types.Style{
		StdAnsi: types.Ansi{ Bg: s.ChooseColor('x', "bg", 0), Fg: s.ChooseColor('y', "fg", 0)},
		Ansi: types.Ansi{ Bg: s.ChooseColor('x', "bg", 0), Fg: s.ChooseColor('y', "fg", 0)},
		Bold:        false,
		Italic:      false,
		Underline:   false,
		Almost:      false,
		All:         false,
		Uppercase:   false,
		Invisible:   false,
	}
}

func (s *Settings) InterpretRunes(style *types.Style, runes []rune) (*types.Style, int) {

	wrap := 0
	switch runes[0] {

	case '§':
		if len(runes) < 3 {
			break
		}
		style.StdAnsi = types.Ansi{Bg: s.ChooseColor(runes[1], "bg", 0.5), Fg: s.ChooseColor(runes[2], "fg", 0.5)}
		style.Ansi = style.StdAnsi
		wrap, _ = strconv.Atoi(string(runes[3:]))

	case '¤':
		if len(runes) < 3 {
			style.Ansi = style.StdAnsi
		} else {
			style.Ansi = types.Ansi{Bg: s.ChooseColor(runes[1], "bg", 0.5), Fg: s.ChooseColor(runes[2], "fg", 0.5)}
		}

	case '¬':
		for _, r := range runes[1:] {
			switch r {
			case 'b':
				style.Bold = !style.Bold
			case 'i':
				style.Italic = !style.Italic
			case 'u':
				style.Underline = !style.Underline
			case 'U':
				style.Uppercase = !style.Uppercase
			case 'a':
				style.Almost = !style.Almost
			case 'A':
				style.All = !style.All
			case 'v':
				style.Invisible = !style.Invisible
			}
		}

	case '‹':
		for _, r := range runes[1:] {
			switch r {
			case 'b':
				style.Bold = true
			case 'i':
				style.Italic = true
			case 'u':
				style.Underline = true
			case 'U':
				style.Uppercase = true
			case 'a':
				style.Almost = true
			case 'A':
				style.All = true
			case 'v':
				style.Invisible = true
			}
		}

	case '›':
		for _, r := range runes[1:] {
			switch r {
			case 'b':
				style.Bold = false
			case 'i':
				style.Italic = false
			case 'u':
				style.Underline = false
			case 'U':
				style.Uppercase = false
			case 'a':
				style.Almost = false
			case 'A':
				style.All = false
			case 'v':
				style.Invisible = false
			}
		}

	}

	return style, wrap
}
