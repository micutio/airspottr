package tuiapp

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/micutio/airspottr/internal"
)

// Error types

var errColumnMismatch = errors.New("number of columns does not match number of format columns")

// Automated Table Formatting

type tableColumnSizingOption int

const (
	// fixed column width, regardless of table width.
	fixed tableColumnSizingOption = iota
	// relative column with, given as percentage of the total table width.
	relative
	// fill columns receive any remaining table space, evenly distributed.
	fill
)

type columnFormat struct {
	option tableColumnSizingOption
	value  float32
}

type tableFormat struct {
	columnSizes        []columnFormat
	fixedWidth         int     // fixedWidth is the total space taken up by all fixed-width columns.
	fillWidthCount     int     // fillWidthCount indicates how many columns have fill width.
	totalRelativeWidth float32 // how much width is taken by relative columns.
}

func newTableFormat(items ...columnFormat) tableFormat {
	var totalRelativeWidth float32
	fixedWidth := 0
	fillWidthCount := 0

	for _, item := range items {
		switch item.option {
		case relative:
			totalRelativeWidth += item.value
			continue
		case fixed:
			fixedWidth += int(item.value)
			continue
		case fill:
			fillWidthCount++
			continue
		}
	}

	return tableFormat{
		columnSizes:        items,
		fixedWidth:         fixedWidth,
		fillWidthCount:     fillWidthCount,
		totalRelativeWidth: totalRelativeWidth,
	}
}

// Integrated Formatted Table Type

type autoFormatTable struct {
	table  table.Model
	format tableFormat
}

// TODO: Take table padding into account!
func (aft *autoFormatTable) resize(newWidth int) error {
	columnCount := len(aft.table.Columns())
	if columnCount != len(aft.format.columnSizes) {
		return fmt.Errorf(
			"table.resize: %w -> %d in table, %d in tableFormat",
			errColumnMismatch,
			columnCount,
			len(aft.format.columnSizes))
	}

	adjustedWidth := newWidth - 1 - columnCount
	aft.table.SetWidth(adjustedWidth)
	totalRelativeWidth := int(float32(adjustedWidth) * aft.format.totalRelativeWidth)
	totalFillWidth := adjustedWidth - totalRelativeWidth - aft.format.fixedWidth
	fillPerColumn := int(float32(totalFillWidth) / float32(aft.format.fillWidthCount))

	for idx := range columnCount {
		format := aft.format.columnSizes[idx]
		switch format.option {
		case fixed:
			aft.table.Columns()[idx].Width = int(format.value)
			continue
		case relative:
			aft.table.Columns()[idx].Width = int(format.value * float32(newWidth))
			continue
		case fill:
			aft.table.Columns()[idx].Width = fillPerColumn
			continue
		}
	}

	return nil
}

func (aft *autoFormatTable) SetHeight(height int) {
	aft.table.SetHeight(height)
}

func newCurrentAircraftTable(tableStyle table.Styles) autoFormatTable {
	dstLen := 7
	fnoLen := 10
	spdLen := 5
	initialTableHeight := 5
	format := newTableFormat(
		columnFormat{fixed, float32(dstLen)},
		columnFormat{fixed, float32(dstLen)},
		columnFormat{fill, 0.0},
		columnFormat{fixed, float32(dstLen)},
		columnFormat{fixed, float32(spdLen)},
		columnFormat{fixed, float32(dstLen)},
	)

	currentAircraftTbl := table.New(
		// table header
		table.WithColumns(
			[]table.Column{
				{Title: "DST", Width: dstLen},
				{Title: "FNO", Width: fnoLen},
				{Title: "TID", Width: 0},
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

	return autoFormatTable{
		table:  currentAircraftTbl,
		format: format,
	}
}

func newTypeRarityTable(tableStyle table.Styles) autoFormatTable {
	countLen := 6
	typeNameLen := 12
	initialTableHeight := 5
	format := newTableFormat(
		columnFormat{fixed, float32(countLen)},
		columnFormat{fill, float32(typeNameLen)},
	)

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

	return autoFormatTable{
		table:  typeRarityTbl,
		format: format,
	}
}

func newOperatorRarityTable(tableStyle table.Styles) autoFormatTable {
	countLen := 6
	operatorNameLen := 12
	initialTableHeight := 5
	format := newTableFormat(
		columnFormat{fixed, float32(countLen)},
		columnFormat{fill, float32(operatorNameLen)},
	)

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

	return autoFormatTable{
		table:  operatorRarityTbl,
		format: format,
	}
}

func newCountryRarityTable(tableStyle table.Styles) autoFormatTable {
	countLen := 6
	countryNameLen := 12
	initialTableHeight := 5
	format := newTableFormat(
		columnFormat{fixed, float32(countLen)},
		columnFormat{fill, float32(countryNameLen)},
	)

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

	return autoFormatTable{
		table:  countryRarityTbl,
		format: format,
	}
}

func aircraftToRow(aircraft *internal.AircraftRecord) table.Row {
	return table.Row{
		fmt.Sprintf("%3.0f", aircraft.CachedDist),
		aircraft.GetFlightNoAsStr(),
		aircraft.CachedType,
		aircraft.GetAltitudeAsStr(),
		fmt.Sprintf("%3.0f", aircraft.GroundSpeed),
		fmt.Sprintf("%3.0f", aircraft.NavHeading),
	}
}

func propertyCountToRow(propCount internal.PropertyCountTuple) table.Row {
	return table.Row{fmt.Sprintf("%5d", propCount.Count), propCount.Property}
}
