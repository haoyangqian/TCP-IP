package network

import (
	"factory"
	"fmt"
	"model"
	"transport"
)

type IpHandler struct {
}

func (handler IpHandler) Handle(packet model.IpPacket, receivedFrom model.VirtualIp) {
	printPacketInfo(packet)

}

func printPacketInfo(packet model.IpPacket) {
	fmt.Println("driver received packet:")
	fmt.Println(packet.IpPacketString())
	tcppacket := transport.ConvertToTcpPacket(packet.Payload)
	fmt.Println(tcppacket.)
	fmt.Print("> ")
}
