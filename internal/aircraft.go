package internal

import (
	"fmt"
	"strings"
	"unicode"
)

// See https://www.adsbexchange.com/version-2-api-wip/
// for further explanations of the fields

// civAircraftResult mirrors the JSON which is returned for aircraft queries within a given distance.
type civAircraftResult struct {
	Now         float64          `json:"now"`         // time this file was generated in [ms]
	ResultCount int              `json:"resultCount"` // total count of aircraft returned
	Ptime       float64          `json:"ptime"`       // server processing time required in [ms]
	Aircraft    []AircraftRecord `json:"aircraft"`    // list of Aircraft records
}

// AircraftRecord is used by both civilian and military aircraft queries.
type AircraftRecord struct {
	Alert           int      `json:"alert"`            // flight status alert bit
	AltBaro         any      `json:"alt_baro"`         // altitude in [feet] or string "ground"
	AltGeom         int      `json:"alt_geom"`         // altitude in [feet]
	BaroRate        float64  `json:"baro_rate"`        // rate of change of baro alt in [feet/minute]
	EmitterCategory string   `json:"category"`         // emitter category to identify aircraft or vehicle classes (A0-D7)
	Emergency       string   `json:"emergency"`        // emergency/priority status, 7X00
	Flight          string   `json:"flight"`           // Flight number
	GroundSpeed     float64  `json:"gs"`               // ground speed in [knots]
	Gva             float64  `json:"gva"`              // geometric vertical accuracy
	Hex             string   `json:"hex"`              // hex code ID for aircraft, assumed to be unique
	Lat             float64  `json:"lat"`              // Latitude in [decimal degrees]
	Lon             float64  `json:"lon"`              // Longitude in [decimal degrees]
	Messages        int      `json:"messages"`         // total number of Mode-S msg received from aircraft
	Mlat            []string `json:"mlat"`             // position calculation arrival time diffs
	NacP            float64  `json:"nac_p"`            // navigation accuracy for position
	NacV            float64  `json:"nac_v"`            // navigation accuracy for velocity
	NavAltitudeMcp  int      `json:"nav_altitude_mcp"` // selected from mode or flight control panel (MCP)/(FCP) or other
	NavHeading      float64  `json:"nav_heading"`      // selected heading (True/Magnetic), magnetic is de-facto standard
	NavQNH          float64  `json:"nav_qnh"`          // altimeter setting (QFE  or QNH/QNE) in [hPa]
	Nic             int      `json:"nic"`              // Navigation Integrity Category
	NicBaro         int      `json:"nic_baro"`         // NIC for barometric altitude
	Registration    string   `json:"r"`                // Registration of the aircraft
	RadiusOfCtn     float64  `json:"rc"`               // Radius of containment, measure of position integrity in [meters]
	Rssi            float64  `json:"rssi"`             // recent average signal power, always negative, in [dbFS]
	Sda             int      `json:"sda"`              // system design assurance
	Seen            float64  `json:"seen"`             // last message received from aircraft in [seconds] from 'now'
	SeenPos         float64  `json:"seen_pos"`         // last update of position from aircraft in [seconds] from 'now'
	Sil             int      `json:"sil"`              // Source integrity level
	SilType         string   `json:"sil_type"`         // Source integrity level type
	Spi             int      `json:"spi"`              // flight status special position identification bit
	Squawk          string   `json:"squawk"`           // Mode A code (Squawk) encoded as 4 octal digits
	IcaoType        string   `json:"t"`                // aircraft ICAO type pulled from database
	Tisb            []string `json:"tisb"`             // list of fields derived from TIS-B data
	Track           float64  `json:"track"`            // true track over ground in degrees (0-359)
	Type            string   `json:"type"`             // type of underlying messages
	Version         int      `json:"version"`          // ADS-B Version number 0,1,2 (3-7 are reserved)
	GeomRate        float64  `json:"geom_rate"`        // Rate of change of geometric (GNSS/INS) altitude in [ft/min]
	DBFlags         int      `json:"dbFlags"`          // bitfield for certain database flags (programming language)
	NavModes        []string `json:"nav_modes"`        // (autopilot, vnav, althold, approach, lnav, tcas)
	TrueHeading     float64  `json:"true_heading"`     // Heading clockwise from true north in [degrees]
	Ias             float64  `json:"ias"`              // indicated airspeed in [knots]
	Mach            float64  `json:"mach"`             // Mach number
	MagHeading      float64  `json:"mag_heading"`      // Heading clockwise from magnetic north in [degrees]
	Oat             float64  `json:"oat"`              // outer air temperature
	Roll            float64  `json:"roll"`             // roll, negative is left, in [degrees]
	Tas             float64  `json:"tas"`              // true airspeed in [knots]
	Tat             float32  `json:"tat"`              // total air temperature, might be inaccurate at lower alt, in [C]
	TrackRate       float64  `json:"track_rate"`       // rate of change of track in [degrees/second]
	WindDirection   float64  `json:"wd"`               // wind direction
	WindSpeed       float64  `json:"ws"`               // wind speed
	GpsOkBefore     float64  `json:"gpsOkBefore"`      // experimental, last timestamp of working GPS
	GpsOkLat        float64  `json:"gpsOkLat"`         // experimental, last timestamp of working Latitude
	GpsOkLon        float64  `json:"gpsOkLon"`         // experimental, last timestamp of working Longitude
	LastPosition    any      `json:"lastPosition"`     // TODO: Type
	RrLat           float64  `json:"rr_lat"`           // rough estimated latitude if no ADS-B or MLAT available
	RrLon           float64  `json:"rr_lon"`           // rough estimated longitude if no ADS-B or MLAT available
	CalcTrack       any      `json:"calc_track"`       // ? TODO
	NavAltitudeFMS  float64  `json:"nav_altitude_fms"` // selected altitude from the flight management system (FMS)
	// found by my own investigation
	OwnOp       string `json:"ownOp"` // owner or operator, only rarely set
	Description string `json:"desc"`  // aircraft type description
	// cached data
	CachedDist float64
	CachedType string
}

// GetAltitudeAsStr reads the altitude of an aircraft and returns it as a string.
// The altitude is stored either as a string 'ground' or as a float denoting the measured
// barometric altitude.
// If the latter is the case, the float will be formatted without any decimal places
// (unnecessary accuracy) and converted to string.
func (ac *AircraftRecord) GetAltitudeAsStr() string {
	if num, numOk := ac.AltBaro.(float64); numOk {
		return fmt.Sprintf("%5.0f", num)
	}

	if str, strOk := ac.AltBaro.(string); strOk {
		return str
	}

	return altitudeUnknown
}

// GetFlightNoAsStr converts the flight number to a unified string of length 8.
// Returns either the full flight number or 'unknown ' if it was not transmitted.
func (ac *AircraftRecord) GetFlightNoAsStr() string {
	if ac.Flight == "" {
		return flightUnknown
	}

	return ac.Flight
}

// GetFlightNoAsIcaoCode trims whitespaces and digits from the flight number,
// resulting in the three-digit icao code for civilian flights and arbitrary length codes
// for military, government and private flights.
func (ac *AircraftRecord) GetFlightNoAsIcaoCode() string {
	if len(ac.Flight) == 0 {
		return flightUnknownCode
	}

	return stripDigits(strings.TrimSpace(ac.Flight))
}

// GetRegistrationPrefix returns the prefix of the registration if it exists,
// otherwise it returns the entire registration.
func (ac *AircraftRecord) GetRegistrationPrefix() string {
	prefixEnd := strings.Index(ac.Registration, "-")

	if prefixEnd != -1 {
		return ac.Registration[0:prefixEnd]
	}
	return ac.Registration
}

func stripDigits(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return -1 // Remove the digit
		}
		return r // Keep the character
	}, str)
}

// ByFlight implements the comparator interface and allows sorting a list of aircraft records
// by flight.
type ByFlight []AircraftRecord

func (a ByFlight) Len() int           { return len(a) }
func (a ByFlight) Less(i, j int) bool { return a[i].Flight < a[j].Flight }
func (a ByFlight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// ByDistance implements the comparator interface and allows sorting a list of aircraft records.
// by distance to a given lon,lat coordinate.
type ByDistance []AircraftRecord

func (a ByDistance) Len() int           { return len(a) }
func (a ByDistance) Less(i, j int) bool { return a[i].CachedDist < a[j].CachedDist }
func (a ByDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
