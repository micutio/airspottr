[<img src="./assets/icon.png" width="100" />](./assets/icon.png)

# Flight Tracker

Testing flight tracking APIs and http/get requests in Go.

Queries various ADS-B APIs every 30 seconds

## Current output

- list of detected military aircraft world-wide (once upon program start)
- fastest aircraft
- highest aircraft

## TODO

- [x] track occurrence by aircraft type
- [x] after a warmup period (default 15 mins):
  - [x] immediate output whenever a rare aircraft is spotted (rarity continually updated)
- [x] closest military whenever there is one within 250NM
- hourly output of top 5 aircraft by:
  - [x] most common types spotted
  - [x] least common type spotted
  - [x] fastest aircraft spotted
  - [x] highest aircraft spotted
- [x] system-native notifications when rare aircraft spotted
- [x] allow notifications for the same aircraft if enough time passed since last contact (24h?)
- [ ] persist all processed responses in history JSON file
- [ ] printing to console, logging to file
- [ ] colored console output
- [ ] graceful shutdown
- [ ] better visual output via some TUI library
- [ ] current aircraft closest to your location
- [ ] "latency", i.e.: time difference between ICAO info timestamp and displaying on screen
- [ ] radar-like visual for nearby aircraft (look up APIs for geographic data)
- [ ] more flight information, e.g.: origin, destination, flight time remaining
- [ ] more unit testing
- [ ] map hex ranges to countries of registration
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
