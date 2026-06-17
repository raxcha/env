package engine

import "time"

func (e *Engine) startOrchestra() {

	go func() {
		for queue := range e.Queue {
			if e.stop != nil {
				close(e.stop)
			}

			e.drainPendingFrames()

			stop := make(chan struct{})
			e.stop = stop

			go e.playQueue(queue, stop)
		}
	}()
}

func (e *Engine) drainPendingFrames() {
	for {
		select {
		case <-e.Frame:
		default:
			return
		}
	}
}

func (e *Engine) playQueue(queue Queue, stop <-chan struct{}) {

	if len(queue.Frames) == 0 {
		return
	}

	for {
		for _, frame := range queue.Frames {

			select {
			case <-stop:
				return

			case e.Frame <- frame:
			}

			if frame.Timeout > 0 {
				timer := time.NewTimer(time.Duration(frame.Timeout) * time.Millisecond)

				select {
				case <-stop:
					timer.Stop()
					return

				case <-timer.C:
				}
			}
		}

		if !queue.Cycle {
			return
		}
	}
}
