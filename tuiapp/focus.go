package tuiapp

// inputFocus determines whether keys target the table(s) or the notify checkbox strip.
type inputFocus int

const (
	focusTable inputFocus = iota
	focusNotifyStrip
)
