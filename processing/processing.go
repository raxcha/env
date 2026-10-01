package processing

import (
	
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/term"

	"env/types"
	"env/utils"
)

type Processing struct {

	oldTerminalState *term.State

	bytes    chan []byte
	escTimer *time.Timer
	meta     bool

	Sizes chan *types.Dimensions
	Input chan *types.Input
}

func CreateProcessing() *Processing {

	p := Processing{

		oldTerminalState: nil,

		bytes:            make(chan []byte),
		escTimer:         time.NewTimer(0),
		meta:             true,

		Sizes:            make(chan *types.Dimensions),
		Input:            make(chan *types.Input),
	}

	p.prepareTerminal()
	p.startResizing()
	p.startInputting()

	return &p
}

func (p *Processing) prepareTerminal() {

	p.oldTerminalState, _ = term.MakeRaw(int(os.Stdin.Fd()))
	fmt.Print("\x1b[?1049h\x1b[?25l\x1b[2J\x1b[H\x1b[?7l\x1b[>1u")
}

func (p *Processing) RestoreTerminal() {

	fmt.Print("\x1b[<u\x1b[?25h\x1b[?1049l")
	if p.oldTerminalState != nil { term.Restore(int(os.Stdin.Fd()), p.oldTerminalState) }
}

func (p *Processing) startResizing() {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	go func() {

		w, h, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			p.Sizes <- utils.NewSizes(w, h)
		}

		for range sig {
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				continue
			}
			p.Sizes <- utils.NewSizes(w, h)
		}
	}()
}

func (p *Processing) startInputting() {

	go func() {
		
		for {
			buf := make([]byte, 32)
			n, err := os.Stdin.Read(buf)

			if err != nil {
				return
			}

			bytes := make([]byte, n)
			copy(bytes, buf[:n])

			if len(bytes) == 1 && bytes[0] == 0x1b {

				p.escTimer.Reset(30 * time.Millisecond)
				continue
			}

			p.bytes <- bytes
		}
	}()

	go func() {

		for {

			select {

			case bytes := <-p.bytes:
				p.parseBytes(bytes)

			case <-p.escTimer.C:

				p.meta = !p.meta
				p.parseBytes([]byte{0x1b})
			}
		}
	}()
}

func (p *Processing) parseBytes(bytes []byte) {

	m := kittyKeyPattern.FindSubmatch(bytes)
	
	if m != nil {
		p.parseKittyKey(m)
		return
	}

	if len(bytes) == 3 && bytes[0] == 27 && bytes[1] == 91 {
		switch bytes[2] {
		case 65:
			p.Input <- &types.Input{Meta: p.meta, Description: "up"}
		case 66:
			p.Input <- &types.Input{Meta: p.meta, Description: "down"}
		case 67:
			p.Input <- &types.Input{Meta: p.meta, Description: "right"}
		case 68:
			p.Input <- &types.Input{Meta: p.meta, Description: "left"}
		}

	} else if len(bytes) == 6 && bytes[0] == 27 && bytes[1] == 91 && bytes[2] == 49 && bytes[3] == 59 && bytes[4] == 53 {
		switch bytes[5] {
		case 65:
			p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+up"}
		case 66:
			p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+down"}
		case 67:
			p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+right"}
		case 68:
			p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+left"}
		}

	} else if len(bytes) == 4 && bytes[0] == 27 && bytes[1] == 91 && bytes[2] == 51 && bytes[3] == 126 {
		p.Input <- &types.Input{Meta: p.meta, Description: "delete"}

	} else if len(bytes) == 6 && bytes[0] == 27 && bytes[1] == 91 && bytes[2] == 51 && bytes[3] == 59 && bytes[4] == 53 && bytes[5] == 126 {
		p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+delete"}

	} else if len(bytes) == 1 {
		switch bytes[0] {
		case 13:
			p.Input <- &types.Input{Meta: p.meta, Description: "enter"}
		case 10:
			p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+enter"}
		case 127:
			p.Input <- &types.Input{Meta: p.meta, Description: "backspace"}
		case 8:
			p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+backspace"}
		case 9:
			p.Input <- &types.Input{Meta: p.meta, Description: "tab"}
		case 27:
			p.Input <- &types.Input{Meta: p.meta, Description: "esc"}
		case 32:
			p.Input <- &types.Input{Meta: p.meta, Description: "char", Char: ' '}
		case 0:
			p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+space"}

		default:
			if bytes[0] >= 97 && bytes[0] <= 122 {
				p.Input <- &types.Input{Meta: p.meta, Description: "char", Char: rune(bytes[0])}

			} else if bytes[0] >= 65 && bytes[0] <= 90 {
				p.Input <- &types.Input{Meta: p.meta, Description: "char", Char: rune(bytes[0])}

			} else if bytes[0] >= 1 && bytes[0] <= 26 {
				p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+" + string(rune('A'+(bytes[0]-1)))}

			} else if bytes[0] >= 48 && bytes[0] <= 57 {
				p.Input <- &types.Input{Meta: p.meta, Description: "number", Char: rune(bytes[0])}

			} else if (bytes[0] >= 33 && bytes[0] <= 47) || (bytes[0] >= 58 && bytes[0] <= 64) || (bytes[0] >= 91 && bytes[0] <= 96) || (bytes[0] >= 123 && bytes[0] <= 126) {
				p.Input <- &types.Input{Meta: p.meta, Description: "char", Char: rune(bytes[0])}

			}
		}
	}
}

var kittyKeyPattern = regexp.MustCompile(`^\x1b\[(\d+)(?:;(\d+))?u$`)

func (p *Processing) parseKittyKey(m [][]byte) {

	keycode, _ := strconv.Atoi(string(m[1]))

	modifier := 1
	if len(m[2]) > 0 {
		modifier, _ = strconv.Atoi(string(m[2]))
	}
	
	ctrl := (modifier-1)&4 != 0

	if keycode == 13 && ctrl {
		p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+enter"}
	} else if (keycode == 127 || keycode == 8) && ctrl {
		p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+backspace"}
	} else if keycode == 27 && !ctrl {
		p.meta = !p.meta
		p.Input <- &types.Input{Meta: p.meta, Description: "esc"}
	} else if ctrl && keycode >= 97 && keycode <= 122 {
		p.Input <- &types.Input{Meta: p.meta, Description: "ctrl+" + string(rune('A'+(keycode-97)))}
	}
}
