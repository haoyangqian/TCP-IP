package model

import "errors"

//import "sync"

// import "sync/atomic"
import "fmt"

type RoutingTable struct {
	RoutingEntries map[VirtualIp]*RoutingEntry
	Neighbors      map[VirtualIp]VirtualIp //nextHop -> ExitIp
	//Table_lock      sync.Mutex
}

func (t *RoutingTable) HasEntry(ip VirtualIp) bool {
	if _, ok := t.RoutingEntries[ip]; ok {
		return true
	} else {
		return false
	}
}

func (t *RoutingTable) HasNeighbor(neighbor VirtualIp) bool {
	if _, ok := t.Neighbors[neighbor]; ok {
		return true
	} else {
		return false
	}
}

func (t *RoutingTable) GetEntry(vip VirtualIp) (*RoutingEntry, error) {
	var entry RoutingEntry
	if !t.HasEntry(vip) {
		return &entry, errors.New("Invalid State: an entry not in the RoutingTable was requested " + vip.Ip)
	}

	return t.RoutingEntries[vip], nil
}

func (t *RoutingTable) PutEntry(entry *RoutingEntry) {
	t.RoutingEntries[entry.Dest] = entry
}

func (t *RoutingTable) DeleteEntry(entry *RoutingEntry) {
	fmt.Println("DeleteEntry")
	if t.HasEntry(entry.Dest) {
		fmt.Println("Entry found, deleting...")
		delete(t.RoutingEntries, entry.Dest)
	}
}

func (t *RoutingTable) PutNeighbor(vip VirtualIp, inter VirtualIp) {
	t.Neighbors[vip] = inter
}

func (t *RoutingTable) GetNeighbor(vip VirtualIp) (VirtualIp, error) {
	var inter VirtualIp
	if !t.HasNeighbor(vip) {
		return inter, errors.New("Invalid State: an vip not in the neighbor was requested ")
	}

	return t.Neighbors[vip], nil
}

func (t *RoutingTable) GetAllEntries() []*RoutingEntry {
	routingentries := make([]*RoutingEntry, 0)
	for _, v := range t.RoutingEntries {
		routingentries = append(routingentries, v)
	}
	return routingentries
}

func (t *RoutingTable) GetAllNeighbors() []VirtualIp {
	neighbors := make([]VirtualIp, 0)
	for k, _ := range t.Neighbors {
		neighbors = append(neighbors, k)
	}
	return neighbors
}

func (t *RoutingTable) GetUpdatedEntries() []*RoutingEntry {
	return make([]*RoutingEntry, len(t.RoutingEntries))
}

func (t *RoutingTable) GetExpiredEntries() []*RoutingEntry {
	expiredRoutes := make([]*RoutingEntry, 0)
	for _, v := range t.RoutingEntries {
		if v.ShouldExpire() || v.Expired() {
			expiredRoutes = append(expiredRoutes, v)
		}
	}
	return expiredRoutes
}

func MakeRoutingTable() RoutingTable {
	RoutingEntries := make(map[VirtualIp]*RoutingEntry)
	Neighbors := make(map[VirtualIp]VirtualIp)
	return RoutingTable{RoutingEntries, Neighbors}
}
