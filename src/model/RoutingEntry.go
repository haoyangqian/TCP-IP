package model

import "time"

type RoutingEntry struct {
	Dest             VirtualIp
	Cost             int
	ExitIP           VirtualIp
	Learn_from       VirtualIp
	Expiration_timer *time.Timer
	Is_updated       bool
	Gc_timer         *time.Timer
}

func (e *RoutingEntry) GetDestIp() VirtualIp {
	return e.Dest
}

func (e *RoutingEntry) GetLearn_from() VirtualIp {
	return e.Learn_from
}

func (e *RoutingEntry) GetExitIP() VirtualIp {
	return e.ExitIP
}

func (e *RoutingEntry) GetCost() int {
	return e.Cost
}

func MakeRoutingEntry(dst VirtualIp, exitip VirtualIp, learn_from VirtualIp, cost int) RoutingEntry {
	timer1 := time.NewTimer(time.Second * 12)
	timer2 := time.NewTimer(time.Second * 12)
	return RoutingEntry{dst, cost, exitip, learn_from, timer1, false, timer2}
}
