package master

import (
	"env/cli/editor"
	"time"
)

// PrepareTemplate keeps pending log edits and leaves persistence to Ctrl+S.
func (m *Master) PrepareTemplate(reference string) bool {
	now := time.Now()
	path := "log/" + now.Format("2006.01.02")
	for i, client := range m.Clients {
		if e, ok := client.(*editor.Editor); ok && e.GetSpec() == path {
			// An editor still loading cannot safely accept an append yet.
			if e.Page == nil {
				return false
			}
			e.AppendTemplate(reference, now)
			m.Focus = i
			m.Layout = "visual"
			m.applySizes()
			return true
		}
	}
	req := m.Filesystem.NewLoading()
	req.Path = path
	req.Depth = 0
	req.CreateDraft = true
	ok, page := m.Filesystem.LoadLocal(req)
	if !ok || page == nil {
		return false
	}
	e := editor.CreateEditorFromPage(m, page)
	e.AppendTemplate(reference, now)
	m.Clients = append(m.Clients, e)
	m.Focus = len(m.Clients) - 1
	m.Layout = "visual"
	m.applySizes()
	return true
}
