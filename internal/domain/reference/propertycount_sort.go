package reference

import "sort"

type PropertyCountTuple struct {
	Property string
	Count    int
}

func comparePropertyCountAscending(count1, count2 PropertyCountTuple, byProperty bool) bool {
	if byProperty {
		if count1.Property != count2.Property {
			return count1.Property < count2.Property
		}
		return count1.Count < count2.Count
	}
	if count1.Count != count2.Count {
		return count1.Count < count2.Count
	}
	return count1.Property < count2.Property
}

// SortedPropertyCounts converts a property-to-count map into a deterministically sorted slice.
// When byProperty is false, it sorts by Count first (breaking ties by Property name).
// When byProperty is true, it sorts by Property name first (breaking ties by Count).
func SortedPropertyCounts(propertyCountMap map[string]int, byProperty, desc bool) []PropertyCountTuple {
	tuples := make([]PropertyCountTuple, 0, len(propertyCountMap))
	for key, value := range propertyCountMap {
		tuples = append(tuples, PropertyCountTuple{Property: key, Count: value})
	}

	sort.SliceStable(tuples, func(i, j int) bool {
		less := comparePropertyCountAscending(tuples[i], tuples[j], byProperty)
		if desc {
			return !less
		}
		return less
	})
	return tuples
}

// GetSortedCountsForProperty returns property counts sorted ascending by count, then property name.
func GetSortedCountsForProperty(propertyCountMap map[string]int) []PropertyCountTuple {
	return SortedPropertyCounts(propertyCountMap, false, false)
}
