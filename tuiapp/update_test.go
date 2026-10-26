//nolint:exhaustruct_v5 // Don't enforce for bubbletea.
package tuiapp

import (
	"io"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	srv "github.com/micutio/airspottr/internal/application/services"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

func TestHandleKeyQuit(t *testing.T) {
	t.Parallel()
	var m model
	cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected non-nil quit command")
	}
}

func TestHandleKeyCtrlC(t *testing.T) {
	t.Parallel()
	var m model
	cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected non-nil quit command")
	}
}

func TestNavigateNotifyStripWrapsAround(t *testing.T) {
	t.Parallel()
	testModel := model{
		inputFocus:     focusNotifyStrip,
		notifyStripIdx: 0,
	}
	// Moving left from index 0 should wrap to index 2, not become -1.
	testModel.navigateNotifyStrip("left")
	if testModel.notifyStripIdx != notifyCountry {
		t.Fatalf("notifyStripIdx after left wrap = %d, want %d", testModel.notifyStripIdx, notifyCountry)
	}

	// Spacebar toggles the selected notify flag.
	testModel.navigateNotifyStrip(" ")
	if !testModel.notifyOnCountry {
		t.Fatal("expected notifyOnCountry to be toggled true")
	}

	// Moving right from index 2 should wrap back to index 0.
	testModel.navigateNotifyStrip("right")
	if testModel.notifyStripIdx != notifyType {
		t.Fatalf("notifyStripIdx after right wrap = %d, want %d", testModel.notifyStripIdx, notifyType)
	}
}

type mockRouteSource struct {
	routes map[string]ref.FlightrouteRecord
}

func (f mockRouteSource) GetFlightroutes(callsigns []string) map[string]ref.FlightrouteRecord {
	out := make(map[string]ref.FlightrouteRecord, len(callsigns))
	for _, cs := range callsigns {
		if r, ok := f.routes[cs]; ok {
			out[cs] = r
		}
	}
	return out
}

func (mockRouteSource) GetPendingCallsigns() []string     { return nil }
func (mockRouteSource) RestorePendingCallsigns([]string) {}

func TestUpdateFetchesFlightRoutesAsynchronously(t *testing.T) {
	t.Parallel()

	route := *ref.GetDefaultFlightrouteRecord()
	route.Callsign = "BAW123"
	route.Origin.IataCode = "LHR"
	route.Destination.IataCode = "JFK"

	theme := getDefaultTheme()
	style := initTableStyle(theme)
	testModel := &model{
		tableStyle:      style,
		theme:           theme,
		tables:          newTuiTables(style),
		dashboard:       srv.NewDashboard(53.55, 9.99, nil, io.Discard),
		flightrouteRepo: mockRouteSource{routes: map[string]ref.FlightrouteRecord{"BAW123": route}},
	}

	records := AircraftResponseMsg{
		obs.AircraftRecord{Hex: "abc123", Flight: "BAW123", Lat: 53.6, Lon: 10.0},
	}

	_, cmd := testModel.Update(records)
	if cmd == nil {
		t.Fatal("expected async flight route command after AircraftResponseMsg")
	}

	msg := cmd()
	routesMsg, ok := msg.(FlightRoutesResponseMsg)
	if !ok {
		t.Fatalf("expected FlightRoutesResponseMsg, got %T", msg)
	}

	testModel.Update(routesMsg)
	sightings := testModel.dashboard.GetCurrentSightings()
	if len(sightings) != 1 || sightings[0].Flightroute.Origin.IataCode != "LHR" {
		t.Fatalf("expected route LHR->JFK assigned after FlightRoutesResponseMsg, got %+v", sightings)
	}
}
