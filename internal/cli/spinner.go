package cli

import (
	"fmt"
	"sync"
	"time"
)

type spinner struct {
	done  chan struct{}
	chars []rune
	index int
	once  sync.Once
}

func newSpinner() *spinner {
	return &spinner{
		chars: []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'},
		done:  make(chan struct{}),
	}
}

func (s *spinner) start(banner string) {
	go func() {
		if banner != "" {
			fmt.Println(banner)
		}

		for {
			select {
			case <-s.done:
				return
			default:
				fmt.Printf("    %c\r", s.chars[s.index])
				s.index = (s.index + 1) % len(s.chars)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (s *spinner) stop() {
	s.once.Do(func() {
		close(s.done)

		// Clear spinner line.
		fmt.Print("\r\033[K")
	})
}
