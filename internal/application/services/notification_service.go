package services

import (
	"fmt"
	"io"

	"github.com/micutio/airspottr/internal/domain/reference"
)

// StdoutNotifier prints console summaries of sighting statistics.
type StdoutNotifier struct {
	out io.Writer
}

func NewStdoutNotifier(consoleOut io.Writer) *StdoutNotifier {
	if consoleOut == nil {
		consoleOut = io.Discard
	}
	return &StdoutNotifier{
		out: consoleOut,
	}
}

// PrintSummary prints the highest, fastest and the most and the least common types.
func (notify *StdoutNotifier) PrintSummary(dash *Dashboard) {
	_, _ = fmt.Fprintln(notify.out, "=== Summary ===")
	notify.listByRarity("aircraft", dash.SeenType)
	notify.listByRarity("operator", dash.SeenOperator)
	notify.listByRarity("country", dash.SeenCountry)
	_, _ = fmt.Fprintln(notify.out, "Fastest Aircraft:")
	fastest := dash.GetFastest()
	_, _ = fmt.Fprintln(notify.out, fastest.SightingToString())
	_, _ = fmt.Fprintln(notify.out, "Highest Aircraft:")
	highest := dash.GetHighest()
	_, _ = fmt.Fprintln(notify.out, highest.SightingToString())
	_, _ = fmt.Fprintln(notify.out, "=== End Summary ===")
}

func (notify *StdoutNotifier) listByRarity(propertyName string, propertyCountMap map[string]int) {
	propertyCounts := reference.GetSortedCountsForProperty(propertyCountMap)

	_, _ = fmt.Fprintf(notify.out, "Rarity from least to most common %s\n", propertyName)
	for j := range propertyCounts {
		_, _ = fmt.Fprintf(notify.out, "%6d - %s\n", propertyCounts[j].Count, propertyCounts[j].Property)
	}
}
