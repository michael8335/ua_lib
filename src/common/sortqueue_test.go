package common

import "testing"
import "sort"
import "fmt"

func TestSort(t *testing.T) {
	pq := make(SortQueue, 0)
	for i := 20; i > 0; i-- {
		t := &SortItem{i, fmt.Sprintf("%03d%03d", i, i)}
		pq = append(pq, t)
	}
	sort.Sort(pq)
	for _, v := range pq {
		fmt.Printf("%v\t%v\n", v.Value, v.Priority)
	}
}
