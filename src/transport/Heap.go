package transport

import (
	"container/heap"
	"fmt"
)

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*PacketInFlight

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].ExpireTimeNanos < pq[j].ExpireTimeNanos
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PacketInFlight)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *PacketInFlight, time int64) {
	item.ExpireTimeNanos = time
	heap.Fix(pq, item.Index)
}

// This example creates a PriorityQueue with some items, adds and manipulates an item,
// and then removes the items in priority order.
func Heaptest() {
	// Insert a new item and then modify its priority.
	pq := make(PriorityQueue, 0)
	var item *PacketInFlight
	item = &PacketInFlight{
		ExpireTimeNanos: 3,
	}
	heap.Push(&pq, item)
	item = &PacketInFlight{
		ExpireTimeNanos: 1,
	}
	heap.Push(&pq, item)
	//pq.update(item, 2, false)
	item = &PacketInFlight{
		ExpireTimeNanos: 2,
	}
	heap.Push(&pq, item)
	// Take the items out; they arrive in decreasing priority order.
	for pq.Len() > 0 {
		fmt.Printf("top %d", pq[pq.Len()-1].ExpireTimeNanos)
		item := heap.Pop(&pq).(*PacketInFlight)
		fmt.Printf("ExpireTimeNanos:%d\n", item.ExpireTimeNanos)
	}

}
