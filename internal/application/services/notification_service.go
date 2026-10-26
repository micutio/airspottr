package services

import (
	"fmt"
	"io"

	"github.com/micutio/airspottr/internal/domain/reference"
)

// Notify prints console summaries of sighting statistics.
type Notify struct {
	out io.Writer
}

func NewNotify(appName string, consoleOut io.Writer) *Notify {
	_ = appName
	if consoleOut == nil {
		consoleOut = io.Discard
	}
	return &Notify{
		out: consoleOut,
	}
}

// PrintSummary prints the highest, fastest and the most and the least common types.
func (notify *Notify) PrintSummary(dash *Dashboard) {
	_, _ = fmt.Fprintln(notify.out, "=== Summary ===")
	notify.listByRarity("aircraft", dash.SeenType)
	notify.listByRarity("operator", dash.SeenOperator)
	notify.listByRarity("country", dash.SeenCountry)
	_, _ = fmt.Fprintln(notify.out, "Fastest Aircraft:")
	fastest := dash.GetFastest()
	_, _ = fmt.Fprintln(notify.out, fastest.AircraftToString())
	_, _ = fmt.Fprintln(notify.out, "Highest Aircraft:")
	highest := dash.GetHighest()
	_, _ = fmt.Fprintln(notify.out, highest.AircraftToString())
	_, _ = fmt.Fprintln(notify.out, "=== End Summary ===")
}

func (notify *Notify) listByRarity(propertyName string, propertyCountMap map[string]int) {
	propertyCounts := reference.GetSortedCountsForProperty(propertyCountMap)

	_, _ = fmt.Fprintf(notify.out, "Rarity from least to most common %s\n", propertyName)
	for j := range propertyCounts {
		_, _ = fmt.Fprintf(notify.out, "%6d - %s\n", propertyCounts[j].Count, propertyCounts[j].Property)
	}
}
