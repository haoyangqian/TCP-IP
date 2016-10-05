package model

//import "sync"

// import "sync/atomic"

type RoutingTable struct {
	RoutingEntries map[VirtualIp]RoutingEntry
	//Table_lock      sync.Mutex
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
