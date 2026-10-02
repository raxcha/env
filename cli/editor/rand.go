package editor

import (
	"env/types"
	"env/utils"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"
)

// syncEventBlocks also updates linked drafts when saving from inside a block.
func (e *Editor) syncEventBlocks() {
	if e.Page == nil || e.Filesystem == nil || !strings.HasPrefix(filepath.ToSlash(filepath.Clean(e.Page.Path)), "log/") {
		return
	}
	for y, line := range e.Content {
		if (line == "" || !strings.ContainsRune("`~>=", rune(line[0]))) || y+1 >= len(e.Content) || !strings.HasPrefix(e.Content[y+1], "id: ") {
			continue
		}
		id := strings.TrimSpace(strings.TrimPrefix(e.Content[y+1], "id: "))
		if len(id) != 12 || strings.ContainsFunc(id, func(r rune) bool {
			return !(r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z')
		}) {
			continue
		}
		root, _, _ := eventTemplate(line[:1])
		e.syncEventBlock(e.Filesystem.CacheEventDraft(root, line[:1], id, eventTitle(line)), y)
	}
}

func (e *Editor) syncEventBlock(page *types.Page, start int) {
	if page.Stage != "draft" {
		return
	}
	bodyStart := min(start+2, len(e.Content))
	end := bodyStart
	for end < len(e.Content) && !blockBoundary(e.Content[end]) {
		end++
	}
	previous := slices.Clone(page.Content)
	content := e.Content[bodyStart:end]
	req := e.Filesystem.NewLoading()
	marker := e.Content[start][:1]
	root, name, _ := eventTemplate(marker)
	req.Path = root + "/." + marker + "abcdef123456$" + name
	req.Depth = 0
	if ok, template := e.Filesystem.LoadLocal(req); ok && template != nil {
		id, _, _ := strings.Cut(strings.TrimPrefix(filepath.Base(page.Path), marker), "$")
		expanded := strings.Join(utils.ExpandTemplate(template.Content, time.Now()), "\n")
		expanded = strings.ReplaceAll(expanded, ".{abcdef123456}", id)
		if value := ideaInteger(content); value != "" {
			expanded = strings.ReplaceAll(expanded, ".{int}", value)
		}
		// Insert the block last so its literal placeholders remain untouched.
		expanded = strings.ReplaceAll(expanded, ".{?content}", strings.Join(content, "\n"))
		content = strings.Split(expanded, "\n")
	}
	e.Filesystem.EditPage(page, content)
	if e.Parent == nil {
		return
	}
	for _, client := range e.Parent.GetClients() {
		if target, ok := client.(*Editor); ok && target != e && target.Page == page && slices.Equal(target.Content, previous) {
			target.Content = slices.Clone(page.Content)
			target.FoldedBlocks = nil
			target.clampcursor()
		}
	}
}

// ideaInteger returns the first whitespace-delimited integer marker as text.
func ideaInteger(content []string) string {
	for _, line := range content {
		for _, field := range strings.Fields(line) {
			if !strings.HasPrefix(field, "!") {
				continue
			}
			value := field[1:]
			digits := value
			if strings.HasPrefix(digits, "+") || strings.HasPrefix(digits, "-") {
				digits = digits[1:]
			}
			if digits == "" || strings.ContainsFunc(digits, func(r rune) bool { return r < '0' || r > '9' }) {
				continue
			}
			return value
		}
	}
	return ""
}

// eventTemplate maps log markers to their page templates.
func eventTemplate(marker string) (root, name string, ok bool) {
	switch marker {
	case "`":
		return "rand", "idea", true
	case "~":
		return "rec", "task", true
	case ">":
		return "rec", "plan", true
	case "=":
		return "rec", "session", true
	}
	return "", "", false
}

func eventTitle(line string) string {
	titleText := line[1:]
	return strings.Join(strings.FieldsFunc(strings.ToLower(titleText), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.'
	}), "-")
}
