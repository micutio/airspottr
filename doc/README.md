# airspottr Developer Documentation

This directory contains developer-facing documentation for the `airspottr` terminal plane spotting application.

## Purpose

This documentation is intended for contributors, maintainers, and engineers who want to understand the architecture, development workflow, and extension points of the project.

## Contents

- `doc/README.md` - Developer overview, build and run instructions, repository layout, and general guidance.
- `doc/architecture.md` - Component and data flow design for the project.
- `doc/ddd-analysis.md` - DDD-focused analysis of the current architecture and domain boundaries.
- `doc/ddd-migration-guide.md` - Step-by-step migration guide to move `airspottr` toward a DDD architecture.

## Quick start

### Requirements

- Go 1.26 installed
- Git

### Build

From the repository root:

```bash
go build ./...
```

### Run

The application supports two modes:

- interactive TUI mode (default)
- line-oriented ticker mode (`--ticker` or `-t`)

Example:

```bash
go run . --location hamburg
```

or to run the ticker:

```bash
go run . --ticker --latlon 40.7128,-74.0060
```

### Test

Run the full Go test suite:

```bash
go test ./...
```

## Repository layout

- `main.go` - Application entrypoint and CLI flag handling.
- `internal/` - Core logic and domain model.
  - `aircraft.go` - Aircraft JSON model and helper methods.
  - `dashboard.go` - Stateful dashboard that tracks seen aircraft, rarity metrics, fastest/highest aircraft, and enriches sightings.
  - `flightroute.go` - Flight route data model and default placeholders.
  - `notification.go` - Desktop and summary notifications using `beeep`.
  - `persistence.go` - Persistent state serialization and restore.
  - `request.go` - HTTP request layer for ADS-B and flight route APIs.
  - `rarity.go`, `sighting.go`, `sort.go` - Supporting types and utility logic.
  - `dash/` - CSV-backed data loaders for ICAO, registration prefixes, country lookup, and distance utilities.
- `tuiapp/` - Bubble Tea-based interactive terminal UI.
- `tickerapp/` - Simple ticker mode printing updates and scheduled summaries without a full TUI.
- `data/` - Supporting static CSV datasets used to resolve aircraft types, airlines, countries, and registry ranges.
- `assets/` - Application icon used for desktop notifications.

## Running locally

### Default behavior

If no location flags are provided, the application uses the default coordinate `0.000000,0.000000`.

It is recommended to provide either a predefined location or latitude/longitude to get useful results.

### Predefined locations

Supported predefined locations in `main.go`:

- `hamburg`
- `new-york`
- `singapore`

Example:

```bash
go run . --location singapore
```

### Custom coordinates

Provide a latitude and longitude pair with `--latlon` or `-l`:

```bash
go run . --latlon 52.5200,13.4050
```

## Extension points

### Adding new UI behavior

- Add new components in `tuiapp/` using Bubble Tea.
- Update models and views via `tuiapp/model.go`, `tuiapp/view.go`, and `tuiapp/update.go`.

### Changing update frequency

Adjust constants in `internal/request.go`:

- `AircraftUpdateInterval` controls how often the aircraft list is refreshed.
- `SummaryInterval` controls how often the summary is printed.
- `DashboardWarmup` controls how long the initial warmup period lasts before rare sightings are reported.

### Supporting a new data source

- Add new request logic in `internal/request.go`.
- Update dashboard enrichment in `../internal/application/dashboard.go`.
- Add any new data source mapping or parsing code under `internal/dash/`.

### Persistence

Persistent state is saved to the user configuration directory by default (`$XDG_CONFIG_HOME` on Linux, or the OS-specific config path).  The state file is named `airspottr_state.json`.

## Notes for maintainers

- Static dataset files in `data/` are required for correct operation and are loaded at startup.
- The project currently uses `beeep` for desktop notifications; desktop notifications may behave differently by OS.
- The ticker mode is designed for piping and scripting, while the TUI is interactive.
- The app is intentionally conservative with external requests: flight route requests are batched and rate-limited by `FlightRouteQueryThreshold`.
