package notifications

import "env/cli"

func CreateNotifications(parent cli.Parent) *Notifications {
	return &Notifications{
		Parent:    parent,
		Utilities: parent.GetUtilities(),
		On:        false,
		Items:     []*Notification{},
		Tick:      make(chan struct{}, 8),
	}
}
