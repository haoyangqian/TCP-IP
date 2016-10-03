package model

import "sync"

// import "sync/atomic"

type RoutingTable struct {
	Routing_entries map[VirtualIp]RoutingEntry
	Table_lock      sync.Mutex
}

func (t RoutingTable) putEntry(entry RoutingEntry) {
	t.Routing_entries[entry.getDestIp()] = entry
}

func (t RoutingTable) getAllEntries() []RoutingEntry {
	return make([]RoutingEntry, len(t.Routing_entries))
}
