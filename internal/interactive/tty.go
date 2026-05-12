package interactive

import (
	"os"

	"github.com/mattn/go-isatty"
)

func IsTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// IsTerminalFn is the function used to detect whether stdin is a terminal.
// Tests can override this to force non-interactive mode.
var IsTerminalFn = IsTerminal
