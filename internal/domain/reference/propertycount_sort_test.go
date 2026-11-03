package reference

import (
	"slices"
	"testing"
)

func TestGetSortedCountsForProperty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]int
		expected []PropertyCountTuple
	}{
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: []PropertyCountTuple{},
		},
		{
			name:     "single item",
			input:    map[string]int{"one": 1},
			expected: []PropertyCountTuple{{Property: "one", Count: 1}},
		},
		{
			name:  "multiple items",
			input: map[string]int{"one": 1, "three": 3, "two": 2},
			expected: []PropertyCountTuple{
				{Property: "one", Count: 1},
				{Property: "two", Count: 2},
				{Property: "three", Count: 3},
			},
		},
		{
			name:  "items with same count are deterministically ordered by property name",
			input: map[string]int{"b": 1, "c": 2, "a": 1},
			expected: []PropertyCountTuple{
				{Property: "a", Count: 1},
				{Property: "b", Count: 1},
				{Property: "c", Count: 2},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := GetSortedCountsForProperty(test.input)
			if !slices.Equal(got, test.expected) {
				t.Errorf("GetSortedCountsForProperty() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestSortedPropertyCountsByPropertyAndDirection(t *testing.T) {
	t.Parallel()

	input := map[string]int{"z": 1, "a": 2, "b": 2}

	byCountAsc := SortedPropertyCounts(input, false, false)
	if byCountAsc[0].Property != "z" || byCountAsc[1].Property != "a" || byCountAsc[2].Property != "b" {
		t.Fatalf("unexpected byCountAsc order: %+v", byCountAsc)
	}

	byNameDesc := SortedPropertyCounts(input, true, true)
	if byNameDesc[0].Property != "z" || byNameDesc[1].Property != "b" || byNameDesc[2].Property != "a" {
		t.Fatalf("unexpected byNameDesc order: %+v", byNameDesc)
	}
}
