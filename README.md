[<img src="./assets/icon.png" width="100" />](./assets/icon.png)

# airspottr - Plane Spotting In Your Terminal

A command line tool which shows real-life aircraft in your surrounding
and keeps track of rare aircraft types, airlines and countries of origin.

## Output

- list of aircraft currently in the vicinity of a given geographical coordinate
- fastest aircraft overall recorded
- highest aircraft overall recorded
- list of aircraft types by rarity
- list of airlines by rarity
- list of countries of origin by rarity

## TODO

- [x] show total uptime in summaries
- [x] show location in TUI and maybe ticker output
- [ ] TUI checkboxes to toggle notifications for type/operator/country individually
- [ ] colored console output
- [ ] graceful shutdown
- [ ] current aircraft closest to your location
- [ ] graphs for aircraft type count over time
- [ ] radar-like visual for nearby aircraft (look up APIs for geographic data)
- [ ] more flight information, e.g.: origin, destination, flight time remaining
- [ ] more unit testing
- [ ] collect additional information about unknown aircraft/types to try and identify later
  - [ ] lat/lon position

## Links for further investigation

### Flight Tracking APIs

Free ADSB APIs:

 - https://opendata.adsb.fi/api/v2/ (https://github.com/adsbfi/opendata)
 - https://github.com/ADSB-One/api
 - https://airplanes.live/api-guide/
   - might shift to feeder only
 - https://api.adsb.lol/docs
 - https://openskynetwork.github.io/opensky-api/rest.html
   - limited to [now] time, but access to bounding boxes and airport ARR/DEP
 - example for antarctica bounding box: https://opensky-network.org/api/states/all?lamin=-90&lomin=-180&lamax=-50&lomax=180

### Airline code to name mappings

- https://github.com/tbouron/MMM-FlightTracker

### Various ICAO code mappings

- https://github.com/rikgale/ICAOList

### Raw ADS-B data parsing in Go

- https://github.com/cjkreklow/go-adsb
