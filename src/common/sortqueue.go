package common

type SortItem struct {
	Value    interface{}
	Priority string
}

type SortQueue []*SortItem

func (pq SortQueue) Len() int { return len(pq) }

func (pq SortQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Priority < pq[j].Priority
}

func (pq SortQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
