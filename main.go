// Package main provides the flight tracking application
package main

import (
	"github.com/micutio/flighttrack/tickerapp"
)

// TODO: Argument parsing, e.g.: `--tui` to run app either as TUI or pure cli.
func main() {
	tickerapp.Run()
}
