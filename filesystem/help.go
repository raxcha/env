package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func calculatePathDepth(path string) int {

	path = filepath.Clean(path)
	if path == "." || path == "" {
		return 0
	}
	return len(strings.Split(path, "/"))
}

func getPaths(path string) (string, string, string) {

	path = filepath.Clean(path)
	root := getRootpath()
	relpath, err := filepath.Rel(root, path)
	if err == nil {
		return root, relpath, filepath.Join(root, relpath)
	}
	return root, path, filepath.Join(root, path)
}

func getRootpath() string {
	if root := strings.TrimSpace(os.Getenv("ENV_ROOT")); root != "" {
		return filepath.Clean(root)
	}

	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "/home/asdf/prsnl.spc"
	}

	return filepath.Join(home, "prsnl.spc")
}

func checkPath(abspath string) (bool, string) {

	info, err := os.Stat(abspath)
	if err != nil {
		return false, ""
	}
	if info.IsDir() {
		return true, "dir"
	}
	return true, "file"
}

func splitIntoTwo(str string, where string) (string, string) {

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

func splitIntoMore(str string, where string) []string {

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

func splitHeader(lines []string) ([]string, []string) {

	for i, line := range lines {
		if isDashedLine(line) {
			return lines[:i], lines[i+1:]
		}
	}
	return lines, []string{}
}

func isDashedLine(str string) bool {

	for _, r := range str {
		if r != '-' {
			return false
		}
	}
	return true
}

func isSeparatorLine(str string) bool {

	for _, r := range str {
		if r != '-' && r != ' ' {
			return false
		}
	}
	return true
}

func trimString(str string) string {

	return strings.Trim(str, ` "'`)
}

func trimStrings(strs []string) []string {

	for i, str := range strs {
		strs[i] = trimString(str)
	}
	return strs
}

func parseFilter(spec string) func(*Page) bool {

	if spec == "" {
		return nil
	}

	key, value := splitIntoTwo(spec, ":")

	switch key {
	case "time":
		_, t := parseTime(value)
		if t.IsZero() {
			return nil
		}
		return func(p *Page) bool {
			pt, ok := p.Metadata["time"].(time.Time)
			if !ok {
				return false
			}
			return !pt.Before(t)
		}
	}

	return nil
}

func parseTime(str string) (string, time.Time) {

	layout := "02.01.2006 15:04"
	t, err := time.Parse(layout, str)
	if err == nil {
		return "full", t
	}

	layout = "02.01.2006"
	t, err = time.Parse(layout, str)
	if err == nil {
		return "date", t
	}

	layout = "15:04"
	t, err = time.Parse(layout, str)
	if err == nil {
		return "time", t
	}

	return "error", time.Time{}
}
