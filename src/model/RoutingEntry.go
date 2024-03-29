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
	IsLocal    bool
}

func MakeRoutingEntry(dst VirtualIp, exitIp VirtualIp, nextHop VirtualIp, cost int, isLocal bool) RoutingEntry {
	ttlTimer := time.Now().Unix() + RIP_ROUTING_ENTRY_TTL_SECONDS
	gcTimer := ttlTimer + RIP_GC_TIMER_SECONDS
	return RoutingEntry{dst, cost, exitIp, nextHop, ttlTimer, false, gcTimer, false, isLocal}
}

func (entry *RoutingEntry) Update(cost int, nextHop VirtualIp) {
	entry.UpdateCost(cost)
	entry.UpdateNextHop(nextHop)
	entry.ResetTtl()
	entry.SetExpired(false)
	entry.ResetGcTimer()
	entry.SetIsUpdated(true)
}

func (entry *RoutingEntry) ExtendTtl() {
	entry.Ttl = time.Now().Unix() + RIP_ROUTING_ENTRY_TTL_SECONDS
	entry.GcTimer = time.Now().Unix() + RIP_GC_TIMER_SECONDS
	entry.HasExpired = false
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

func (entry *RoutingEntry) Reachable() bool {
	return entry.Cost < RIP_INFINITY
}

func (entry *RoutingEntry) ShouldExpire() bool {
	// retrun true if ttl is expired but the entry hasn't been marked as expired
	if entry.IsLocal && entry.Cost == 0 {
		// a route to a local destination should never expire
		return false
	}
	return (time.Now().Unix() > entry.Ttl) && !entry.Expired()
}

func (entry *RoutingEntry) ShouldGC() bool {
	if entry.IsLocal && entry.Cost == 0 {
		return false
	}
	return (time.Now().Unix() > entry.GcTimer) && entry.Expired()
}

func (entry *RoutingEntry) MarkAsExpired() {
	entry.UpdateCost(RIP_INFINITY)
	entry.SetExpired(true)
	entry.SetIsUpdated(true)
}
