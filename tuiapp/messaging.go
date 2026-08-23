package tuiapp

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
	adsb "github.com/micutio/airspottr/internal/infrastructure/adsb"
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
		adsb.AircraftUpdateInterval,
		func(t time.Time) tea.Msg {
			return AircraftQueryTickMsg(t)
		},
	)
}

type AircraftResponseMsg []obs.AircraftRecord

func requestAircraftDataCmd(request *adsb.Request) tea.Cmd {
	return func() tea.Msg {
		aircraftData := request.RequestAircraft()
		return AircraftResponseMsg(aircraftData)
	}
}

type FlightRoutesResponseMsg []ref.FlightRouteRecord

func requestFlightRouteDataCmd(request *adsb.Request, callsigns []string) tea.Cmd {
	return func() tea.Msg {
		flightRoutes := request.RequestFlightRoutesForCallsigns(callsigns)
		return FlightRoutesResponseMsg(flightRoutes)
	}
}
