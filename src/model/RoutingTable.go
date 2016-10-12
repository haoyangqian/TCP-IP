package model

import "errors"

import "sync"

// import "fmt"

// import "sync/atomic"

type RoutingTable struct {
	RoutingEntries map[VirtualIp]*RoutingEntry
	Neighbors      map[VirtualIp]VirtualIp //nextHop -> ExitIp
	Lock           *sync.RWMutex
}

func (t *RoutingTable) ReadLock() {
	t.Lock.RLock()
}

func (t *RoutingTable) ReadUnLock() {
	t.Lock.RUnlock()
}

func (t *RoutingTable) WriteLock() {
	t.Lock.Lock()
}

func (t *RoutingTable) WriteUnLock() {
	t.Lock.Unlock()
}
func (t *RoutingTable) HasEntry(ip VirtualIp) bool {
	t.ReadLock()
	// fmt.Println("[RoutingTable] HasEntry Acquired read lock")
	// fmt.Println("[RoutingTable] HasEntry relased read lock")
	defer t.ReadUnLock()
	if _, ok := t.RoutingEntries[ip]; ok {
		return true
	} else {
		return false
	}
}

func (t *RoutingTable) HasNeighbor(neighbor VirtualIp) bool {
	t.ReadLock()
	// fmt.Println("[RoutingTable] HasNeighbor Acquired read lock")
	// fmt.Println("[RoutingTable] HasNeighbor relased read lock")
	defer t.ReadUnLock()
	if _, ok := t.Neighbors[neighbor]; ok {
		return true
	} else {
		return false
	}
}

func (t *RoutingTable) GetEntry(vip VirtualIp) (*RoutingEntry, error) {
	var entry RoutingEntry
	// // fmt.Println("calling hasEntry in GetEntry")
	if !t.HasEntry(vip) {
		// // fmt.Println("hasEntry in GetEntry returned")
		return &entry, errors.New("Invalid State: an entry not in the RoutingTable was requested " + vip.Ip)
	}
	t.ReadLock()
	// fmt.Println("[RoutingTable] GetEntry Acquired read lock")
	// fmt.Println("[RoutingTable] GetEntry relased read lock")
	defer t.ReadUnLock()
	return t.RoutingEntries[vip], nil
}

func (t *RoutingTable) PutEntry(entry *RoutingEntry) {
	// fmt.Println("[RoutingTable] PutEntry WANTS write lock")
	t.WriteLock()
	// fmt.Println("[RoutingTable] PutEntry ACQUIRED write lock")
	t.RoutingEntries[entry.Dest] = entry
	t.WriteUnLock()
	// fmt.Println("[RoutingTable] PutEntry RELEASED write lock")
}

func (t *RoutingTable) DeleteEntry(entry *RoutingEntry) {
	// // fmt.Println("calling hasEntry in DeleteEntry")
	if t.HasEntry(entry.Dest) {
		// // fmt.Println("hasEntry returned")
		// fmt.Println("[RoutingTable] DeleteEntry WANTS write lock")
		t.WriteLock()
		// fmt.Println("[RoutingTable] DeleteEntry ACQUIRED write lock")
		delete(t.RoutingEntries, entry.Dest)
		t.WriteUnLock()
		// fmt.Println("[RoutingTable] DeleteEntry RELEASED write lock")
	}
}

func (t *RoutingTable) PutNeighbor(neighborVip VirtualIp, selfVip VirtualIp) {
	t.WriteLock()
	t.Neighbors[neighborVip] = selfVip
	t.WriteUnLock()
}

func (t *RoutingTable) DeleteNeighbor(neighborVip VirtualIp) {
	if t.HasNeighbor(neighborVip) {
		t.WriteLock()
		delete(t.Neighbors, neighborVip)
		t.WriteUnLock()
	}
}

func (t *RoutingTable) GetNeighbor(vip VirtualIp) (VirtualIp, error) {

	var inter VirtualIp
	if !t.HasNeighbor(vip) {
		return inter, errors.New("Invalid State: an vip not in the neighbor was requested ")
	}
	t.ReadLock()
	defer t.ReadUnLock()
	return t.Neighbors[vip], nil
}

func (t *RoutingTable) GetAllNeighbors() []VirtualIp {
	t.ReadLock()
	defer t.ReadUnLock()
	neighbors := make([]VirtualIp, 0)
	for k, _ := range t.Neighbors {
		neighbors = append(neighbors, k)
	}
	return neighbors
}

func (t *RoutingTable) GetAllEntries() []*RoutingEntry {
	t.ReadLock()
	defer t.ReadUnLock()
	routingentries := make([]*RoutingEntry, 0)
	for _, v := range t.RoutingEntries {
		routingentries = append(routingentries, v)
	}
	return routingentries
}

func (t *RoutingTable) GetUpdatedEntries() []*RoutingEntry {
	// fmt.Println("[RoutingTable] getUpdatedEntries")
	t.ReadLock()
	defer t.ReadUnLock()
	routingentries := make([]*RoutingEntry, 0)
	for _, v := range t.RoutingEntries {
		if v.IsUpdated == true {
			routingentries = append(routingentries, v)
		}
	}
	// fmt.Println("[RoutingTable] getUpdatedEntries DONE")
	return routingentries
}

func (t *RoutingTable) GetExpiredEntries() []*RoutingEntry {
	t.ReadLock()
	defer t.ReadUnLock()
	expiredRoutes := make([]*RoutingEntry, 0)
	for _, v := range t.RoutingEntries {
		if v.ShouldExpire() || v.Expired() {
			expiredRoutes = append(expiredRoutes, v)
		}
	}
	return expiredRoutes
}

func (t *RoutingTable) ExpireRoutesByExitIp(ip VirtualIp) {
	for _, v := range t.RoutingEntries {
		if v.ExitIp == ip && !v.Expired() {
			// fmt.Println("[RoutingTable] ExpireRoutesByExitIp WANTS write lock")
			t.WriteLock()
			// fmt.Println("[RoutingTable] ExpireRoutesByExitIp ACQUIRED write lock")
			v.MarkAsExpired()
			t.WriteUnLock()
			// fmt.Println("[RoutingTable] ExpireRoutesByExitIp RELEASED write lock")
		}
	}
}

func MakeRoutingTable() RoutingTable {
	RoutingEntries := make(map[VirtualIp]*RoutingEntry)
	Neighbors := make(map[VirtualIp]VirtualIp)
	return RoutingTable{RoutingEntries, Neighbors, &sync.RWMutex{}}
}
