package tuiapp

type uiState int

const (
	mainPage        uiState = iota     // first page on startup, showing current aircraft
	aircraftDetails uiState = iota + 1 // current aircraft, overlaid by details of selected
	globalStats     uiState = iota + 2 // second page, showing type, operator and country rarity
)
