package utils

import (
	"crypto/rand"
	"math/big"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"env/types"
)

var PatchTime = time.Duration(5000) * time.Millisecond

func SplitLines(lines []string) ([]string, []string) {

	for i, line := range lines {
		if DashedLine(line) {
			return lines[:i], lines[i+1:]
		}
	}
	return lines, []string{}
}

func DashedLine(line string) bool {

	none := true
	for _, r := range line {
		if r != '-' {
			none = false
		}
		if r != ' ' && r != '-' {
			return false
		}
	}
	return none
}

func SplitTwo(str string, where string) (string, string) {

	if where == "" {
		return strings.TrimSpace(str), ""
	}

	idxfirst := strings.Index(str, where)
	if idxfirst == -1 {
		return strings.TrimSpace(str), ""
	}

	part1 := strings.TrimSpace(str[:idxfirst])

	idxlast := idxfirst
	for idxlast < len(str) && strings.ContainsRune(where, rune(str[idxlast])) {
		idxlast += len(where)
	}

	part2 := ""
	if idxlast < len(str) {
		part2 = strings.TrimSpace(str[idxlast:])
	}

	return part1, part2
}

func SplitMore(str string, where string) []string {

	if where == "" {
		return strings.Fields(str)
	}

	parts := strings.FieldsFunc(str, func(r rune) bool {
		return strings.ContainsRune(where, r)
	})

	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}

	return parts
}

func PathDepth(path string) int {

	path = filepath.Clean(path)
	if path == "." || path == "" {
		return 0
	}
	return len(strings.Split(path, "/"))
}

var weekdaySuffix = regexp.MustCompile(`\s+\([^()]+\)$`)

func ParseTime(str string) time.Time {

	str = strings.Join(strings.Fields(str), " ")
	str = weekdaySuffix.ReplaceAllString(str, "")

	layouts := []string{"2006.01.02 15:04", "15:04 2006.01.02", "2006.01.02"}
	for _, layout := range layouts {
		t, err := time.Parse(layout, str)
		if err == nil {
			return t
		}
	}

	t, _ := time.Parse(layouts[0], "2002.10.25 00:00")
	return t
}

func CompareContent(one, two []string) bool {

	if len(one) != len(two) {
		return false
	}
	for i := range one {
		if one[i] != two[i] {
			return false
		}
	}
	return true
}

func NewSizes(w, h int) *types.Dimensions {

	return &types.Dimensions{
		Full: types.Size{w, h},
		Pos:  types.Pos{0, 0},
		Size: types.Size{w, h},
	}
}

func CompareSize(one, two types.Size) bool {

	if len(one) != 2 || len(two) != 2 {
		return false
	}
	if one[0] != two[0] {
		return false
	}
	if one[1] != two[1] {
		return false
	}
	return true
}

func CompareCells(one, two *types.Cell) bool {

	if one.Bold != two.Bold {
		return false
	}
	if one.Italic != two.Italic {
		return false
	}
	if one.Underline != two.Underline {
		return false
	}
	if one.Char != two.Char {
		return false
	}
	if one.Invisible != two.Invisible {
		return false
	}
	if one.Ansi.Bg != two.Ansi.Bg {
		return false
	}
	if one.Ansi.Fg != two.Ansi.Fg {
		return false
	}
	return true

}

func NextSpace(runes []rune) int {

	for i, r := range runes {
		if r == ' ' {
			return i
		}
	}
	return len(runes)
}

func VisibleLength(line string) int {

	size := 0
	runes := []rune(line)
	for j := 0; j < len(runes); j++ {
		switch runes[j] {
		case '§', '¤', '¬', '‹', '›', '∅':
			j += NextSpace(runes[j:])
		default:
			size++
		}
	}

	return size
}

func VisibleSlice(line string, start int, end int) string {

	var sb strings.Builder
	size := 0
	runes := []rune(line)
	for j := 0; j < len(runes); j++ {
		switch runes[j] {
		case '§', '¤', '¬', '‹', '›', '∅':
			k := NextSpace(runes[j:])
			if end < 0 || size < end {
				sb.WriteString(string(runes[j : j+k+1]))
			}
			j += k
		default:
			if size >= start && (end < 0 || size < end) {
				sb.WriteRune(runes[j])
			}
			size++
		}
	}

	return sb.String()
}

func RealIndex(line string, visible int) int {

	size := 0
	runes := []rune(line)
	for j := 0; j < len(runes); j++ {

		switch runes[j] {
		case '§', '¤', '¬', '‹', '›', '∅':
			j += NextSpace(runes[j:])
			continue
		}

		if size == visible {
			return j
		}

		size++
	}

	return len(runes)
}

var superscriptMap = map[rune]rune{
	'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴',
	'5': '⁵', '6': '⁶', '7': '⁷', '8': '⁸', '9': '⁹',
	'+': '⁺', '-': '⁻', '=': '⁼', '(': '⁽', ')': '⁾',

	'a': 'ᵃ', 'b': 'ᵇ', 'c': 'ᶜ', 'd': 'ᵈ', 'e': 'ᵉ',
	'f': 'ᶠ', 'g': 'ᵍ', 'h': 'ʰ', 'i': 'ⁱ', 'j': 'ʲ',
	'k': 'ᵏ', 'l': 'ˡ', 'm': 'ᵐ', 'n': 'ⁿ', 'o': 'ᵒ',
	'p': 'ᵖ', 'q': 'ᵠ', 'r': 'ʳ', 's': 'ˢ', 't': 'ᵗ',
	'u': 'ᵘ', 'v': 'ᵛ', 'w': 'ʷ', 'x': 'ˣ', 'y': 'ʸ',
	'z': 'ᶻ',

	'A': 'ᴬ', 'B': 'ᴮ', 'D': 'ᴰ', 'E': 'ᴱ', 'G': 'ᴳ',
	'H': 'ᴴ', 'I': 'ᴵ', 'J': 'ᴶ', 'K': 'ᴷ', 'L': 'ᴸ',
	'M': 'ᴹ', 'N': 'ᴺ', 'O': 'ᴼ', 'P': 'ᴾ', 'R': 'ᴿ',
	'S': 'ˢ', 'T': 'ᵀ', 'U': 'ᵁ', 'V': 'ⱽ', 'W': 'ᵂ',
}

func SuperscriptString(s string) string {

	var sb strings.Builder
	for _, r := range s {
		if sup, ok := superscriptMap[r]; ok {
			sb.WriteRune(sup)
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func DrawBox(lines []string, margin []int) []string {

	top, right, bot, left := margin[0], margin[1], margin[2], margin[3]

	width := 0
	for _, line := range lines {
		if w := VisibleLength(line); w > width {
			width = w
		}
	}

	inner := left + width + right
	blank := "│" + strings.Repeat(" ", inner) + "│"

	box := []string{"┌" + strings.Repeat("─", inner) + "┐"}

	for i := 0; i < top; i++ {
		box = append(box, blank)
	}

	for _, line := range lines {
		padded := line + strings.Repeat(" ", width-VisibleLength(line))
		box = append(box, "│"+strings.Repeat(" ", left)+padded+strings.Repeat(" ", right)+"│")
	}

	for i := 0; i < bot; i++ {
		box = append(box, blank)
	}

	box = append(box, "└"+strings.Repeat("─", inner)+"┘")

	return box
}

var StageIcons = map[string]string{
	"move":     "↪",
	"ghost":    "󰊠",
	"draft":    "󱩽",
	"edit":     "󰤀",
	"local":    "",
	"api":      "󱂛",
	"*api":     "*󱂛",
	"doom":     "󱂥",
	"conflict": "",
}

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func GenerateId(n int) string {

	b := make([]byte, n)

	for i := range b {
		x, err := rand.Int(rand.Reader, big.NewInt(62))
		if err != nil {
			panic(err)
		}
		b[i] = alphabet[x.Int64()]
	}

	return string(b)
}

func ReplacePaired(s string, old string, first string, second string) string {

	if old == "" {
		return s
	}

	var sb strings.Builder
	count := 0

	for {
		idx := strings.Index(s, old)
		if idx == -1 {
			sb.WriteString(s)
			break
		}

		sb.WriteString(s[:idx])
		if count%2 == 0 {
			sb.WriteString(first)
		} else {
			sb.WriteString(second)
		}
		count++

		s = s[idx+len(old):]
	}

	return sb.String()
}

func ReplaceRegex(s string, pattern string, first string, second string) string {

	re, err := regexp.Compile(pattern)
	if err != nil {
		return s
	}

	return re.ReplaceAllStringFunc(s, func(match string) string {
		return first + match + second
	})
}
