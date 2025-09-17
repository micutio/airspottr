package internal

import "sort"

type PropertyCountTuple struct {
	Property string
	Count    int
}

type ByCount []PropertyCountTuple

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Less(i, j int) bool { return a[i].Count < a[j].Count }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func GetSortedCountsForProperty(propertyCountMap map[string]int) []PropertyCountTuple {
	propertyCounts := make([]PropertyCountTuple, len(propertyCountMap))
	i := 0
	for key, value := range propertyCountMap {
		propertyCounts[i] = PropertyCountTuple{Property: key, Count: value}
		i++
	}

	sort.Sort(ByCount(propertyCounts))
	return propertyCounts
}
