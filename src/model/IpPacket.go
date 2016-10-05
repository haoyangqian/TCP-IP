package model

import (
	"../ipv4"
	"fmt"
	"net"
)

type IpPacket struct {
	Ipheader ipv4.Header
	Payload  []byte
}

func MakeIpPacket(message []byte, protocol int, src VirtualIp, dst VirtualIp) IpPacket {
	h := ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: 20 + len(message),
		TTL:      16,
		Protocol: protocol,
		Dst:      net.ParseIP(dst.Ip),
		Src:      net.ParseIP(src.Ip),
		Options:  []byte{},
		// ID, Src and Checksum will be set for us by the kernel
	}
	return IpPacket{h, message}
}

func (Ip *IpPacket) IpPacketString() string {
	returnstring := fmt.Sprintf("  src_ip:%v\n  dst_ip:%v\n  body_len:%d\n  headr:\n    tos:%d\n    id:%d\n    prot:%d\n  payload:%s\n", Ip.Ipheader.Src, Ip.Ipheader.Dst, Ip.Ipheader.TotalLen-Ip.Ipheader.Len, Ip.Ipheader.TOS, Ip.Ipheader.ID, Ip.Ipheader.Protocol, string(Ip.Payload[:]))

	return returnstring
}

func (Ip *IpPacket) ConvertToBuffer() []byte {
	buffer, error := Ip.Ipheader.Marshal()
	if error != nil {
		fmt.Println(error)
	}
	buffer = append(buffer, Ip.Payload...)
	return buffer
}

func ConvertToIpPacket(buffer []byte) IpPacket {
	rHeader, error := ipv4.ParseHeader(buffer[0:20])
	if error != nil {
		fmt.Println(error)
	}
	payload := buffer[20:]
	return IpPacket{*rHeader, payload}
}
