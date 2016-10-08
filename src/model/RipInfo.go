package model

import (
	"encoding/binary"
	"errors"
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

/*
function :   Add one entrie to RipInfo,and increment num of entries
parameter:   RipEntie
return   :   void
*/

func (r *RipInfo) AddEntrie(entry RipEntrie) {
	r.Entries = append(r.Entries, entry)
	r.NumEntries += 1
}

/*
function :  Convert routing table to RipInfo
            if command = 1,which indicates a request,set num = 0
            if command = 2,which indicates a response,iterate the routing table and add every entry to the RipInfo
parameter:   routingTable,command(int)
return   :   RipInfo
*/

func RoutingTable2RipInfo(rtable RoutingTable, command int) (RipInfo, error) {
	if command != 1 && command != 2 {
		return RipInfo{}, errors.New("Wrong command type")
	}

	if command == 1 {
		ripentries := make([]RipEntrie, 0)
		ripinfo := RipInfo{uint32(command), uint32(0), ripentries}
		return ripinfo, nil
	}

	if len(rtable.RoutingEntries) == 0 {
		return RipInfo{}, errors.New("Routing table has o entries")
	}
	ripentries := make([]RipEntrie, 0)
	ripinfo := RipInfo{uint32(command), uint32(0), ripentries}
	for _, v := range rtable.RoutingEntries {
		ripinfo.AddEntrie(RipEntrie{uint32(v.Cost), v.Dest})
	}

	return ripinfo, nil
}

/*
function :  Convert RipInfo to String,
parameter:  RipInfo
return   :  String
*/

func (r *RipInfo) String() string {
	returnstring := fmt.Sprintf("  command:%d\n  num:%d\n", r.Command, r.NumEntries)
	for i := uint32(0); i < r.NumEntries; i++ {
		returnstring += fmt.Sprintf("    cost:%d  address:%v\n", r.Entries[i].Cost, r.Entries[i].Address)
	}

	return returnstring
}

/*
function :  Convert RipInfo to []byte,
parameter:  RipInfo
return   :  []byte
*/
func (r *RipInfo) Marshal() ([]byte, error) {
	if r == nil {
		return nil, syscall.EINVAL
	}
	riplen := 4 + 8*r.NumEntries
	b := make([]byte, riplen)
	binary.BigEndian.PutUint16(b[0:2], uint16(r.Command))
	binary.BigEndian.PutUint16(b[2:4], uint16(r.NumEntries))
	for i := uint32(0); i < r.NumEntries; i++ {
		binary.BigEndian.PutUint32(b[4+8*i:8+8*i], r.Entries[i].Cost)
		vipbyte4 := r.Entries[i].Address.Vip2Int()
		b[8+8*i] = vipbyte4[0]
		b[9+8*i] = vipbyte4[1]
		b[10+8*i] = vipbyte4[2]
		b[11+8*i] = vipbyte4[3]
	}
	return b, nil
}

/*
function :  Convert []byte to RipInfo,
parameter:  []byte
return   :  RipInfo
*/
func UnmarshalForInfo(b []byte) (RipInfo, error) {
	if len(b) < 4 {
		return RipInfo{}, errors.New("byte too short")
	}
	command := uint32(binary.BigEndian.Uint16(b[0:2]))
	num := uint32(binary.BigEndian.Uint16(b[2:4]))
	entries := make([]RipEntrie, 0)
	for i := uint32(0); i < num; i++ {
		cost := uint32(binary.BigEndian.Uint32(b[4+8*i : 8+8*i]))
		address := Int2Vip(net.IPv4(b[8+8*i], b[9+8*i], b[10+8*i], b[11+8*i]))
		entries = append(entries, RipEntrie{cost, address})
	}
	return RipInfo{command, num, entries}, nil
}
