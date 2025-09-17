package internal

const (
	// typeRarityThreshold denotes the maximum rate an aircraft type is seen to be considered rare.
	typeRarityThreshold = 0.001
	// operatorRarityThreshold denotes the maximum rate an operator is seen to be considered rare.
	operatorRarityThreshold = 0.001
	// countryRarityThreshold denotes the maximum rate a country is seen to be considered rare.
	countryRarityThreshold = 0.001
)

type RarityFlag int

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
