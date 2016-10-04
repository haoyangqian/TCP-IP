package model

import (
	"../ipv4"
	"net"
)

type IpPacket struct {
	Ipheader ipv4.Header
	Payload  []byte
}

func MakeIpPacket(message []byte,protocol int,src VirtualIp,dst VirtualIp) IpPacket {
	h := ipv4.Header{
		Version:  4,
		Len:      len(message),
        TOS:      0,
		TotalLen: 20 + len(message), 
		TTL:      16,
		Protocol: protocol, 
		Dst:      net.ParseIP(VirtualIp.Ip)
        Src:      net.ParseIP(VirtualIp.Ip)
        Options:  []byte{}
		// ID, Src and Checksum will be set for us by the kernel
	}
	return IpPacket(h, message)
}
