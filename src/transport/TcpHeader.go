package transport

import (
	"bytes"
	"encoding/binary"
)

const (
	FIN = 1  // 00 0001
	SYN = 2  // 00 0010
	RST = 4  // 00 0100
	PSH = 8  // 00 1000
	ACK = 16 // 01 0000
	URG = 32 // 10 0000
)

type TCPHeader struct {
	Source      int //uint16
	Destination int //uint16
	SeqNum      int //uint32
	AckNum      int //uint32
	DataOffset  int //uint8 // 4 bits
	Reserved    int //uint8 // 3 bits
	ECN         int //uint8 // 3 bits
	Ctrl        int //uint8 // 6 bits
	Window      int //uint16
	Checksum    int //uint16 // Kernel will set this if it's 0
	Urgent      int //uint16
	Options     []TCPOption
}

type TCPOption struct {
	Kind   uint8
	Length uint8
	Data   []byte
}

func (tcp *TCPHeader) HasFlag(flagBit int) bool {
	return tcp.Ctrl&flagBit != 0
}

func MakeTcpHeader(srcport int,
	dstport int,
	seqnum int,
	acknum int,
	ctrl int,
	ws int) TCPHeader {
	h := TCPHeader{
		Source:      srcport,
		Destination: dstport,
		SeqNum:      seqnum,
		AckNum:      acknum,
		DataOffset:  5,
		Ctrl:        ctrl,
		Window:      ws,
		Checksum:    0,
		Options:     []TCPOption{},
	}
	return h
}

func (tcp *TCPHeader) Marshal() []byte {

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint16(tcp.Source))
	binary.Write(buf, binary.BigEndian, uint16(tcp.Destination))
	binary.Write(buf, binary.BigEndian, uint32(tcp.SeqNum))
	binary.Write(buf, binary.BigEndian, uint32(tcp.AckNum))

	var mix uint16
	mix = uint16(tcp.DataOffset)<<12 | // top 4 bits
		uint16(tcp.Reserved)<<9 | // 3 bits
		uint16(tcp.ECN)<<6 | // 3 bits
		uint16(tcp.Ctrl) // bottom 6 bits
	binary.Write(buf, binary.BigEndian, mix)
	binary.Write(buf, binary.BigEndian, uint16(tcp.Window))
	binary.Write(buf, binary.BigEndian, uint16(tcp.Checksum))
	binary.Write(buf, binary.BigEndian, uint16(tcp.Urgent))

	for _, option := range tcp.Options {
		binary.Write(buf, binary.BigEndian, option.Kind)
		if option.Length > 1 {
			binary.Write(buf, binary.BigEndian, option.Length)
			binary.Write(buf, binary.BigEndian, option.Data)
		}
	}

	out := buf.Bytes()

	// Pad to min tcp header size, which is 20 bytes (5 32-bit words)
	if len(out) > 20 {
		pad := 24 - len(out)
		for i := 0; i < pad; i++ {
			out = append(out, 0)
		}
	}

	return out
}

// Parse packet into TCPHeader structure
func Unmarshal(data []byte) *TCPHeader {
	var tcp TCPHeader
	r := bytes.NewReader(data)
	var src uint16
	var dst uint16
	var seq uint32
	var ack uint32
	binary.Read(r, binary.BigEndian, &src)
	binary.Read(r, binary.BigEndian, &dst)
	binary.Read(r, binary.BigEndian, &seq)
	binary.Read(r, binary.BigEndian, &ack)
	tcp.Source = int(src)
	tcp.Destination = int(dst)
	tcp.SeqNum = int(seq)
	tcp.AckNum = int(ack)

	var mix uint16
	binary.Read(r, binary.BigEndian, &mix)
	tcp.DataOffset = int(mix >> 12)  // top 4 bits
	tcp.Reserved = int(mix >> 9 & 7) // 3 bits
	tcp.ECN = int(mix >> 6 & 7)      // 3 bits
	tcp.Ctrl = int(mix & 0x3f)       // bottom 6 bits

	var window uint16
	var checksum uint16
	var urgent uint16
	binary.Read(r, binary.BigEndian, &window)
	binary.Read(r, binary.BigEndian, &checksum)
	binary.Read(r, binary.BigEndian, &urgent)
	tcp.Window = int(window)
	tcp.Checksum = int(checksum)
	tcp.Urgent = int(urgent)

	return &tcp
}

// TCP Checksum
func Csum(data []byte, srcip, dstip []byte) uint16 {

	pseudoHeader := []byte{
		srcip[0], srcip[1], srcip[2], srcip[3],
		dstip[0], dstip[1], dstip[2], dstip[3],
		0,                  // zero
		6,                  // protocol number (6 == TCP)
		0, byte(len(data)), // TCP length (16 bits), not inc pseudo header
	}

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
		sum += uint32(sumThis[len(sumThis)-1])
	}

	// Add back any carry, and any carry from adding the carry
	sum = (sum >> 16) + (sum & 0xffff)
	sum = sum + (sum >> 16)

	// Bitwise complement
	return uint16(^sum)
}
