package tuiapp

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	obs "github.com/micutio/airspottr/internal/domain/observation"
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
		return m, tea.Batch(requestAircraftDataCmd(m.aircraftRepo), aircraftQueryTick())
	case AircraftResponseMsg:
		m.processAircraftResponse(thisMsg)
	case FlightRoutesResponseMsg:
		m.processFlightRouteResponse(thisMsg)
	}
	return m, nil
}

func (m *model) processAircraftResponse(msg AircraftResponseMsg) {
	m.lastUpdate = time.Now()
	aircraftRecords := []obs.AircraftRecord(msg)
	m.refreshUseCase.RefreshBatch(aircraftRecords, obs.RarityNotifyToggles{
		Type:     m.notifyOnType,
		Operator: m.notifyOnOp,
		Country:  m.notifyOnCountry,
	})
	m.updateAllTables()
}

func (m *model) processFlightRouteResponse(msg FlightRoutesResponseMsg) {
	m.refreshUseCase.AssignFlightRoutes(msg)
	m.updateAllTables()
}
