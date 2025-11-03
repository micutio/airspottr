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

type ADSBResponseMsg []byte

func requestADSBDataCmd(opts internal.RequestOptions) tea.Cmd {
	return func() tea.Msg {
		body, err := internal.RequestAndProcessCivAircraft(opts)
		if err != nil {
			// TODO: Log error
			return nil
		}
		return ADSBResponseMsg(body)
	}
}
