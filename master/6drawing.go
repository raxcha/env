package master

import "env/engine"

func (m *Master) Draw() {

	queues := []engine.Queue{}

	switch m.Mode {
	case "fibonacci":
		for _, client := range m.Clients {
			q := client.Draw()
			if q != nil {
				queues = append(queues, *q)
			}
		}
	default:
		if len(m.Clients) > 0 && m.Focus >= 0 && m.Focus < len(m.Clients) {
			q := m.Clients[m.Focus].Draw()
			if q != nil {
				queues = append(queues, *q)
			}
		}
	}

	if m.Mode == "fibonacci" {
		if q := m.fibonacciDividersQueue(); len(q.Frames) > 0 {
			queues = append(queues, q)
		}
	}
	if m.Tabs.On {
		queues = append(queues, m.Tabs.Draw())
	}
	if m.Menu.On {
		queues = append(queues, m.Menu.Draw())
	}
	if m.Status.On {
		q := m.Status.Draw()
		if q != nil {
			queues = append(queues, *q)
		}
	}
	if m.Notifications.On {
		q := m.Notifications.Draw()
		if q != nil {
			queues = append(queues, *q)
		}
	}

	finalqueue := m.Utilities.MergeQueues(queues...)

	select {
	case m.Engine.Queue <- finalqueue:
	default:
	}
}
