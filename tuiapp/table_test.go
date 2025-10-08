package tuiapp

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestTableFormat(t *testing.T) {
	tests := []struct {
		name                       string
		format                     tableFormat
		expectedFixedWidth         int
		expectedFillWidthCount     int
		expectedTotalRelativeWidth float32
	}{
		{
			name:                       "singleFixed",
			format:                     newTableFormat(columnFormat{fixed, 10.0}),
			expectedFixedWidth:         10,
			expectedFillWidthCount:     0,
			expectedTotalRelativeWidth: 0.0,
		},
		{
			name:                       "singleRelative",
			format:                     newTableFormat(columnFormat{relative, 0.254}),
			expectedFixedWidth:         0,
			expectedFillWidthCount:     0,
			expectedTotalRelativeWidth: 0.254,
		},
		{
			name:                       "singleFill",
			format:                     newTableFormat(columnFormat{fill, 0.0}),
			expectedFixedWidth:         0,
			expectedFillWidthCount:     1,
			expectedTotalRelativeWidth: 0.0,
		},
		{
			name: "fixedAndRelative",
			format: newTableFormat(
				columnFormat{fixed, 90},
				columnFormat{relative, 0.67},
			),
			expectedFixedWidth:         90,
			expectedFillWidthCount:     0,
			expectedTotalRelativeWidth: 0.67,
		},
		{
			name: "multiFill",
			format: newTableFormat(
				columnFormat{fill, 0},
				columnFormat{fixed, 90},
				columnFormat{fill, 0},
				columnFormat{relative, 0.67},
				columnFormat{fill, 0},
			),
			expectedFixedWidth:         90,
			expectedFillWidthCount:     3,
			expectedTotalRelativeWidth: 0.67,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectedFixedWidth != test.format.fixedWidth {
				t.Errorf(
					"Expected fixedWidth %d, got %d",
					test.expectedFixedWidth,
					test.format.fixedWidth)
			}

			if test.expectedFillWidthCount != test.format.fillWidthCount {
				t.Errorf(
					"Expected fillWidthCount %d, got %d",
					test.expectedFillWidthCount,
					test.format.fillWidthCount)
			}

			if test.expectedTotalRelativeWidth != test.format.totalRelativeWidth {
				t.Errorf(
					"Expected totalRelativeWidth %f, got %f",
					test.expectedTotalRelativeWidth,
					test.format.totalRelativeWidth)
			}
		})
	}
}

func TestAutoFormatTableInit(t *testing.T) {
	tests := []struct {
		name                            string
		tableModel                      table.Model
		tableFormat                     tableFormat
		resizeWidth                     int
		expectedTableWidthAfterResize   int
		expectedColumnWidthsAfterResize []int
	}{
		{
			name: "SingleColumnFixed",
			tableModel: table.New(
				table.WithColumns(
					[]table.Column{
						{Title: "A", Width: 10},
					},
				),
			),
			tableFormat: newTableFormat(
				columnFormat{fixed, 10.0},
			),
			resizeWidth:                     20,
			expectedTableWidthAfterResize:   18,
			expectedColumnWidthsAfterResize: []int{9},
		},
		{
			name: "SingleColumnRelative",
			tableModel: table.New(
				table.WithColumns(
					[]table.Column{
						{Title: "A", Width: 5},
					},
				),
			),
			tableFormat: newTableFormat(
				columnFormat{relative, .5},
			),
			resizeWidth:                     40,
			expectedTableWidthAfterResize:   38,
			expectedColumnWidthsAfterResize: []int{18},
		},
		{
			name: "SingleColumnFill",
			tableModel: table.New(
				table.WithColumns(
					[]table.Column{
						{Title: "A", Width: 10},
					},
				),
			),
			tableFormat: newTableFormat(
				columnFormat{fill, .0},
			),
			resizeWidth:                     15,
			expectedTableWidthAfterResize:   13,
			expectedColumnWidthsAfterResize: []int{12},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			aft := autoFormatTable{
				table:  test.tableModel,
				format: test.tableFormat,
			}

			err := aft.resize(test.resizeWidth)
			if err != nil {
				t.Errorf(
					"resize(%d) failed: %v",
					test.resizeWidth,
					err)
			}

			if aft.table.Width() != test.expectedTableWidthAfterResize {
				t.Errorf(
					"resized table width -> expected: %d, got: %d",
					test.expectedTableWidthAfterResize,
					aft.table.Width())
			}

			for i, col := range aft.table.Columns() {
				if col.Width != test.expectedColumnWidthsAfterResize[i] {
					t.Errorf(
						"resized table col '%s' width -> expected: %d, got: %d",
						test.tableModel.Columns()[i].Title,
						test.expectedColumnWidthsAfterResize[i],
						col.Width)
				}
			}
		})
	}
}
