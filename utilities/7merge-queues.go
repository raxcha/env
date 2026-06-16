package utilities

import (
	"env/engine"
	"math"
)

func (u *Utilities) MergeQueues(queues ...engine.Queue) engine.Queue {
	
	validQueues := make([]engine.Queue, 0, len(queues))
	for _, q := range queues {
		if len(q.Frames) > 0 {
			validQueues = append(validQueues, q)
		}
	}
	queues = validQueues

	if len(queues) == 0 {
		return engine.Queue{}
	}
	if len(queues) == 1 {
		return queues[0]
	}

	type qState struct {
		frameIdx   int
		timeLeft   int
		isFinished bool
	}

	states := make([]qState, len(queues))
	anyCycle := false

	for i := range queues {
		if len(queues[i].Frames) > 0 {
			states[i].timeLeft = queues[i].Frames[0].Timeout
		} else {
			states[i].isFinished = true
		}
		if queues[i].Cycle { anyCycle = true }
	}

	var resultFrames []engine.Frame
	maxIterations := 200

	for i := 0; i < maxIterations; i++ {

		allDone := true
		for j, q := range queues {
			if !states[j].isFinished || q.Cycle {
				allDone = false
				break
			}
		}
		if allDone && i > 0 {
			break
		}

		minStep := math.MaxInt
		for i := range states {
			if !states[i].isFinished && states[i].timeLeft > 0 {
				if states[i].timeLeft < minStep {
					minStep = states[i].timeLeft
				}
			}
		}
		if minStep == math.MaxInt {
			minStep = 1000
		}

		finalFrame := queues[0].Frames[states[0].frameIdx]
		
		for j := 1; j < len(queues); j++ {
			// MergeFrames(topo, fundo)
			finalFrame = u.MergeFrames(queues[j].Frames[states[j].frameIdx], finalFrame)
		}
		
		finalFrame.Timeout = minStep
		resultFrames = append(resultFrames, finalFrame)

		for j := range states {
			if !states[j].isFinished {
				states[j].timeLeft -= minStep
				if states[j].timeLeft <= 0 {
					states[j].frameIdx++
					if states[j].frameIdx >= len(queues[j].Frames) {
						if queues[j].Cycle {
							states[j].frameIdx = 0
						} else {
							states[j].isFinished = true
							states[j].frameIdx = len(queues[j].Frames) - 1
						}
					}
					states[j].timeLeft = queues[j].Frames[states[j].frameIdx].Timeout
				}
			}
		}
	}

	return engine.Queue{
		Size:   queues[0].Size,
		Frames: resultFrames,
		Cycle:  anyCycle,
	}
}