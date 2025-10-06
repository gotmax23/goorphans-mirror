package cmds

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

// colorToStderrF prints informational text to stderr, only if the output is a
// terminal
func colorToStderrF(c color.Attribute, format string, a ...any) {
	// If stdout is piped to a file, don't bother printing anything
	// to avoid noise when trying to pipe output to other commands.
	if color.NoColor {
		return
	}
	_, err := color.New(c).Fprintf(os.Stderr, format, a...)
	if err != nil {
		panic(err)
	}
}

// colorToStderrForce prints informational text to stderr whether or not it's a
// terminal
func colorToStderrForce(c color.Attribute, format string, a ...any) {
	var err error
	if isatty.IsTerminal(os.Stderr.Fd()) {
		_, err = color.New(c).Fprintf(os.Stderr, format, a...)
	} else {
		_, err = fmt.Fprintf(os.Stderr, format, a...)
	}
	if err != nil {
		panic(err)
	}
}
