package services

import (
	"io"
	"log" //nolint:depguard // Don't feel like using slog

	"github.com/micutio/airspottr/internal/domain/reference"
)

// Notify prints console summaries of sighting statistics.
type Notify struct {
	Stdout log.Logger
}

func NewNotify(appName string, consoleOut io.Writer) *Notify {
	_ = appName
	return &Notify{
		Stdout: *log.New(consoleOut, "", 0),
	}
}

// PrintSummary prints the highest, fastest and the most and the least common types.
func (notify *Notify) PrintSummary(dash *Dashboard) {
	notify.Stdout.Println("=== Summary ===")
	notify.listByRarity("aircraft", dash.SeenType)
	notify.listByRarity("operator", dash.SeenOperator)
	notify.listByRarity("country", dash.SeenCountry)
	notify.Stdout.Println("Fastest Aircraft:")
	fastest := dash.GetFastest()
	notify.Stdout.Println(fastest.AircraftToString())
	notify.Stdout.Println("Highest Aircraft:")
	highest := dash.GetHighest()
	notify.Stdout.Println(highest.AircraftToString())
	notify.Stdout.Println("=== End Summary ===")
}

func (notify *Notify) listByRarity(propertyName string, propertyCountMap map[string]int) {
	propertyCounts := reference.GetSortedCountsForProperty(propertyCountMap)

	notify.Stdout.Printf("Rarity from least to most common %s", propertyName)
	for j := range propertyCounts {
		notify.Stdout.Printf("%6d - %s\n", propertyCounts[j].Count, propertyCounts[j].Property)
	}
}
