package editor

import (
	"strings"
	"unicode"
)

// Dollars embedded in path components are literal, not formatting delimiters.
func highlightDollars(line string, selected bool) string {
	runes := []rune(line)
	var out strings.Builder
	opening := true
	pathChar := func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("/\\._-~@#>=$", r)
	}
	for i, r := range runes {
		if r != '$' || (i > 0 && i+1 < len(runes) && pathChar(runes[i-1]) && pathChar(runes[i+1])) {
			out.WriteRune(r)
			continue
		}
		if opening {
			if selected {
				out.WriteRune('$')
			}
			out.WriteString("‹a ")
		} else {
			out.WriteString("›a ")
			if selected {
				out.WriteRune('$')
			}
		}
		opening = !opening
	}
	return out.String()
}
