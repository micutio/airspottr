# airspottr Architecture

This document describes the internal architecture and design of the `airspottr` application.

## High-level architecture

`airspottr` is built as a Go application with a shared core engine and two frontends:

- `tuiapp/` - a terminal user interface built with `bubbletea`.
- `tickerapp/` - a line-oriented ticker mode for scripted or non-interactive use.

Both frontends reuse the same core `internal/` package for request handling, aircraft processing, enrichment, rarity detection, notification, and persistence.

## Core components

### main.go

The main entrypoint wires together command-line options and selects the frontend.

- `--ticker` / `-t` switches to ticker mode.
- `--latlon` / `-l` supplies custom latitude and longitude.
- `--location` / `-L` selects from predefined coordinates.

If a predefined location is selected, `main.go` replaces any supplied `--latlon` coordinates with the matching latitude and longitude.

### internal.Request

Located in `internal/request.go`, the `Request` type is responsible for:

- Building and validating attractive API URLs.
- Fetching aircraft data from `https://opendata.adsb.fi/api/v2/lat/{lat}/lon/{lon}/dist/250`.
- Fetching flight route details from `https://api.adsbdb.com/v0/callsign/{callsign}`.
- Validating host names to prevent unintended requests.
- Batching and throttling flight route requests via `FlightRouteQueryThreshold`.

The request layer uses a 25-second HTTP client timeout and only permits configured hosts.

### internal.Dashboard

Located in `../internal/application/dashboard.go`, the `Dashboard` holds runtime state and performs enrichment and statistics.

It tracks:

- Current aircraft sightings
- Fastest and highest aircraft seen
- Rare sightings by type, operator, and country
- Seen counts for types, operators, and countries
- Flight route cache
- History of aircraft sightings via `AircraftSighting`

The dashboard also loads CSV-backed reference data from `internal/dash/`:

- ICAO aircraft type mapping from `data/ICAOList.csv`
- ICAO airline/operator mapping from `data/Airlines.csv`
- Registration prefix country mapping from `data/RegPrefixList.csv`
- ICAO hex range country mapping from `data/ICAOHexRange.csv`
- Military operator lookup from `data/MilICAOOperatorLookUp.csv`

### internal/dash

The `internal/dash/` package provides parsing for CSV-based lookups and distance calculations.

- `GetIcaoToAircraftMap()` parses aircraft type metadata.
- `GetIcaoToAirlineMap()` parses airline/operator metadata.
- `GetRegPrefixMap()` parses registration prefix country assignments.
- `GetHexRangeToCountryMap()` parses assigned ICAO hex ranges to countries.
- `GetMilCodeToOperatorMap()` parses military callsign operator mappings.
- `Distance()` computes haversine distance between geographic coordinates.

### internal/notification

Located in `../internal/infrastructure/notify/notification.go`, the notification layer handles summary printing and desktop notifications.

Notifications are triggered for rare sightings using a combination of rarity flags:

- rare aircraft type
- rare operator
- rare country

`beeep` is used to emit desktop notifications, and the icon at `assets/icon.png` is included where supported.

### internal/persistence

Located in `internal/persistence.go`, persistence stores runtime state in a JSON file.

#### State file

- Filename: `airspottr_state.json`
- Directory: OS-specific user config directory (`os.UserConfigDir()`)
- Data includes dashboard state, cached flight routes, aircraft sighting history, and pending callsigns.

State is loaded on startup and saved on graceful shutdown.

### Ticker frontend

`tickerapp/tickerapp.go` implements a lightweight text-first run mode.

It uses:

- periodic tickers for aircraft updates and summary reports
- a goroutine to perform recurring fetch and enrich cycles
- signal handling for `SIGINT` and `SIGTERM`
- `internal.Notify` for summary output and desktop notifications

The ticker loop:

1. Fetch aircraft.
2. Process aircraft in the dashboard.
3. Emit rarity notifications.
4. Assign cached flight routes or request missing route data.
5. Print hourly summaries.

### TUI frontend

The `tuiapp/` package provides an interactive terminal UI built on `bubbletea`.

Key responsibilities:

- `run.go` for application startup and shutdown handling
- `model.go` for Bubble Tea model state
- `view.go` for rendering tables, panels, and notify controls
- `update.go` for handling user input and periodic updates
- `table.go` and `table_bundle.go` for table creation, sorting, and formatting
- `keymap.go` and `focus.go` for navigation and input focus

The TUI displays current aircraft data, rarity panels, and allows toggling notification dimensions.

## Data flow

1. The CLI parses the location and mode.
2. `main.go` creates `internal.Request` and `internal.Dashboard`.
3. Backend updates fetch aircraft from ADS-B APIs.
4. The dashboard enriches records from local data and computes rarity metrics.
5. Flight routes are requested for callsigns that lack cached route data.
6. The selected frontend renders or prints the enriched results.

## Important constants

- `AircraftUpdateInterval` - 30 seconds
- `SummaryInterval` - 1 hour
- `DashboardWarmup` - 1 hour
- `FlightRouteQueryThreshold` - 10 concurrent flight route requests

## Practical developer notes

- The core application assumes the static `data/` CSV files are located relative to the repository root.
- `internal.Request` validates host names explicitly and does not permit arbitrary URLs.
- `internal.Dashboard` only restores persisted state if the stored location coordinates match the configured `lat` and `lon`.
- Rarity is calculated using a logarithmic threshold over seen counts.
- `tuiapp` uses `bubbletea.WithAltScreen()` to render in the terminal alternate screen buffer.

## Findings and extension suggestions

- If you add a new external API, keep HTTP wrapper concerns in `internal/request.go` and enrichment in `../internal/application/dashboard.go`.
- If you add new UI state or controls, follow the existing `tuiapp` pattern: `model.go` for state, `view.go` for rendering, `update.go` for event handling.
- Consider adding a dedicated configuration or environment-driven data path if you need to support non-root repository execution.
- Use the existing test packages as a guide for domain-level unit tests in `internal/` and UI-level tests in `tuiapp/`.
