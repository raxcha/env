package utilities

import (
	"env/engine"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func findNextSpace (runes []rune) int {

	for i, r := range runes {
		if r == ' ' { return i }
	}
	return -1
}

func markupSpan(runes []rune, i int) int {
	switch runes[i] {
	case '§', '¤':
		for j := i + 1; j < len(runes); j++ {
			if runes[j] == ' ' {
				return j - i + 1
			}
		}
		return len(runes) - i
	case '¬', '‹', '›':
		j := i + 1
		for j < len(runes) {
			if runes[j] == ' ' {
				j++
				break
			}
			if !strings.ContainsRune("biuUaAv", runes[j]) {
				break
			}
			j++
		}
		return j - i
	case '¦', '¶':
		return 1
	}
	return 0
}

func (u *Utilities) VisibleLength(str string) int {
	runes := []rune(str)
	count := 0
	for i := 0; i < len(runes); {
		if span := markupSpan(runes, i); span > 0 {
			i += span
		} else {
			count++
			i++
		}
	}
	return count
}

func (u *Utilities) ActualIndex(str string, visualIndex int) int {
	if visualIndex <= 0 {
		return 0
	}
	runes := []rune(str)
	visible := 0
	for i := 0; i < len(runes); {
		if span := markupSpan(runes, i); span > 0 {
			i += span
		} else {
			if visible == visualIndex {
				return i
			}
			visible++
			i++
		}
	}
	return len(runes)
}

func NewRGB(r int, g int, b int) engine.RGB {
	return engine.RGB{
		R: Clamp(r),
		G: Clamp(g),
		B: Clamp(b),
	}
}

func Clamp(v int) int {
	if v < 0 {
		return 0
	}

	if v > 255 {
		return 255
	}

	return v
}

func HexToRGB(hex string) engine.RGB {
	hex = strings.TrimSpace(hex)

	if !ValidHex(hex) {
		return engine.RGB{R: 255, G: 255, B: 255}
	}

	hex = strings.TrimPrefix(hex, "#")

	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)

	return NewRGB(int(r), int(g), int(b))
}

func ValidHex(hex string) bool {
	if len(hex) != 7 {
		return false
	}

	if hex[0] != '#' {
		return false
	}

	for _, r := range hex[1:] {
		if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
			return false
		}
	}

	return true
}

func Hex(c engine.RGB) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func ANSIBg(c engine.RGB) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", c.R, c.G, c.B)
}

func ANSIFg(c engine.RGB) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c.R, c.G, c.B)
}

func Lighten(amount float64) ColorMod {
	return func(c engine.RGB) engine.RGB {
		return MixRGB(c, engine.RGB{R: 255, G: 255, B: 255}, amount)
	}
}

func Darken(amount float64) ColorMod {
	return func(c engine.RGB) engine.RGB {
		return MixRGB(c, engine.RGB{R: 0, G: 0, B: 0}, amount)
	}
}

func Mix(other engine.RGB, amount float64) ColorMod {
	return func(c engine.RGB) engine.RGB {
		return MixRGB(c, other, amount)
	}
}

func AlphaOver(bg engine.RGB, alpha float64) ColorMod {
	return func(fg engine.RGB) engine.RGB {
		return MixRGB(bg, fg, alpha)
	}
}

func Saturate(amount float64) ColorMod {
	return func(c engine.RGB) engine.RGB {
		h, s, l := RGBToHSL(c)
		s += amount
		if s > 1 {
			s = 1
		}
		return HSLToRGB(h, s, l)
	}
}

func Desaturate(amount float64) ColorMod {
	return func(c engine.RGB) engine.RGB {
		h, s, l := RGBToHSL(c)
		s -= amount
		if s < 0 {
			s = 0
		}
		return HSLToRGB(h, s, l)
	}
}

func EnsureContrast(bg engine.RGB, minRatio float64) ColorMod {
	return func(fg engine.RGB) engine.RGB {
		if ContrastRatio(fg, bg) >= minRatio {
			return fg
		}

		if IsDark(bg) {
			for i := 0; i < 10; i++ {
				fg = MixRGB(fg, engine.RGB{R: 255, G: 255, B: 255}, 0.15)
				if ContrastRatio(fg, bg) >= minRatio {
					return fg
				}
			}
		} else {
			for i := 0; i < 10; i++ {
				fg = MixRGB(fg, engine.RGB{R: 0, G: 0, B: 0}, 0.15)
				if ContrastRatio(fg, bg) >= minRatio {
					return fg
				}
			}
		}

		return fg
	}
}

func MixRGB(a engine.RGB, b engine.RGB, amount float64) engine.RGB {
	amount = ClampFloat(amount, 0, 1)

	r := int(float64(a.R)*(1-amount) + float64(b.R)*amount)
	g := int(float64(a.G)*(1-amount) + float64(b.G)*amount)
	bl := int(float64(a.B)*(1-amount) + float64(b.B)*amount)

	return NewRGB(r, g, bl)
}

func ClampFloat(v float64, min float64, max float64) float64 {
	if v < min {
		return min
	}

	if v > max {
		return max
	}

	return v
}

func Luma(c engine.RGB) float64 {
	return 0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)
}

func IsDark(c engine.RGB) bool {
	return Luma(c) < 128
}

func BestTextOn(bg engine.RGB) engine.RGB {
	if IsDark(bg) {
		return engine.RGB{R: 255, G: 255, B: 255}
	}

	return engine.RGB{R: 0, G: 0, B: 0}
}

func ContrastRatio(a engine.RGB, b engine.RGB) float64 {
	la := RelativeLuminance(a)
	lb := RelativeLuminance(b)

	if la < lb {
		la, lb = lb, la
	}

	return (la + 0.05) / (lb + 0.05)
}

func RelativeLuminance(c engine.RGB) float64 {
	r := ChannelLuminance(float64(c.R) / 255.0)
	g := ChannelLuminance(float64(c.G) / 255.0)
	b := ChannelLuminance(float64(c.B) / 255.0)

	return 0.2126*r + 0.7152*g + 0.0722*b
}

func ChannelLuminance(c float64) float64 {
	if c <= 0.03928 {
		return c / 12.92
	}

	return math.Pow((c+0.055)/1.055, 2.4)
}

func RGBToHSL(c engine.RGB) (float64, float64, float64) {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0

	maxValue := math.Max(r, math.Max(g, b))
	minValue := math.Min(r, math.Min(g, b))

	h := 0.0
	s := 0.0
	l := (maxValue + minValue) / 2.0

	if maxValue == minValue {
		return h, s, l
	}

	d := maxValue - minValue

	if l > 0.5 {
		s = d / (2.0 - maxValue - minValue)
	} else {
		s = d / (maxValue + minValue)
	}

	switch maxValue {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}

	h /= 6

	return h, s, l
}

func HSLToRGB(h float64, s float64, l float64) engine.RGB {
	var r float64
	var g float64
	var b float64

	if s == 0 {
		r = l
		g = l
		b = l
	} else {
		q := 0.0
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}

		p := 2*l - q

		r = HueToRGB(p, q, h+1.0/3.0)
		g = HueToRGB(p, q, h)
		b = HueToRGB(p, q, h-1.0/3.0)
	}

	return NewRGB(
		int(math.Round(r*255)),
		int(math.Round(g*255)),
		int(math.Round(b*255)),
	)
}

func HueToRGB(p float64, q float64, t float64) float64 {
	if t < 0 {
		t += 1
	}

	if t > 1 {
		t -= 1
	}

	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}

	if t < 1.0/2.0 {
		return q
	}

	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}

	return p
}

func Blend(a, b engine.RGB, amount float64) engine.RGB {
	return MixRGB(a, b, amount)
}

func SoftOn(base engine.RGB, amount float64) ColorMod {
	return func(c engine.RGB) engine.RGB {
		return Blend(base, c, amount)
	}
}

func Calm(amount float64) ColorMod {
	return func(c engine.RGB) engine.RGB {
		gray := uint8((uint16(c.R) + uint16(c.G) + uint16(c.B)) / 3)

		return engine.RGB{
			R: int(float64(c.R)*(1-amount) + float64(gray)*amount),
			G: int(float64(c.G)*(1-amount) + float64(gray)*amount),
			B: int(float64(c.B)*(1-amount) + float64(gray)*amount),
		}
	}
}

func (u *Utilities) CutVisibleFrom(str string, start int) string {
	if start <= 0 {
		return str
	}
	runes := []rune(str)
	out := make([]rune, 0, len(runes))
	visible := 0
	for i := 0; i < len(runes); {
		span := markupSpan(runes, i)
		if span > 0 {
			if visible >= start {
				out = append(out, runes[i:i+span]...)
			}
			i += span
		} else {
			if visible >= start {
				out = append(out, runes[i])
			}
			visible++
			i++
		}
	}
	return string(out)
}