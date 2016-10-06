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
