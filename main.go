// Package main provides the flight tracking application
package main

import (
	"github.com/micutio/airspottr/internal"
	"github.com/micutio/airspottr/tickerapp"
	"github.com/micutio/airspottr/tuiapp"
	"github.com/spf13/pflag"
)

const (
	// thisAppName is the name of this application as shown on notifications.
	thisAppName = "airspottr"
)

func main() {
	var argIsUseTicker bool
	var argLatLon []float32

	setupCommandLineFlags(&argIsUseTicker, &argLatLon)

	// Parse all arguments provided to the program on launch.
	pflag.Parse()

	options := internal.RequestOptions{
		Lat: argLatLon[0],
		Lon: argLatLon[1],
	}

	if argIsUseTicker {
		tickerapp.Run(thisAppName, options)
	} else {
		tuiapp.Run(thisAppName, options)
	}
}

// TODO: Predefine some locations to make launching the app less cumbersome:
// - Singapore, New York, Hamburg, ...
func setupCommandLineFlags(argIsUseTicker *bool, argLatLon *[]float32) {
	// Whether to launch the Ticker or TUI app.
	pflag.BoolVarP(
		argIsUseTicker,
		"ticker",
		"t",
		false,
		"print plane spotting information on the command line without TUI")
	pflag.Lookup("ticker").NoOptDefVal = "true"

	// Location to plane spot, provided as lat,lon coordinates
	pflag.Float32SliceVarP(
		argLatLon,
		"latlon",
		"l",
		[]float32{0, 0},
		"define the location where to spot planes")
}
