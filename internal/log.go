package internal

import (
	"io"
)

// LogParams contains the parameters for logging console output and errors.
// These will vary depending on whether the flight tracker runs in ticker or tui mode.
// # Ticker mode
// - console output goes to stdout
// - error logs go to stderr
// # TUI mode
// - console output goes to `dev/null` or os-dependent equivalents
// - error logs go to log file `/flighttrack.log`
// .
type LogParams struct {
	ConsoleOut io.Writer
	ErrorOut   io.Writer
}
