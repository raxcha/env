package notifications

import (
	"env/cli"
	"env/engine"
	"env/utilities"
	"sync"
	"time"
)

type Notification struct {
	ID         string
	Message    string
	Deadline   time.Time
	OnCommit   func()
	OnCancel   func()
	cancel     chan struct{}
	closeOnce  sync.Once
	actionOnce sync.Once
}

type Notifications struct {
	Parent    cli.Parent
	Bounds    *engine.Boundaries
	Utilities *utilities.Utilities

	On    bool
	Items []*Notification

	mu     sync.Mutex
	Tick   chan struct{}
	nextID int
}
