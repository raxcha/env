package cli

import (
	"env/apiclient"
	"env/cli/menu"
	"env/filesystem"
	"env/utilities"
	"time"
)

type Parent interface {
	GetClients() []Cli
	GetFocus() int
	GetTheme() utilities.Theme
	GetUtilities() *utilities.Utilities
	GetFilesystem() *filesystem.Filesystem
	GetSyncClient() *apiclient.Client
	GetInteractionMode() string
	AddClients(path string, typ string)
	OpenLauncher(title string, items []menu.Item)
	RunTemplateAction(templatePath string)

	SetFocus(int)
	SetClients(Cli, int)
	MoveClient(from int, to int)

	QueueDelayedAction(label string, duration time.Duration, onCommit func(), onCancel func())
}
