package model

import "time"

type RoutingEntry struct {
	Dest        VirtualIp
	Cost        int
	ExitIp      VirtualIp
	LearnedFrom VirtualIp
	Ttl         *time.Timer
	IsUpdated   bool
	GcTimer     *time.Timer
}

func (e *RoutingEntry) GetDestIp() VirtualIp {
	return e.Dest
}

func (e *RoutingEntry) GetLearnedFrom() VirtualIp {
	return e.LearnedFrom
}

func (e *RoutingEntry) GetExitIp() VirtualIp {
	return e.ExitIp
}

func (e *RoutingEntry) GetCost() int {
	return e.Cost
}

func MakeRoutingEntry(dst VirtualIp, exitIp VirtualIp, learnedFrom VirtualIp, cost int) RoutingEntry {
	ttlTimer := time.NewTimer(time.Second * 12)
	gcTimer := time.NewTimer(time.Second * 12)
	return RoutingEntry{dst, cost, exitip, LearnedFrom, ttlTimer, false, gcTimer}
}
