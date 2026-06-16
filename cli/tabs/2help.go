package tabs

import (
	"env/cli"
	"strings"
)

var switchKeys = []rune("asdfghjklqwertyuiopzxcvbnm")

func SwitchKeyForIndex(idx int) (rune, bool) {
	if idx < 0 || idx >= len(switchKeys) {
		return 0, false
	}

	return switchKeys[idx], true
}

func SwitchPrefixForIndex(idx int) string {
	key, ok := SwitchKeyForIndex(idx)
	if !ok {
		return ""
	}

	return string(key)
}

func NormalClientLabel(client cli.Cli, maxW int) string {
	kind := client.GetKind()
	spec := client.GetSpec()
	symbol := ClientSymbol(kind, false)

	if kind == "editor" {
		return symbol + " " + AbbrevPath(spec, maxW-visibleRuneLen(symbol)-1)
	}

	name := client.GetName()
	if name == "" {
		name = client.GetTitle()
	}
	if name == "" {
		name = kind
	}

	return symbol + " " + AbbrevText(name, maxW-visibleRuneLen(symbol)-1)
}

func NormalClientLabelStyled(client cli.Cli, maxW int, selected bool) string {
	kind := client.GetKind()
	spec := client.GetSpec()
	symbol := ClientSymbol(kind, selected)

	if kind == "editor" {
		return symbol + " " + AbbrevPath(spec, maxW-visibleRuneLen(symbol)-1)
	}

	name := client.GetName()
	if name == "" {
		name = client.GetTitle()
	}
	if name == "" {
		name = kind
	}

	return symbol + " " + AbbrevText(name, maxW-visibleRuneLen(symbol)-1)
}

func SwitchClientLabel(idx int, client cli.Cli, maxW int) string {
	key := SwitchPrefixForIndex(idx)

	if key == "" {
		key = "?"
	}

	kind := client.GetKind()
	spec := client.GetSpec()
	if spec == "." || spec == "" {
		spec = "prsnl.spc"
	}

	prefix := key + " " + kind + ":"

	if maxW > 0 {
		specW := maxW - runeLen(prefix)
		if specW < 1 {
			specW = 1
		}

		spec = AbbrevPath(spec, specW)
	}

	label := prefix + spec
	if maxW > 0 && runeLen(label) > maxW {
		return AbbrevText(label, maxW)
	}

	return label
}

func ClientSymbol(kind string, selected bool) string {
	symbol := clientSymbolGlyph(kind)
	bg := "8"
	reset := "¤8F "

	return "¤" + bg + clientSymbolColor(kind) + " " + symbol + reset
}

func clientSymbolGlyph(kind string) string {
	switch kind {
	case "editor":
		return ""
	case "picker":
		return "⌕"
	case "projects":
		return "󱌾"
	case "zettelkasten":
		return "󰀼"
	default:
		return "•"
	}
}

func clientSymbolColor(kind string) string {
	switch kind {
	case "editor":
		return "b"
	case "picker":
		return "e"
	case "projects":
		return "d"
	case "zettelkasten":
		return "a"
	default:
		return "F"
	}
}

func AbbrevPath(path string, maxW int) string {
	if maxW <= 0 {
		return ""
	}

	path = strings.TrimPrefix(path, "prsnl.spc/")

	if runeLen(path) <= maxW {
		return path
	}

	parts := strings.Split(path, "/")

	if len(parts) >= 2 {
		first := parts[0]
		last := parts[len(parts)-1]

		prefix := first + "/.../"
		rest := maxW - runeLen(prefix)

		if rest >= 1 {
			return prefix + AbbrevText(last, rest)
		}

		return cutRootPrefix(first, maxW)
	}

	return AbbrevText(path, maxW)
}

func AbbrevText(value string, maxW int) string {
	if maxW <= 0 {
		return ""
	}

	if runeLen(value) <= maxW {
		return value
	}

	return cutMiddleWithDots(value, maxW)
}

func cutRootPrefix(root string, maxW int) string {
	if maxW <= 0 {
		return ""
	}

	rootW := runeLen(root)
	if rootW >= maxW {
		return cutWithDots(root, maxW)
	}

	remaining := maxW - rootW
	if remaining <= 3 {
		return root + strings.Repeat(".", remaining)
	}

	dots := "/..."
	if remaining < runeLen(dots) {
		return root + strings.Repeat(".", remaining)
	}

	return root + dots
}

func cutLeft(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}

	runes := []rune(s)

	if len(runes) <= maxW {
		return s
	}

	return string(runes[len(runes)-maxW:])
}

func cutMiddleWithDots(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= maxW {
		return s
	}

	if maxW <= 3 {
		return string(runes[:maxW])
	}

	if maxW <= 8 {
		return string(runes[:maxW-3]) + "..."
	}

	mid := maxW - 6
	start := (len(runes) - mid) / 2
	return "..." + string(runes[start:start+mid]) + "..."
}

func cutWithDots(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}

	runes := []rune(s)

	if len(runes) <= maxW {
		return s
	}

	if maxW <= 3 {
		return string(runes[:maxW])
	}

	return string(runes[:maxW-3]) + "..."
}

func runeLen(s string) int {
	return len([]rune(s))
}

func visibleRuneLen(s string) int {
	runes := []rune(s)
	count := 0
	for i := 0; i < len(runes); {
		if runes[i] == '¤' {
			i++
			for i < len(runes) && runes[i] != ' ' {
				i++
			}
			if i < len(runes) {
				i++
			}
			continue
		}
		count++
		i++
	}
	return count
}

func findClosestStart(startings []int, pos int) int {
	if len(startings) == 0 {
		return -1
	}

	closest := startings[0]
	smallestDiff := abs(startings[0] - pos)

	for i := 1; i < len(startings); i++ {
		diff := abs(startings[i] - pos)

		if diff < smallestDiff {
			smallestDiff = diff
			closest = startings[i]
		}
	}

	return closest
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}
