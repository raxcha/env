package master

import (
	"env/apiclient"
	"env/cli"
	"env/cli/menu"
	"env/cli/notifications"
	"env/cli/status"
	"env/cli/tabs"
	"env/engine"
	"env/filesystem"
	"env/routines"
	"env/utilities"
)

type Master struct {
	Routines   *routines.Routines
	Filesystem *filesystem.Filesystem
	Engine     *engine.Engine
	Utilities  *utilities.Utilities

	Size routines.Bound

	Tabs          *tabs.Tabs
	Menu          *menu.Menu
	Status        *status.Status
	Notifications *notifications.Notifications
	SyncClient    *apiclient.Client

	Mode    string
	Stage   string
	Profile string
	Focus   int
	Clients []cli.Cli

	EmptySkullDirection int
}
