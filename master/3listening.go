package master

func (m *Master) startListening() {

	go func () {
		for {
			select {
				case newfullsize := <- m.Routines.Size:
					m.updateBoundaries(newfullsize)

				case newinput := <- m.Routines.Input:
					m.updateInput(newinput)

				case newreading := <- m.Filesystem.Page:
					m.updatePage(newreading)

				case <- m.Notifications.Tick:
					m.Draw()
			}
		}
	} ()
}

