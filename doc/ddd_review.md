# Domain-Driven Design review: airspottr

**Date:** 2026-09-25  
**Stack:** Go 1.26, Bubble Tea TUI, ADS-B HTTP APIs, CSV reference data, JSON file persistence  
**Verdict:** Mid-migration. Package layout and repository ports are real. The behavioral half (rich model, aggregates, events, a single use case) is still unfinished. The app is closer to **layered architecture + ports** than to a rich DDD model.

Existing docs (`ddd-analysis.md`, `ddd-migration-guide.md`, `strategic-ddd.md`) still describe pre-move paths (`internal/request.go`, `internal/dash/`). This review is against the **current** tree.

---

## Scorecard

| Aspect | Assessment |
|--------|------------|
| Strategic design / bounded contexts | Emerging: observation + reference packages; adapters for intake and UI |
| Tactical building blocks | Types and repo interfaces present; **no aggregates, no domain events** |
| Ubiquitous language | Strong spotting vocabulary; polluted by ADS-B JSON field names |
| Dependency rule | **Domain is clean**; application and UI still leak infrastructure |
| Rich vs anemic model | Mostly **anemic**; rarity and identity rules live on `Dashboard` |
| Ports and adapters | Useful and partial; composition is duplicated in ticker and TUI |
| Documentation | Intent is clear; several paths and TODOs are stale |

---

## Current structure

```
main.go                         # CLI: mode + location only
tickerapp/  tuiapp/             # presentation (also composition + refresh loop)
internal/
  domain/
    observation/                # sighting, aircraft DTO, rarity flags
    reference/                  # geo, ICAO, flightroute shapes
    repositories/               # ports only
  application/
    interfaces/                 # SpottingService, NotificationService
    services/                   # Dashboard, Notify, persistence DTOs
  infrastructure/
    adsb/  data/  observation/  persistence/
```

`domain` does not import application, infrastructure, or UI. That is the strongest DDD win so far.

---

## What already adheres to DDD

**Layered packages.** Domain / application / infrastructure / presentation exist and match `doc/target_architecture.md` in spirit (though there is no `cmd/` composition root and no `interface` HTTP layer — this is a TUI app).

**Ports in the domain.** `AircraftRepository`, `FlightrouteRepository`, `SightingRepo`, `AircraftTypeRepo`, `OperatorRepository`, `CountryRepository` are interfaces-only (`internal/domain/repositories`). CSV, HTTP, and in-memory stores implement them.

**Constructor injection.** `NewDashboard` takes repository interfaces, not CSV maps or HTTP clients.

**Ubiquitous language.** Sighting, rarity (type / operator / country), flight / callsign, flightroute, operator, country, ICAO type, warmup, fastest / highest. Package docs state observation vs reference boundaries.

**Some real value-object behavior.** `Coordinates`, haversine `Distance`, and `Direction` live in `reference` with tests. `RarityFlag` bitflags match the domain.

**Two UIs share one application service.** Ticker and TUI both drive `Dashboard` / `Notify` rather than duplicating classification math.

---

## Bounded contexts (intent vs code)

The migration plan names four contexts. Code maps to them only loosely.

| Proposed context | Current home | Maturity |
|------------------|--------------|----------|
| Observation / Sighting | `domain/observation` + rules in `Dashboard` | Types extracted; **rules still application** |
| Reference / Classification | `domain/reference` + `infrastructure/data` | Good split of types vs CSV |
| Flight data intake | `infrastructure/adsb` behind repo ports | Solid adapter shape |
| Presentation | `tuiapp`, `tickerapp`; Notify still in application | UI OK; **beeep** not isolated |

`domain/repositories` is a shared kernel of ports, not a context. Location-scoped history (README high-priority TODO) is not modeled as its own context or aggregate.

---

## Tactical building blocks

### Aggregates

None are declared or enforced. The de-facto consistency boundary is `Dashboard`: location, current sightings, fastest/highest, seen-count maps, warmup. Fields are public (`IsWarmup`, `SeenTypeCount`, …). There is no invariant protection, no identity for a “spotting session,” and no private collection of sightings.

### Entities

- `AircraftSighting` — identity is ICAO hex via the repo map; mutable; mostly data plus `SightingToString()`. JSON tags couple it to persistence.
- `AircraftRecord` — treated as domain, but it is an **ADS-B API DTO** (~50 JSON fields, including `NacP`, `Sil`, `Tisb`).

### Value objects

Informal structs, rarely with constructors or invariants. `RarityNotifyToggles` is a **presentation policy** sitting in the domain. `IcaoAircraftSpec` / `IcaoOperator` / `HexRange` are thin lookup records. `FlightrouteRecord` includes HTTP response wrappers.

### Domain events

None. Rarity is a flag on the sighting; UI/ticker scan flags and call `EmitRarityNotifications`. Docs still talk about future `RareSightingDetected`.

### Domain services

Not extracted. Logarithmic rarity (`log(total) - RarityConstant`), operator/country resolution (ICAO → mil → `ownOp`; operator → hex → registration), and new-vs-updated flight identity all live on `Dashboard`.

### Application services

`Dashboard` implements `SpottingService` but is a **fat God object**, not a thin use case. `SpottingService` documents a full refresh workflow (fetch → process → routes → notify → save) that **neither the interface nor `Dashboard` actually owns**. That loop is duplicated in `tickerapp` and `tuiapp/update.go`.

---

## Anemic model and where the rules live

Domain packages have helpers (`GetFlightNoAsStr`, altitude formatting, geo). Core rules are outside:

- `Dashboard.ProcessAircraftRecords` decides new vs updated flights and applies rarity flags.
- `updateType` / `updateOperator` / `updateCountry` own classification strategy and the logarithmic rarity threshold.
- Sighting factory defaults and direction live in `infrastructure/observation.SightingRepo.GetOrCreateSighting`, which is the wrong layer for domain construction.

The planned `EvaluateBatch(state, records, classifier) → (newState, rareSightings)` from `strategic-ddd.md` is **not implemented**.

---

## Dependency rule — leaks

```
domain  →  stdlib + domain only          CLEAN
application → domain + beeep             LEAK
infrastructure → domain
               → application/services    inward coupling (persistence DTOs)
tuiapp / tickerapp → application + domain + concrete infrastructure
main.go → adsb.RequestOptions + UI       not a true composition root
```

| Leak | Where |
|------|--------|
| ADS-B JSON shape | `domain/observation/aircraft.go` (`AircraftResult`, tagged `AircraftRecord`) |
| HTTP flightroute envelope | `domain/reference/flightroute.go` |
| Persistence JSON on the entity | `AircraftSighting` tags |
| Desktop notify | `application/services/notification_service.go` (`beeep`, `./assets/icon.png`) |
| UI sort types in domain | `ByFlight`, `ByDistance`, `propertycount_sort` |
| Persistence imports application | `infrastructure/persistence` → `services.PersistentState` |
| Dual composition roots | ticker and TUI construct adapters and run the use case |

`target_architecture.md` says interface must never import infrastructure. Both UIs import `adsb`, `data`, `observation`, and `persistence` directly.

---

## Recommended next steps (priority)

These align with the existing migration recipes; most of them are still open.

1. **Move rarity, classification strategy, and flight-identity rules into `domain/observation`** (pure `EvaluateBatch` or a spotting-session aggregate). Keep `Dashboard` as orchestrator only.
2. **Anti-corruption layer** in `infrastructure/adsb`: map JSON DTOs there; keep a lean domain `Aircraft` / `FlightRoute` without ADS-B noise fields.
3. **Notification adapter** under `infrastructure/notify`. Application delivers already-decided rare sightings or events; no `beeep` in application.
4. **One application use case** for the refresh cycle (`Fetch → Process → Routes → Notify → Save`) so ticker and TUI only schedule and render.
5. **Single composition root** (`cmd/airspottr` or `main.go` wiring only). UI depends on `SpottingService` / `NotificationService`, not concrete repos.
6. **Explicit aggregate** (spotting session at a location) with private counts and invariants — also the natural place for per-location history.
7. **Domain events** (`RareSightingDetected`, optionally `FlightRouteAssigned`) instead of polling `RarityFlag` in the UI.
8. **State repository port** so persistence does not import application DTO types.
9. **Constructors / validated VOs** for hex, callsign, coordinates; stop treating `GetOrCreateSighting` as the domain factory.
10. **Refresh docs** so `doc/README.md` and `architecture.md` match `internal/domain|application|infrastructure`.

Smaller consistency issues worth fixing while touching this: `AssignFlightRoutes` keying vs callsign cache; `FlightrouteRepository` naming (`flighroute_repo.go`); warmup (`IsWarmup`) vs TUI finishing warmup immediately.

---

## Bottom line

**Structural DDD is largely in place. Behavioral DDD is not.**

Do not add an event bus, generic unit-of-work, or extra repository ceremony until rarity and sighting identity live in the domain and `Dashboard` shrinks. The highest-leverage move is still recipe 1 from `strategic-ddd.md`: extract observation evaluation from `ProcessAircraftRecords`.
