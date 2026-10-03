package services

import rep "github.com/micutio/airspottr/internal/domain/repositories"

// AirspottrState is kept as a compatibility alias while the persisted-state model lives in the
// repositories layer.
type AirspottrState = rep.AirspottrState

// PersistedRareSighting is kept as a compatibility alias.
type PersistedRareSighting = rep.PersistedRareSighting

// FlightrouteRepoState is kept as a compatibility alias.
type FlightrouteRepoState = rep.FlightrouteRepoState

// PersistentState is kept as a compatibility alias.
type PersistentState = rep.PersistentState

// DashboardState is kept as a compatibility alias.
type DashboardState = rep.DashboardState
