package routines

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"golang.org/x/term"
)

var originalTerminalState *term.State

func (r *Routines) prepareTerminal() {

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		os.Exit(1)
	}

	state, err := term.MakeRaw(fd)
	if err != nil {
		os.Exit(1)
	}
	originalTerminalState = state
	fmt.Print("\x1b[?1049h") // ... alternate screen ...
	fmt.Print("\x1b[?25l")   // ... hide cursor ...
	fmt.Print("\x1b[2J")     // ... clear screen ...
	fmt.Print("\x1b[H")      // ... move cursor home ...
	r.handleTerminalExit()
}

func (r *Routines) restoreTerminal() {

	if originalTerminalState != nil {
		term.Restore(int(os.Stdin.Fd()), originalTerminalState)
	}
	fmt.Print("\x1b[?25h")   // ... show cursor ...
	fmt.Print("\x1b[?1049l") // ... leave alternate screen ...
	fmt.Print("\x1b[0m")     // ... reset styling ...
}

func (r *Routines) handleTerminalExit() {

	sig := make(chan os.Signal, 1)
	signal.Notify(
		sig,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	go func() {
		<-sig
		r.restoreTerminal()
		os.Exit(0)
	}()
}

func (r *Routines) RestoreAndExit() {

	r.restoreTerminal()
	os.Exit(0)
}
