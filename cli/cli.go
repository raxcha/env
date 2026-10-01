package cli

import (
	"env/engine"
	"env/filesystem"
	"env/settings"
	"env/types"
)

type Client interface {
	GetType() string
	GetSpec() string
	GetName() string
	GetShortLabel() string
	GetLongLabel() string

	Resize(types.Dimensions)
	Input(types.Input)
	Draw() *types.Queue

	GetTokens() []string
	GetStage() string
}

type Parent interface {
	GetClients() []Client
	GetFocus() int
	GetMode() string
	GetLayout() string
	GetFocusedClient() Client
	GetSizes() *types.Dimensions

	SetFocus(int)
	GetClientsSize() int

	MoveClient(int, int)
	CloseClient(int)
	AddClients(string, string, bool)

	GetSettings() *settings.Settings
	GetFilesystem() *filesystem.Filesystem
	GetEngine() *engine.Engine
	GetClipboard() string
	SetClipboard(string)

}
