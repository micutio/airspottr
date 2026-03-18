// Package tuiapp provides the TUI app which displays flight tracking data, updates continuously
// and can be interacted with.
//
// Layout idea:
//
//	+-------------------------------------------------+
//	| last update time: 00:00:00                      |
//	|                                                 |
//	| Highest Aircraft                                |
//	| ALT: ... FNO: ... Type: ... REG: ...            |
//	| Fastest Aircraft                                |
//	| SPD: ... FNO: ... Type: ... REG: ...            |
//	|  ________________________       ______________  |
//	| | current aircraft table |     | rarity table | |
//	| | entry 0                |     | entry 0      | |
//	| | ...                    |     | ...          | |
//	| | entry N                |     | entry M      | |
//	|  ------------------------       --------------  |
//	+-------------------------------------------------+
package tuiapp
