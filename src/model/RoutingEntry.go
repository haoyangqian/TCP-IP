package model

import "time"

type RoutingEntry struct {
	Dest       VirtualIp
	Cost       int
	ExitIp     VirtualIp
	NextHop    VirtualIp
	Ttl        int64
	IsUpdated  bool
	GcTimer    int64
	HasExpired bool
}

func MakeRoutingEntry(dst VirtualIp, exitIp VirtualIp, nextHop VirtualIp, cost int) RoutingEntry {
	ttlTimer := time.Now().Unix() + RIP_ROUTING_ENTRY_TTL_SECONDS
	gcTimer := ttlTimer + RIP_GC_TIMER_SECONDS
	return RoutingEntry{dst, cost, exitIp, nextHop, ttlTimer, false, gcTimer, false}
}

func (entry *RoutingEntry) Update(cost int, nextHop VirtualIp) {
	entry.UpdateCost(cost)
	entry.UpdateNextHop(nextHop)
	entry.ResetTtl()
	entry.ResetGcTimer()
	entry.SetIsUpdated(true)
}

func (entry *RoutingEntry) ResetTtl() {
	entry.Ttl = time.Now().Unix() + RIP_ROUTING_ENTRY_TTL_SECONDS
	entry.SetIsUpdated(true)
}

func (entry *RoutingEntry) ResetGcTimer() {
	entry.GcTimer = time.Now().Unix() + RIP_GC_TIMER_SECONDS
	entry.SetIsUpdated(true)
}

func (entry *RoutingEntry) UpdateCost(cost int) {
	entry.Cost = cost
	entry.SetIsUpdated(true)
}

func (entry *RoutingEntry) UpdateNextHop(nextHop VirtualIp) {
	entry.NextHop = nextHop
	entry.SetIsUpdated(true)
}

func (entry *RoutingEntry) SetIsUpdated(isUpdated bool) {
	entry.IsUpdated = isUpdated
}

func (entry *RoutingEntry) SetExpired(expired bool) {
	entry.HasExpired = expired
	entry.SetIsUpdated(true)
}

func (entry *RoutingEntry) Expired() bool {
	return entry.HasExpired
}

func (entry *RoutingEntry) ShouldExpire() bool {
	// retrun true if ttl is expired but the entry hasn't been marked as expired
	return time.Now().UnixNano() > entry.Ttl && !entry.Expired()
}

func (entry *RoutingEntry) ShouldGC() bool {
	return time.Now().UnixNano() > entry.GcTimer
}

func (entry *RoutingEntry) MarkAsExpired() {
	entry.Cost = RIP_INFINITY
	entry.SetExpired(true)
	entry.SetIsUpdated(true)
}
