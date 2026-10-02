package master

import (
	"env/cli"
	"env/cli/launcher"
	"env/cli/notif"
	"env/cli/status"
	"env/cli/tabs"
	"env/types"
	"testing"
)

type inputClient struct {
	cli.Client
	inputs []types.Input
}

func (c *inputClient) Input(input types.Input) { c.inputs = append(c.inputs, input) }
func (c *inputClient) GetShortLabel() string   { return "test" }
func (c *inputClient) GetLongLabel() string    { return "test" }

func TestSpecialModeTabManagement(t *testing.T) {
	first, second := &inputClient{}, &inputClient{}
	m := &Master{Clients: []cli.Client{first, second}, Notif: &notif.Notif{}, Status: &status.Status{}}
	m.Tabs = tabs.CreateTabs(m)
	m.Launcher = launcher.CreateLauncher(m)

	send := func(description string, char rune, meta bool) {
		m.Inputting(&types.Input{Description: description, Char: char, Meta: meta})
	}
	send("esc", 0, true)
	send("right", 0, true)
	if m.Focus != 1 || !m.Tabs.On {
		t.Fatal("Esc mode should immediately enable tab navigation")
	}
	send("char", 'a', true)
	send("right", 0, true)
	if m.Focus != 1 || !m.Tabs.On {
		t.Fatal("quick switching should keep tab management active")
	}
	send("ctrl+left", 0, true)
	if m.Focus != 0 || m.Clients[0] != second {
		t.Fatal("tab reordering should be available in Esc mode")
	}
	send("tab", 0, true)
	if m.Focus != 0 || !m.Tabs.On {
		t.Fatal("Tab should not change tab management")
	}
	send("enter", 0, true)
	if !m.Launcher.On {
		t.Fatal("Enter should open the menu while managing tabs")
	}
	send("esc", 0, false)
	if m.Launcher.On || m.Tabs.On {
		t.Fatal("leaving Esc mode should close the menu and disable tab management")
	}
	send("tab", 0, false)
	if len(second.inputs) != 1 || second.inputs[0].Description != "tab" {
		t.Fatal("Tab should reach the client outside Esc mode")
	}
}
