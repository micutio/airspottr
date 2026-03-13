package internal

type RarityFlag int

const (
	RarityConstant float64 = 6.0
)

const (
	NoRarity                RarityFlag = 0b000
	RareType                RarityFlag = 0b001
	RareOperator            RarityFlag = 0b010
	RareCountry             RarityFlag = 0b100
	RareTypeAndOperator     RarityFlag = 0b011
	RareTypeAndCountry      RarityFlag = 0b101
	RareOperatorAndCountry  RarityFlag = 0b110
	RareTypeOperatorCountry RarityFlag = 0b111
)
