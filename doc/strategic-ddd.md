# Strategic DDD Migration Plan for airspottr

This guide is a pragmatic, incremental refactoring plan for the airspottr codebase. It is designed for a human developer to execute manually in short, bite-sized sessions of 30–45 minutes each.

## Step 1: Codebase audit

### Coupling hotspots

These are the places where infrastructure or I/O is currently mixed with business rules and core state:

- [internal/dashboard.go](../internal/application/dashboard.go)
  - `Dashboard.ProcessAircraftRecords` is the main hotspot. It mixes:
    - state transitions for sightings,
    - type/operator/country classification,
    - rarity calculation,
    - distance calculation,
    - route assignment,
    - and mutation of shared Dashboard state.
  - `Dashboard.updateType`, `Dashboard.updateOperator`, `Dashboard.updateCountry`, `AssignRouteToCallsigns`, `AssignFlightRoutes`, and `recomputeFastestAndHighest` are all part of the same crowded responsibility cluster.

- [internal/request.go](../internal/request.go)
  - `Request.RequestAircraft`, `Request.RequestFlightRoutesForCallsigns`, `createAircraftReqURL`, `createFlightRouteRequestURL`, and `sendRequest` combine:
    - URL construction,
    - HTTP client usage,
    - JSON parsing,
    - concurrency,
    - pending-callsign queue management,
    - and shaping of data into domain-facing records.
  - This is infrastructure code, but it is currently doing too much orchestration around the domain flow.

- [internal/persistence.go](../internal/persistence.go)
  - `SaveState` and `LoadState` mix:
    - domain objects such as `AircraftSighting` and `RareSighting`,
    - filesystem operations,
    - JSON serialization,
    - and app-state restoration.
  - The persistence layer is also responsible for translating between runtime objects and persisted shapes.

- [internal/notification.go](../internal/infrastructure/notify/notification.go)
  - `Notify.EmitRarityNotifications` and the `notifyRare*` helpers mix:
    - domain event interpretation,
    - formatting,
    - and desktop notification delivery via `beeep`.
  - The business decision of what is rare is still embedded in the delivery layer.

- [tickerapp/tickerapp.go](../tickerapp/tickerapp.go) and [tuiapp/model.go](../tuiapp/model.go)
  - These packages coordinate runtime behavior, but they also sit on top of the same stateful dashboard and request structures. They are application entrypoints, but they still carry too much orchestration and state wiring directly.

### Proposed bounded contexts

A small number of boundaries will be enough for this codebase. I would start with these four:

1. Observation / Sighting Context
   - Best fit for:
     - [internal/sighting.go](../internal/sighting.go)
     - [internal/rarity.go](../internal/rarity.go)
     - the state transition logic in [internal/dashboard.go](../internal/application/dashboard.go)
   - Purpose:
     - capture what was observed,
     - track sighting history,
     - decide whether a sighting is rare,
     - and keep the “current known state” of an aircraft.

2. Aircraft Classification / Reference Data Context
   - Best fit for:
     - [internal/dash/icao.go](../internal/dash/icao.go)
     - [internal/dash/geo.go](../internal/dash/geo.go)
   - Purpose:
     - resolve type, operator, country, and distance from reference datasets and coordinate math.
   - This is not core business behavior in itself; it is support logic that should be hidden behind a small interface.

3. Flight Data Intake Context
   - Best fit for:
     - [internal/request.go](../internal/request.go)
     - [internal/flightroute.go](../internal/flightroute.go)
   - Purpose:
     - fetch aircraft and flight-route data from external systems.
   - Keep this as an adapter layer that produces plain domain records for the rest of the app.

4. Presentation / Delivery Context
   - Best fit for:
     - [internal/notification.go](../internal/infrastructure/notify/notification.go)
     - [tickerapp/tickerapp.go](../tickerapp/tickerapp.go)
     - [tuiapp/model.go](../tuiapp/model.go)
   - Purpose:
     - render observations to the terminal, TUI, or desktop notifications.
   - This layer should not own the domain rules.

### Leaky dependencies

These types and modules should become pure or much less coupled:

- `AircraftSighting` and `RareSighting` should become plain domain data structures rather than being tightly coupled to persistence and runtime concerns.
- `AircraftRecord` is currently a transport-shaped struct with JSON tags and cached fields; it should become either a clear domain DTO or a thin input model at the boundary.
- `FlightRouteRecord` and its nested value objects should be treated as input/output data rather than part of the core domain behavior.
- `Dashboard` currently depends directly on CSV-backed reference maps and on concrete HTTP-oriented flow; it should depend on small abstractions instead.
- `Request` is coupled to `net/http`, JSON, and logging; it should become an adapter that yields domain data rather than being a domain object itself.
- `Notify` depends on the `beeep` package and should become a presentation adapter that receives already-decided events.
- `SaveState` and `LoadState` should be infrastructure concerns, not part of the core observation rules.

## Step 2: Bite-sized refactoring recipes

### 1. Extract the pure sighting model and rarity rules

- Target context: Observation / Sighting
- Estimated time: 35–45 minutes
- Goal: Separate the core sighting lifecycle from the current Dashboard mutation logic.

- Human recipe:
  - Create a small domain-focused module for observation concepts.
  - Move `AircraftSighting`, `RareSighting`, `RarityFlag`, and the related constants from [internal/sighting.go](../internal/sighting.go) and [internal/rarity.go](../internal/rarity.go) into that module.
  - Extract the core decision logic into a function shaped like `EvaluateBatch(state, records, classifier) -> (newState, rareSightings)`.
  - Leave [internal/dashboard.go](../internal/application/dashboard.go) as a thin wrapper that adapts the result back into the existing dashboard fields.

- Before:

  ```go
  func (db *Dashboard) ProcessAircraftRecords(records []AircraftRecord) {
      // mutates db directly
      // calls updateType/updateOperator/updateCountry
  }
  ```

- After:

  ```go
  func EvaluateBatch(state ObservationState, records []AircraftRecord, classifier Classifier) ([]RareSighting, ObservationState) {
      // pure-ish: transforms state + inputs only
  }
  ```

- Verification:
  - Add one unit test for: one new aircraft, one repeated flight, and one rare case.
  - Run: `go test ./internal`

### 2. Extract classification behind a small interface

- Target context: Aircraft Classification / Reference Data
- Estimated time: 30–40 minutes
- Goal: Remove direct dependence on CSV maps from the core observation logic.

- Human recipe:
  - Introduce a tiny `Classifier` interface with methods such as:
    - `TypeFor(icaoType string) string`
    - `OperatorFor(flightNo string) string`
    - `CountryFor(flightNo, reg, hex string) string`
  - Implement it in a small adapter around [internal/dash/icao.go](../internal/dash/icao.go).
  - Change the dashboard’s `updateType`, `updateOperator`, and `updateCountry` methods to use the interface instead of reaching into `IcaoToAircraft`, `IcaoToAirline`, `regPrefixToCountry`, and `hexRangeToCountry` directly.

- Before:

  ```go
  aType := db.IcaoToAircraft[aircraft.IcaoType].Make
  ```

- After:

  ```go
  aType := classifier.TypeFor(aircraft.IcaoType)
  ```

- Verification:
  - Add a stub classifier test that proves the dashboard still behaves correctly when the classifier returns known values.
  - Run: `go test ./internal`

### 3. Isolate the external flight-data adapter

- Target context: Flight Data Intake
- Estimated time: 35–45 minutes
- Goal: Keep HTTP, JSON, and request orchestration out of the core flow.

- Human recipe:
  - Extract the logic in [internal/request.go](../internal/request.go) behind a narrow interface such as:
    - `AircraftProvider`
    - `FlightRouteProvider`
  - Keep `Request` as the concrete infrastructure implementation that knows about `net/http`, JSON, TLS, and hosts.
  - Make the observation service depend on the interface instead of the concrete request type.

- Before:

  ```go
  func (r *Request) RequestAircraft() []AircraftRecord
  ```

- After:

  ```go
  type AircraftProvider interface {
      Aircrafts() []AircraftRecord
      FlightRoutes(callsigns []string) []FlightRouteRecord
  }
  ```

- Verification:
  - Add a fake provider test that returns one aircraft and one route and confirm the observation state updates correctly.
  - Run: `go test ./internal`

### 4. Move persistence to a thin infrastructure layer

- Target context: State Persistence
- Estimated time: 30–40 minutes
- Goal: Separate “what state exists” from “how it is stored on disk”.

- Human recipe:
  - Keep [internal/persistence.go](../internal/persistence.go) as an adapter, but make the boundary explicit.
  - Define a narrow state contract such as:
    - `SaveState(path string, state AppState) error`
    - `LoadState(path string) (AppState, error)`
  - Let the JSON and filesystem logic stay there, but stop letting persistence be the place where the domain logic lives.

- Before:

  ```go
  func SaveState(filePath string, db *Dashboard, req *Request) error
  ```

- After:

  ```go
  func SaveState(filePath string, state AppState) error
  ```

- Verification:
  - Re-run the existing persistence test in [internal/persistence_test.go](../internal/persistence_test.go).
  - Run: `go test ./internal`

### 5. Make notifications a presentation adapter

- Target context: Presentation / Delivery
- Estimated time: 30–40 minutes
- Goal: Keep notification delivery outside the core domain.

- Human recipe:
  - Treat [internal/notification.go](../internal/infrastructure/notify/notification.go) as a delivery-specific layer.
  - Keep the domain side as simple events such as `RareSightingDetected`.
  - Let the notification layer format and emit those events, rather than deciding what counts as rare itself.

- Before:

  ```go
  func (notify *Notify) EmitRarityNotifications(sightings []RareSighting, toggles RarityNotifyToggles)
  ```

- After:

  ```go
  func (notify *Notify) Deliver(events []RareSightingEvent)
  ```

- Verification:
  - Add one test for formatting a rare-event payload without touching the desktop notification backend.
  - Run: `go test ./internal`

### 6. Make the app entrypoints composition roots

- Target context: Application Composition
- Estimated time: 35–45 minutes
- Goal: Keep startup wiring simple and explicit.

- Human recipe:
  - In [tickerapp/tickerapp.go](../tickerapp/tickerapp.go) and [tuiapp/model.go](../tuiapp/model.go), keep the role of glue only:
    - create the classifier,
    - create the data provider,
    - create the notifier,
    - create the persistence adapter,
    - and pass them into a small application service.
  - Do not let the app shell own the domain rules.

- Before:

  ```go
  dashboard.ProcessAircraftRecords(...)
  notify.EmitRarityNotifications(...)
  ```

- After:

  ```go
  service.HandleBatch(...)
  notifier.Deliver(...)
  ```

- Verification:
  - Run the full suite: `go test ./...`

### 7. Keep the dependency rule simple and explicit

- Target context: Architecture hygiene
- Estimated time: 15–20 minutes
- Goal: Prevent the refactor from drifting into an over-engineered design.

- Human recipe:
  - Adopt one clear rule:
    - domain code may depend on basic language primitives and domain types only,
    - infrastructure code may depend on OS, network, JSON, CSV, and external libraries,
    - application code wires the two together.
  - Avoid introducing a generic repository layer, event bus, or enterprise wrapper until the domain boundaries are stable.

- Verification:
  - Re-check the package boundaries mentally and run: `go test ./...`

## Recommended order

For the lowest-risk path, do the work in this order:

1. Sighting model and rarity rules
2. Classification extraction
3. Flight-data adapter extraction
4. Persistence adapter extraction
5. Notification adapter extraction
6. App wiring cleanup

That sequence gives the best return on each 30–45 minute session while keeping the codebase understandable and testable.
