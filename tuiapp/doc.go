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
//
// Table sort (while the table is focused):
//   [ / ]     previous / next sort column
//   r         reverse ascending/descending
//   1–8       aircraft table: sort by column (1=DST … 8=HDG)
//   1 / 2     rarity tables: sort by count / by name (Type, operator, country)
package tuiapp
