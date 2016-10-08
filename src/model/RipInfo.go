package model

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
)

type RipEntrie struct {
	Cost    uint32
	Address VirtualIp
}

type RipInfo struct {
	Command    uint32
	NumEntries uint32
	Entries    []RipEntrie
}

func (r *RipInfo) AddEntrie(entry RipEntrie) {
	r.Entries = append(r.Entries, entry)
	r.NumEntries += 1
}

func (r *RipInfo) String() string {
	returnstring := fmt.Sprintf("  command:%d\n  num:%d\n", r.Command, r.NumEntries)
	for i := 0; i < r.NumEntries; i++ {
		returnstring += fmt.Sprintf("cost:%d  address:%v\n", r.Entries[i].Cost, r.Entries[i].Address)
	}

	return returnstring
}

func (r *RipInfo) Marshal() ([]byte, error) {
	if r == nil {
		return nil, syscall.EINVAL
	}
	riplen := 4 + 8*r.NumEntries
	b := make([]byte, riplen)
	binary.BigEndian.PutUint16(b[0:2], uint16(r.Command))
	binary.BigEndian.PutUint16(b[2:4], uint16(r.NumEntries))
	for i := 0; i < r.NumEntries; i++ {
		binary.BigEndian.PutUint32(b[4+8*i:8+8*i], r.Entries[i].Cost)
		binary.BigEndian.PutUint32(b[8+8*i:12+8*i], r.Entries[i].Address.Vip2Int())
	}
	return b, nil
}

func UnmarshalForInfo(b []byte) (RipInfo, error) {
	if len(b) < 4 {
		return nil, "byte too short"
	}
	command := int(binary.BigEndian.Uint16(b[0:2]))
	num := int(binary.BigEndian.Uint16(b[2:4]))
	entries := make([]RipEntrie)
	for i := 0; i < num; i++ {
		cost := int(binary.BigEndian.Uint32(b[4+8*i : 8+8*i]))
		address := Int2Vip(net.IP{binary.BigEndian.Uint32(b[8+8*i : 12+8*i])})
		append(entries, RipEntrie{cost, address})
	}
	return RipInfo{command, num, entries}
}
