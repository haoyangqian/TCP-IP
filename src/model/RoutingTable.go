package model

//import "sync"

// import "sync/atomic"

type RoutingTable struct {
	Routing_entries map[VirtualIp]RoutingEntry
	//Table_lock      sync.Mutex
}

func (t *RoutingTable) PutEntry(entry RoutingEntry) {
	t.Routing_entries[entry.GetDestIp()] = entry
}

func (t *RoutingTable) GetAllEntries() []RoutingEntry {
	return make([]RoutingEntry, len(t.Routing_entries))
}

func MakeRoutingTable() RoutingTable {
	Routing_entries := make(map[VirtualIp]RoutingEntry)
	return RoutingTable{Routing_entries}
}
