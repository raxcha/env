package master

import "strings"

var stageOrder = []string{
	"master",
	"tabs",
	"editor",
	"picker",
	"sidebar",
	"menu",
	"status",
	"notifications",
	"projects",
	"zettelkasten",
}

func NormalizeStage(stage string) string {
	stage = strings.ToLower(strings.TrimSpace(stage))
	if stage == "" || stage == "all" || stage == "full" {
		return stageOrder[len(stageOrder)-1]
	}

	for _, known := range stageOrder {
		if stage == known {
			return known
		}
	}

	return stageOrder[len(stageOrder)-1]
}

func (m *Master) stageEnabled(name string) bool {
	name = NormalizeStage(name)
	current := NormalizeStage(m.Stage)

	return stageIndex(current) >= stageIndex(name)
}

func stageIndex(name string) int {
	for i, known := range stageOrder {
		if name == known {
			return i
		}
	}

	return len(stageOrder) - 1
}

func (m *Master) applyStage() {
	m.Stage = NormalizeStage(m.Stage)

	if m.Tabs != nil {
		m.Tabs.On = m.stageEnabled("tabs")
	}

	if m.Menu != nil {
		m.Menu.On = false
	}

	if m.Status != nil {
		m.Status.On = m.stageEnabled("status")
	}

	if m.Notifications != nil && !m.stageEnabled("notifications") {
		m.Notifications.On = false
	}
}

func clientStage(name string) string {
	switch name {
	case "today", "tomorrow":
		return "editor"
	case "editor", "picker", "projects", "zettelkasten":
		return name
	default:
		return name
	}
}
