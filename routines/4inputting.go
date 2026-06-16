package routines

import (
	"os"
	"time"
	"golang.org/x/sys/unix"
)

func (r *Routines) startInputting() {

	go func() {

		buf := make([]byte, 1)

		for {

			_, err := os.Stdin.Read(buf)

			if err != nil { return }

			b := buf[0]

			switch b {
				
			case 9:
				r.Input <- &Input{Key: "tab"}

			case 127:
				r.Input <- &Input{Key: "backspace"}

			case 8:
				r.Input <- &Input{Key: "ctrl+backspace"}

			case 13:
				r.Input <- &Input{Key: "enter"}

			case 10:
				r.Input <- &Input{Key: "ctrl+enter"}

			case 4:
				r.Input <- &Input{Key: "ctrl+d"}

			case 24:
				r.Input <- &Input{Key: "ctrl+x"}

			case 22:
				r.Input <- &Input{Key: "ctrl+v"}

			case 26:
				r.Input <- &Input{Key: "ctrl+z"}

			case 25:
				r.Input <- &Input{Key: "ctrl+y"}

			case 19:
				r.Input <- &Input{Key: "ctrl+s"}

			case 20:
				r.Input <- &Input{Key: "ctrl+t"}

			case 16:
				r.Input <- &Input{Key: "ctrl+p"}

			case 17:
				r.Input <- &Input{Key: "ctrl+q"}
				
			case 0:
				r.Input <- &Input{Key: "ctrl+space"}
				
			case 15:
				r.Input <- &Input{Key: "ctrl+o"}

			case 27:

				if !stdinHasData(25 * time.Millisecond) {
					r.Input <- &Input{Key: "esc"}
					break
				}

				seq := []byte{}

				for {

					if !stdinHasData(10 * time.Millisecond) {
						break
					}


					tmp := make([]byte, 1)
					_, err := os.Stdin.Read(tmp)
					if err != nil {
						r.Input <- &Input{Key: "esc"}
						break
					}

					seq = append(seq, tmp[0])

					if (tmp[0] >= 'A' && tmp[0] <= 'Z') ||
						(tmp[0] >= 'a' && tmp[0] <= 'z') ||
						tmp[0] == '~' {
						break
					}
				}

				s := string(seq)

				switch s {
				case "[A":
					r.Input <- &Input{Key: "up"}
				case "[B":
					r.Input <- &Input{Key: "down"}
				case "[C":
					r.Input <- &Input{Key: "right"}
				case "[D":
					r.Input <- &Input{Key: "left"}

				case "[1;5A":
					r.Input <- &Input{Key: "ctrl+up"}
				case "[1;5B":
					r.Input <- &Input{Key: "ctrl+down"}
				case "[1;5C":
					r.Input <- &Input{Key: "ctrl+right"}
				case "[1;5D":
					r.Input <- &Input{Key: "ctrl+left"}

				case "[3~":
					r.Input <- &Input{Key: "delete"}
				case "[3;5~":
					r.Input <- &Input{Key: "ctrl+delete"}
				case "[13;5u":
					r.Input <- &Input{Key: "ctrl+t"}

				default:
					r.Input <- &Input{Key: "esc"}
				}


			default:
				r.Input <- &Input{ 
					Key:  "char",
					Char: rune(b),	
				}
			}
		}
	}()
}

func stdinHasData(timeout time.Duration) bool {
	
	fds := []unix.PollFd{
		{
			Fd:     int32(os.Stdin.Fd()),
			Events: unix.POLLIN,
		},
	}

	ms := int(timeout / time.Millisecond)

	n, err := unix.Poll(fds, ms)
	if err != nil {
		return false
	}

	return n > 0 && (fds[0].Revents&unix.POLLIN) != 0
}