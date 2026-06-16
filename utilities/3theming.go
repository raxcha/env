package utilities

import (
	"encoding/json"
	"env/engine"
	"os"
)

type ColorMod func(engine.RGB) engine.RGB

func (t Theme) Color(r rune, mods ...ColorMod) *engine.RGB {
	color := t.colorRaw(r)

	for _, mod := range mods {
		color = mod(color)
	}

	return &color
}

func mapTheme (t *rawTheme) *Theme {

	return &Theme {
		Name:    t.Name,
		Author:  t.Author,
		Variant: t.Variant,
		Raw:     *t,

		Black:   HexToRGB(t.Color01),
		Red:     HexToRGB(t.Color02),
		Green:   HexToRGB(t.Color03),
		Yellow:  HexToRGB(t.Color04),
		Blue:    HexToRGB(t.Color05),
		Magenta: HexToRGB(t.Color06),
		Cyan:    HexToRGB(t.Color07),
		White:   HexToRGB(t.Color08),

		BrightBlack:   HexToRGB(t.Color09),
		BrightRed:     HexToRGB(t.Color10),
		BrightGreen:   HexToRGB(t.Color11),
		BrightYellow:  HexToRGB(t.Color12),
		BrightBlue:    HexToRGB(t.Color13),
		BrightMagenta: HexToRGB(t.Color14),
		BrightCyan:    HexToRGB(t.Color15),
		BrightWhite:   HexToRGB(t.Color16),

		Background: HexToRGB(t.Background),
		Foreground: HexToRGB(t.Foreground),
		Cursor:     HexToRGB(t.Cursor),
	}
}

func (t *Theme) colorRaw(r rune) engine.RGB {
	switch r {
	case '0':
		return t.Black
	case '1':
		return t.Red
	case '2':
		return t.Green
	case '3':
		return t.Yellow
	case '4':
		return t.Blue
	case '5':
		return t.Magenta
	case '6':
		return t.Cyan
	case '7':
		return t.White

	case '8':
		return t.BrightBlack
	case '9':
		return t.BrightRed
	case 'a':
		return t.BrightGreen
	case 'b':
		return t.BrightYellow
	case 'c':
		return t.BrightBlue
	case 'd':
		return t.BrightMagenta
	case 'e':
		return t.BrightCyan
	case 'f':
		return t.BrightWhite

	case 'A':
		return t.Background
	case 'B':
		return t.Foreground
	case 'C':
		return t.Cursor

	default:
		return t.Foreground
	}
}

func LoadRawThemes(path string) ([]*rawTheme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var themes []*rawTheme
	if err := json.Unmarshal(data, &themes); err != nil {
		return nil, err
	}

	return themes, nil
}

func LoadThemes(path string) ([]*Theme, error) {
	rawThemes, err := LoadRawThemes(path)
	if err != nil {
		return nil, err
	}

	mapped := make([]*Theme, 0, len(rawThemes))

	for _, raw := range rawThemes {
		mapped = append(mapped, mapTheme(raw))
	}

	return mapped, nil
}

func LoadThemesMap(path string) (map[string]*Theme, error) {
	rawThemes, err := LoadRawThemes(path)
	if err != nil {
		return nil, err
	}

	mapped := make(map[string]*Theme, len(rawThemes))

	for _, raw := range rawThemes {
		theme := mapTheme(raw)

		if theme.Name == "" {
			continue
		}

		mapped[theme.Name] = theme
	}

	return mapped, nil
}