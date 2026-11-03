package observation

type RarityFlag int

const (
	RarityConstant float64 = 6.0
)

const (
	NoRarity     RarityFlag = 0b000
	RareType     RarityFlag = 0b001
	RareOperator RarityFlag = 0b010
	RareCountry  RarityFlag = 0b100
)

// RarityNotifyToggles selects which rarity dimensions may trigger desktop notifications.
type RarityNotifyToggles struct {
	Type, Operator, Country bool
}
