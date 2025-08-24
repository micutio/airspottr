// Package main provides the flight tracking application
package main

import (
	"fmt"

	"github.com/micutio/flighttrack/tickerapp"
	"github.com/spf13/pflag"
)

// TODO: Argument parsing, e.g.: `--tui` to run app either as TUI or pure cli.
// TODO: Change title of the console!
func main() {
	// Define the flag with a long name ("name") and a short name ("n").

	isLaunchTicker := pflag.BoolP(
		"ticker",
		"t",
		false,
		"print plane spotting information on the command line without TUI")

	pflag.Lookup("ticker").NoOptDefVal = "true"
	pflag.Parse()

	if *isLaunchTicker {
		tickerapp.Run()
	} else {
		fmt.Println("TUI feature coming soon!")
	}
}
