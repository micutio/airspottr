// Package main provides the flight tracking application
package main

import (
	"github.com/gen2brain/beeep"
	"github.com/micutio/airspottr/tickerapp"
	"github.com/micutio/airspottr/tuiapp"
	"github.com/spf13/pflag"
)

const (
	// thisAppName is the name of this application as shown on notifications.
	thisAppName = "FLTRK"
)

// TODO: Argument parsing, e.g.: `--tui` to run app either as TUI or pure cli.
// TODO: Change title of the console!
func main() {
	beeep.AppName = thisAppName //nolint:reassign // This is the only way to set app name in beeep.
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
		tuiapp.Run()
	}
}
