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
