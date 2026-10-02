package utils

import (
	"regexp"
	"strings"
	"time"
)

var templateTime = regexp.MustCompile(`\.\{yyyy\.mm\.dd[^{}]*\}`)

// ExpandTemplate replaces supported time markers without changing the source.
func ExpandTemplate(content []string, now time.Time) []string {
	result := make([]string, len(content))
	for i, line := range content {
		result[i] = templateTime.ReplaceAllStringFunc(line, func(marker string) string {
			format := strings.Join(strings.Fields(marker[2:len(marker)-1]), " ")
			var layout string
			switch format {
			case "yyyy.mm.dd", "yyyy.mm.dd (weekday)":
				layout = "2006.01.02"
			case "yyyy.mm.dd hh:mm", "yyyy.mm.dd hh:mm (weekday)":
				layout = "2006.01.02 15:04"
			default:
				return marker
			}
			value := now.Format(layout)
			if strings.HasSuffix(format, "(weekday)") {
				value += " " + strings.ToLower(now.Weekday().String())
			}
			return value
		})
	}
	return result
}
