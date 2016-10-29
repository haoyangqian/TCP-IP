package transport

import (
	"fmt"
)

type TcpPacket struct {
	tcpheader TCPHeader
	Payload   []byte
}

func MakeTcpPacket(message []byte, h TCPHeader) TcpPacket {
	return TcpPacket{h, message}
}

func (Tcp *TcpPacket) TcpPacketString() string {
	returnstring := fmt.Sprintf("  src_port:   %d\n  dst_port:   %d\n   payload:  %s\n", Tcp.tcpheader.Source, Tcp.tcpheader.Destination, string(Tcp.Payload[:]))
	return returnstring
}

func (Tcp *TcpPacket) ConvertToBuffer() []byte {
	buffer := Tcp.tcpheader.Marshal()
	buffer = append(buffer, Tcp.Payload...)
	return buffer
}

func ConvertToTcpPacket(buffer []byte) TcpPacket {
	tcpHeader := Unmarshal(buffer[0:20])
	index := 4 * int(tcpHeader.DataOffset)
	payload := buffer[index:]
	return TcpPacket{*tcpHeader, payload}
}
