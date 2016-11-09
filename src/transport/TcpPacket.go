package transport

import (
	"encoding/binary"
	//"fmt"
	"logging"
)

type TcpPacket struct {
	Tcpheader TCPHeader
	Payload   []byte
}

func MakeTcpPacket(message []byte, h TCPHeader) TcpPacket {
	return TcpPacket{h, message}
}

func (Tcp *TcpPacket) PrintTcpPacketString() {
	logging.Logger.Printf("[IpHandler][TcpPacket] PrintTcpPacketString tcpheader:%+v, payload: %d\n", Tcp.Tcpheader, len(Tcp.Payload))
}

func (Tcp *TcpPacket) ConvertToBuffer() []byte {
	buffer := Tcp.Tcpheader.Marshal()
	buffer = append(buffer, Tcp.Payload...)
	return buffer
}

func ConvertToTcpPacket(buffer []byte) TcpPacket {
	tcpHeader := Unmarshal(buffer[0:20])
	index := 4 * int(tcpHeader.DataOffset)
	payload := buffer[index:]
	return TcpPacket{*tcpHeader, payload}
}

// TCP Checksum
func Csum(data []byte, srcip, dstip []byte) int {

	pseudoHeader := []byte{
		srcip[0], srcip[1], srcip[2], srcip[3],
		dstip[0], dstip[1], dstip[2], dstip[3],
		0,    // zero
		6,    // protocol number (6 == TCP)
		0, 0, // TCP length (16 bits), not inc pseudo header
	}
	//fmt.Println("pseudo header length:", len(pseudoHeader))
	binary.BigEndian.PutUint16(pseudoHeader[10:12], uint16(len(data)))
	sumThis := make([]byte, 0, len(pseudoHeader)+len(data))
	sumThis = append(sumThis, pseudoHeader...)
	sumThis = append(sumThis, data...)
	//fmt.Printf("% x\n", sumThis)

	lenSumThis := len(sumThis)
	var nextWord uint16
	var sum uint32
	for i := 0; i+1 < lenSumThis; i += 2 {
		nextWord = uint16(sumThis[i])<<8 | uint16(sumThis[i+1])
		sum += uint32(nextWord)
	}
	if lenSumThis%2 != 0 {
		//fmt.Println("Odd byte")
		sum += uint32(sumThis[lenSumThis-1])
	}

	// Add back any carry, and any carry from adding the carry
	sum = (sum >> 16) + (sum & 0xffff)
	sum = sum + (sum >> 16)
	answer := uint16(0xffffffff ^ sum)
	// Bitwise complement
	return int(answer)
}
