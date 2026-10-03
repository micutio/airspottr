package services

import (
	appint "github.com/micutio/airspottr/internal/application/interfaces"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// RefreshUseCase coordinates the standard aircraft refresh loop.
// It owns the orchestration boundary without forcing callers to know the
// dashboard or route-fetch details.
type RefreshUseCase struct {
	dashboard         *Dashboard
	notifier          appint.NotificationService
	flightRouteLookup func([]string) map[string]ref.FlightrouteRecord
}

func NewRefreshUseCase(
	dashboard *Dashboard,
	notifier appint.NotificationService,
	flightRouteLookup func([]string) map[string]ref.FlightrouteRecord,
) *RefreshUseCase {
	return &RefreshUseCase{
		dashboard:         dashboard,
		notifier:          notifier,
		flightRouteLookup: flightRouteLookup,
	}
}

func (uc *RefreshUseCase) RefreshBatch(
	records []obs.AircraftRecord,
	toggles obs.RarityNotifyToggles,
) []obs.AircraftSighting {
	if uc == nil || uc.dashboard == nil {
		return nil
	}

	uc.dashboard.ProcessAircraftRecords(records)
	currentSightings := uc.dashboard.GetCurrentSightings()
	if uc.notifier != nil {
		uc.notifier.EmitRarityNotifications(currentSightings, toggles)
	}

	if uc.flightRouteLookup != nil {
		callsignsWithoutRoute := GetCallsignsRequiringRoutes(currentSightings)
		if len(callsignsWithoutRoute) > 0 {
			routes := uc.flightRouteLookup(callsignsWithoutRoute)
			if len(routes) > 0 {
				uc.dashboard.AssignFlightRoutes(routes)
			}
		}
	}

	return currentSightings
}

func (uc *RefreshUseCase) AssignFlightRoutes(routes map[string]ref.FlightrouteRecord) {
	if uc == nil || uc.dashboard == nil || len(routes) == 0 {
		return
	}

	uc.dashboard.AssignFlightRoutes(routes)
}
