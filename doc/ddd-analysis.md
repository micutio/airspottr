# Domain-Driven Design Analysis for airspottr

This document analyzes the current architecture of `airspottr` and identifies how it can be migrated toward a Domain-Driven Design (DDD) structure.

## Current architecture summary

`airspottr` currently has the following core modules:

- `main.go` - CLI parsing and frontend selection.
- `internal/` - shared application logic, including domain concepts, infrastructure, and orchestration.
- `internal/dash/` - CSV-backed lookup and geographic utility helpers.
- `tuiapp/` - Bubble Tea-based terminal UI.
- `tickerapp/` - line-oriented ticker mode.
- `data/` - static data files for ICAO mappings, registration prefixes, and hex ranges.
- `assets/` - icon asset for desktop notifications.

### Mixed responsibilities

The current `internal/` package mixes several concerns:

- Domain entities and behavior: `AircraftRecord`, `Dashboard`, `AircraftSighting`, `RareSighting`, `FlightRouteRecord`.
- Infrastructure: HTTP request construction and execution (`internal/request.go`), CSV file loading (`internal/dash/`), desktop notification emission (`../internal/infrastructure/notify/notification.go`), JSON persistence (`internal/persistence.go`).
- Application logic: coordination of periodic updates, warmup behavior, route assignment, and notification emission.

This mixture makes it harder to reason about the domain model separately from the technical mechanisms that support it.

## What needs to become explicit in a DDD architecture

To move `airspottr` toward DDD, the codebase should separate the following layers:

- Domain layer: core business concepts and invariants.
- Application layer: use cases, orchestration, and transaction boundaries.
- Infrastructure layer: external adapters such as HTTP API clients, file persistence, dataset loaders, and notification services.
- Interface/Adapter layer: UI and CLI concerns.

### Domain concepts in airspottr

The application already has clear domain concepts that map well to DDD:

- `Aircraft` / `AircraftRecord`: the sighted aircraft with flight information and metadata.
- `Sighting` / `AircraftSighting`: the observation event and tracked history for each aircraft.
- `RareSighting`: a detected rarity event derived from a sighting.
- `Dashboard`: an aggregate that processes aircraft sightings and maintains rarity statistics.
- `FlightRoute`: enriched route metadata for a callsign.
- `Coordinates` and `Distance`: geographic value objects.
- `Operator` and `Country`: domain concepts used for rarity and enrichment.

### Existing domain behavior

Important domain behavior currently implemented in `../internal/application/dashboard.go` and across the `internal` package includes:

- tracking fastest/highest aircraft
- updating type/operator/country rarity counts
- deriving operator and country via flight code, registration prefix, and hex range
- assigning flight route metadata to existing sightings
- tracking whether a flight is new or updated

These behaviors should remain in the domain layer, but be isolated from infrastructure details.

### Technical/infrastructure responsibilities

The current code also includes technical responsibilities that should be migrated out of the domain layer:

- HTTP fetching and API URL construction (`internal/request.go`)
- CSV parsing and dataset loading (`internal/dash/`)
- persistence to the user config directory (`internal/persistence.go`)
- desktop notifications through `beeep` (`../internal/infrastructure/notify/notification.go`)
- UI rendering and keyboard handling (`tuiapp/` and `tickerapp/`)

## Domain boundaries and potential bounded contexts

The application is small enough that a single bounded context is plausible, but it still benefits from a clean separation into these modules:

- `domain/spotting` - core domain model for aircraft observations, rarity detection, and route enrichment.
- `domain/reference` - domain abstractions for airline, ICAO, registration, and country lookup.
- `application` - use cases like `SpotAircraft`, `RequestFlightRoutes`, `EmitRarityNotifications`, `SaveState`, and `LoadState`.
- `infrastructure/adsb` - ADS-B request client and route lookup client.
- `infrastructure/data` - CSV dataset loader implementations.
- `infrastructure/notify` - desktop notification adapter.
- `infrastructure/persistence` - JSON state repository.
- `ui/tui` and `ui/ticker` - frontends and CLI adapter code.

## Key DDD migration goals

1. Extract the domain model from `internal/` into explicit domain packages.
2. Define repository and service interfaces in the domain/application layers.
3. Move implementation details into infrastructure adapters.
4. Keep UI packages dependent on the application layer only.
5. Preserve existing behavior while introducing clearer boundaries.

## Design risks in the current code

- `../internal/application/dashboard.go` depends directly on CSV-backed maps and file-loading semantics, coupling the domain to a specific persistence format.
- `internal/request.go` validates hosts and performs network calls, but the domain only needs an abstraction that fetches aircraft and route data.
- `../internal/infrastructure/notify/notification.go` uses desktop notification side effects directly from domain logic, making domain testing difficult.
- Persistence and state restoration logic is entangled with domain state shaping.

## What a DDD-aligned architecture will enable

- easier unit testing of business logic without file, network, or UI dependencies
- clearer dependency injection for external adapters
- a more maintainable package structure as the app grows
- cleaner separation between `spotting` use cases and technical concerns
- a natural path to extract reusable domain libraries or support additional frontends
