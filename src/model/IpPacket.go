package model

import (
	"encoding/binary"
	"fmt"
	"ipv4"
	"net"
)

type IpPacket struct {
	Ipheader ipv4.Header
	Payload  []byte
}

func MakeIpPacket(message []byte, protocol int, src VirtualIp, dst VirtualIp) IpPacket {
	h := ipv4.Header{
		Version:  IP_VERSION,
		Len:      IP_DEFAUTL_HEADER_LEN,
		TOS:      IP_DEFAULT_TOS,
		TotalLen: IP_DEFAUTL_HEADER_LEN + len(message),
		// ID:       IP_DEFAULT_ID,
		// Flags:    IP_DEFAULT_FLAGS,
		// Offset:   IP_DEFAULT_OFFSET,
		TTL:      IP_DEFAULT_TTL,
		Protocol: protocol,
		Dst:      net.ParseIP(dst.Ip),
		Src:      net.ParseIP(src.Ip),
		Options:  []byte{},
	}
	h.Checksum = int(IpSum(h))
	return IpPacket{h, message}
}

func (Ip *IpPacket) IpPacketString() string {
	returnstring := fmt.Sprintf("  src_ip:   %v\n  dst_ip:   %v\n  body_len: %d\n  headr:\n    tos:    %d\n    id:     %d\n    prot:   %d\n  payload:  %s\n", Ip.Ipheader.Src, Ip.Ipheader.Dst, Ip.Ipheader.TotalLen-Ip.Ipheader.Len, Ip.Ipheader.TOS, Ip.Ipheader.ID, Ip.Ipheader.Protocol, string(Ip.Payload[:]))

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

func IpSum(header ipv4.Header) int {
	buffer, _ := header.Marshal()
	n := len(buffer)
	sum := uint32(0)
	for i := 0; i < n-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(buffer[i : i+2]))
	}
	if n%2 == 1 {
		sum += uint32(buffer[n])
	}
	sum = (sum >> 16) + (sum & 0xffff) /* add hi 16 to low 16 */
	sum += (sum >> 16)                 /* add carry */
	answer := uint16(0xffffffff ^ sum)
	return int(answer)
}
