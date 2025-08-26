// Package main provides the flight tracking application
package main

import (
	"github.com/gen2brain/beeep"
	"github.com/micutio/airspottr/internal"
	"github.com/micutio/airspottr/tickerapp"
	"github.com/micutio/airspottr/tuiapp"
	"github.com/spf13/pflag"
)

const (
	// thisAppName is the name of this application as shown on notifications.
	thisAppName = "airspottr"
)

var (
	argIsUseTicker bool
	argLatLon      []float32
)

// TODO: Predefine some locations to make launching the app less cumbersome:
// - Singapore, New York, Hamburg, ...
func setupCommandLineFlags() {
	// Whether to launch the Ticker or TUI app.
	pflag.BoolVarP(
		&argIsUseTicker,
		"ticker",
		"t",
		false,
		"print plane spotting information on the command line without TUI")
	pflag.Lookup("ticker").NoOptDefVal = "true"

	// Location to plane spot, provided as lat,lon coordinates
	pflag.Float32SliceVarP(
		&argLatLon,
		"latlon",
		"l",
		[]float32{0, 0},
		"define the location where to spot planes")
}

func main() {
	beeep.AppName = thisAppName //nolint:reassign // This is the only way to set app name in beeep.

	setupCommandLineFlags()

	// Parse all arguments provided to the program on launch.
	pflag.Parse()

	options := internal.RequestOptions{
		Lat: argLatLon[0],
		Lon: argLatLon[1],
	}

	if argIsUseTicker {
		tickerapp.Run(options)
	} else {
		tuiapp.Run(options)
	}
}
