package model

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"syscall"
)

type RipEntry struct {
	Cost    int
	Address VirtualIp
}

type RipInfo struct {
	Command    int
	NumEntries int
	Entries    []RipEntry
}

/*
function :   Add one entrie to RipInfo,and increment num of entries
parameter:   RipEntie
return   :   void
*/

func (r *RipInfo) AddEntry(entry RipEntry) {
	r.Entries = append(r.Entries, entry)
	r.NumEntries += 1
}

/*
function :  Convert RipInfo to String,
parameter:  RipInfo
return   :  String
*/

func (r *RipInfo) String() string {
	returnstring := fmt.Sprintf("  command:%d\n  num:%d\n", r.Command, r.NumEntries)
	for i := 0; i < r.NumEntries; i++ {
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
	for i := 0; i < r.NumEntries; i++ {
		binary.BigEndian.PutUint32(b[4+8*i:8+8*i], uint32(r.Entries[i].Cost))
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
	command := int(binary.BigEndian.Uint16(b[0:2]))
	num := int(binary.BigEndian.Uint16(b[2:4]))
	entries := make([]RipEntry, 0)
	for i := 0; i < num; i++ {
		cost := int(binary.BigEndian.Uint32(b[4+8*i : 8+8*i]))
		address := Int2Vip(net.IPv4(b[8+8*i], b[9+8*i], b[10+8*i], b[11+8*i]))
		entries = append(entries, RipEntry{cost, address})
	}
	return RipInfo{command, num, entries}, nil
}
