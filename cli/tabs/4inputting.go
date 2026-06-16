package tabs

import "env/routines"

func (t *Tabs) Input(newinput *routines.Input) {

	// no clients, no tabs ...
	if len(t.Parent.GetClients()) == 0 {
		return
	}

	// in case of swtiching mode ...
	if t.Switch {
		// characters into quick switch ...
		if newinput.Key == "char" {
			if t.focusBySwitchKey(newinput.Char) {
				t.Switch = false
			}

			return
		}

		switch newinput.Key {

		// moving ...
		case "left":
			t.moveLeft()

		case "right":
			t.moveRight()

		// shifting places ...
		case "ctrl+left":
			t.moveToLeft()

		case "ctrl+right":
			t.moveToRight()

		}
	}
}

func (t *Tabs) moveLeft() {
	focus := t.Parent.GetFocus() - 1
	if focus < 0 {
		focus = 0
	}
	t.Parent.SetFocus(focus)
}

func (t *Tabs) moveRight() {
	focus := t.Parent.GetFocus() + 1
	if focus >= len(t.Parent.GetClients()) {
		focus = len(t.Parent.GetClients()) - 1
	}
	t.Parent.SetFocus(focus)
}

func (t *Tabs) moveToLeft() {
	focus := t.Parent.GetFocus()
	if focus > 0 {
		t.Parent.MoveClient(focus, focus-1)
	}
}

func (t *Tabs) moveToRight() {
	focus := t.Parent.GetFocus()
	clients := t.Parent.GetClients()
	if focus < len(clients)-1 {
		t.Parent.MoveClient(focus, focus+1)
	}
}

func (t *Tabs) focusBySwitchKey(char rune) bool {
	for i := range t.Parent.GetClients() {
		key, ok := SwitchKeyForIndex(i)
		if !ok {
			continue
		}

		if char == key {
			t.Parent.SetFocus(i)
			return true
		}
	}

	return false
}
