package transport

import (
	"container/heap"
	"fmt"
)

type PacketInFlight struct {
	ExpireTimeNanos int64
	//Packet          TcpPacket
	HasAcked bool
	index    int
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*PacketInFlight

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].ExpireTimeNanos < pq[j].ExpireTimeNanos
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PacketInFlight)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *PacketInFlight, time int64, hasAcked bool) {
	item.ExpireTimeNanos = time
	item.HasAcked = hasAcked
	heap.Fix(pq, item.index)
}

// This example creates a PriorityQueue with some items, adds and manipulates an item,
// and then removes the items in priority order.
func Heaptest() {
	// Insert a new item and then modify its priority.
	pq := make(PriorityQueue, 0)
	var item *PacketInFlight
	item = &PacketInFlight{
		ExpireTimeNanos: 1,
		HasAcked:        false,
	}
	heap.Push(&pq, item)
	//pq.update(item, 2, false)
	item = &PacketInFlight{
		ExpireTimeNanos: 2,
		HasAcked:        false,
	}
	heap.Push(&pq, item)
	// Take the items out; they arrive in decreasing priority order.
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*PacketInFlight)
		fmt.Printf("ExpireTimeNanos:%d HasAcked:%t\n", item.ExpireTimeNanos, item.HasAcked)
	}

}
