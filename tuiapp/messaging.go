package tuiapp

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
	repo "github.com/micutio/airspottr/internal/domain/repositories"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
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

func requestAircraftDataCmd(frr repo.AircraftRepository) tea.Cmd {
	return func() tea.Msg {
		aircraftData := frr.RequestAircraft()
		return AircraftResponseMsg(aircraftData)
	}
}

type FlightRoutesResponseMsg []ref.FlightRouteRecord

func requestFlightRouteDataCmd(frr repo.FlightrouteRepository, callsigns []string) tea.Cmd {
	return func() tea.Msg {
		flightRoutes := frr.RequestFlightroutesForCallsigns(callsigns)
		return FlightRoutesResponseMsg(flightRoutes)
	}
}
