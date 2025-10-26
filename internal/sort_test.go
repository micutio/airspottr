package internal

import (
	"reflect"
	"testing"
)

func TestGetSortedCountsForProperty(t *testing.T) {
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
			name:  "items with same count",
			input: map[string]int{"a": 1, "c": 2, "b": 1},
			// The order of items with the same count is not guaranteed, so we accept either
			expected: []PropertyCountTuple{
				{Property: "a", Count: 1},
				{Property: "b", Count: 1},
				{Property: "c", Count: 2},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := GetSortedCountsForProperty(test.input)
			// For the case with same counts, the order is not guaranteed.
			// We can't do a deep equal directly.
			if len(got) != len(test.expected) {
				t.Errorf("GetSortedCountsForProperty() = %v, want %v", got, test.expected)
			}

			// A simple check for this specific test case. For more complex scenarios,
			// a more robust comparison would be needed.
			if test.name == "items with same count" {
				if (!reflect.DeepEqual(got[0], test.expected[0]) || !reflect.DeepEqual(got[1], test.expected[1])) &&
					(!reflect.DeepEqual(got[0], test.expected[1]) || !reflect.DeepEqual(got[1], test.expected[0])) {
					t.Errorf("GetSortedCountsForProperty() got = %v, want %v", got, test.expected)
				}
				if !reflect.DeepEqual(got[2], test.expected[2]) {
					t.Errorf("GetSortedCountsForProperty() got = %v, want %v", got, test.expected)
				}
			} else if len(test.expected) > 0 && !reflect.DeepEqual(got, test.expected) {
				t.Errorf("GetSortedCountsForProperty() = %v, want %v", got, test.expected)
			}
		})
	}
}
