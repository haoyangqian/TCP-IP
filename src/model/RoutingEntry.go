package model

import "time"

type RoutingEntry struct {
	Dest      VirtualIp
	Cost      int
	ExitIp    VirtualIp
	NextHop   VirtualIp
	Ttl       *time.Timer
	IsUpdated bool
	GcTimer   *time.Timer
}

func MakeRoutingEntry(dst VirtualIp, exitIp VirtualIp, learnedFrom VirtualIp, cost int) RoutingEntry {
	ttlTimer := time.NewTimer(time.Second * 12)
	gcTimer := time.NewTimer(time.Second * 12)
	return RoutingEntry{dst, cost, exitIp, learnedFrom, ttlTimer, false, gcTimer}
}

func (entry *RoutingEntry) Update(cost int, nextHop VirtualIp) {
	entry.UpdateCost(cost)
	entry.UpdateNextHop(nextHop)
	entry.ResetTtl()
	entry.ResetGcTimer()
	entry.SetIsUpdated(true)
}

func (entry *RoutingEntry) ResetTtl() {
	entry.Ttl = time.NewTimer(RIP_ROUTING_ENTRY_TTL_SECONDS * time.Second)
}

func (entry *RoutingEntry) ResetGcTimer() {
	entry.Ttl = time.NewTimer(RIP_GC_TIMER_SECONDS * time.Second)
}

func (entry *RoutingEntry) UpdateCost(cost int) {
	entry.Cost = cost
}

func (entry *RoutingEntry) UpdateNextHop(nextHop VirtualIp) {
	entry.NextHop = nextHop
}

func (entry *RoutingEntry) SetIsUpdated(isUpdated bool) {
	entry.IsUpdated = isUpdated
}
