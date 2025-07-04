package main

// Also see the following for description of fields: https://www.adsbexchange.com/version-2-api-wip/
// Example result record:

/* Aircraft response
    "now":
	"aircraft": [list of aircraft structs]
*/

/* Aircraft struct:

     "alert": 0,
     "alt_baro": 0,
     "alt_geom": 0,
     "baro_rate": 0,
     "category": "string",
     "emergency": "string",
     "flight": "string",
     "gs": 0,
     "gva": 0,
     "hex": "string",
     "lat": 0,
     "lon": 0,
     "messages": 0,
     "mlat": [
       "string"
     ],
     "nac_p": 0,
     "nac_v": 0,
     "nav_altitude_mcp": 0,
     "nav_heading": 0,
     "nav_qnh": 0,
     "nic": 0,
     "nic_baro": 0,
     "r": "string",
     "rc": 0,
     "rssi": 0,
     "sda": 0,
     "seen": 0,
     "seen_pos": 0,
     "sil": 0,
     "sil_type": "string",
     "spi": 0,
     "squawk": "string",
     "t": "string",
     "tisb": [
       "string"
     ],
     "track": 0,
     "type": "string",
     "version": 0,
     "geom_rate": 0,
     "dbFlags": 0,
     "nav_modes": [
       "string"
     ],
     "true_heading": 0,
     "ias": 0,
     "mach": 0,
     "mag_heading": 0,
     "oat": 0,
     "roll": 0,
     "tas": 0,
     "tat": 0,
     "track_rate": 0,
     "wd": 0,
     "ws": 0,
     "gpsOkBefore": 0,
     "gpsOkLat": 0,
     "gpsOkLon": 0,
     "lastPosition": {
       "lat": 0,
       "lon": 0,
       "nic": 0,
       "rc": 0,
       "seen_pos": 0
     },
     "rr_lat": 0,
     "rr_lon": 0,
     "calc_track": 0,
     "nav_altitude_fms": 0
   }
*/
