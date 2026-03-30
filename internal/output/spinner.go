package output

import (
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// Spinner displays a simple progress indicator during API calls.
type Spinner struct {
	message string
	done    chan struct{}
	mu      sync.Mutex
	active  bool
}

// NewSpinner creates a spinner with the given message.
// Returns nil if stdout is not a terminal (non-interactive mode).
func NewSpinner(message string) *Spinner {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return nil
	}
	return &Spinner{message: message, done: make(chan struct{})}
}

// Start begins the spinner animation.
func (s *Spinner) Start() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.active = true
	s.mu.Unlock()
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-s.done:
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%s %s", frames[i%len(frames)], s.message)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// Stop halts the spinner and clears the line.
func (s *Spinner) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active {
		return
	}
	s.active = false
	close(s.done)
	fmt.Fprintf(os.Stderr, "\r\033[K") // clear line
}
