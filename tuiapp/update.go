package tuiapp

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/micutio/airspottr/internal"
)

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
		return m, tea.Batch(requestAircraftDataCmd(m.request), aircraftQueryTick())
	case AircraftResponseMsg:
		return m, m.processAircraftResponse(thisMsg)
	case FlightRoutesResponseMsg:
		return m, m.processFlightRouteResponse(thisMsg)
	}
	return m, nil
}

func (m *model) processAircraftResponse(msg AircraftResponseMsg) tea.Cmd {
	m.lastUpdate = time.Now()
	aircraftRecords := []internal.AircraftRecord(msg)
	m.dashboard.ProcessAircraftRecords(aircraftRecords)
	m.notify.EmitRarityNotifications(m.dashboard.RareSightings, internal.RarityNotifyToggles{
		Type:     m.notifyOnType,
		Operator: m.notifyOnOp,
		Country:  m.notifyOnCountry,
	})

	callsignsWithoutRoute := m.dashboard.AssignRouteToCallsigns()
	if callsignsWithoutRoute != nil {
		return requestFlightRouteDataCmd(m.request, callsignsWithoutRoute)
	}

	m.updateAllTables()
	return nil
}

func (m *model) processFlightRouteResponse(msg FlightRoutesResponseMsg) tea.Cmd {
	flightRoutes := []internal.FlightRouteRecord(msg)
	m.dashboard.AssignFlightRoutes(flightRoutes)

	// Check if there are more callsigns without routes and request them
	callsignsWithoutRoute := m.dashboard.AssignRouteToCallsigns()
	if callsignsWithoutRoute != nil {
		return requestFlightRouteDataCmd(m.request, callsignsWithoutRoute)
	}

	m.updateAllTables()
	return nil
}
