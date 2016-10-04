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
		Dst:      net.IPv4(dst.A, dst.B, dst.C, dst.D),
        Src:      net.IPv4(src.A, src.B, src.C, src.D),
        Options:  []byte{}
		// ID, Src and Checksum will be set for us by the kernel
	}
	return IpPacket(h, message)
}
