package tuiapp

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/micutio/airspottr/internal"
)

type UpdateTickMsg time.Time

func updateTick() tea.Cmd {
	return tea.Every(
		time.Second,
		func(t time.Time) tea.Msg {
			return UpdateTickMsg(t)
		},
	)
}

type AircraftQueryTickMsg time.Time

func aircraftQueryTick() tea.Cmd {
	return tea.Every(
		internal.AircraftUpdateInterval,
		func(t time.Time) tea.Msg {
			return AircraftQueryTickMsg(t)
		},
	)
}

type AircraftResponseMsg []internal.AircraftRecord

func requestAircraftDataCmd(request *internal.Request) tea.Cmd {
	return func() tea.Msg {
		aircraftData := request.RequestAircraft()
		return AircraftResponseMsg(aircraftData)
	}
}

type FlightRoutesResponseMsg []internal.FlightRouteRecord

func requestFlightRouteDataCmd(request *internal.Request, callsigns []string) tea.Cmd {
	return func() tea.Msg {
		flightRoutes := request.RequestFlightRoutesForCallsigns(callsigns)
		return FlightRoutesResponseMsg(flightRoutes)
	}
}
