package tuiapp

import (
	"github.com/charmbracelet/bubbles/table"
)

func newCurrentAircraftTable(tableStyle table.Styles, maxTypeNameLen int) table.Model {
	dstLen := 7
	fnoLen := 10
	spdLen := 5
	initialTableHeight := 5

	currentAircraftTbl := table.New(
		// table header
		table.WithColumns(
			[]table.Column{
				{Title: "DST", Width: dstLen},
				{Title: "FNO", Width: fnoLen},
				{Title: "TID", Width: maxTypeNameLen},
				{Title: "ALT", Width: dstLen},
				{Title: "SPD", Width: spdLen},
				{Title: "HDG", Width: spdLen},
			},
		),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(initialTableHeight),
		table.WithStyles(tableStyle),
	)

	return currentAircraftTbl
}

func newTypeRarityTable(tableStyle table.Styles) table.Model {
	countLen := 6
	typeNameLen := 12
	initialTableHeight := 5

	// Create a new table with specified columns and initial empty rows.
	typeRarityTbl := table.New(
		// table header
		table.WithColumns(
			[]table.Column{
				{Title: "Count", Width: countLen},
				{Title: "Type", Width: typeNameLen},
			},
		),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(initialTableHeight),
		table.WithStyles(tableStyle),
	)

	return typeRarityTbl
}

func newOperatorRarityTable(tableStyle table.Styles) table.Model {
	countLen := 6
	operatorNameLen := 12

	initialTableHeight := 5

	// Create a new table with specified columns and initial empty rows.
	operatorRarityTbl := table.New(
		// table header
		table.WithColumns(
			[]table.Column{
				{Title: "Count", Width: countLen},
				{Title: "Operator", Width: operatorNameLen},
			},
		),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(initialTableHeight),
		table.WithStyles(tableStyle),
	)

	return operatorRarityTbl
}

func newCountryRarityTable(tableStyle table.Styles) table.Model {
	countLen := 6
	countryNameLen := 12
	initialTableHeight := 5

	// Create a new table with specified columns and initial empty rows.
	countryRarityTbl := table.New(
		// table header
		table.WithColumns(
			[]table.Column{
				{Title: "Count", Width: countLen},
				{Title: "Country", Width: countryNameLen},
			},
		),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(initialTableHeight),
		table.WithStyles(tableStyle),
	)

	return countryRarityTbl
}
