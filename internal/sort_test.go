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
			name:     "multiple items",
			input:    map[string]int{"one": 1, "three": 3, "two": 2},
			expected: []PropertyCountTuple{{Property: "one", Count: 1}, {Property: "two", Count: 2}, {Property: "three", Count: 3}},
		},
		{
			name:  "items with same count",
			input: map[string]int{"a": 1, "c": 2, "b": 1},
			// The order of items with the same count is not guaranteed, so we accept either
			expected: []PropertyCountTuple{{Property: "a", Count: 1}, {Property: "b", Count: 1}, {Property: "c", Count: 2}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSortedCountsForProperty(tt.input)
			// For the case with same counts, the order is not guaranteed.
			// We can't do a deep equal directly.
			if len(got) != len(tt.expected) {
				t.Errorf("GetSortedCountsForProperty() = %v, want %v", got, tt.expected)
			}

			// A simple check for this specific test case. For more complex scenarios,
			// a more robust comparison would be needed.
			if tt.name == "items with same count" {
				if !((reflect.DeepEqual(got[0], tt.expected[0]) && reflect.DeepEqual(got[1], tt.expected[1])) ||
					(reflect.DeepEqual(got[0], tt.expected[1]) && reflect.DeepEqual(got[1], tt.expected[0]))) {
					t.Errorf("GetSortedCountsForProperty() got = %v, want %v", got, tt.expected)
				}
				if !reflect.DeepEqual(got[2], tt.expected[2]) {
					t.Errorf("GetSortedCountsForProperty() got = %v, want %v", got, tt.expected)
				}
			} else if len(tt.expected) > 0 && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetSortedCountsForProperty() = %v, want %v", got, tt.expected)
			}
		})
	}
}
