package model

import "errors"

//import "sync"

// import "sync/atomic"

type RoutingTable struct {
	RoutingEntries map[VirtualIp]RoutingEntry
	//Table_lock      sync.Mutex
}

func (t *RoutingTable) HasEntry(ip VirtualIp) bool {
	if _, ok := t.RoutingEntries[ip]; ok {
		return true
	} else {
		return false
	}
}

func (t *RoutingTable) GetEntry(vip VirtualIp) (RoutingEntry, error) {
	var entry RoutingEntry
	if !t.HasEntry(vip) {
		return entry, errors.New("Invalid State: an entry not in the RoutingTable was requested " + vip.Ip)
	}

	return t.RoutingEntries[vip], nil
}

func (t *RoutingTable) PutEntry(entry RoutingEntry) {
	t.RoutingEntries[entry.Dest] = entry
}

func (t *RoutingTable) GetAllEntries() []RoutingEntry {
	return make([]RoutingEntry, len(t.RoutingEntries))
}

func MakeRoutingTable() RoutingTable {
	RoutingEntries := make(map[VirtualIp]RoutingEntry)
	return RoutingTable{RoutingEntries}
}
