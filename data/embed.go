// Package data embeds the static ICAO, airline, military, and registration reference CSV files.
package data

import "embed"

// Files holds the embedded CSV reference datasets.
//
//go:embed *.csv
var Files embed.FS
