package model

import "time"

type RoutingEntry struct {
	Dest             VirtualIp
	Cost             int
	Next_hop         VirtualIp
	Learned_from     VirtualIp
	Expiration_timer *time.Timer
	Is_updated       bool
	Gc_timer         *time.Timer
}

func (e RoutingEntry) getDestIp() VirtualIp {
	return e.Dest
}
