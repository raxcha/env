package notifications

import (
	"env/engine"
	"strconv"
	"time"
)

func (n *Notifications) Add(message string, duration time.Duration, onCommit func(), onCancel func()) {
	n.mu.Lock()
	n.nextID++
	id := strconv.Itoa(n.nextID)
	notif := &Notification{
		ID:       id,
		Message:  message,
		Deadline: time.Now().Add(duration),
		OnCommit: onCommit,
		OnCancel: onCancel,
		cancel:   make(chan struct{}),
	}
	n.Items = append(n.Items, notif)
	n.On = true
	n.mu.Unlock()
	n.tick()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-notif.cancel:
				notif.actionOnce.Do(func() {
					if notif.OnCancel != nil {
						notif.OnCancel()
					}
					n.remove(id)
					n.tick()
				})
				return
			case t := <-ticker.C:
				if !t.Before(notif.Deadline) {
					notif.actionOnce.Do(func() {
						if notif.OnCommit != nil {
							notif.OnCommit()
						}
						n.remove(id)
						n.tick()
					})
					return
				}
				n.tick()
			}
		}
	}()
}

func (n *Notifications) CancelLatest() bool {
	n.mu.Lock()
	if len(n.Items) == 0 {
		n.mu.Unlock()
		return false
	}
	last := n.Items[len(n.Items)-1]
	n.mu.Unlock()
	last.closeOnce.Do(func() {
		close(last.cancel)
	})
	return true
}

func (n *Notifications) Resize(b *engine.Boundaries) {
	n.Bounds = b
}

func (n *Notifications) remove(id string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for i, item := range n.Items {
		if item.ID == id {
			n.Items = append(n.Items[:i], n.Items[i+1:]...)
			break
		}
	}
	if len(n.Items) == 0 {
		n.On = false
	}
}

func (n *Notifications) tick() {
	select {
	case n.Tick <- struct{}{}:
	default:
	}
}
