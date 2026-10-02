package editor

import (
	"env/utils"
	"strings"
	"time"
)

// AppendTemplate inserts the template with expanded dates into the log buffer.
// Saving creates the session through ensureLogEvent.
func (e *Editor) AppendTemplate(reference string, now time.Time) {
	var content []string
	if e.Filesystem != nil && strings.Contains(reference, "/") {
		req := e.Filesystem.NewLoading()
		req.Path = "proj/" + reference
		req.Depth = 0
		if ok, page := e.Filesystem.LoadLocal(req); ok && page != nil {
			content = utils.ExpandTemplate(page.Content, now)
		}
	}
	project, template, hasTemplate := strings.Cut(reference, "/")
	project, _, _ = strings.Cut(project, "$")
	reference = project
	if hasTemplate {
		reference += "." + strings.TrimPrefix(template, ".")
	}
	e.savestate()
	if len(e.Content) > 0 && e.Content[len(e.Content)-1] != "" {
		e.Content = append(e.Content, "")
	}
	start := len(e.Content)
	e.Content = append(e.Content, now.Format("15:04"), "= "+reference, "id: "+utils.GenerateId(12))
	e.Content = append(e.Content, content...)
	e.Cursor = [3]int{2, start + 1, 0}
	e.EditingPath = false
	e.FoldedBlocks = nil
	e.Redo = nil
	e.savestate()
}
