package master

import "env/filesystem"

func (m *Master) updatePage(page *filesystem.Page) {

	for _, client := range m.Clients {
		if client.AcceptsPage(page) {
			client.SetPage(page)
		}
	}

	m.Draw()
}
