package openapi

import "sort"

var locationsOrder = map[string]int{
	"path":   0,
	"query":  1,
	"header": 2,
	"cookie": 3,
}

type paramLessFunc func(p1, p2 *ParameterOrRef) bool

// paramsMultiSorter implements the Sort interface,
// sorting the operation parameters within.
type paramsMultiSorter struct {
	params []*ParameterOrRef
	less   []paramLessFunc
}

// Sort sorts the argument slice according to the less
// functions passed to paramsOrderedBy.
func (pms *paramsMultiSorter) Sort(params []*ParameterOrRef) {
	pms.params = params
	sort.Sort(pms)
}

// Len implements the Sort interface for paramsMultiSorter.
func (pms *paramsMultiSorter) Len() int { return len(pms.params) }

// Swap implements the Sort interface for paramsMultiSorter.
func (pms *paramsMultiSorter) Swap(i, j int) {
	pms.params[i], pms.params[j] = pms.params[j], pms.params[i]
}

// Less implements the Sort interface for paramsMultiSorter.
// It is implemented by looping along the less functions until
// it finds a comparison that discriminates between the two items
// (one is less than the other).
func (pms *paramsMultiSorter) Less(i, j int) bool {
	p, q := pms.params[i], pms.params[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(pms.less)-1; k++ {
		less := pms.less[k]
		switch {
		case less(p, q): // p < q, so we have a decision.
			return true
		case less(q, p): // p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just
	// return whatever the final comparison reports.
	return pms.less[k](p, q)
}

// paramsOrderedBy returns a Sorter that sorts using the less functions,
// in order. Call its Sort method to sort the data.
func paramsOrderedBy(less ...paramLessFunc) *paramsMultiSorter {
	return &paramsMultiSorter{
		less: less,
	}
}
