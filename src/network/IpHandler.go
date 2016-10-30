package network

import (
	"fmt"
	"model"
	"transport"
)

type IpHandler struct {
	RecvCh chan<- model.IpPacket
}

func (handler IpHandler) Handle(packet model.IpPacket, receivedFrom model.VirtualIp) {
	printPacketInfo(packet)

}

func printPacketInfo(packet model.IpPacket) {
	fmt.Println("driver received packet:")
	fmt.Println(packet.IpPacketString())
	tcppacket := transport.ConvertToTcpPacket(packet.Payload)
	fmt.Println(tcppacket.TcpPacketString())
	fmt.Print("> ")
}

func (handler IpHandler) handlePacket(packet model.IpPacket) {
	handler.RecvCh <- packet
}
