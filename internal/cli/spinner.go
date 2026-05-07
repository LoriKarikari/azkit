package cli

import (
	"fmt"
	"io"
	"time"
)

type terminalSpinner struct {
	frames []string
	index  int
	out    io.Writer
}

func newSpinner(out io.Writer) *terminalSpinner {
	return &terminalSpinner{frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}, out: out}
}

func (s *terminalSpinner) tick(label string, elapsed time.Duration) {
	frame := s.frames[s.index%len(s.frames)]
	s.index++
	fmt.Fprintf(s.out, "\r\033[K%s %s (%s)", frame, label, elapsed.Round(time.Second))
}

func (s *terminalSpinner) done(message string) {
	fmt.Fprintf(s.out, "\r\033[K%s\n", message)
}
