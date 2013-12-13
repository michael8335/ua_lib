package common

import "testing"
import "container/heap"
import "fmt"
import "strconv"

func TestHeap(t *testing.T) {
	pq := make(PriorityQueue, 0)
	for i := 20; i > 0; i-- {
		t := &Item{i, strconv.Itoa(i), i}
		heap.Push(&pq, t)
	}
	for _, v := range pq {
		fmt.Printf("%v\n", v)
	}
}
