package tuiapp

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	srv "github.com/micutio/airspottr/internal/application/services"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
)

type UpdateTickMsg time.Time

func updateTick() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return UpdateTickMsg(t)
	})
}

type AircraftQueryTickMsg time.Time

func aircraftQueryTick() tea.Cmd {
	return tea.Every(adsb.AircraftUpdateInterval, func(t time.Time) tea.Msg {
		return AircraftQueryTickMsg(t)
	})
}

type AircraftResponseMsg []obs.AircraftRecord

func requestAircraftDataCmd(frr srv.AircraftDataSource) tea.Cmd {
	return func() tea.Msg {
		return AircraftResponseMsg(frr.RequestAircraft())
	}
}

type FlightRoutesResponseMsg map[string]ref.FlightrouteRecord

func requestFlightRoutesCmd(frr srv.FlightRouteDataSource, callsigns []string) tea.Cmd {
	if frr == nil || len(callsigns) == 0 {
		return nil
	}
	return func() tea.Msg {
		routes := frr.GetFlightroutes(callsigns)
		if len(routes) == 0 {
			return nil
		}
		return FlightRoutesResponseMsg(routes)
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:ireturn // tea.Model interface
	switch thisMsg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = thisMsg.Height
		m.width = thisMsg.Width
		m.resizeTables()
	case tea.KeyMsg:
		return m, m.handleKey(thisMsg)
	case UpdateTickMsg:
		return m, updateTick()
	case AircraftQueryTickMsg:
		return m, tea.Batch(requestAircraftDataCmd(m.aircraftRepo), aircraftQueryTick())
	case AircraftResponseMsg:
		return m, m.processAircraftResponse(thisMsg)
	case FlightRoutesResponseMsg:
		m.processFlightRouteResponse(thisMsg)
	}
	return m, nil
}

func (m *model) processAircraftResponse(msg AircraftResponseMsg) tea.Cmd {
	m.lastUpdate = time.Now()
	aircraftRecords := []obs.AircraftRecord(msg)
	currentSightings := m.dashboard.RefreshBatch(aircraftRecords, obs.RarityNotifyToggles{
		Type:     m.notifyOnType,
		Operator: m.notifyOnOp,
		Country:  m.notifyOnCountry,
	})
	m.updateAllTables()

	missingCallsigns := srv.GetCallsignsRequiringRoutes(currentSightings)
	return requestFlightRoutesCmd(m.flightrouteRepo, missingCallsigns)
}

func (m *model) processFlightRouteResponse(msg FlightRoutesResponseMsg) {
	m.dashboard.AssignFlightRoutes(msg)
	m.updateAllTables()
}
