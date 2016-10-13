package model

import "errors"
import "sync"

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
	defer t.ReadUnLock()
	if _, ok := t.RoutingEntries[ip]; ok {
		return true
	} else {
		return false
	}
}

func (t *RoutingTable) HasNeighbor(neighbor VirtualIp) bool {
	t.ReadLock()
	defer t.ReadUnLock()
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
	t.ReadLock()
	defer t.ReadUnLock()
	return t.RoutingEntries[vip], nil
}

func (t *RoutingTable) PutEntry(entry *RoutingEntry) {
	t.WriteLock()
	t.RoutingEntries[entry.Dest] = entry
	t.WriteUnLock()
}

func (t *RoutingTable) DeleteEntry(entry *RoutingEntry) {
	if t.HasEntry(entry.Dest) {
		t.WriteLock()
		delete(t.RoutingEntries, entry.Dest)
		t.WriteUnLock()
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
	t.ReadLock()
	defer t.ReadUnLock()
	routingentries := make([]*RoutingEntry, 0)
	for _, v := range t.RoutingEntries {
		if v.IsUpdated == true {
			routingentries = append(routingentries, v)
		}
	}
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

func (t *RoutingTable) GetLocalRoutes() []*RoutingEntry {
	t.ReadLock()
	defer t.ReadUnLock()
	localRoutes := make([]*RoutingEntry, 0)
	for _, v := range t.RoutingEntries {
		if v.IsLocal {
			localRoutes = append(localRoutes, v)
		}
	}
	return localRoutes
}

func (t *RoutingTable) ExpireRoutesByExitIp(ip VirtualIp) {
	for _, v := range t.RoutingEntries {
		if v.ExitIp == ip && !v.Expired() {
			t.WriteLock()
			v.MarkAsExpired()
			t.WriteUnLock()
		}
	}
}

func MakeRoutingTable() RoutingTable {
	RoutingEntries := make(map[VirtualIp]*RoutingEntry)
	Neighbors := make(map[VirtualIp]VirtualIp)
	return RoutingTable{RoutingEntries, Neighbors, &sync.RWMutex{}}
}
