package cmds

import (
	"os"

	"github.com/fatih/color"
)

// colorToStderrF prints informational text to stderr, only if the output is a
// terminal
func colorToStderrF(c color.Attribute, fmt string, a ...any) {
	// If stdout is piped to a file, don't bother printing anything
	// to avoid noise when trying to pipe output to other commands.
	if color.NoColor {
		return
	}
	_, err := color.New(c).Fprintf(os.Stderr, fmt, a...)
	if err != nil {
		panic(err)
	}
}
