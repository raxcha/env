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
	"os"
	"path/filepath"
	"strings"
)

const defaultSyncAddr = "https://prsnlspc.xyz"

func CreateMaster(arg string) *Master {
	return CreateMasterWithStage(arg, "")
}

func CreateMasterWithStage(arg string, stage string) *Master {

	m := &Master{
		Routines:   &routines.Routines{},
		Filesystem: &filesystem.Filesystem{},
		Engine:     &engine.Engine{},
		Utilities:  &utilities.Utilities{},

		Size:    routines.Bound{0, 0},
		Mode:    "monocle",
		Stage:   NormalizeStage(stage),
		Focus:   0,
		Clients: []cli.Cli{},
	}

	m.Routines = routines.CreateRoutines()
	m.Filesystem = filesystem.CreateFilesystem()
	m.Engine = engine.CreateEngine()
	m.Utilities = utilities.CreateUtilities()

	m.Tabs = tabs.CreateTabs(m)
	m.Menu = menu.CreateMenu(m.Utilities)
	m.Status = status.CreateStatus(m)
	m.Notifications = notifications.CreateNotifications(m)
	m.applyStage()

	if addr := syncAddr(); addr != "" {
		token := syncToken()
		m.SyncClient = apiclient.New(addr, token)
		m.Filesystem.Api = m.SyncClient
	}

	m.startListening()
	m.AddClients(arg, "end")
	m.initGhostCache()

	return m
}

func syncAddr() string {
	addr := strings.TrimSpace(os.Getenv("ENV_ADDR"))
	switch strings.ToLower(addr) {
	case "off", "none", "no", "false", "0":
		return ""
	case "":
		return defaultSyncAddr
	default:
		return addr
	}
}

func syncToken() string {
	for _, key := range []string{"ENV_TOKEN", "API_TOKEN"} {
		if token := strings.TrimSpace(os.Getenv(key)); token != "" {
			return token
		}
	}

	return envFileValue("API_TOKEN", apiEnvFileCandidates()...)
}

func apiEnvFileCandidates() []string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		home = "/home/asdf"
	}

	return []string{
		filepath.Join(home, "api", "env-api.env"),
		"/home/asdf/api/env-api.env",
	}
}

func envFileValue(key string, paths ...string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			name, value, ok := strings.Cut(line, "=")
			if !ok || strings.TrimSpace(name) != key {
				continue
			}
			return strings.Trim(strings.TrimSpace(value), `"'`)
		}
	}

	return ""
}
