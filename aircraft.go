package main

// See https://www.adsbexchange.com/version-2-api-wip/
// for further explanations of the fields

type dataType uint

// Type types
const (
	T_INT dataType = 1
	T_STR dataType = 2
)

type AircraftRecord struct {
	Now         float64    `json:"now"`         // time this file was generated in ms
	ResultCount int        `json:"resultCount"` // total aircraft returned
	Ptime       float64    `json:"ptime"`       // server processing time required in ms
	Aircraft    []Aircraft `json:"aircraft"`    // list of Aircraft records
}

type IntOrString struct {
	Type   dataType
	intVal int
	StrVal string
}

type Aircraft struct {
	Alert           int      `json:"alert"`            // flight status alert bit
	AltBaro         any      `json:"alt_baro"`         // altitude in [feet] or string "ground"
	AltGeom         int      `json:"alt_geom"`         // altitude in [feet]
	BaroRate        float64  `json:"baro_rate"`        // rate of change of baro alt in [feet/minute]
	EmitterCategory string   `json:"category"`         // emitter category to identify aircraft or vehicle classes (A0-D7)
	Emergency       string   `json:"emergency"`        // emergency/priority status, 7X00
	Flight          string   `json:"flight"`           // Flight number
	GroundSpeed     float64  `json:"gs"`               // ground speed in [knots]
	Gva             float64  `json:"gva"`              // geometric vertical accuracy
	Hex             string   `json:"hex"`              // ? TODO
	Lat             float32  `json:"lat"`              // Latitude in [decimal degrees]
	Lon             float32  `json:"lon"`              // Longitude in [decimal degrees]
	Messages        int      `json:"messages"`         // total number of Mode-S msg received from aircraft
	Mlat            []string `json:"mlat"`             // position calculation arrival time diffs, TODO: clarify
	NacP            float64  `json:"nac_p"`            // navigation accuracy for position
	NacV            float64  `json:"nac_v"`            // navigation accuracy for velocity
	NavAltitudeMcp  int      `json:"nav_altitude_mcp"` // selected from mode or flight control panel (MCP)/(FCP) or other
	NavHeading      float64  `json:"nav_heading"`      // selected heading (True/Magnetic), magnetic is de-facto standard
	NavQNH          float64  `json:"nav_qnh"`          // altimeter setting (QFE  or QNH/QNE) in [hPa]
	Nic             int      `json:"nic"`              // Navigation Integrity Category
	NicBaro         int      `json:"nic_baro"`         // NIC for barometric altitude
	R               string   `json:"r"`                // ? TODO
	RadiusOfCtn     float64  `json:"rc"`               // Radius of containment, measure of position integrity in [meters]
	Rssi            float64  `json:"rssi"`             // recent average signal power, always negative, in [dbFS]
	Sda             int      `json:"sda"`              // system design assurance
	Seen            float64  `json:"seen"`             // last message received from this aircraft in [seconds] from 'now'
	SeenPos         float64  `json:"seen_pos"`         // last update of position from this aircraft in [seconds] from 'now'
	Sil             int      `json:"sil"`              // Source integrity level
	SilType         string   `json:"sil_type"`         // Source integrity level type
	Spi             int      `json:"spi"`              // flight status special position identification bit
	Squawk          string   `json:"squawk"`           // Mode A code (Squawk) encoded as 4 octal digits
	IcaoType        string   `json:"t"`                // aircraft ICAO type pulled from database
	Tisb            []string `json:"tisb"`             // list of fields derived from TIS-B data
	Track           float64  `json:"track"`            // true track over ground in degrees (0-359)
	Type            string   `json:"type"`             // type of underlying messages
	Version         int      `json:"version"`          // ADS-B Version number 0,1,2 (3-7 are reserved)
	GeomRate        float64  `json:"geom_rate"`        // Rate of change of geometric (GNSS/INS) altitude in [feet/minute]
	DbFlags         int      `json:"dbFlags"`          // bitfield for certain database flags (check doc for programming lang)
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
	GpsOkBefore     string   `json:"gpsOkBefore"`      // experimental, last timestamp of working GPS
	GpsOkLat        string   `json:"gpsOkLat"`
	GpsOkLon        string   `json:"gpsOkLon"`
	LastPosition    any      `json:"lastPosition"`     // TODO: Type
	RrLat           float64  `json:"rr_lat"`           // rough estimated latitude if no ADS-B or MLAT available
	RrLon           float64  `json:"rr_lon"`           // rough estimated longitude if no ADS-B or MLAT available
	CalcTrack       any      `json:"calc_track"`       // ? TODO
	NavAltitudeFMS  float64  `json:"nav_altitude_fms"` // selected altitude from the flight management system (FMS)
}
