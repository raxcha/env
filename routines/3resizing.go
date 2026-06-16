package routines

import (
	"os"
	"os/signal"
	"syscall"
	"golang.org/x/term"
)

func (r *Routines) startResizing() {

	sigchan := make(chan os.Signal, 1)

	signal.Notify(sigchan, syscall.SIGWINCH)

	go func () {
		defer signal.Stop(sigchan)

		if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			r.Size <- &Bound{ w, h }
		}

		for range sigchan {

			width, height, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				continue
			}

			r.Size <- &Bound{ width, height }
		}
	} ()
}