# DDD Migration Guide for airspottr

This migration guide describes a step-by-step process to migrate `airspottr` from its current structure into a Domain-Driven Design architecture.

## Migration objective

The goal is to maintain the existing behavior of `airspottr` while shifting the codebase into clearly separated layers:

- domain
- application
- infrastructure
- interface / adapters

This makes the project easier to extend, test, and evolve.

## Recommended package layout

Consider reorganizing the repository into the following logical packages:

- `domain/spotting`
- `domain/reference`
- `application`
- `infrastructure/adsb`
- `infrastructure/data`
- `infrastructure/notify`
- `infrastructure/persistence`
- `ui/tui`
- `ui/ticker`
- `cmd/airspottr` (optional) or keep `main.go` at repository root

### Domain layer

The domain layer should define the core business model and invariants.

Suggested packages:

- `domain/spotting`
  - `aircraft.go` - entity and value object definitions for aircraft, flight numbers, and location.
  - `sighting.go` - aggregates and domain behavior for tracking sightings and rarity.
  - `rarity.go` - rarity flag definitions and detection rules.
  - `dashboard.go` - the `Dashboard` aggregate root and its behavior.
- `domain/reference`
  - `airline.go` - operator and airline value objects.
  - `country.go` - country lookup abstractions.
  - `icao.go` - ICAO type and prefix metadata abstractions.

### Application layer

The application layer should define use case orchestrators and service interfaces.

Suggested packages:

- `application/service` or `application/usecase`
  - `spotting.go` - high-level use cases such as `ProcessAircraftSightings`, `AssignFlightRoutes`, `PrintSummary`, `SaveState`, and `LoadState`.
  - `interfaces.go` - repository and service interfaces used by application services.

### Infrastructure layer

Implementation details for external systems should live here.

Suggested packages:

- `infrastructure/adsb`
  - `request.go` - ADS-B and flight route HTTP client implementations.
  - `request_test.go` - tests for URL generation and response handling.
- `infrastructure/data`
  - `csv_loader.go` - CSV data loaders for ICAO list, airline list, registration prefixes, hex ranges, and military code lookup.
- `infrastructure/notify`
  - `notify.go` - `beeep` desktop notification adapter.
- `infrastructure/persistence`
  - `state_repo.go` - JSON state persistence and restore.

### UI/adapter layer

Keep the existing `tuiapp/` and `tickerapp/` packages here.

- `ui/ticker` - orchestrates the ticker use case using application services.
- `ui/tui` - renders the TUI and drives event handling.

## Step-by-step migration plan

### 1. Define domain interfaces and entities

1. Create `domain/spotting` and `domain/reference`.
2. Move domain types from `internal/aircraft.go`, `internal/sighting.go`, `internal/flightroute.go`, and `internal/dashboard.go` into the new domain packages.
3. Keep behavior such as `GetFlightNoAsStr`, `GetFlightNoAsIcaoCode`, `GetAltitudeAsStr`, and `Distance` inside domain/value object code.
4. Extract `Dashboard` methods that are pure domain behavior:
   - `ProcessAircraftRecords`
   - `updateType`
   - `updateOperator`
   - `updateCountry`
   - `AssignRouteToCallsigns`
   - `AssignFlightRoutes`
   - `recomputeFastestAndHighest`

### 2. Introduce repository and service abstractions

Create interfaces in `application/interfaces.go` such as:

- `AircraftSource` or `AircraftRepository` for fetching aircraft records.
- `FlightRouteRepository` for obtaining flight route metadata.
- `AirportReferenceRepository` or `ReferenceDataProvider` for looking up ICAO/operator/country data.
- `StateRepository` for saving and loading persisted runtime state.
- `NotificationSender` for emitting summary and rare-sighting notifications.

These interfaces enable the domain/application layers to depend on abstractions rather than concrete HTTP, file, and notification implementations.

### 3. Move infrastructure details into adapters

Move the concrete implementations to infrastructure packages:

- `infrastructure/adsb.Request` becomes `infrastructure/adsb.AdsbClient`.
- `internal/dash` CSV loaders become `infrastructure/data/csv_loader.go`.
- `internal/persistence` becomes `infrastructure/persistence/state_repo.go`.
- `internal/notification` becomes `infrastructure/notify/beeep_notifier.go`.

Use constructor functions like `NewAdsbClient`, `NewCsvReferenceProvider`, `NewJsonStateRepository`, and `NewBeeepNotifier`.

### 4. Refactor the application orchestrators

Create use case handlers in `application/spotting.go`:

- `SpottingService` or `SpottingOrchestrator` that coordinates:
  - fetching aircraft data
  - processing the dashboard
  - requesting route data for missing callsigns
  - saving state
  - emitting notifications

The service should depend only on interfaces, not on concrete implementations.

Example abstract flow:

1. `aircraftRecords := aircraftSource.FetchNearbyAircraft(lat, lon)`
2. `dashboard.ProcessAircraft(aircraftRecords)`
3. `callsigns := dashboard.CallsignsNeedingRoutes()`
4. `routes := flightRouteRepository.FetchRoutes(callsigns)`
5. `dashboard.AssignRoutes(routes)`
6. `notifications.EmitRareSightings(dashboard.RareSightings())`
7. `stateRepo.Save(dashboard, pendingCallsigns)`

### 5. Update frontends to use application services

Refactor `tickerapp/` and `tuiapp/` to depend on application services:

- `tickerapp` should create the service and call its methods on each tick. It should remain responsible only for runtime scheduling and shutdown handling.
- `tuiapp` should create the same service and use it to update model state.

This isolates UI code from domain process logic.

### 6. Extract reference data lookups into their own module

The current `internal/dash` package loads datasets directly from path constants such as `./data/ICAOList.csv`.

In DDD, these lookups should become infrastructure adapters that satisfy a domain `ReferenceDataProvider` interface.

- Create a `ReferenceDataProvider` interface in `application/interfaces.go` or `domain/reference`.
- Implement CSV-backed provider in `infrastructure/data/csv_reference_provider.go`.
- Optionally add a constructor that accepts a path or filesystem abstraction to avoid root-relative dependencies.

### 7. Isolate state persistence

Move persistence out of the domain aggregate and into the infrastructure layer:

- `application/StateService` should orchestrate save/load using `StateRepository`.
- `infrastructure/persistence` should implement state serialization to JSON.
- `domain/spotting.Dashboard` should expose a serializable state snapshot, but not perform file I/O directly.

### 8. Keep tests relevant and add boundary tests

As you migrate packages:

- Keep existing unit tests passing.
- Add tests for interfaces and application services.
- Create tests targeting the domain package without file or network dependencies.
- Add mocks/fakes for `AircraftSource`, `FlightRouteRepository`, `StateRepository`, and `NotificationSender`.

### 9. Gradually remove the `internal/` package

A steady migration should minimize churn:

1. Create the new DDD package layout while leaving existing code intact.
2. Implement new interfaces and adapters alongside the old code.
3. Incrementally switch callers from `internal/` to the new packages.
4. Remove the old internal code only after the new architecture is fully covered.

## Suggested refactor increments

### Increment 1: Domain extraction

- Move types from `internal/aircraft.go`, `internal/sighting.go`, `internal/flightroute.go`, and `internal/dashboard.go` to `domain/spotting`.
- Move geographic helpers from `internal/dash/geo.go` to `domain/common` or `domain/spotting`.
- Keep the moved code behaviorally equivalent.

### Increment 2: Infrastructure interface definitions

- Define `AircraftSource`, `FlightRouteSource`, `ReferenceDataProvider`, `StateRepository`, and `NotificationSender`.
- Implement these interfaces in supporting packages.

### Increment 3: Application use case orchestration

- Introduce an application service that orchestrates a single refresh cycle.
- Update `tickerapp` and `tuiapp` to invoke the service rather than `Dashboard` and `Request` directly.

### Increment 4: Persistence and startup

- Move `LoadState`/`SaveState` into infrastructure and the application service.
- Wire up state persistence at app startup and graceful shutdown.

### Increment 5: UI isolation and cleanup

- Keep `tuiapp` and `tickerapp` as thin adapters.
- Remove any remaining domain logic from UI packages.
- Update `main.go` or add `cmd/airspottr` to bootstrap dependency injection.

## Practical considerations

### Preserve static data loading

- Keep `data/` files intact.
- Create adapters that accept dataset directory paths, so the application can be launched from any working directory.
- If helpful, add a `--data-dir` CLI option for local development.

### Preserve persistent state path

- Keep using `os.UserConfigDir()` for the default state file path.
- Introduce an interface to allow unit tests to use an in-memory or temp file repository.

### Avoid premature generalization

- The current codebase is small; avoid inventing overly broad domain abstractions.
- Prefer explicit, concrete domain concepts with a small, stable interface.

### Keep existing behavior intact

- Preserve all reported features: fastest/highest aircraft, rarity detection, route assignment, warmup behavior, notifications, and TUI/ticker frontends.
- Preserve the location selection model using `--latlon` and `--location`.

## Recommended DDD alignment improvements

To better match the `go-ddd` style without changing the core migration plan, consider adding these refinements:

- Explicitly codify the onion dependency rule:
  - `domain` must not depend on `application`, `infrastructure`, or `interface`.
  - `application` may depend on `domain` only.
  - `infrastructure` may depend on `domain` and `application`, but not `interface`.
  - `interface` may depend on `application` and `domain` types only.
- Use constructors and validation for domain objects:
  - `NewAircraft`, `NewSighting`, `NewDashboard`, and value object constructors should enforce invariants.
  - Avoid public struct literals for domain entities outside their package.
- Introduce validated types at trust boundaries:
  - Accept validated domain or DTO types in repository and service methods rather than raw primitives where needed.
- Define sentinel domain errors:
  - Use `errors.Is` with wrapped sentinel errors such as `ErrInvalidFlight`, `ErrEntityNotFound`, or `ErrValidation`.
- Consider lightweight CQRS separation:
  - Keep write-oriented use cases in application services and read-oriented projections or queries simple and isolated.
- Treat domain events as future evolution:
  - You can add `RareSightingEmitted` or `FlightRouteAssigned` events later once the core aggregate boundaries are stable.

## Example package mappings

| Current file | New package | Notes |
| --- | --- | --- |
| `internal/aircraft.go` | `domain/spotting/aircraft.go` | entity model and helpers |
| `internal/sighting.go` | `domain/spotting/sighting.go` | aggregate and event model |
| `internal/dashboard.go` | `domain/spotting/dashboard.go` | aggregate root and domain invariants |
| `internal/flightroute.go` | `domain/spotting/flightroute.go` | route value objects |
| `internal/dash/icao.go` | `infrastructure/data/csv_reference_provider.go` | reference data adapter |
| `internal/dash/geo.go` | `domain/common/geo.go` | physical distance value objects |
| `internal/request.go` | `infrastructure/adsb/client.go` | HTTP adapters |
| `internal/persistence.go` | `infrastructure/persistence/json_state_repo.go` | state repository |
| `internal/notification.go` | `infrastructure/notify/desktop_notifier.go` | notification adapter |

## Migration checklist

- [ ] Create domain and application package structure.
- [ ] Extract existing domain model into `domain/spotting` and `domain/reference`.
- [ ] Define repository/service interfaces in application packages.
- [ ] Implement concrete adapters in `infrastructure`.
- [ ] Refactor frontends to use application services.
- [ ] Add or update tests for clean boundaries.
- [ ] Update documentation to describe the new architecture.
- [ ] Remove legacy `internal/` package when the migration is complete.

## Next steps

After completing the basic migration, the project will be better positioned to support future enhancements:

- new ADS-B or flight route providers
- additional UI frontends (web, desktop, headless)
- configurable persistence backends
- richer domain analytics and event-driven alerts

This guide should be used as the source of truth for the migration and as a checklist for ongoing refactors.
